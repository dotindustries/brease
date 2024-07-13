package api

import (
	v1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/context/v1"
	v11 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"connectrpc.com/connect"
	"context"
	"fmt"

	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/models"
	"go.uber.org/zap"
)

func (b *BreaseHandler) CreateRule(ctx context.Context, c *connect.Request[v1.CreateRuleRequest]) (*connect.Response[v11.VersionedRule], error) {
	orgID := auth.CtxString(ctx, auth.ContextOrgKey)
	ctxID := c.Msg.ContextId
	rule := c.Msg.Rule

	if !auth.HasPermission(ctx, auth.PermissionWrite) {
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
