package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"go.dot.industries/brease/env"
	"go.uber.org/zap"
	"time"
)

type Token struct {
	logger   *zap.Logger
	secret   string
	duration time.Duration
}

func NewToken(logger *zap.Logger) Token {
	jwtSecret := env.Getenv("JWT_SECRET", "")
	if jwtSecret == "" {
		logger.Fatal("🔥 JWT_SECRET is not set")
	}
	jwtDurationEnv := env.Getenv("JWT_EXPIRATION", "5m")
	jwtDuration, err := time.ParseDuration(jwtDurationEnv)
	if err != nil {
		logger.Fatal("🔥 configured JWT_EXPIRATION is invalid", zap.Error(err))
	}

	return Token{
		logger:   logger,
		secret:   jwtSecret,
		duration: jwtDuration,
	}
}

func (t Token) Sign(sub string, userID string, exp time.Duration) (string, error) {
	if exp == 0 {
		exp = t.duration // apply default duration
	}
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = sub
	claims["exp"] = time.Now().Add(exp).Unix()
	claims[ContextUserIDKey] = userID

	return token.SignedString([]byte(t.secret))
}

func (t Token) Parse(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.secret), nil
	})
}
