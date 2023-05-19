package code

import (
	"context"
	"fmt"
	"strconv"

	"go.dot.industries/brease/cache"
	"go.dot.industries/brease/models"
	"go.dot.industries/brease/pb"
	"go.dot.industries/brease/rref"
	"go.dot.industries/brease/worker"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

type Assembler struct {
	logger *zap.Logger
	cache  cache.Cache
}

func NewAssembler(logger *zap.Logger, c cache.Cache) *Assembler {
	return &Assembler{
		logger: logger,
		cache:  c,
	}
}

func (a *Assembler) BuildCode(ctx context.Context, rules []models.Rule) (string, error) {
	ctx, span := trace.StartSpan(ctx, "code")
	defer span.End()

	key := cache.SimpleHash(rules)
	code := a.cache.Get(ctx, key).(string)
	if code != "" {
		return code, nil
	}

	assembled, err := a.assemble(ctx, rules)
	if err != nil {
		a.logger.Error("code assembly failed", zap.Error(err))
		return "", err
	}

	if assembled == "" {
		return "", fmt.Errorf("assembled code is empty")
	}

	if !a.cache.Set(ctx, key, assembled) {
		a.logger.Error("cannot cache assembled code", zap.String("code", assembled))
		return "", fmt.Errorf("cannot cache assembled code")
	}

	return assembled, nil
}

func (a *Assembler) assemble(ctx context.Context, rules []models.Rule) (string, error) {
	ctx, span := trace.StartSpan(ctx, "assemble")
	defer span.End()

	relevantRules := make([]models.Rule, 0, len(rules))
	for i := 0; i < len(rules); i++ {
		if rules[i].Expression == nil {
			continue
		}
		relevantRules = append(relevantRules, rules[i])
	}

	if rref.IsConfigured() {
		// FIXME: do we need to replace base references before building code blocks?
		//   should this entire feature be client only?
		relevantRules = a.lookupReferences(ctx, relevantRules)
	} else {
		a.logger.Warn("Lookup for reference valued expressions is off. The dref library is not configured")
	}

	code, err := a.parseRules(ctx, relevantRules)
	if err != nil {
		return "", fmt.Errorf("failed to parse rules: %v", err)
	}

	return code, nil
}

func (a *Assembler) lookupReferences(ctx context.Context, rules []models.Rule) []models.Rule {
	ctx, span := trace.StartSpan(ctx, "references")
	defer span.End()

	jobs := make([]worker.Job, len(rules))
	for i := 0; i < len(rules); i++ {
		jobs[i] = worker.Job{
			Descriptor: worker.JobDescriptor{
				ID:    worker.JobID(fmt.Sprintf("%v", i)),
				JType: "lookup",
			},
			ExecFn: a.lookupReferencesExec,
			Args:   rules[i],
		}
	}

	refLookupPool := worker.New(len(rules))
	refLookupPool.GenerateFrom(jobs)
	go refLookupPool.Run(ctx)

	results := make([]models.Rule, len(rules))
	for {
		select {
		case r, ok := <-refLookupPool.Results():
			if !ok {
				continue
			}

			i, _ := strconv.ParseInt(string(r.Descriptor.ID), 10, 64)

			results[i] = r.Value.(models.Rule)
		case <-refLookupPool.Done:
			// worker finished
			break
		}
	}
}

func (a *Assembler) lookupReferencesExec(ctx context.Context, args interface{}) (interface{}, error) {
	rule := args.(models.Rule)
	ctx, span := trace.StartSpan(ctx, fmt.Sprintf("lookup-%s", rule.ID))
	defer span.End()

	expr, err := models.ValidateExpression(rule.Expression)
	if err != nil {
		return nil, err
	}

	lookupReferenceValues(ctx, expr)

	// return modified rule
	return rule, nil
}

// lookupReferenceValues is a recursive function which fills in condition base values
// from remote references using the rref library
func lookupReferenceValues(ctx context.Context, expr *pb.Expression) {
	switch expr.Expr.(type) {
	case *pb.Expression_And:
		e := expr.GetAnd().Expression
		for _, e2 := range e {
			lookupReferenceValues(ctx, e2)
		}
	case *pb.Expression_Or:
		e := expr.GetOr().Expression
		for _, e2 := range e {
			lookupReferenceValues(ctx, e2)
		}
	case *pb.Expression_Condition:
		condition := expr.GetCondition()
		switch condition.Base.(type) {
		// we only fill in references
		case *pb.Condition_Ref:
			ref := condition.GetRef()
			ref.Value = rref.LookupReferenceValue(ctx, ref)
		}
	}
}
