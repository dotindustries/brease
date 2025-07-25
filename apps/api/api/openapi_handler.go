package api

import (
	authv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/auth/v1"
	contextv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/context/v1"
	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/janvaclavik/govar"
	"go.dot.industries/brease/auth"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type openApiHandler struct {
	handler *BreaseHandler
}

func (o *openApiHandler) GetObjectSchema(ctx context.Context, request *contextv1.GetObjectSchemaRequest) (*contextv1.GetObjectSchemaResponse, error) {
	ctx = o.forwardMetadata(ctx)
	r, err := o.handler.GetObjectSchema(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) ReplaceObjectSchema(ctx context.Context, request *contextv1.ReplaceObjectSchemaRequest) (*contextv1.ReplaceObjectSchemaResponse, error) {
	ctx = o.forwardMetadata(ctx)
	r, err := o.handler.ReplaceObjectSchema(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) GetToken(ctx context.Context, empty *emptypb.Empty) (*authv1.TokenPair, error) {
	ctx = o.forwardMetadata(ctx)
	r, err := o.handler.GetToken(ctx, connect.NewRequest(empty))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) RefreshToken(ctx context.Context, request *authv1.RefreshTokenRequest) (*authv1.TokenPair, error) {
	ctx = o.forwardMetadata(ctx)
	r, err := o.handler.RefreshToken(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) ListRules(ctx context.Context, request *contextv1.ListRulesRequest) (*contextv1.ListRulesResponse, error) {
	ctx = o.forwardMetadata(ctx)
	govar.Dump(ctx)
	r, err := o.handler.ListRules(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) GetRule(ctx context.Context, request *contextv1.GetRuleRequest) (*rulev1.VersionedRule, error) {
	ctx = o.forwardMetadata(ctx)
	r, err := o.handler.GetRule(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) GetRuleVersions(ctx context.Context, request *contextv1.ListRuleVersionsRequest) (*contextv1.ListRuleVersionsResponse, error) {
	ctx = o.forwardMetadata(ctx)
	r, err := o.handler.GetRuleVersions(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) CreateRule(ctx context.Context, request *contextv1.CreateRuleRequest) (*rulev1.VersionedRule, error) {
	ctx = o.forwardMetadata(ctx)
	r, err := o.handler.CreateRule(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) UpdateRule(ctx context.Context, request *contextv1.UpdateRuleRequest) (*rulev1.VersionedRule, error) {
	ctx = o.forwardMetadata(ctx)
	r, err := o.handler.UpdateRule(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) DeleteRule(ctx context.Context, request *contextv1.DeleteRuleRequest) (*emptypb.Empty, error) {
	ctx = o.forwardMetadata(ctx)
	r, err := o.handler.DeleteRule(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) Evaluate(ctx context.Context, request *contextv1.EvaluateRequest) (*contextv1.EvaluateResponse, error) {
	ctx = o.forwardMetadata(ctx)
	r, err := o.handler.Evaluate(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) forwardMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	if orgIds := md.Get(auth.ContextOrgKey); len(orgIds) > 0 {
		ctx = context.WithValue(ctx, auth.ContextOrgKey, orgIds[0])
	}
	if userIds := md.Get(auth.ContextUserIDKey); len(userIds) > 0 {
		ctx = context.WithValue(ctx, auth.ContextUserIDKey, userIds[0])
	}

	return ctx
}
