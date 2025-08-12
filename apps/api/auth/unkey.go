package auth

import (
	unkey "github.com/unkeyed/sdks/api/go/v2"
	"go.dot.industries/brease/env"
)

var unkeyClient = unkey.New(
	unkey.WithSecurity(env.Getenv("UNKEY_TOKEN", "")),
)
