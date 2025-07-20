package api

import (
	contextv1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/context/v1"
	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/kaptinlin/jsonschema"
	"go.dot.industries/brease/cache"
	"go.dot.industries/brease/code"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/juju/errors"
	"go.dot.industries/brease/auth"
)

func (b *BreaseHandler) Evaluate(ctx context.Context, c *connect.Request[contextv1.EvaluateRequest]) (*connect.Response[contextv1.EvaluateResponse], error) {
	orgID, _, _, cErr := permissionCheck(ctx, auth.PermissionEvaluate, auth.PermissionReadRule)
	if cErr != nil {
		return nil, cErr
	}

	schema, err := b.db.GetObjectSchema(ctx, orgID, c.Msg.ContextId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to fetch object schema: %v", err))
	}
	if schema != "" {
		compiledSchema, schemaErr := b.compiledObjectSchema(ctx, schema)
		if schemaErr != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(schemaErr, fmt.Errorf("invalid json schema")))
		}
		res := compiledSchema.ValidateStruct(c.Msg.Object)
		if !res.Valid {
			es := res.ToList()
			errStr := ""
			for _, e := range es.Errors {
				if len(errStr) > 0 {
					errStr += ", "
				}
				errStr += e
			}
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(schemaErr, fmt.Errorf("invalid object shape: %s", errStr)))
		}
	}

	codeBlock, err := b.findCode(ctx, c.Msg, orgID)
	if err != nil {
		return nil, err
	}

	results, err := b.run(ctx, codeBlock, c.Msg.Object)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&contextv1.EvaluateResponse{
		Results: results,
	}), nil
}

func (b *BreaseHandler) compiledObjectSchema(ctx context.Context, schema string) (*jsonschema.Schema, error) {
	compiled := b.cache.Get(ctx, cache.SimpleHash(schema))
	if compiled != nil && compiled != "" {
		return compiled.(*jsonschema.Schema), nil
	}
	compiledSchema, err := b.jsonSchemaCompiler.Compile([]byte(schema))
	if err != nil {
		return nil, err
	}
	// if we can't set it to cache, at worst it's gonna cause a delay on the next call
	_ = b.cache.Set(ctx, cache.SimpleHash(schema), compiledSchema)
	return compiledSchema, err
}

func (b *BreaseHandler) run(ctx context.Context, codeBlock string, object *structpb.Struct) ([]*rulev1.EvaluationResult, error) {
	compiledScript, err := b.findScript(ctx, codeBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to compile code: %v", err)
	}

	run, err := code.NewRun(ctx, b.logger, object)
	if err != nil {
		return nil, fmt.Errorf("failed to create run context: %v", err)
	}

	return run.Execute(ctx, compiledScript)
}

// Override code takes precedence
func (b *BreaseHandler) findCode(ctx context.Context, r *contextv1.EvaluateRequest, orgID string) (string, error) {
	if r.OverrideCode != "" {
		return r.OverrideCode, nil
	}

	rules, err := b.findRules(ctx, orgID, r.ContextId, r.OverrideRules)
	if err != nil {
		return "", err
	}

	c, err := b.assembler.BuildCode(ctx, rules)
	if err != nil {
		return "", errors.Errorf("Failed to assemble code: %v", err)
	}

	return c, nil
}

// Override rules take precedence
func (b *BreaseHandler) findRules(ctx context.Context, orgID string, contextID string, overrideRules []*rulev1.Rule) ([]*rulev1.VersionedRule, error) {
	if overrideRules != nil {
		vRs := make([]*rulev1.VersionedRule, len(overrideRules))
		for i, rule := range overrideRules {
			vRs[i] = &rulev1.VersionedRule{
				Id:          rule.Id,
				Version:     0, // override rules don't have versioning
				Description: rule.Description,
				Actions:     rule.Actions,
				Expression:  rule.Expression,
			}
		}
		return vRs, nil
	}

	rules, err := b.db.Rules(ctx, orgID, contextID, 0, "")
	if err != nil {
		return nil, fmt.Errorf("rules not found for context: %s", contextID)
	}

	return rules, nil
}

func (b *BreaseHandler) findScript(ctx context.Context, codeBlock string) (*code.Script, error) {
	script, err := b.compiler.CompileCode(ctx, codeBlock)
	if err != nil {
		return nil, errors.Errorf("Failed to compile code block: %v", err)
	}
	return script, nil
}
