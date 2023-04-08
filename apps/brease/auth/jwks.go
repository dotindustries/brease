package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/s12v/go-jwks"
	"go.dot.industries/brease/env"
	"go.dot.industries/brease/log"
	"go.uber.org/zap"
)

var jwksClient jwks.JWKSClient

func InitJWKS() {
	workspaceID := env.Getenv("SPEAKEASY_WORKSPACE_ID", "")
	if workspaceID == "" {
		return
	}
	logger, _, _ := log.Logger()

	url := fmt.Sprintf("https://app.speakeasyapi.dev/v1/auth/oauth/%s/.well-known/jwks.json", workspaceID)
	jwksSource := jwks.NewWebSource(url, http.DefaultClient)
	jwksClient = jwks.NewDefaultClient(
		jwksSource,
		time.Minute*5, // Refresh keys every 1 hour
		12*time.Hour,  // Expire keys after 12 hours
	)
	logger.Info("Configured Speakeasy JWKS for JWT verification.", zap.String("source", url))
}
