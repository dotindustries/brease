package api

import (
	contextv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/context/v1"
	"connectrpc.com/connect"
	"context"
	"go.dot.industries/brease/auth"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (b *BreaseHandler) DeleteRule(ctx context.Context, c *connect.Request[contextv1.DeleteRuleRequest]) (*connect.Response[emptypb.Empty], error) {
	orgID := auth.CtxString(ctx, auth.ContextOrgKey)
	_ = b.db.RemoveRule(ctx, orgID, c.Msg.ContextId, c.Msg.RuleId)
	// we don't expose whether we succeeded
	return connect.NewResponse(&emptypb.Empty{}), nil
}
