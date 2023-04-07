package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ApiKeyAuthMiddleware(logger *zap.Logger) gin.HandlerFunc {
	rootApiKey := getenv("ROOT_API_KEY", "")
	useSpeakeasy := getenv("SPEAKEASY_API_KEY", "") != ""

	if rootApiKey == "" && !useSpeakeasy {
		logger.Fatal("🔥 Neither ROOT_API_KEY nor SPEAKEASY_API_KEY are specified. You have to choose one.")
	}

	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "API Key not set"})
			c.Abort()
			return
		}
		if !useSpeakeasy {
			if token != rootApiKey {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid root API Key"})
				c.Abort()
				return
			}
		} else {
			apiKey, ok := strings.CutPrefix(token, "JWT ")
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid API Key"})
				c.Abort()
				return
			}

			logger.Debug("TODO: Verifying api key", zap.String("apiKey", apiKey))
			// TODO: grab the thing from speakeasy and check if the apiKey is a valid one
		}

		c.Next()
	}

}
