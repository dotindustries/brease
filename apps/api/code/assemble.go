package code

import (
	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"context"
	"fmt"
	"go.dot.industries/brease/cache"
	"go.dot.industries/brease/trace"
	"go.uber.org/zap"
)

type Assembler struct {
	logger *zap.Logger
	cache  cache.Cache
}

func NewAssembler(logger *zap.Logger, c cache.Cache) *Assembler {
	return &Assembler{
		logger: logger,
		cache:  c,
	}
}

func (a *Assembler) BuildCode(ctx context.Context, rules []*rulev1.VersionedRule) (string, error) {
	ctx, span := trace.Tracer.Start(ctx, "code")
	defer span.End()

	key := cache.SimpleHash(rules)
	if a.cache != nil {
		code := a.cache.Get(ctx, key).(string)
		if code != "" {
			return code, nil
		}
	}

	assembled, err := a.assemble(ctx, rules)
	if err != nil {
		a.logger.Error("code assembly failed", zap.Error(err))
		return "", err
	}

	if assembled == "" {
		return "", fmt.Errorf("assembled code is empty")
	}

	if a.cache != nil && !a.cache.Set(ctx, key, assembled) {
		a.logger.Error("cannot cache assembled code", zap.String("code", assembled))
		return "", fmt.Errorf("cannot cache assembled code")
	}

	return assembled, nil
}

func (a *Assembler) assemble(ctx context.Context, rules []*rulev1.VersionedRule) (string, error) {
	ctx, span := trace.Tracer.Start(ctx, "assemble")
	defer span.End()

	relevantRules := make([]*rulev1.VersionedRule, 0, len(rules))
	for i := 0; i < len(rules); i++ {
		if rules[i].Expression == nil {
			continue
		}
		relevantRules = append(relevantRules, rules[i])
	}

	code, err := a.parseRules(ctx, relevantRules)
	if err != nil {
		return "", fmt.Errorf("failed to parse rules: %v", err)
	}

	return code, nil
}
