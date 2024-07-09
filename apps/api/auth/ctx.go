package auth

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
)

func CtxString(c context.Context, key string) (s string) {
	if val := c.Value(key); val != nil {
		s, _ = val.(string)
	}
	return
}

func CtxJWTToken(c context.Context, key string) (token *jwt.Token) {
	if val := c.Value(key); val != nil {
		token, _ = val.(*jwt.Token)
	}
	return nil
}
