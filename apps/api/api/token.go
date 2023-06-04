package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/juju/errors"
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/models"
)

type RefreshTokenPairRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

func (b *BreaseHandler) generateTokenPair(ownerID string, userID string) (tp models.TokenPair, err error) {
	t, err := b.token.Sign(ownerID, userID, 0)
	if err != nil {
		return tp, fmt.Errorf("failed to generate accessToken: %w", err)
	}

	rt, err := b.token.Sign(ownerID, userID, time.Hour*24)
	if err != nil {
		return tp, fmt.Errorf("failed to generate refreshToken: %w", err)
	}

	tp.AccessToken = t
	tp.RefreshToken = rt

	return tp, nil
}

func (b *BreaseHandler) GenerateTokenPair(c *gin.Context) (*models.TokenPair, error) {
	ownerID := c.GetString(auth.ContextOrgKey)
	userID := c.GetString(auth.ContextUserIDKey)

	tp, err := b.generateTokenPair(ownerID, userID)
	if err != nil {
		return nil, err
	}

	if err = b.db.SaveAccessToken(c.Request.Context(), ownerID, tp); err != nil {
		return nil, fmt.Errorf("failed to save tokens to database: %w", err)
	}

	return &tp, nil
}

func (b *BreaseHandler) RefreshTokenPair(c *gin.Context, r *RefreshTokenPairRequest) (*models.TokenPair, error) {
	token, err := b.token.Parse(r.RefreshToken)
	if err != nil {
		return nil, errors.BadRequestf("invalid refreshToken: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.BadRequestf("invalid refreshToken")
	}

	orgID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.BadRequestf("invalid refreshToken sub")
	}

	oldTokenPairs, err := b.db.GetAccessTokens(c.Request.Context(), orgID)
	if err != nil {
		return nil, errors.BadRequestf("refreshToken not found")
	}

	unknown := true
	for _, tp := range oldTokenPairs {
		if tp.RefreshToken == r.RefreshToken {
			unknown = false
			break
		}
	}
	if unknown {
		return nil, errors.BadRequestf("unknown refreshToken")
	}

	tp, err := b.generateTokenPair(orgID, "")
	if err != nil {
		return nil, err
	}

	if err = b.db.SaveAccessToken(c.Request.Context(), orgID, tp); err != nil {
		return nil, fmt.Errorf("failed to save tokens to database: %w", err)
	}

	return &tp, nil
}
