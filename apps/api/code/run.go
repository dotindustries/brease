package code

import (
	"context"
	"github.com/d5/tengo/v2"
	"github.com/juju/errors"
	"go.dot.industries/brease/models"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

const (
	resultVariable = "results"
)

type Run struct {
	object map[string]interface{}
	logger *zap.Logger
}

type Object struct {
	tengo.ObjectImpl
}

func NewRun(_ context.Context, logger *zap.Logger, object map[string]interface{}) (*Run, error) {
	return &Run{
		object: object,
		logger: logger,
	}, nil
}

func (r *Run) Execute(ctx context.Context, script *Script) ([]models.EvaluationResult, error) {
	ctx, span := trace.StartSpan(ctx, "exec")
	defer span.End()

	c := script.compiled.Clone()
	start := time.Now()
	err := c.Set(objectVariable, r.object)
	if err != nil {
		r.logger.Error("Failed to set object on script", zap.Error(err))
		return nil, errors.Errorf("failed to setup run: %v", err)
	}
	if err = c.RunContext(ctx); err != nil {
		r.logger.Error("Failed to execute run", zap.Error(err))
		return nil, errors.Errorf("failed to execute run: %v", err)
	}
	r.logger.Info("Run execution finished", zap.Duration("time", time.Since(start)))
	resVar := script.compiled.Get(resultVariable)
	results := r.parseResults(resVar)
	r.logger.Debug("Run results", zap.Array("results", resultsArray(results)))
	return results, nil
}

type resultsArray []models.EvaluationResult

func (r resultsArray) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range r {
		if e := arr.AppendObject(r[i]); e != nil {
			return e
		}
	}
	return nil
}

func (r *Run) parseResults(result *tengo.Variable) (results []models.EvaluationResult) {
	// transform results for sending back
	rawResults := result.Array()
	for _, raw := range rawResults {
		r := raw.(map[string]interface{})
		results = append(results, models.EvaluationResult{
			Action:     r["action"].(string),
			TargetID:   r["targetID"].(string),
			TargetType: r["targetType"].(string),
			Value:      r["value"].(string),
		})
	}
	return
}
