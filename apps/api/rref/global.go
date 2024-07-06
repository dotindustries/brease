package rref

import (
	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"context"

	"go.dot.industries/brease/cache"
	"go.dot.industries/brease/cache/memory"
	"go.dot.industries/brease/trace"
)

var localCache = memory.New()

// IsConfigured returns true if rref service has been configured for remote data access
func IsConfigured() bool {
	return false
}

func LookupReferenceValue(ctx context.Context, ref *rulev1.ConditionBaseRef) []byte {
	ctx, span := trace.Tracer.Start(ctx, "reference-query")
	defer span.End()

	key := cache.SimpleHash(ref)
	value := localCache.Get(ctx, key)
	if value != nil {
		// fetch in the background for next caller
		go func() {
			_, _ = fetchReferenceValue(ctx, ref)
		}()
		return value.([]byte)
	}

	ch := make(chan []byte, 1)
	go func() {
		newValue, err := fetchReferenceValue(ctx, ref)
		if err == nil {
			localCache.Set(ctx, key, newValue)
			ch <- newValue
		} else {
			ch <- nil
		}
	}()

	select {
	case newValue := <-ch:
		if newValue != nil {
			return newValue
		}
	case <-ctx.Done():
		return nil
	}

	return nil
}

func fetchReferenceValue(ctx context.Context, ref *rulev1.ConditionBaseRef) ([]byte, error) {
	// TODO: actually call the remote service to fetch the value
	return nil, nil
}
