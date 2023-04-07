package api

import (
	"fmt"

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

type AllRulesRequest struct {
	PathParams
	CompileCode bool `json:"compileCode"`
}

type AllRulesResponse struct {
	Rules []Rule `json:"rules"`
	Code  string `json:"code"`
}

type EvaluateRulesRequest struct {
	PathParams
	Object        map[string]interface{} `json:"object" validate:"required"`
	OverrideRules []Rule                 `json:"overrideRules"`
	OverrideCode  string                 `json:"overrideCode"`
}

type EvaluateRulesResponse struct {
	Results []EvaluationResult `json:"results"`
}

type AddRuleRequest struct {
	PathParams
	Rule Rule `json:"rule"`
}

type AddRuleResponse struct {
	Rule Rule `json:"rule"`
}

type RemoveRuleRequest struct {
	PathParams
	ID string `json:"-" validate:"required" path:"id"`
}

type ReplaceRuleRequest struct {
	PathParams
	ID   string `json:"-" validate:"required" path:"id"`
	Rule Rule   `json:"rule" validate:"required"`
}

type ReplaceRuleResponse struct {
	Rule Rule `json:"rule"`
}

func (b *BreaseHandler) AllRules(c *gin.Context, r *AllRulesRequest) (AllRulesResponse, error) {
	return AllRulesResponse{}, nil
}

func (b *BreaseHandler) ExecuteRules(c *gin.Context, r *EvaluateRulesRequest) (EvaluateRulesResponse, error) {
	return EvaluateRulesResponse{}, fmt.Errorf("not yet implemented")
}

func (b *BreaseHandler) AddRule(c *gin.Context, r *AddRuleRequest) (AddRuleResponse, error) {
	return AddRuleResponse{}, fmt.Errorf("not yet implemented")
}

func (b *BreaseHandler) ReplaceRule(c *gin.Context, r *ReplaceRuleRequest) (ReplaceRuleResponse, error) {
	return ReplaceRuleResponse{}, fmt.Errorf("not yet implemented")
}

func (b *BreaseHandler) RemoveRule(c *gin.Context, r *RemoveRuleRequest) error {
	return nil
}
