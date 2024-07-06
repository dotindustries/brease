package api

import (
	"buf.build/gen/go/dot/brease/grpc/go/brease/auth/v1/authv1grpc"
	"buf.build/gen/go/dot/brease/grpc/go/brease/context/v1/contextv1grpc"
	authv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/auth/v1"
	contextv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/context/v1"
	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"connectrpc.com/connect"
	"context"
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/cache"
	"go.dot.industries/brease/code"
	"go.dot.industries/brease/storage"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OpenApiHandler interface {
	contextv1grpc.ContextServiceServer
	authv1grpc.AuthServiceServer
}

type BreaseHandler struct {
	db        storage.Database
	logger    *zap.Logger
	assembler *code.Assembler
	compiler  *code.Compiler
	token     auth.Token
	OpenApi   OpenApiHandler
}

func NewHandler(db storage.Database, c cache.Cache, logger *zap.Logger) *BreaseHandler {
	if db == nil {
		panic("database cannot be nil")
	}
	bh := &BreaseHandler{
		db:        db,
		logger:    logger,
		assembler: code.NewAssembler(logger, c),
		compiler:  code.NewCompiler(logger),
		token:     auth.NewToken(logger),
	}

	bh.OpenApi = &openApiHandler{handler: bh}

	return bh
}

type openApiHandler struct {
	handler *BreaseHandler
}

func (o *openApiHandler) GetToken(ctx context.Context, empty *emptypb.Empty) (*authv1.TokenPair, error) {
	r, err := o.handler.GetToken(ctx, connect.NewRequest(empty))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) RefreshToken(ctx context.Context, request *authv1.RefreshTokenRequest) (*authv1.TokenPair, error) {
	r, err := o.handler.RefreshToken(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) ListRules(ctx context.Context, request *contextv1.ListRulesRequest) (*contextv1.ListRulesResponse, error) {
	r, err := o.handler.ListRules(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) GetRule(ctx context.Context, request *contextv1.GetRuleRequest) (*rulev1.VersionedRule, error) {
	r, err := o.handler.GetRule(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) GetRuleVersions(ctx context.Context, request *contextv1.ListRuleVersionsRequest) (*contextv1.ListRuleVersionsResponse, error) {
	r, err := o.handler.GetRuleVersions(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) CreateRule(ctx context.Context, request *contextv1.CreateRuleRequest) (*rulev1.VersionedRule, error) {
	r, err := o.handler.CreateRule(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) UpdateRule(ctx context.Context, request *contextv1.UpdateRuleRequest) (*rulev1.VersionedRule, error) {
	r, err := o.handler.UpdateRule(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) DeleteRule(ctx context.Context, request *contextv1.DeleteRuleRequest) (*emptypb.Empty, error) {
	r, err := o.handler.DeleteRule(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}

func (o *openApiHandler) Evaluate(ctx context.Context, request *contextv1.EvaluateRequest) (*contextv1.EvaluateResponse, error) {
	r, err := o.handler.Evaluate(ctx, connect.NewRequest(request))
	if err != nil {
		return nil, err
	}
	return r.Msg, nil
}
