package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/juju/errors"
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/env"
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

type RefreshTokenPairRequest struct {
	RefreshToken string `json:"refreshToken"`
}

var jwtSecret = env.Getenv("JWT_SECRET", "")

func (b *BreaseHandler) generateTokenPair(ownerID string) (*models.TokenPair, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = ownerID
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()

	t, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate accessToken: %v", err)
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["sub"] = ownerID
	rtClaims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	rt, err := refreshToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refreshToken: %v", err)
	}

	tp := &models.TokenPair{
		AccessToken:  t,
		RefreshToken: rt,
	}

	return tp, nil
}

func (b *BreaseHandler) GenerateTokenPair(c *gin.Context) (*models.TokenPair, error) {
	ownerID := c.GetString(auth.ContextOrgKey)

	tp, err := b.generateTokenPair(ownerID)
	if err != nil {
		return nil, err
	}

	if err = b.db.SaveAccessToken(ownerID, tp); err != nil {
		return nil, fmt.Errorf("failed to save tokens to database: %v", err)
	}

	return tp, nil
}

func (b *BreaseHandler) RefreshTokenPair(_ *gin.Context, r *RefreshTokenPairRequest) (*models.TokenPair, error) {
	token, err := jwt.Parse(r.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, errors.BadRequestf("invalid refreshToken: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || token.Valid {
		return nil, errors.BadRequestf("invalid refreshToken")
	}

	orgID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.BadRequestf("invalid refreshToken sub")
	}

	oldTokenPair, err := b.db.GetAccessToken(orgID)
	if err != nil {
		return nil, errors.BadRequestf("refreshToken not found")
	}

	if oldTokenPair.RefreshToken != r.RefreshToken {
		return nil, errors.BadRequestf("unknown refreshToken")
	}

	tp, err := b.generateTokenPair(orgID)
	if err != nil {
		return nil, err
	}

	if err = b.db.SaveAccessToken(orgID, tp); err != nil {
		return nil, fmt.Errorf("failed to save tokens to database: %v", err)
	}

	return tp, nil
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

func (b *BreaseHandler) EvaluateRules(c *gin.Context, r *EvaluateRulesRequest) (*EvaluateRulesResponse, error) {
	orgID := c.GetString(auth.ContextOrgKey)
	return nil, errors.NotImplementedf("EvaluateRules not yet ready: %s", orgID)
}

func (b *BreaseHandler) AddRule(c *gin.Context, r *AddRuleRequest) (*AddRuleResponse, error) {
	orgID := c.GetString(auth.ContextOrgKey)
	rule := r.Rule

	// duplicate check
	if exists, err := b.db.Exists(orgID, r.ContextID, rule.ID); exists && err == nil {
		b.logger.Warn("Rule already exists. Rejecting to add", zap.String("contextID", r.ContextID), zap.String("ruleID", rule.ID))
		return nil, errors.AlreadyExistsf("rule with ID '%s'", rule.ID)
	} else if err != nil {
		return nil, err
	}
	err := b.validateExpression(rule.Expression)
	if err != nil {
		return nil, errors.NewBadRequest(err, "invalid expression")
	}
	err = b.db.AddRule(orgID, r.ContextID, rule)
	if err != nil {
		return nil, fmt.Errorf("failed to add rule: %v", err)
	}
	return &AddRuleResponse{
		Rule: rule,
	}, nil
}

func (b *BreaseHandler) ReplaceRule(c *gin.Context, r *ReplaceRuleRequest) (*ReplaceRuleResponse, error) {
	orgID := c.GetString(auth.ContextOrgKey)

	err := b.db.ReplaceRule(orgID, r.ContextID, r.Rule)
	if err != nil {
		return nil, fmt.Errorf("failed to update rule: %v", err)
	}
	return &ReplaceRuleResponse{
		Rule: r.Rule,
	}, nil
}

func (b *BreaseHandler) RemoveRule(c *gin.Context, r *RemoveRuleRequest) error {
	orgID := c.GetString(auth.ContextOrgKey)

	_ = b.db.RemoveRule(orgID, r.ContextID, r.ID)
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
