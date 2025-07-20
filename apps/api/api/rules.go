package api

import (
	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"connectrpc.com/connect"
	"context"
	"fmt"
	"go.uber.org/zap"

	v1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/context/v1"
	"go.dot.industries/brease/auth"
)

func (b *BreaseHandler) ListRules(ctx context.Context, c *connect.Request[v1.ListRulesRequest]) (*connect.Response[v1.ListRulesResponse], error) {
	orgID, _, _, cErr := permissionCheck(ctx, auth.PermissionReadRule)
	if cErr != nil {
		b.logger.Warn("ListRules", zap.String("contextID", c.Msg.ContextId), zap.String("orgID", orgID))
		return nil, cErr
	}
	pageToken := c.Msg.PageToken
	pageSize := c.Msg.PageSize

	rules, err := b.db.Rules(ctx, orgID, c.Msg.ContextId, int(pageSize), pageToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to fetch rules: %v", err))
	}
	code := ""
	if c.Msg.CompileCode {
		code, err = b.assembler.BuildCode(ctx, rules)
		if err != nil {
			b.logger.Warn("Failed to assemble code", zap.Error(err))
		} else {
			b.logger.Debug("Assembled code", zap.String("code", code))
		}
	}
	return connect.NewResponse(&v1.ListRulesResponse{
		Rules:         rules,
		Code:          code,
		NextPageToken: "",
	}), nil
}

func (b *BreaseHandler) GetRule(ctx context.Context, c *connect.Request[v1.GetRuleRequest]) (*connect.Response[rulev1.VersionedRule], error) {
	orgID := auth.CtxString(ctx, auth.ContextOrgKey)
	ruleID := c.Msg.RuleId
	if ruleID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("missing rule_id"))
	}
	rules, err := b.db.Rules(ctx, orgID, c.Msg.ContextId, 1, ruleID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to fetch rules: %v", err))
	}
	if l := len(rules); l == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("rule_id not found: %s", ruleID))
	} else if l > 1 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("rule_id is corrupted: %s", ruleID))
	}

	return connect.NewResponse(rules[0]), nil
}
