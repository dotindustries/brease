package api

import (
	"fmt"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/models"
)

type AllRulesRequest struct {
	PathParams
	CompileCode bool `query:"compileCode"`
}

type AllRulesResponse struct {
	Rules []models.VersionedRule `json:"rules" validate:"required"`
	Code  string                 `json:"code"`
}

func (b *BreaseHandler) AllRules(c *gin.Context, r *AllRulesRequest) (*AllRulesResponse, error) {
	orgID := c.GetString(auth.ContextOrgKey)
	rules, err := b.db.Rules(c.Request.Context(), orgID, r.ContextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rules: %v", err)
	}
	code := ""
	if r.CompileCode {
		code, err = b.assembler.BuildCode(c.Request.Context(), rules)
		if err != nil {
			b.logger.Warn("Failed to assemble code", zap.Error(err))
		} else {
			b.logger.Debug("Assembled code", zap.String("code", code))
		}
	}
	return &AllRulesResponse{
		Rules: rules,
		Code:  code,
	}, nil
}
