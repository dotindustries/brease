package code

import (
	"context"
	"fmt"
	"github.com/d5/tengo/v2"
	"go.dot.industries/brease/cache"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"time"
)

type Compiler struct {
	logger *zap.Logger
	cache  cache.Cache
}

func NewCompiler(logger *zap.Logger, c cache.Cache) *Compiler {
	return &Compiler{
		logger: logger,
		cache:  c,
	}
}

func (c *Compiler) CompileCode(ctx context.Context, codeBlock string) (*Script, error) {
	ctx, span := trace.StartSpan(ctx, "code")
	defer span.End()

	key := cache.SimpleHash(codeBlock)
	cachedScript := c.cache.Get(ctx, key)
	script, ok := cachedScript.(*tengo.Compiled)
	if !ok {
		c.logger.Error("Funky cached value of compiled script", zap.Any("cachedValue", cachedScript))
	}
	if script != nil {
		return &Script{compiled: script}, nil
	}

	compiled, err := c.compile(ctx, codeBlock)
	if err != nil {
		c.logger.Error("Code compilation failed", zap.Error(err))
		return nil, err
	}

	if c.cache.Set(ctx, key, compiled) {
		return nil, fmt.Errorf("failed to cache compiled script")
	}

	return &Script{compiled: compiled}, nil
}

func (c *Compiler) compile(_ context.Context, codeBlock string) (*tengo.Compiled, error) {
	ts := tengo.NewScript([]byte(codeBlock))
	ts.SetImports(moduleMaps())
	ts.Add(objectVariable, nil)
	start := time.Now()
	compiled, err := ts.Compile()
	if err != nil {
		c.logger.Error("Failed to compile script", zap.Error(err))
		return nil, err
	}
	c.logger.Info("Compiled script", zap.Duration("time", time.Since(start)))
	return compiled, nil
}
