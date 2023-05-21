package code

import (
	"context"
	"github.com/d5/tengo/v2"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"log"
	"time"
)

type Compiler struct {
	logger *zap.Logger
}

func NewCompiler(logger *zap.Logger) *Compiler {
	return &Compiler{
		logger: logger,
	}
}

func (c *Compiler) CompileCode(ctx context.Context, codeBlock string) (*Script, error) {
	ctx, span := trace.StartSpan(ctx, "code")
	defer span.End()

	compiled, err := c.compile(ctx, codeBlock)
	if err != nil {
		c.logger.Error("Code compilation failed", zap.Error(err))
		return nil, err
	}

	return &Script{compiled: compiled, codeBlock: codeBlock}, nil
}

func (c *Compiler) compile(_ context.Context, codeBlock string) (*tengo.Compiled, error) {
	ts := tengo.NewScript([]byte(codeBlock))
	ts.SetImports(moduleMaps())
	ts.Add(objectVariable, nil)
	start := time.Now()
	compiled, err := ts.Compile()
	if err != nil {
		log.Println(extractCodeSection(err.Error(), codeBlock, 2))
		c.logger.Error("Failed to compile script", zap.Error(err))
		return nil, err
	}
	c.logger.Info("Compiled script", zap.Duration("time", time.Since(start)))
	return compiled, nil
}
