package api

import (
	contextv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/context/v1"
	"connectrpc.com/connect"
	"context"
)

func (b *BreaseHandler) GetObjectSchema(ctx context.Context, c *connect.Request[contextv1.GetObjectSchemaRequest]) (*connect.Response[contextv1.GetObjectSchemaResponse], error) {
	//TODO implement me
	panic("implement me")
}

func (b *BreaseHandler) ReplaceObjectSchema(ctx context.Context, c *connect.Request[contextv1.ReplaceObjectSchemaRequest]) (*connect.Response[contextv1.ReplaceObjectSchemaResponse], error) {
	//TODO implement me
	panic("implement me")
}
