package api

import (
	contextv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/context/v1"
	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"connectrpc.com/connect"
	"context"
	"fmt"
	"go.uber.org/zap"

	"go.dot.industries/brease/auth"
)

func (b *BreaseHandler) UpdateRule(ctx context.Context, c *connect.Request[contextv1.UpdateRuleRequest]) (*connect.Response[rulev1.VersionedRule], error) {
	orgID, _, _, cErr := permissionCheck(ctx, auth.PermissionCreateRule)
	if cErr != nil {
		b.logger.Warn("UpdateRule", zap.String("contextID", c.Msg.ContextId), zap.String("orgID", orgID))
		return nil, cErr
	}
	updatedRule, err := b.db.ReplaceRule(ctx, orgID, c.Msg.ContextId, c.Msg.Rule)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update rule: %v", err))
	}
	return connect.NewResponse(updatedRule), nil
}
