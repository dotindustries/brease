package storage

import (
	"context"

	"go.dot.industries/brease/models"
)

type Database interface {
	Close() error
	AddRule(ctx context.Context, ownerID string, contextID string, rule models.Rule) error
	Rules(ctx context.Context, ownerID string, contextID string) ([]models.Rule, error)
	RemoveRule(ctx context.Context, ownerID string, contextID string, ruleID string) error
	ReplaceRule(ctx context.Context, ownerID string, contextID string, rule models.Rule) error
	Exists(ctx context.Context, ownerID string, contextID string, ruleID string) (exists bool, err error)
	SaveAccessToken(ctx context.Context, ownerID string, tokenPair models.TokenPair) error
	GetAccessTokens(ctx context.Context, ownerID string) ([]models.TokenPair, error)
}
