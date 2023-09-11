package api

import (
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/cache"
	"go.dot.industries/brease/code"
	"go.dot.industries/brease/storage"
	"go.uber.org/zap"
)

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
	return &BreaseHandler{
		db:        db,
		logger:    logger,
		assembler: code.NewAssembler(logger, c),
		compiler:  code.NewCompiler(logger),
		token:     auth.NewToken(logger),
	}
}
