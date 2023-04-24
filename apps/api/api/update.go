package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/models"
)

type ReplaceRuleRequest struct {
	PathParams
	ID   string      `json:"-" validate:"required" path:"id"`
	Rule models.Rule `json:"rule" validate:"required"`
}

type ReplaceRuleResponse struct {
	Rule models.Rule `json:"rule"`
}

func (b *BreaseHandler) ReplaceRule(c *gin.Context, r *ReplaceRuleRequest) (*ReplaceRuleResponse, error) {
	orgID := c.GetString(auth.ContextOrgKey)

	err := b.db.ReplaceRule(c.Request.Context(), orgID, r.ContextID, r.Rule)
	if err != nil {
		return nil, fmt.Errorf("failed to update rule: %v", err)
	}
	return &ReplaceRuleResponse{
		Rule: r.Rule,
	}, nil
}
