package storage

import (
	authv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/auth/v1"
	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"context"
)

type Database interface {
	Close() error
	AddRule(ctx context.Context, ownerID string, contextID string, rule *rulev1.Rule) (*rulev1.VersionedRule, error)
	Rules(ctx context.Context, ownerID string, contextID string, pageSize int, pageToken string) ([]*rulev1.VersionedRule, error)
	RemoveRule(ctx context.Context, ownerID string, contextID string, ruleID string) error
	ReplaceRule(ctx context.Context, ownerID string, contextID string, rule *rulev1.Rule) (*rulev1.VersionedRule, error)
	Exists(ctx context.Context, ownerID string, contextID string, ruleID string) (exists bool, err error)
	RuleVersions(ctx context.Context, ownerID string, contextID string, ruleID string, pageSize int, pageToken string) ([]*rulev1.VersionedRule, error)
	SaveAccessToken(ctx context.Context, ownerID string, tokenPair *authv1.TokenPair) error
	GetAccessTokens(ctx context.Context, ownerID string) ([]*authv1.TokenPair, error)
	GetObjectSchema(ctx context.Context, ownerID string, contextID string) (string, error)
	ReplaceObjectSchema(ctx context.Context, ownerID string, contextID string, schema string) error
}
