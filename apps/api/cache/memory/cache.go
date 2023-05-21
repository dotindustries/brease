package memory

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"
	"go.dot.industries/brease/cache"
	"go.dot.industries/brease/trace"
)

type cacheContainer struct {
	ch  *ristretto.Cache
	ttl time.Duration
}

func New() cache.Cache {
	ch, err := ristretto.NewCache(&ristretto.Config{
		// 10M
		NumCounters: 1e7,
		// 1GB
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}

	return &cacheContainer{
		ch:  ch,
		ttl: cache.Ttl(),
	}
}

func (c *cacheContainer) Get(ctx context.Context, key string) any {
	_, span := trace.Tracer.Start(ctx, "cache")
	defer span.End()

	if script, ok := c.ch.Get(key); ok {
		return script
	}
	return ""
}

func (c *cacheContainer) Set(ctx context.Context, key string, value any) bool {
	_, span := trace.Tracer.Start(ctx, "cache")
	defer span.End()

	return c.ch.SetWithTTL(key, value, 0, c.ttl)
}
