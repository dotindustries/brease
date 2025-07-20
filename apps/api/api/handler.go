package api

import (
	"buf.build/gen/go/dot/brease/grpc/go/brease/auth/v1/authv1grpc"
	"buf.build/gen/go/dot/brease/grpc/go/brease/context/v1/contextv1grpc"
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/cache"
	"go.dot.industries/brease/code"
	"go.dot.industries/brease/storage"
	"go.uber.org/zap"
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

	return bh
}
