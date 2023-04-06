package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type headers struct {
	RequestID      string `header:"X-Request-Id,required"`
	Authentication string `header:"Authentication,required"`
}

type BreaseHandler struct {
}

type PathParams struct {
	ContextID string `path:"contextID"`
}

func (b *BreaseHandler) AllRules(c *gin.Context) error {
	contextID := c.Param("contextID")
	c.String(http.StatusOK, fmt.Sprintf("Rules for: %s", contextID))
	return nil
}

func (b *BreaseHandler) ExecuteRules(c *gin.Context) error {
	return nil
}

func (b *BreaseHandler) AddRule(c *gin.Context) error {
	return nil
}

func (b *BreaseHandler) ReplaceRule(c *gin.Context) error {
	return nil
}

func (b *BreaseHandler) DeleteRule(c *gin.Context) error {
	return nil
}
