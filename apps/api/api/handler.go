package api

import (
	"go.dot.industries/brease/storage"
	"go.uber.org/zap"
)

type PathParams struct {
	ContextID string `path:"contextID"`
}

type BreaseHandler struct {
	db     storage.Database
	logger *zap.Logger
}

func NewHandler(db storage.Database, logger *zap.Logger) *BreaseHandler {
	if db == nil {
		panic("database cannot be nil")
	}
	return &BreaseHandler{
		db:     db,
		logger: logger,
	}
}
