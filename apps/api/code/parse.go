package code

import (
	expressionv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/expression/v1"
	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"context"
	"fmt"
	"strings"
	"sync"

	"go.dot.industries/brease/trace"
	"go.dot.industries/brease/worker"
)

type parserArgs struct {
	rule     *rulev1.VersionedRule
	appendFn func(section string, ID string)
}

func (a *Assembler) parseRules(ctx context.Context, rules []*rulev1.VersionedRule) (string, error) {
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
				ID:    worker.JobID(rule.Id),
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

	expression := parseExpression(ctx, rule.Expression)

	actions := ""
	for _, action := range rule.Actions {
		actions += fmt.Sprintf("\naction(\"%s\", \"%s\", \"%s\", \"%s\", \"%s\")\n", action.Kind, action.Target.Kind, action.Target.Id, action.Target.Value, rule.Id)

	}
	codeSection := fmt.Sprintf(`if %s {%s}`, expression, actions)

	pArgs.appendFn(codeSection, fmt.Sprintf("%s@v%d: %s", rule.Id, rule.Version, rule.Description))

	return nil, nil
}

func parseExpression(ctx context.Context, expr *expressionv1.Expression) string {
	switch expr.Expr.(type) {
	case *expressionv1.Expression_And:
		and := expr.GetAnd()
		if and == nil || len(and.Expression) == 0 {
			return ""
		}
		return joinExpressions(deepDiveFn(ctx, and.Expression), andJoin)
	case *expressionv1.Expression_Or:
		or := expr.GetOr()
		if or == nil || len(or.Expression) == 0 {
			return ""
		}
		return joinExpressions(deepDiveFn(ctx, or.Expression), orJoin)
	case *expressionv1.Expression_Condition:
		condition := expr.GetCondition()
		if condition == nil {
			return ""
		}
		return conditionToScript(condition)
	default:
		return ""
	}
}

func deepDiveFn(ctx context.Context, e []*expressionv1.Expression) (expressions []string) {
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
