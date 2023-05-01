package buntdb

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/goccy/go-json"
	"github.com/tidwall/buntdb"
	"go.dot.industries/brease/models"
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
	db, err := buntdb.Open(opts.Path)
	if err != nil {
		return nil, err
	}

	b := &buntdbContainer{
		db:     db,
		logger: opts.Logger,
		rulePool: sync.Pool{
			New: func() interface{} {
				return new(models.Rule)
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

func (b *buntdbContainer) GetAccessToken(_ context.Context, orgID string) (*models.TokenPair, error) {
	atKey := fmt.Sprintf("access:%s", orgID)

	val := ""
	err := b.db.View(func(tx *buntdb.Tx) error {
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

	tp := &models.TokenPair{}
	if err = json.Unmarshal([]byte(val), tp); err != nil {
		return nil, fmt.Errorf("failed to read tokenPair: %v", err)
	}

	return tp, nil
}

func (b *buntdbContainer) SaveAccessToken(_ context.Context, orgID string, tokenPair *models.TokenPair) error {
	atKey := fmt.Sprintf("access:%s", orgID)

	bts, err := json.Marshal(tokenPair)
	if err != nil {
		return err
	}
	return b.db.Update(func(tx *buntdb.Tx) error {
		_, _, ierr := tx.Set(atKey, string(bts), nil)
		return ierr
	})
}

func (b *buntdbContainer) Exists(_ context.Context, orgID string, contextID string, ruleID string) (exists bool, err error) {
	err = b.db.View(func(tx *buntdb.Tx) error {
		_, ierr := tx.Get(ruleKey(orgID, contextID, ruleID), true)
		if errors.Is(ierr, buntdb.ErrNotFound) {
			exists = false
		} else if err == nil {
			exists = true
		} else {
			return ierr
		}
		return nil
	})
	return
}

func (b *buntdbContainer) ReplaceRule(_ context.Context, orgID string, contextID string, rule models.Rule) error {
	rk := ruleKey(orgID, contextID, rule.ID)
	ruleJson, jsonErr := json.Marshal(rule)
	if jsonErr != nil {
		return jsonErr
	}
	return b.db.Update(func(tx *buntdb.Tx) error {
		if oldVal, replaced, err := tx.Set(rk, string(ruleJson), nil); replaced {
			b.logger.Info("Successfully updated rule in context.", zap.String("key", rk), zap.String("old", oldVal), zap.String("new", string(ruleJson)))
		} else {
			return err
		}

		return nil
	})
}

func (b *buntdbContainer) RemoveRule(_ context.Context, orgID string, contextID string, ruleID string) error {
	rk := ruleKey(orgID, contextID, ruleID)
	return b.db.Update(func(tx *buntdb.Tx) error {
		oldVal, err := tx.Delete(rk)
		if err == nil {
			b.logger.Info("Successfully removed rule from context.", zap.String("key", rk), zap.String("value", oldVal))
		}
		return err
	})
}

func (b *buntdbContainer) Rules(_ context.Context, orgID string, contextID string) (rules []models.Rule, err error) {
	rk := ruleKey(orgID, contextID, "*")

	var rawRules []string
	err = b.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(rk, func(key, val string) bool {
			rawRules = append(rawRules, val)
			return true
		})
	})

	for _, r := range rawRules {
		rule := b.rulePool.Get().(*models.Rule)
		umErr := json.Unmarshal([]byte(r), rule)
		if umErr != nil {
			return nil, umErr
		}
		b.rulePool.Put(rule)
		rules = append(rules, *rule)
	}

	return
}

func (b *buntdbContainer) AddRule(_ context.Context, orgID string, contextID string, rule models.Rule) error {
	rk := ruleKey(orgID, contextID, rule.ID)
	ruleJson, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	err = b.db.Update(func(tx *buntdb.Tx) error {
		_, _, txErr := tx.Set(rk, string(ruleJson), nil)
		if txErr != nil {
			return txErr
		}
		return nil
	})

	return err
}

func (b *buntdbContainer) Close() error {
	return b.db.Close()
}

func (b *buntdbContainer) createIndices() {
	err := b.db.CreateIndex("contexts", contextKey("*", "*"))
	if err != nil {
		panic(err)
	}

	err = b.db.CreateIndex("rules", ruleKey("*", "*", "*"), buntdb.IndexJSONCaseSensitive("id"))
	if err != nil {
		panic(err)
	}
}

func contextKey(orgID string, contextID string) string {
	return fmt.Sprintf("org:%s:context:%s", orgID, contextID)
}

func ruleKey(orgID string, contextID, ruleID string) string {
	return fmt.Sprintf("org:%s:context:%s:rule:%s", orgID, contextID, ruleID)
}