package auditlogstore

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"
	"go.dot.industries/brease/auditlog"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
)

type Options struct {
	URL    string
	Logger *zap.Logger
}

func NewRedis(opts Options) (auditlog.Store, error) {
	if opts.URL == "" {
		opts.URL = "rediss://default@localhost:6379"
	}
	opt, err := redis.ParseURL(opts.URL)
	if err != nil {
		return nil, err
	}
	db := redis.NewClient(opt)
	db.AddHook(redisotel.NewTracingHook(redisotel.WithAttributes(semconv.NetSockPeerAddrKey.String(opt.Addr))))

	r := &redisContainer{
		db:     db,
		logger: opts.Logger,
	}

	return r, nil
}

type redisContainer struct {
	db     *redis.Client
	logger *zap.Logger
}

func (r *redisContainer) Store(entry auditlog.Entry) error {
	ck := contextKey(entry.OrgID, entry.ContextID)

	bts, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = r.db.RPush(context.TODO(), ck, bts).Result()
	if err != nil {
		return err
	}

	return nil
}

func contextKey(orgID string, contextID string) string {
	ctx := ""
	if contextID != "" {
		ctx = fmt.Sprintf(":context:%s", contextID)
	}
	return fmt.Sprintf("org:%s%s", orgID, ctx)
}
