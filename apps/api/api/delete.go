package api

import (
	contextv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/context/v1"
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"go.dot.industries/brease/auth"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (b *BreaseHandler) DeleteRule(ctx context.Context, c *connect.Request[contextv1.DeleteRuleRequest]) (*connect.Response[emptypb.Empty], error) {
	orgID := auth.CtxString(ctx, auth.ContextOrgKey)
	if !auth.HasPermission(ctx, auth.PermissionWrite) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("permission denied"))
	}
	err := b.db.RemoveRule(ctx, orgID, c.Msg.ContextId, c.Msg.RuleId)
	if err != nil {
		sentry.CaptureException(err)
		b.logger.Error("Failed to delete rule", zap.Error(err))
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}
