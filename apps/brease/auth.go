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
			c.JSON(http.StatusBadRequest, gin.H{"message": "API Key not set"})
			return
		}
		token, ok := strings.CutPrefix(token, "Bearer ")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid API Key"})
			return
		}
		if !useSpeakeasy && token != rootApiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid API Key"})
			return
		}

		if useSpeakeasy {
			// TODO: grab the thing from speakeasy and check if the token is a valid one
		}

		c.Next()
	}

}
