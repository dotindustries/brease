package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	errors2 "github.com/juju/errors"
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
			_ = c.AbortWithError(http.StatusUnauthorized, errors2.Unauthorizedf("API key not set"))
			return
		}

		// allow for root api key to bypass jwt check
		if rootApiKey != "" && authHeader == rootApiKey {
			c.Next()
			return
		}

		if !useSpeakeasy {
			if authHeader != rootApiKey {
				_ = c.AbortWithError(http.StatusUnauthorized, errors2.Unauthorizedf("Invalid root API key"))
				return
			}
		} else {
			apiKey, ok := strings.CutPrefix(authHeader, "JWT ")
			if !ok {
				_ = c.AbortWithError(http.StatusUnauthorized, errors2.Unauthorizedf("Invalid API key"))
				return
			}

			token, _, err := new(jwt.Parser).ParseUnverified(apiKey, jwt.MapClaims{})
			if err != nil {
				_ = c.AbortWithError(http.StatusUnauthorized, errors2.NewUnauthorized(err, "Invalid JWT"))
				return
			}

			kid, ok := token.Header["kid"].(string)
			if !ok {
				_ = c.AbortWithError(http.StatusUnauthorized, errors2.Unauthorizedf("Invalid JWT: kid not present"))
				return
			}

			// don't use the request's context because it's short life will prevent the underlying jwks from refreshing
			key, err := jwksClient.GetKey(context.Background(), kid, "kid")
			if err != nil {
				_ = c.AbortWithError(http.StatusUnauthorized, errors2.NewUnauthorized(err, "Invalid JWT"))
				return
			}

			// verify the token
			_, err = jwt.Parse(apiKey, func(token *jwt.Token) (interface{}, error) {
				return key.Key, nil
			})
			if err != nil {
				logger.Error("Failed to verify the JWT.\nError: %s", zap.Error(err))
				_ = c.AbortWithError(http.StatusUnauthorized, errors2.NewUnauthorized(err, "Invalid API key"))
				return
			}
		}

		c.Next()
	}
}
