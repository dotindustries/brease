package auth

import (
	unkey "github.com/unkeyed/unkey/sdks/golang"
	"go.dot.industries/brease/env"
)

var unkeyClient = unkey.New(
	unkey.WithSecurity(env.Getenv("UNKEY_TOKEN", "")),
)
