package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/models"
)

type AllRulesRequest struct {
	PathParams
	CompileCode bool `json:"compileCode"`
}

type AllRulesResponse struct {
	Rules []models.Rule `json:"rules"`
	Code  string        `json:"code"`
}

func (b *BreaseHandler) AllRules(c *gin.Context, r *AllRulesRequest) (*AllRulesResponse, error) {
	orgID := c.GetString(auth.ContextOrgKey)
	rules, err := b.db.Rules(orgID, r.ContextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rules: %v", err)
	}
	return &AllRulesResponse{
		Rules: rules,
		Code:  "",
	}, nil
}
