package auth

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

func CtxStringArr(c context.Context, key string) (s []string) {
	if val := c.Value(key); val != nil {
		s, _ = val.([]string)
	}
	return
}

func CtxString(c context.Context, key string) (s string) {
	if val := c.Value(key); val != nil {
		s, _ = val.(string)
	}
	return
}

func CtxInt(c context.Context, key string) (i int) {
	if val := c.Value(key); val != nil {
		i, _ = val.(int)
	}
	return
}

func CtxJWTToken(c context.Context, key string) (token *jwt.Token) {
	if val := c.Value(key); val != nil {
		token, _ = val.(*jwt.Token)
	}
	return nil
}

func UserIDFromContext(c context.Context) (userID string) {
	return CtxString(c, ContextUserIDKey)
}

func OrgIDFromContext(c context.Context) (orgID string) {
	return CtxString(c, ContextOrgKey)
}

func PermissionsFromContext(c context.Context) (permissions []string) {
	if val := c.Value(ContextPermissionsKey); val != nil {
		permissions, _ = val.([]string)
	}
	return
}

func HasPermission(c context.Context, permission string) bool {
	permissions := PermissionsFromContext(c)
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	// no permission
	return false
}

func AuthFromContext(c context.Context) (authed bool, userID string, orgID string, permissions []string) {
	userID = UserIDFromContext(c)
	orgID = OrgIDFromContext(c)
	permissions = PermissionsFromContext(c)
	authed = userID != "" && orgID != "" && len(permissions) > 0
	return
}
