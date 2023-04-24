package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/juju/errors"
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/env"
	"go.dot.industries/brease/models"
)

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

	if err = b.db.SaveAccessToken(c.Request.Context(), ownerID, tp); err != nil {
		return nil, fmt.Errorf("failed to save tokens to database: %v", err)
	}

	return tp, nil
}

func (b *BreaseHandler) RefreshTokenPair(c *gin.Context, r *RefreshTokenPairRequest) (*models.TokenPair, error) {
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

	oldTokenPair, err := b.db.GetAccessToken(c.Request.Context(), orgID)
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

	if err = b.db.SaveAccessToken(c.Request.Context(), orgID, tp); err != nil {
		return nil, fmt.Errorf("failed to save tokens to database: %v", err)
	}

	return tp, nil
}
