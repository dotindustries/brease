package redis

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
	"go.dot.industries/brease/models"
	"go.dot.industries/brease/storage"
	"go.uber.org/zap"
)

type Options struct {
	Endpoint string
	Password string
	DB       int
	Logger   *zap.Logger
}

func NewDatabase(opts Options) (storage.Database, error) {
	if opts.Endpoint == "" {
		opts.Endpoint = "localhost:6379"
	}
	db := redis.NewClient(&redis.Options{
		Addr:     opts.Endpoint,
		Password: opts.Password,
		DB:       opts.DB,
	})

	r := &redisContainer{
		db:     db,
		logger: opts.Logger,
	}

	return r, nil
}

type redisContainer struct {
	db       *redis.Client
	logger   *zap.Logger
	rulePool sync.Pool
}

func (r *redisContainer) AddRule(_ context.Context, ownerID string, contextID string, rule models.Rule) error {
	//TODO implement me
	panic("implement me")
}

func (r *redisContainer) Close() error {
	return r.db.Close()
}

func (r *redisContainer) Rules(_ context.Context, ownerID string, contextID string) ([]models.Rule, error) {
	//TODO implement me
	panic("implement me")
}

func (r *redisContainer) RemoveRule(_ context.Context, ownerID string, contextID string, ruleID string) error {
	//TODO implement me
	panic("implement me")
}

func (r *redisContainer) ReplaceRule(_ context.Context, ownerID string, contextID string, rule models.Rule) error {
	//TODO implement me
	panic("implement me")
}

func (r *redisContainer) Exists(ctx context.Context, ownerID string, contextID string, ruleID string) (exists bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (r *redisContainer) SaveAccessToken(ctx context.Context, ownerID string, tokenPair *models.TokenPair) error {
	//TODO implement me
	panic("implement me")
}

func (r *redisContainer) GetAccessToken(ctx context.Context, ownerID string) (*models.TokenPair, error) {
	//TODO implement me
	panic("implement me")
}
