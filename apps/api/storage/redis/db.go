package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"
	"go.dot.industries/brease/models"
	"go.dot.industries/brease/storage"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"sync"
)

type Options struct {
	URL    string
	Logger *zap.Logger
}

func NewDatabase(opts Options) (storage.Database, error) {
	if opts.URL == "" {
		opts.URL = "rediss://default@localhost:6379"
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
				return new(models.Rule)
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

func (r *redisContainer) AddRule(ctx context.Context, ownerID string, contextID string, rule models.Rule) error {
	ck := contextKey(ownerID, contextID)
	rk := ruleKey(ownerID, contextID, rule.ID)
	bts, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	length, err := r.db.RPush(ctx, ck, bts).Result()
	if err != nil {
		return err
	}
	return r.db.Set(ctx, rk, length-1, 0).Err()
}

func (r *redisContainer) Rules(ctx context.Context, ownerID string, contextID string) (rules []models.Rule, err error) {
	ck := contextKey(ownerID, contextID)
	rawRules, err := r.db.LRange(ctx, ck, 0, 100).Result()
	if err != nil {
		return nil, err
	}

	for _, rr := range rawRules {
		rule := r.rulePool.Get().(*models.Rule)
		umErr := json.Unmarshal([]byte(rr), rule)
		if umErr != nil {
			return nil, umErr
		}
		r.rulePool.Put(rule)
		rules = append(rules, *rule)
	}

	return
}

func (r *redisContainer) RemoveRule(ctx context.Context, ownerID string, contextID string, ruleID string) error {
	ck := contextKey(ownerID, contextID)
	rk := ruleKey(ownerID, contextID, ruleID)
	idx, err := r.db.Get(ctx, rk).Int64()
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
	err = r.db.Del(ctx, rk).Err()
	if err != nil {
		return err
	}
	r.logger.Info("Successfully removed rule from context.", zap.String("key", rk))
	return nil
}

func (r *redisContainer) ReplaceRule(ctx context.Context, ownerID string, contextID string, rule models.Rule) error {
	ck := contextKey(ownerID, contextID)
	rk := ruleKey(ownerID, contextID, rule.ID)
	idx, err := r.db.Get(ctx, rk).Int64()
	if err != nil {
		r.logger.Warn("Couldn't find rule to delete", zap.String("rule", rk))
		// swallow not found error
		return nil
	}
	ruleJSON, jsonErr := json.Marshal(rule)
	if jsonErr != nil {
		return jsonErr
	}
	err = r.db.LSet(ctx, ck, idx, ruleJSON).Err()
	if err != nil {
		return err
	}
	r.logger.Info("Successfully updated rule in context.", zap.String("key", rk), zap.String("new", string(ruleJSON)))
	return nil
}

func (r *redisContainer) Exists(ctx context.Context, ownerID string, contextID string, ruleID string) (exists bool, err error) {
	rk := ruleKey(ownerID, contextID, ruleID)
	err = r.db.Get(ctx, rk).Err()
	if errors.Is(err, redis.Nil) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
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
		return nil, fmt.Errorf("tokenPair not found")
	}

	if err = json.Unmarshal(bts, tp); err != nil {
		return nil, fmt.Errorf("failed to read tokenPairs: %v", err)
	}
	return tp, nil
}

func contextKey(orgID string, contextID string) string {
	return fmt.Sprintf("org:%s:context:%s", orgID, contextID)
}

func ruleKey(orgID string, contextID, ruleID string) string {
	return fmt.Sprintf("org:%s:context:%s:rule:%s", orgID, contextID, ruleID)
}
