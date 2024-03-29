package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"
	errors2 "github.com/pkg/errors"
	"go.dot.industries/brease/models"
	"go.dot.industries/brease/storage"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"sync"
)

const kLatestVersionField = "latest_version"
const kIndexField = "index"

var ruleFields = []string{kLatestVersionField, kIndexField}

type Options struct {
	URL    string
	Logger *zap.Logger
}

func NewDatabase(opts Options) (storage.Database, error) {
	if opts.URL == "" {
		opts.URL = "redis://default@localhost:6379"
	}
	if opts.Logger != nil {
		opts.Logger.Debug("Using redis for database", zap.String("dsn", opts.URL))
	}
	opt, err := redis.ParseURL(opts.URL)
	if err != nil {
		return nil, err
	}
	db := redis.NewClient(opt)
	db.AddHook(redisotel.NewTracingHook(redisotel.WithAttributes(semconv.NetSockPeerAddrKey.String(opt.Addr))))

	r := &redisContainer{
		db:     db,
		logger: opts.Logger,
		rulePool: sync.Pool{
			New: func() interface{} {
				return new(models.VersionedRule)
			},
		},
	}

	return r, nil
}

type redisContainer struct {
	db       *redis.Client
	logger   *zap.Logger
	rulePool sync.Pool
}

func (r *redisContainer) Close() error {
	return r.db.Close()
}

// AddRule persists a rule to redis by using 3 keys
// array of all rules within a context - RPUSH org:${orgID}:context:${contextID} = [org:${orgID}:context:${contextID}:rule:${ruleID}]
// key for rule index within the context - HSET org:${orgID}:context:${contextID}:rule:${ruleID}[index] = 0
// key for rule version data - HSET org:${orgID}:context:${contextID}:rule:${ruleID}:v${version}[json_data] = ruleDATA
// hash key for rule latest version - HSET  org:${orgID}:context:${contextID}:rule:${ruleID}[latest_version] = org:${orgID}:context:${contextID}:rule:${ruleID}:v${version}
func (r *redisContainer) AddRule(ctx context.Context, ownerID string, contextID string, rule models.Rule) (models.VersionedRule, error) {
	vRule := models.VersionedRule{
		Rule:    rule,
		Version: 1,
	}
	ck := storage.ContextKey(ownerID, contextID)
	rk := storage.RuleKey(ownerID, contextID, vRule.ID)
	vk := storage.VersionKey(ownerID, contextID, vRule.ID, vRule.Version)

	data, err := json.Marshal(vRule)
	if err != nil {
		return models.VersionedRule{}, errors2.Wrap(err, "failed to marshal rule")
	}

	// rule in context array
	length, err := r.db.RPush(ctx, ck, rk).Result()
	if err != nil {
		return models.VersionedRule{}, errors2.Wrap(err, "failed to save rule to context array")
	}

	// index of rule in context
	err = r.db.HSet(ctx, rk, kIndexField, length-1).Err()
	if err != nil {
		return models.VersionedRule{}, errors2.Wrap(err, "failed save index of rule within context")
	}

	// object versions
	err = r.db.HSet(ctx, rk, vk, string(data)).Err()
	if err != nil {
		return models.VersionedRule{}, errors2.Wrap(err, "failed to save object version")
	}

	// Update rule key to point to the latest version key
	err = r.db.HSet(ctx, rk, kLatestVersionField, vk).Err()
	if err != nil {
		return models.VersionedRule{}, errors2.Wrap(err, "failed to update latest version pointer")
	}

	return vRule, nil
}

