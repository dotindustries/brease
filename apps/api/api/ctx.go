package api

import "context"

func CtxString(c context.Context, key string) (s string) {
	if val := c.Value(key); val != nil {
		s, _ = val.(string)
	}
	return
}
