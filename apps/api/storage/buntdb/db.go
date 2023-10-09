package buntdb

import (
	authv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/auth/v1"
	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/goccy/go-json"
	"github.com/tidwall/buntdb"
	"go.dot.industries/brease/storage"
	"go.uber.org/zap"
)

type Options struct {
	Path   string
	Logger *zap.Logger
}

func NewDatabase(opts Options) (storage.Database, error) {
	if opts.Path == "" {
		opts.Path = ":memory:"
	}
	if opts.Logger != nil {
		opts.Logger.Debug("Using buntdb for database", zap.String("dsn", opts.Path))
	}
	db, err := buntdb.Open(opts.Path)
	if err != nil {
		return nil, err
	}

	b := &buntdbContainer{
		db:     db,
		logger: opts.Logger,
		rulePool: sync.Pool{
			New: func() interface{} {
				return new(*rulev1.VersionedRule)
			},
		},
	}

	b.createIndices()

	return b, nil
}

type buntdbContainer struct {
	db       *buntdb.DB
	logger   *zap.Logger
	rulePool sync.Pool
}

func (b *buntdbContainer) Close() error {
	return b.db.Close()
}

func (b *buntdbContainer) AddRule(_ context.Context, ownerID string, contextID string, rule *rulev1.Rule) (*rulev1.VersionedRule, error) {
	vRule := &rulev1.VersionedRule{
		Id:          rule.Id,
		Version:     1,
		Description: rule.Description,
		Actions:     rule.Actions,
		Expression:  rule.Expression,
	}
	rk := storage.RuleKey(ownerID, contextID, vRule.Id)
	vk := storage.VersionKey(ownerID, contextID, vRule.Id, vRule.Version)

	ruleJSON, err := json.Marshal(vRule)
	if err != nil {
		return nil, err
	}
	err = b.db.Update(func(tx *buntdb.Tx) error {
		_, _, txErr := tx.Set(rk, vk, nil)
		if txErr != nil {
			return txErr
		}
		_, _, txErr = tx.Set(vk, string(ruleJSON), nil)
		if txErr != nil {
			return txErr
		}
		return nil
	})

	return vRule, err
}

func (b *buntdbContainer) Rules(_ context.Context, ownerID string, contextID string, pageSize int, pageToken string) (rules []*rulev1.VersionedRule, err error) {
	rkSearch := storage.RuleKey(ownerID, contextID, "*")

	var ruleKeys []string
	err = b.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(rkSearch, func(key, val string) bool {
			if storage.IsVersionKey(val) {
				ruleKeys = append(ruleKeys, val)
			}

			return true
		})
	})

	err = b.db.View(func(tx *buntdb.Tx) error {
		for _, vk := range ruleKeys {
			latestVersionData, vErr := tx.Get(vk)
			if vErr != nil {
				return vErr
			}

			rule := b.rulePool.Get().(*rulev1.VersionedRule)
			umErr := json.Unmarshal([]byte(latestVersionData), rule)
			if umErr != nil {
				return umErr
			}
			rules = append(rules, rule)
			b.rulePool.Put(rule)
		}
		return nil
	})

	return
}

func (b *buntdbContainer) RuleVersions(_ context.Context, ownerID string, contextID string, ruleID string, pageSize int, pageToken string) (rules []*rulev1.VersionedRule, err error) {
	vkSearch := storage.RuleKey(ownerID, contextID, ruleID) + ":v*"
	var ruleVersions []string
	err = b.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(vkSearch, func(key, val string) bool {
			ruleVersions = append(ruleVersions, val)
			return true
		})
	})

	err = b.db.View(func(tx *buntdb.Tx) error {
		for _, versionData := range ruleVersions {
			rule := b.rulePool.Get().(*rulev1.VersionedRule)
			umErr := json.Unmarshal([]byte(versionData), rule)
			if umErr != nil {
				return umErr
			}
			rules = append(rules, rule)
			b.rulePool.Put(rule)
		}
		return nil
	})

	return
}

