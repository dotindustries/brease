package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/models"
	"go.uber.org/zap"
)

type AddRuleRequest struct {
	PathParams
	Rule models.Rule `json:"rule"`
}

type AddRuleResponse struct {
	Rule models.Rule `json:"rule"`
}

func (b *BreaseHandler) AddRule(c *gin.Context, r *AddRuleRequest) (*AddRuleResponse, error) {
	orgID := c.GetString(auth.ContextOrgKey)
	rule := r.Rule

	// duplicate check
	if exists, err := b.db.Exists(c.Request.Context(), orgID, r.ContextID, rule.ID); exists && err == nil {
		b.logger.Warn("Rule already exists. Rejecting to add", zap.String("contextID", r.ContextID), zap.String("ruleID", rule.ID))
		return nil, errors.AlreadyExistsf("rule with ID '%s'", rule.ID)
	} else if err != nil {
		return nil, err
	}
	err := b.validateExpression(rule.Expression)
	if err != nil {
		return nil, errors.NewBadRequest(err, "invalid expression")
	}
	err = b.db.AddRule(c.Request.Context(), orgID, r.ContextID, rule)
	if err != nil {
		return nil, fmt.Errorf("failed to add rule: %v", err)
	}
	return &AddRuleResponse{
		Rule: rule,
	}, nil
}
