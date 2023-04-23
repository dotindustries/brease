package api

import (
	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/models"
)

type EvaluateRulesRequest struct {
	PathParams
	Object        map[string]interface{} `json:"object" validate:"required"`
	OverrideRules []models.Rule          `json:"overrideRules"`
	OverrideCode  string                 `json:"overrideCode"`
}

type EvaluateRulesResponse struct {
	Results []models.EvaluationResult `json:"results"`
}

func (b *BreaseHandler) EvaluateRules(c *gin.Context, r *EvaluateRulesRequest) (*EvaluateRulesResponse, error) {
	orgID := c.GetString(auth.ContextOrgKey)
	return nil, errors.NotImplementedf("EvaluateRules not yet ready: %s", orgID)
}
