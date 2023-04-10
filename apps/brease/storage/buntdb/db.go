package buntdb

import (
	"errors"
	"fmt"
	"sync"

	"github.com/goccy/go-json"
	"github.com/tidwall/buntdb"
	"go.dot.industries/brease/models"
	"go.dot.industries/brease/storage"
	"go.uber.org/zap"
)

type BuntDbOptions struct {
	Path   string
	Logger *zap.Logger
}

func NewDatabase(opts BuntDbOptions) (storage.Database, error) {
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
	}

	b.createIndices()

	return b, nil
}

type buntdbContainer struct {
	db     *buntdb.DB
	logger *zap.Logger
}

func (b *buntdbContainer) Exists(contextID string, ruleID string) (exists bool, err error) {
	err = b.db.View(func(tx *buntdb.Tx) error {
		_, ierr := tx.Get(ruleKey(contextID, ruleID), true)
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

func (b *buntdbContainer) ReplaceRule(contextID string, rule models.Rule) error {
	rk := ruleKey(contextID, rule.ID)
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

func (b *buntdbContainer) RemoveRule(contextID string, ruleID string) error {
	rk := ruleKey(contextID, ruleID)
	return b.db.Update(func(tx *buntdb.Tx) error {
		oldVal, err := tx.Delete(rk)
		if err == nil {
			b.logger.Info("Successfully removed rule from context.", zap.String("key", rk), zap.String("value", oldVal))
		}
		return err
	})
}

func (b *buntdbContainer) Rules(contextID string) (rules []models.Rule, err error) {
	rk := ruleKey(contextID, "*")

	var rawRules []string
	err = b.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(rk, func(key, val string) bool {
			rawRules = append(rawRules, val)
			return true
		})
	})

	var rulePool = sync.Pool{
		New: func() interface{} {
			return new(models.Rule)
		},
	}
	for _, r := range rawRules {
		rule := rulePool.Get().(*models.Rule)
		umErr := json.Unmarshal([]byte(r), rule)
		if umErr != nil {
			return nil, umErr
		}
		rulePool.Put(rule)
		rules = append(rules, *rule)
	}

	return
}

func (b *buntdbContainer) AddRule(contextID string, rule models.Rule) error {
	rk := ruleKey(contextID, rule.ID)
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
	err := b.db.CreateIndex("contexts", contextKey("*"))
	if err != nil {
		panic(err)
	}

	err = b.db.CreateIndex("rules", ruleKey("*", "*"), buntdb.IndexJSONCaseSensitive("id"))
	if err != nil {
		panic(err)
	}
}

func contextKey(contextID string) string {
	return fmt.Sprintf("context:%s", contextID)
}

func ruleKey(contextID, ruleID string) string {
	return fmt.Sprintf("context:%s:rule:%s", contextID, ruleID)
}
