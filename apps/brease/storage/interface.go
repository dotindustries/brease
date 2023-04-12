package storage

import (
	"go.dot.industries/brease/models"
)

type Database interface {
	AddRule(ownerID string, contextID string, rule models.Rule) error
	Close() error
	Rules(ownerID string, contextID string) ([]models.Rule, error)
	RemoveRule(ownerID string, contextID string, ruleID string) error
	ReplaceRule(ownerID string, contextID string, rule models.Rule) error
	Exists(ownerID string, contextID string, ruleID string) (exists bool, err error)
	SaveAccessToken(ownerID string, tokenPair *models.TokenPair) error
	GetAccessToken(ownerID string) (*models.TokenPair, error)
}
