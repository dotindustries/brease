package code

import (
	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"context"
	"github.com/d5/tengo/v2"
	"github.com/juju/errors"
	"go.dot.industries/brease/trace"
	"go.uber.org/zap"
	"log"
	"time"
)

const (
	resultVariable = "results"
)

type Run struct {
	object interface{}
	logger *zap.Logger
}

type Object struct {
	tengo.ObjectImpl
}

func NewRun(_ context.Context, logger *zap.Logger, object interface{}) (*Run, error) {
	return &Run{
		object: object,
		logger: logger,
	}, nil
}

func (r *Run) Execute(ctx context.Context, script *Script) ([]*rulev1.EvaluationResult, error) {
	ctx, span := trace.Tracer.Start(ctx, "exec")
	defer span.End()

	runner := script.compiled.Clone()
	start := time.Now()
	err := runner.Set(objectVariable, r.object)
	if err != nil {
		r.logger.Error("Failed to set object on script", zap.Error(err))
		return nil, errors.Errorf("failed to setup run: %v", err)
	}
	if err = runner.RunContext(ctx); err != nil {
		log.Println(extractCodeSection(err.Error(), script.codeBlock, 2))
		r.logger.Error("Failed to execute run", zap.Error(err))
		return nil, err
	}
	r.logger.Info("Run execution finished", zap.Duration("time", time.Since(start)))
	resVar := runner.Get(resultVariable)
	results := r.parseResults(resVar)
	r.logger.Debug("Run results", zap.Any("results", results))
	return results, nil
}

func (r *Run) parseResults(result *tengo.Variable) (results []*rulev1.EvaluationResult) {
	// transform results for sending back
	for _, raw := range result.Array() {
		res, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		// if result structure changes to dynamic value types, use
		//   github.com/mitchellh/mapstructure
		target := res["target"].(map[string]interface{})
		results = append(results, &rulev1.EvaluationResult{
			Action: res["action"].(string),
			Target: &rulev1.Target{
				Kind:  target["kind"].(string),
				Id:    target["target"].(string),
				Value: target["value"].([]byte),
			},
			By: res["by"].(*rulev1.RuleRef),
		})
	}
	return
}
