package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.dot.industries/brease/env"
	"go.uber.org/zap"
)

var (
	// ErrKID indicates that the JWT had an invalid kid.
	ErrKID = errors.New("the JWT has an invalid kid")
)

func ApiKeyAuthMiddleware(logger *zap.Logger) gin.HandlerFunc {
	rootApiKey := env.Getenv("ROOT_API_KEY", "")
	useSpeakeasy := env.Getenv("SPEAKEASY_API_KEY", "") != ""

	if rootApiKey == "" && !useSpeakeasy {
		logger.Fatal("🔥 Neither ROOT_API_KEY nor SPEAKEASY_API_KEY are specified. You have to choose one.")
	}

	if useSpeakeasy && jwksClient == nil {
		logger.Fatal("🔥 JWKS client is not configured. Make sure SPEAKEASY_WORKSPACE_ID is set.")
	}

	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "API Key not set"})
			c.Abort()
			return
		}
		if !useSpeakeasy {
			if authHeader != rootApiKey {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid root API Key"})
				c.Abort()
				return
			}
		} else {
			apiKey, ok := strings.CutPrefix(authHeader, "JWT ")
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid API Key"})
				c.Abort()
				return
			}

			token, _, err := new(jwt.Parser).ParseUnverified(apiKey, jwt.MapClaims{})
			if err != nil {
				_ = c.AbortWithError(http.StatusUnauthorized, err)
				return
			}

			kid, ok := token.Header["kid"].(string)
			if !ok {
				_ = c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid: kid not present in API key"))
				return
			}

			// don't use the request's context because it's short life will prevent the underlying jwks from refreshing
			key, err := jwksClient.GetKey(context.Background(), kid, "kid")
			if err != nil {
				_ = c.AbortWithError(http.StatusUnauthorized, err)
				return
			}

			// verify the token
			_, err = jwt.Parse(apiKey, func(token *jwt.Token) (interface{}, error) {
				return key.Key, nil
			})
			if err != nil {
				logger.Error("Failed to verify the JWT.\nError: %s", zap.Error(err))
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid API key"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
