package api

import (
	contextv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/context/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/pkg/errors"
	"go.dot.industries/brease/auth"
	"go.uber.org/zap"
)

func (b *BreaseHandler) GetObjectSchema(ctx context.Context, c *connect.Request[contextv1.GetObjectSchemaRequest]) (*connect.Response[contextv1.GetObjectSchemaResponse], error) {
	orgID, _, _, cErr := permissionCheck(ctx, auth.PermissionSchemaRead)
	if cErr != nil {
		b.logger.Warn("GetObjectSchema", zap.String("contextID", c.Msg.ContextId), zap.String("orgID", orgID))
		return nil, cErr
	}
	schema, err := b.db.GetObjectSchema(ctx, orgID, c.Msg.ContextId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.Wrap(err, "failed to fetch object schema"))
	}
	return connect.NewResponse(&contextv1.GetObjectSchemaResponse{
		Schema: schema,
	}), nil
}

func (b *BreaseHandler) ReplaceObjectSchema(ctx context.Context, c *connect.Request[contextv1.ReplaceObjectSchemaRequest]) (*connect.Response[contextv1.ReplaceObjectSchemaResponse], error) {
	orgID, _, _, cErr := permissionCheck(ctx, auth.PermissionSchemaEdit)
	if cErr != nil {
		b.logger.Warn("ReplaceObjectSchema", zap.String("contextID", c.Msg.ContextId), zap.String("orgID", orgID))
		return nil, cErr
	}

	// verify schema validity
	_, err := b.jsonSchemaCompiler.Compile([]byte(c.Msg.Schema))
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid json schema"))
	}

	err = b.db.ReplaceObjectSchema(ctx, orgID, c.Msg.ContextId, c.Msg.Schema)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.Wrap(err, "failed to replace object schema"))
	}

	return connect.NewResponse(&contextv1.ReplaceObjectSchemaResponse{}), nil
}
