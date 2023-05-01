package cache

import (
	"context"
	"fmt"
	"time"

	"go.dot.industries/brease/env"
)

const defaultTTL = "15m"

type Cache interface {
	Get(ctx context.Context, key string) any
	Set(ctx context.Context, key string, value any) bool
}

func Ttl() time.Duration {
	ttlEnv := env.Getenv("BREASE_ASSEMBLY_CACHE_TTL", defaultTTL)
	duration, err := time.ParseDuration(ttlEnv)
	if err != nil {
		panic(fmt.Errorf("invalid cache ttl duration: %s", ttlEnv))
	}
	return duration
}
