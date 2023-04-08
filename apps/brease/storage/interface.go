package storage

import (
	"go.dot.industries/brease/models"
)

type Database interface {
	AddRule(contextID string, rule models.Rule) error
	Close() error
	Rules(contextID string) ([]models.Rule, error)
	RemoveRule(contextID string, ruleID string) error
	ReplaceRule(contextID string, rule models.Rule) error
}
