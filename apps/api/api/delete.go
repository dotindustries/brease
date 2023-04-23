package api

import (
	"github.com/gin-gonic/gin"
	"go.dot.industries/brease/auth"
)

type RemoveRuleRequest struct {
	PathParams
	ID string `json:"-" validate:"required" path:"id"`
}

func (b *BreaseHandler) RemoveRule(c *gin.Context, r *RemoveRuleRequest) error {
	orgID := c.GetString(auth.ContextOrgKey)

	_ = b.db.RemoveRule(orgID, r.ContextID, r.ID)
	// we don't expose whether we succeeded
	return nil
}
