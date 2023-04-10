package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/juju/errors"
	"go.dot.industries/brease/models"
	"go.dot.industries/brease/pb"
	"go.dot.industries/brease/storage"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
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

func (b *BreaseHandler) AllRules(_ *gin.Context, r *AllRulesRequest) (*AllRulesResponse, error) {
	rules, err := b.db.Rules(r.ContextID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rules: %v", err)
	}
	return &AllRulesResponse{
		Rules: rules,
		Code:  "",
	}, nil
}

func (b *BreaseHandler) EvaluateRules(_ *gin.Context, r *EvaluateRulesRequest) (*EvaluateRulesResponse, error) {
	return nil, errors.NotImplementedf("EvaluateRules not yet ready")
}

func (b *BreaseHandler) AddRule(c *gin.Context, r *AddRuleRequest) (*AddRuleResponse, error) {
	rule := r.Rule

	// duplicate check
	if exists, err := b.db.Exists(r.ContextID, rule.ID); exists && err == nil {
		b.logger.Warn("Rule already exists. Rejecting to add", zap.String("contextID", r.ContextID), zap.String("ruleID", rule.ID))
		return nil, errors.AlreadyExistsf("rule with ID '%s'", rule.ID)
	} else if err != nil {
		return nil, err
	}
	err := b.validateExpression(rule.Expression)
	if err != nil {
		return nil, errors.NewBadRequest(err, "invalid expression")
	}
	err = b.db.AddRule(r.ContextID, rule)
	if err != nil {
		return nil, fmt.Errorf("failed to add rule: %v", err)
	}
	return &AddRuleResponse{
		Rule: rule,
	}, nil
}

func (b *BreaseHandler) ReplaceRule(_ *gin.Context, r *ReplaceRuleRequest) (*ReplaceRuleResponse, error) {
	err := b.db.ReplaceRule(r.ContextID, r.Rule)
	if err != nil {
		return nil, fmt.Errorf("failed to update rule: %v", err)
	}
	return &ReplaceRuleResponse{
		Rule: r.Rule,
	}, nil
}

func (b *BreaseHandler) RemoveRule(_ *gin.Context, r *RemoveRuleRequest) error {
	_ = b.db.RemoveRule(r.ContextID, r.ID)
	// we don't expose whether we succeeded
	return nil
}

func (b *BreaseHandler) validateExpression(expression map[string]interface{}) error {
	exprBytes, err := json.Marshal(expression)
	if err != nil {
		b.logger.Error("expression is not base64 encoded", zap.Error(err), zap.Any("expression", expression))
		return fmt.Errorf("expression is not base64 encoded: %v", err)
	}
	expr := &pb.Expression{}
	if unmarshalErr := protojson.Unmarshal(exprBytes, expr); err != nil {
		b.logger.Error("expression cannot be read", zap.Error(unmarshalErr), zap.Any("expression", expression))
		return fmt.Errorf("expression cannot be read: %v", unmarshalErr)
	}
	b.logger.Debug("Valid expression", zap.Any("expression", expr))
	return nil
}
