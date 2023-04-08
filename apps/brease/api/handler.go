package api

import (
	"encoding/base64"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.dot.industries/brease/models"
	"go.dot.industries/brease/pb"
	"go.dot.industries/brease/storage"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type BreaseHandler struct {
	db     storage.Database
	logger *zap.Logger
}

func NewHandler(db storage.Database, logger *zap.Logger) *BreaseHandler {
	if db == nil {
		panic("database cannot be nil")
	}
	return &BreaseHandler{
		db:     db,
		logger: logger,
	}
}

type PathParams struct {
	ContextID string `path:"contextID"`
}

type AllRulesRequest struct {
	PathParams
	CompileCode bool `json:"compileCode"`
}

type AllRulesResponse struct {
	Rules []models.Rule `json:"rules"`
	Code  string        `json:"code"`
}

type EvaluateRulesRequest struct {
	PathParams
	Object        map[string]interface{} `json:"object" validate:"required"`
	OverrideRules []models.Rule          `json:"overrideRules"`
	OverrideCode  string                 `json:"overrideCode"`
}

type EvaluateRulesResponse struct {
	Results []models.EvaluationResult `json:"results"`
}

type AddRuleRequest struct {
	PathParams
	Rule models.Rule `json:"rule"`
}

type AddRuleResponse struct {
	Rule models.Rule `json:"rule"`
}

type RemoveRuleRequest struct {
	PathParams
	ID string `json:"-" validate:"required" path:"id"`
}

type ReplaceRuleRequest struct {
	PathParams
	ID   string      `json:"-" validate:"required" path:"id"`
	Rule models.Rule `json:"rule" validate:"required"`
}

type ReplaceRuleResponse struct {
	Rule models.Rule `json:"rule"`
}

func (b *BreaseHandler) AllRules(_ *gin.Context, r *AllRulesRequest) (AllRulesResponse, error) {
	rules, err := b.db.Rules(r.ContextID)
	if err != nil {
		return AllRulesResponse{}, fmt.Errorf("failed to fetch rules: %v", err)
	}
	return AllRulesResponse{
		Rules: rules,
		Code:  "",
	}, nil
}

func (b *BreaseHandler) EvaluateRules(_ *gin.Context, r *EvaluateRulesRequest) (EvaluateRulesResponse, error) {
	return EvaluateRulesResponse{}, fmt.Errorf("not yet implemented")
}

func (b *BreaseHandler) AddRule(_ *gin.Context, r *AddRuleRequest) (AddRuleResponse, error) {
	rule := r.Rule

	exprBytes, err := base64.StdEncoding.DecodeString(rule.Expression)
	if err != nil {
		b.logger.Error("invalid: expression is not base64 encoded", zap.Error(err), zap.Any("expression", rule.Expression))
		return AddRuleResponse{}, fmt.Errorf("invalid: expression is not base64 encoded: %v", err)
	}
	expr := &pb.Expression{}
	if unmarshalErr := proto.Unmarshal(exprBytes, expr); err != nil {
		b.logger.Error("invalid: expression cannot be read", zap.Error(unmarshalErr), zap.Any("expression", rule.Expression))
		return AddRuleResponse{}, fmt.Errorf("invalid: expression cannot be read: %v", unmarshalErr)
	}
	b.logger.Debug("Valid expression", zap.Any("expression", expr))

	err = b.db.AddRule(r.ContextID, rule)
	if err != nil {
		return AddRuleResponse{}, fmt.Errorf("failed to add rule: %v", err)
	}
	return AddRuleResponse{
		Rule: rule,
	}, nil
}

func (b *BreaseHandler) ReplaceRule(_ *gin.Context, r *ReplaceRuleRequest) (ReplaceRuleResponse, error) {
	err := b.db.ReplaceRule(r.ContextID, r.Rule)
	if err != nil {
		return ReplaceRuleResponse{}, fmt.Errorf("failed to update rule: %v", err)
	}
	return ReplaceRuleResponse{
		Rule: r.Rule,
	}, nil
}

func (b *BreaseHandler) RemoveRule(_ *gin.Context, r *RemoveRuleRequest) error {
	_ = b.db.RemoveRule(r.ContextID, r.ID)
	// we don't expose whether we succeeded
	return nil
}
