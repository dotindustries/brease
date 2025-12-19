package auth

import (
	unkey "github.com/unkeyed/sdks/api/go/v2"
	"go.dot.industries/brease/env"
)

var unkeyClient *unkey.Unkey

func Unkey() *unkey.Unkey {
	if unkeyClient == nil {
		unkeyClient = unkey.New(
			unkey.WithSecurity(env.Getenv("UNKEY_TOKEN", "")),
		)
	}
	return unkeyClient
}
