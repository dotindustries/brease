package api

import (
	"context"
	"fmt"
	"go.dot.industries/brease/code"

	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/models"
)

type EvaluateRulesRequest struct {
	PathParams
	Object        interface{}   `json:"object" validate:"required"`
	OverrideRules []models.Rule `json:"overrideRules"`
	OverrideCode  string        `json:"overrideCode"`
}

type EvaluateRulesResponse struct {
	Results []models.EvaluationResult `json:"results"`
}

func (b *BreaseHandler) EvaluateRules(c *gin.Context, r *EvaluateRulesRequest) (*EvaluateRulesResponse, error) {
	orgID := c.GetString(auth.ContextOrgKey)
	ctx := c.Request.Context()

	codeBlock, err := b.findCode(ctx, r, orgID)
	if err != nil {
		return nil, err
	}

	compiledScript, err := b.findScript(ctx, codeBlock)
	if err != nil {
		return nil, fmt.Errorf("Failed to compile code: %v", err)
	}

	run, err := code.NewRun(ctx, b.logger, r.Object)
	if err != nil {
		return nil, fmt.Errorf("Failed to create run context: %v", err)
	}

	results, err := run.Execute(ctx, compiledScript)
	if err != nil {
		return nil, err
	}

	return &EvaluateRulesResponse{Results: results}, nil
}

// Override code takes precedence
func (b *BreaseHandler) findCode(ctx context.Context, r *EvaluateRulesRequest, orgID string) (string, error) {
	if r.OverrideCode != "" {
		return r.OverrideCode, nil
	}

	rules, err := b.findRules(ctx, r, orgID)
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
func (b *BreaseHandler) findRules(ctx context.Context, r *EvaluateRulesRequest, orgID string) ([]models.Rule, error) {
	if r.OverrideRules != nil {
		return r.OverrideRules, nil
	}

	rules, err := b.db.Rules(ctx, orgID, r.ContextID)
	if err != nil {
		return nil, errors.BadRequestf("Rules not found for context: %s", r.ContextID)
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
