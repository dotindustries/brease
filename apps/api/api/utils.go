package api

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"go.dot.industries/brease/auth"
)

func permissionCheck(ctx context.Context, requiredPermissions ...string) (string, string, []string, *connect.Error) {
	orgID := auth.CtxString(ctx, auth.ContextOrgKey)
	if orgID == "" {
		return "", "", nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("missing orgID"))
	}

	userID := auth.CtxString(ctx, auth.ContextUserIDKey)
	if userID == "" {
		return "", "", nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("missing userID"))
	}

	for _, p := range requiredPermissions {
		if !auth.HasPermission(ctx, p) {
			return "", "", nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("permission denied"))
		}
	}

	permissions := auth.PermissionsFromContext(ctx)

	return orgID, userID, permissions, nil
}
