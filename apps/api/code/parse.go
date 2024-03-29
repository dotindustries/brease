package code

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.dot.industries/brease/models"
	"go.dot.industries/brease/pb"
	"go.dot.industries/brease/trace"
	"go.dot.industries/brease/worker"
)

type parserArgs struct {
	rule     models.VersionedRule
	appendFn func(section string, ID string)
}

func (a *Assembler) parseRules(ctx context.Context, rules []models.VersionedRule) (string, error) {
	ctx, span := trace.Tracer.Start(ctx, "parse")
	defer span.End()

	code := strings.Builder{}
	code.WriteString(codeHeader)
	pool := worker.New(len(rules))
	mux := sync.Mutex{}

	appendFn := func(section, summary string) {
		mux.Lock()
		code.WriteString("// Rule: " + summary + "\n" + section + "\n\n")
		mux.Unlock()
	}

	jobs := make([]worker.Job, len(rules))
	for i := range rules {
		rule := rules[i]
		jobs[i] = worker.Job{
			Descriptor: worker.JobDescriptor{
				ID:    worker.JobID(rule.ID),
				JType: "parser",
			},
			ExecFn: generateCodeForRule,
			Args: parserArgs{
				rule:     rule,
				appendFn: appendFn,
			},
		}
	}
	pool.GenerateFrom(jobs)
	pool.Run(ctx)

	return code.String(), nil
}

func generateCodeForRule(ctx context.Context, args interface{}) (interface{}, error) {
	pArgs := args.(parserArgs)
	rule := pArgs.rule
	if rule.Expression == nil {
		return nil, nil // nothing to do
	}

	expr, err := models.ValidateExpression(rule.Expression)
	if err != nil {
		return nil, err
	}
	expression := parseExpression(ctx, expr)

	actions := ""
	for _, action := range rule.Actions {
		actions += fmt.Sprintf("\naction(\"%s\", \"%s\", \"%s\", \"%s\", \"%s\")\n", action.Action, action.Target.Kind, action.Target.Target, action.Target.Value, rule.ID)

	}
	codeSection := fmt.Sprintf(`if %s {%s}`, expression, actions)

	pArgs.appendFn(codeSection, fmt.Sprintf("%s@v%d: %s", rule.ID, rule.Version, rule.Description))

	return nil, nil
}

func parseExpression(ctx context.Context, expr *pb.Expression) string {
	switch expr.Expr.(type) {
	case *pb.Expression_And:
		and := expr.GetAnd()
		if and == nil || len(and.Expression) == 0 {
			return ""
		}
		return joinExpressions(deepDiveFn(ctx, and.Expression), andJoin)
	case *pb.Expression_Or:
		or := expr.GetOr()
		if or == nil || len(or.Expression) == 0 {
			return ""
		}
		return joinExpressions(deepDiveFn(ctx, or.Expression), orJoin)
	case *pb.Expression_Condition:
		condition := expr.GetCondition()
		if condition == nil {
			return ""
		}
		return conditionToScript(condition)
	default:
		return ""
	}
}

func deepDiveFn(ctx context.Context, e []*pb.Expression) (expressions []string) {
	for _, ex := range e {
		cex := parseExpression(ctx, ex)
		if cex != "" {
			expressions = append(expressions, cex)
		}
	}
	return
}

type joinType string

const (
	andJoin joinType = " && "
	orJoin  joinType = " || "
)

func joinExpressions(expressions []string, sep joinType) string {
	return "(" + strings.Join(expressions, string(sep)) + ")"
}
