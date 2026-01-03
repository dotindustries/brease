package auth

import (
	"fmt"

	unkey "github.com/unkeyed/sdks/api/go/v2"
	"go.dot.industries/brease/env"
)

var unkeyClient *unkey.Unkey

func Unkey() *unkey.Unkey {
	if unkeyClient == nil {
		unkeyToken := env.Getenv("UNKEY_TOKEN", "")
		if unkeyToken == "" {
			panic(fmt.Errorf("UNKEY_TOKEN is not set"))
		}
		unkeyClient = unkey.New(
			unkey.WithSecurity(unkeyToken),
		)
	}
	return unkeyClient
}