func (r *redisContainer) Rules(ctx context.Context, ownerID string, contextID string) (rules []models.VersionedRule, err error) {
	ck := storage.ContextKey(ownerID, contextID)
	ruleKeys, err := r.db.LRange(ctx, ck, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	latestVersionKeys, err := r.getLatestVersionKeys(ctx, ruleKeys)
	if err != nil {
		return nil, err
	}
	latestVersionData, err := r.getLatestVersionData(ctx, latestVersionKeys)
	if err != nil {
		return nil, err
	}

	for vk, versionData := range latestVersionData {
		rule := r.rulePool.Get().(*models.VersionedRule)
		umErr := json.Unmarshal([]byte(versionData), rule)
		if umErr != nil {
			return nil, errors2.Wrapf(umErr, "couldn't unmarshal versionData for %s", vk)
		}
		rules = append(rules, *rule)
		r.rulePool.Put(rule)
	}

	return
}

func (r *redisContainer) RuleVersions(ctx context.Context, ownerID string, contextID string, ruleID string) (rules []models.VersionedRule, err error) {
	rk := storage.RuleKey(ownerID, contextID, ruleID)
	ruleVersionKeys, err := r.db.HGetAll(ctx, rk).Result()
	if err != nil {
		return nil, err
	}

	for key, jsonData := range ruleVersionKeys {
		if slices.Contains(ruleFields, key) {
			continue
		}
		rule := r.rulePool.Get().(*models.VersionedRule)
		umErr := json.Unmarshal([]byte(jsonData), rule)
		if umErr != nil {
			return nil, umErr
		}
		rules = append(rules, *rule)
		r.rulePool.Put(rule)
	}

	return
}

func (r *redisContainer) RemoveRule(ctx context.Context, ownerID string, contextID string, ruleID string) error {
	ck := storage.ContextKey(ownerID, contextID)
	rk := storage.RuleKey(ownerID, contextID, ruleID)
	idx, err := r.db.HGet(ctx, rk, kIndexField).Int64()
	if err != nil {
		r.logger.Warn("Couldn't find rule to delete", zap.String("rule", rk))
		// swallow not found error
		return nil
	}
	r.logger.Debug("Found index of rule to delete", zap.Int64("index", idx))
	delValue := fmt.Sprintf("DELETE:%d", idx)
	err = r.db.LSet(ctx, ck, idx, delValue).Err()
	if err != nil {
		return err
	}
	r.logger.Debug("Set deletion value for rule to delete", zap.String("value", delValue))
	err = r.db.LRem(ctx, ck, 1, delValue).Err()
	if err != nil {
		return err
	}

	// delete version tracking, but not versions?
	err = r.db.HDel(ctx, rk, kLatestVersionField).Err()
	if err != nil {
		return err
	}

	// delete stored index
	err = r.db.Del(ctx, rk).Err()
	if err != nil {
		return err
	}
	r.logger.Info("Successfully removed rule from context.", zap.String("key", rk))
	return nil
}

func (r *redisContainer) ruleLatestVersion(ctx context.Context, ruleKey string) (int64, error) {
	vk, err := r.db.HGet(ctx, ruleKey, kLatestVersionField).Result()
	if err != nil {
		return -1, err
	}
	version := storage.VersionFromVersionKey(vk)
	return version, nil
}

func (r *redisContainer) ReplaceRule(ctx context.Context, ownerID string, contextID string, rule models.Rule) (models.VersionedRule, error) {
	rk := storage.RuleKey(ownerID, contextID, rule.ID)
	currentVersion, err := r.ruleLatestVersion(ctx, rk)
	if err != nil {
		return models.VersionedRule{}, err
	}
	vRule := models.VersionedRule{
		Rule:    rule,
		Version: currentVersion + 1,
	}
	vk := storage.VersionKey(ownerID, contextID, vRule.ID, vRule.Version)

	ruleJSON, err := json.Marshal(vRule)
	if err != nil {
		return models.VersionedRule{}, err
	}

	err = r.db.HSet(ctx, rk, vk, ruleJSON).Err()
	if err != nil {
		return models.VersionedRule{}, err
	}
	err = r.db.HSet(ctx, rk, kLatestVersionField, vk).Err()
	if err != nil {
		return models.VersionedRule{}, err
	}
	r.logger.Info("Successfully updated rule in context.", zap.String("key", rk), zap.String("new", string(ruleJSON)))
	return vRule, nil
}

func (r *redisContainer) Exists(ctx context.Context, ownerID string, contextID string, ruleID string) (exists bool, err error) {
	rk := storage.RuleKey(ownerID, contextID, ruleID)
	return r.db.HExists(ctx, rk, kLatestVersionField).Result()
}

func (r *redisContainer) SaveAccessToken(ctx context.Context, ownerID string, tokenPair models.TokenPair) error {
	atKey := fmt.Sprintf("access:%s", ownerID)

	tokens, err := r.GetAccessTokens(ctx, ownerID)
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	tokens = append(tokens, tokenPair)

	bts, err := json.Marshal(tokens)
	if err != nil {
		return err
	}

	return r.db.Set(ctx, atKey, bts, 0).Err()
}

func (r *redisContainer) GetAccessTokens(ctx context.Context, ownerID string) (tp []models.TokenPair, err error) {
	atKey := fmt.Sprintf("access:%s", ownerID)

	bts, err := r.db.Get(ctx, atKey).Bytes()
	if err != nil {
		return nil, fmt.Errorf("tokenPair not found: %w", err)
	}

	if err = json.Unmarshal(bts, &tp); err != nil {
		return nil, fmt.Errorf("failed to read tokenPairs: %v", err)
	}
	return tp, nil
}

func (r *redisContainer) getLatestVersionKeys(ctx context.Context, ruleKeys []string) (latestVersionKeys map[string]string, err error) {
	pipe := r.db.Pipeline()
	returnMap := make(map[int]string)
	latestVersionKeys = make(map[string]string)
	i := 0
	for _, rk := range ruleKeys {
		pipe.HGet(ctx, rk, kLatestVersionField)
		returnMap[i] = rk
		i++
	}
	res, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	for idx, cmd := range res {
		rk := returnMap[idx]
		vk, vkErr := cmd.(*redis.StringCmd).Result()
		if vkErr != nil {
			return nil, errors2.Wrapf(vkErr, "latest version not found for %s", rk)
		}
		latestVersionKeys[rk] = vk
	}
	return
}

func (r *redisContainer) getLatestVersionData(ctx context.Context, latestVersionKeys map[string]string) (latestVersionData map[string]string, err error) {
	pipe := r.db.Pipeline()
	returnMap := make(map[int]string)
	latestVersionData = make(map[string]string)
	i := 0

	for rk, vk := range latestVersionKeys {
		pipe.HGet(ctx, rk, vk)
		returnMap[i] = vk
		i++
	}
	res, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	for idx, cmd := range res {
		vk := returnMap[idx]
		versionData, vdErr := cmd.(*redis.StringCmd).Result()
		if vdErr != nil {
			return nil, errors2.Wrapf(vdErr, "version data not found for %s", vk)
		}
		latestVersionData[vk] = versionData
	}

	return
}
