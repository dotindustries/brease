package api

import (
	contextv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/context/v1"
	"connectrpc.com/connect"
	"context"
	"go.dot.industries/brease/auth"
)

func (b *BreaseHandler) ListContexts(ctx context.Context, c *connect.Request[contextv1.ListContextsReqeust]) (*connect.Response[contextv1.ListContextsResponse], error) {
	orgID, _, _, cErr := permissionCheck(ctx, auth.PermissionListContext)
	if cErr != nil {
		return nil, cErr
	}

	// read contexts from db
	list, err := b.db.ListContexts(ctx, orgID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&contextv1.ListContextsResponse{ContextIds: list}), nil
}
