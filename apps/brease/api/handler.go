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

type sharedURLParams struct {
	ContextID string `json:"contextID"`
}

func (b *BreaseHandler) AllRules(c *gin.Context) {
	contextID := c.Param("contextID")
	c.String(http.StatusOK, fmt.Sprintf("Rules for: %s", contextID))
}

func (b *BreaseHandler) ExecuteRules(c *gin.Context) {

}

func (b *BreaseHandler) AddRule(c *gin.Context) {

}

func (b *BreaseHandler) ReplaceRule(c *gin.Context) {

}

func (b *BreaseHandler) DeleteRule(c *gin.Context) {

}
