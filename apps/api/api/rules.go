package api

import (
	"context"
	"fmt"

	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"connectrpc.com/connect"
	"go.dot.industries/brease/models"
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

func (b *BreaseHandler) CreateRule(ctx context.Context, c *connect.Request[v1.CreateRuleRequest]) (*connect.Response[rulev1.VersionedRule], error) {
	orgID := auth.CtxString(ctx, auth.ContextOrgKey)
	ctxID := c.Msg.ContextId
	rule := c.Msg.Rule

	if !auth.HasPermission(ctx, auth.PermissionCreateRule) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("permission denied"))
	}

	if rule == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("new rule cannot be missing"))
	}

	// fill rule id if not set
	// FIXME: we really should convince the community to use business meaningful ids
	if rule.Id == "" {
		id := models.NewRuleID()
		rule.Id = id.String()
	}

	// duplicate check
	if exists, err := b.db.Exists(ctx, orgID, ctxID, rule.Id); exists && err == nil {
		b.logger.Warn("Rule already exists. Aborting create rule", zap.String("context_id", ctxID), zap.String("rule_id", rule.Id))
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("duplicate rule ID '%s'", rule.Id))
	} else if err != nil {
		// early exit, failed to check if already exists, better stop
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	newRule, err := b.db.AddRule(ctx, orgID, ctxID, rule)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(newRule), nil
}