func (b *buntdbContainer) RemoveRule(_ context.Context, ownerID string, contextID string, ruleID string) error {
	rk := storage.RuleKey(ownerID, contextID, ruleID)
	vkSearch := storage.RuleKey(ownerID, contextID, ruleID) + ":v*"
	var ruleVersions []string
	err := b.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(vkSearch, func(key, val string) bool {
			ruleVersions = append(ruleVersions, key)
			return true
		})
	})
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *buntdb.Tx) error {
		// all versions
		for _, vk := range ruleVersions {
			oldVal, e := tx.Delete(vk)
			if e == nil {
				b.logger.Info("Successfully removed rule from context.", zap.String("key", vk), zap.String("value", oldVal))
			}
		}
		// latest pointer
		oldVal, e := tx.Delete(rk)
		if e == nil {
			b.logger.Info("Successfully removed rule from context.", zap.String("key", rk), zap.String("value", oldVal))
		}
		return err
	})
}

func (b *buntdbContainer) ruleLatestVersion(ruleKey string) (version uint64, err error) {
	version = 0
	err = b.db.View(func(tx *buntdb.Tx) error {
		vk, gErr := tx.Get(ruleKey)
		if gErr != nil {
			return gErr
		}
		version = storage.VersionFromVersionKey(vk)
		return nil
	})
	return
}

func (b *buntdbContainer) ReplaceRule(_ context.Context, ownerID string, contextID string, rule *rulev1.Rule) (*rulev1.VersionedRule, error) {
	rk := storage.RuleKey(ownerID, contextID, rule.Id)
	currentVersion, err := b.ruleLatestVersion(rk)
	if err != nil {
		return nil, err
	}
	vRule := &rulev1.VersionedRule{
		Id:          rule.Id,
		Version:     currentVersion + 1,
		Description: rule.Description,
		Actions:     rule.Actions,
		Expression:  rule.Expression,
	}
	vk := storage.VersionKey(ownerID, contextID, vRule.Id, vRule.Version)
	ruleJSON, jsonErr := json.Marshal(vRule)
	if jsonErr != nil {
		return nil, jsonErr
	}

	err = b.db.Update(func(tx *buntdb.Tx) error {
		_, _, sErr := tx.Set(rk, vk, nil)
		if sErr != nil {
			return sErr
		}
		_, _, vErr := tx.Set(vk, string(ruleJSON), nil)
		if vErr != nil {
			return vErr
		}
		return nil
	})

	return vRule, err
}

func (b *buntdbContainer) Exists(_ context.Context, ownerID string, contextID string, ruleID string) (exists bool, err error) {
	err = b.db.View(func(tx *buntdb.Tx) error {
		_, ierr := tx.Get(storage.RuleKey(ownerID, contextID, ruleID), true)
		switch {
		case errors.Is(ierr, buntdb.ErrNotFound):
			exists = false
		case err == nil:
			exists = true
		default:
			return ierr
		}
		return nil
	})
	return
}

func (b *buntdbContainer) SaveAccessToken(c context.Context, ownerID string, tokenPair *authv1.TokenPair) error {
	atKey := fmt.Sprintf("access:%s", ownerID)

	tokens, err := b.GetAccessTokens(c, ownerID)
	if err != nil && !errors.Is(err, buntdb.ErrNotFound) {
		return err
	}

	tokens = append(tokens, tokenPair)

	bts, err := json.Marshal(tokens)
	if err != nil {
		return err
	}
	return b.db.Update(func(tx *buntdb.Tx) error {
		_, _, ierr := tx.Set(atKey, string(bts), nil)
		return ierr
	})
}

func (b *buntdbContainer) GetAccessTokens(_ context.Context, ownerID string) (tp []*authv1.TokenPair, err error) {
	atKey := fmt.Sprintf("access:%s", ownerID)

	val := ""
	err = b.db.View(func(tx *buntdb.Tx) error {
		v, ierr := tx.Get(atKey)
		if ierr != nil {
			return ierr
		}
		val = v
		return nil
	})
	if err != nil {
		return nil, err
	}

	if val == "" {
		return nil, fmt.Errorf("tokenPair not found")
	}

	if err = json.Unmarshal([]byte(val), &tp); err != nil {
		return nil, fmt.Errorf("failed to read tokenPairs: %w", err)
	}

	return tp, nil
}

func (b *buntdbContainer) createIndices() {
	err := b.db.CreateIndex("contexts", storage.ContextKey("*", "*"))
	if err != nil {
		panic(err)
	}

	err = b.db.CreateIndex("rules", storage.RuleKey("*", "*", "*"), buntdb.IndexJSONCaseSensitive("id"))
	if err != nil {
		panic(err)
	}
}
