package api

import (
	contextv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/context/v1"
	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"connectrpc.com/connect"
	"context"
	"fmt"
	"go.dot.industries/brease/code"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/juju/errors"
	"go.dot.industries/brease/auth"
)

func (b *BreaseHandler) Evaluate(ctx context.Context, c *connect.Request[contextv1.EvaluateRequest]) (*connect.Response[contextv1.EvaluateResponse], error) {
	orgID := auth.CtxString(ctx, auth.ContextOrgKey)
	if !auth.HasPermission(ctx, auth.PermissionEvaluate) && !auth.HasPermission(ctx, auth.PermissionRead) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("permission denied"))
	}
	codeBlock, err := b.findCode(ctx, c.Msg, orgID)
	if err != nil {
		return nil, err
	}

	results, err := b.run(ctx, codeBlock, c.Msg.Object)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&contextv1.EvaluateResponse{
		Results: results,
	}), nil
}

func (b *BreaseHandler) run(ctx context.Context, codeBlock string, object *structpb.Struct) ([]*rulev1.EvaluationResult, error) {
	compiledScript, err := b.findScript(ctx, codeBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to compile code: %v", err)
	}

	run, err := code.NewRun(ctx, b.logger, object)
	if err != nil {
		return nil, fmt.Errorf("failed to create run context: %v", err)
	}

	return run.Execute(ctx, compiledScript)
}

// Override code takes precedence
func (b *BreaseHandler) findCode(ctx context.Context, r *contextv1.EvaluateRequest, orgID string) (string, error) {
	if r.OverrideCode != "" {
		return r.OverrideCode, nil
	}

	rules, err := b.findRules(ctx, orgID, r.ContextId, r.OverrideRules)
	if err != nil {
		return "", err
	}

	c, err := b.assembler.BuildCode(ctx, rules)
	if err != nil {
		return "", errors.Errorf("Failed to assemble code: %v", err)
	}

	return c, nil
}

// Override rules take precedence
func (b *BreaseHandler) findRules(ctx context.Context, orgID string, contextID string, overrideRules []*rulev1.Rule) ([]*rulev1.VersionedRule, error) {
	if overrideRules != nil {
		vRs := make([]*rulev1.VersionedRule, len(overrideRules))
		for i, rule := range overrideRules {
			vRs[i] = &rulev1.VersionedRule{
				Id:          rule.Id,
				Version:     0, // override rules don't have versioning
				Description: rule.Description,
				Actions:     rule.Actions,
				Expression:  rule.Expression,
			}
		}
		return vRs, nil
	}

	rules, err := b.db.Rules(ctx, orgID, contextID, 0, "")
	if err != nil {
		return nil, fmt.Errorf("rules not found for context: %s", contextID)
	}

	return rules, nil
}

func (b *BreaseHandler) findScript(ctx context.Context, codeBlock string) (*code.Script, error) {
	script, err := b.compiler.CompileCode(ctx, codeBlock)
	if err != nil {
		return nil, errors.Errorf("Failed to compile code block: %v", err)
	}
	return script, nil
}
