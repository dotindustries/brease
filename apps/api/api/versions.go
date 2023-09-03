package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/models"
)

type RuleVersionsRequest struct {
	PathParams
	ID string `json:"-" validate:"required" path:"id"`
}

type RuleVersionsResponse struct {
	Versions []models.VersionedRule `json:"versions" validate:"required"`
}

func (b *BreaseHandler) GetRuleVersions(c *gin.Context, r *RuleVersionsRequest) (*RuleVersionsResponse, error) {
	orgID := c.GetString(auth.ContextOrgKey)
	rules, err := b.db.RuleVersions(c.Request.Context(), orgID, r.ContextID, r.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rule versions: %v", err)
	}
	return &RuleVersionsResponse{
		Versions: rules,
	}, nil
}
