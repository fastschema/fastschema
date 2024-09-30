package caches

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	redisstore "github.com/eko/gocache/store/redis/v4"
	"github.com/fastschema/fastschema/fs"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	name   string
	config *redis.Options
	cache  *cache.Cache[any]
}

func NewRedis(name string, config *redis.Options) (fs.Cache, error) {
	rdb := redis.NewClient(config)
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	redisStore := redisstore.NewRedis(rdb, store.WithExpiration(10*time.Minute))
	redis := &RedisCache{
		name:   name,
		config: config,
		cache:  cache.New[any](redisStore),
	}

	return redis, nil
}

func (c *RedisCache) Name() string {
	return c.name
}

func (c *RedisCache) Driver() string {
	return "redis"
}

func (c *RedisCache) Get(ctx context.Context, key any, binds ...any) (any, error) {
	retrievedData, err := c.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var jsonBytes []byte

	switch v := retrievedData.(type) {
	case string:
		jsonBytes = []byte(v)
	case []byte:
		jsonBytes = v
	default:
		return nil, fmt.Errorf("unexpected cache data type %v", reflect.TypeOf(v))
	}

	if len(binds) > 0 {
		if err := json.Unmarshal(jsonBytes, &binds[0]); err != nil {
			return nil, err
		}

		return binds[0], nil
	}

	return jsonBytes, nil
}

func (c *RedisCache) Set(ctx context.Context, key any, value any) error {
	jsonData, err := json.Marshal(value)

	if err != nil {
		panic(err)
	}

	return c.cache.Set(ctx, key, jsonData)
}

func (c *RedisCache) Delete(ctx context.Context, key any) error {
	return c.cache.Delete(ctx, key)
}

func (c *RedisCache) Clear(ctx context.Context) error {
	return c.cache.Clear(ctx)
}
