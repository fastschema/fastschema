package caches

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	redisstore "github.com/eko/gocache/store/redis/v4"
	"github.com/fastschema/fastschema/fs"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
)

type CacheCloneRedisConfig struct {
	Driver   string `json:"driver"`
	Address  string `json:"address"`
	Password string `json:"password"`
	Database int    `json:"database"`
}

type CacheCloneRedis struct {
	CacheManager *cache.Cache[any]
}

func NewRedis(config *CacheCloneRedisConfig) (fs.Cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Address,  // use default Addr
		Password: config.Password, // no password set
		DB:       config.Database, // use default DB
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()

	if err != nil {
		panic(err)
	}

	redisStore := redisstore.NewRedis(rdb, store.WithExpiration(10*time.Minute))

	cacheManager := cache.New[any](redisStore)
	redis := &CacheCloneRedis{
		CacheManager: cacheManager,
	}

	return redis, nil
}

func (c *CacheCloneRedis) Name() string {
	return "redis"
}

func (c *CacheCloneRedis) Get(ctx context.Context, key any) (any, error) {
	retrievedData, err := c.CacheManager.Get(ctx, key)

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
		log.Info("Unexpected type: %s\n", reflect.TypeOf(retrievedData))
	}

	if key == "roles" {
		roles := []*fs.Role{}

		err = json.Unmarshal([]byte(jsonBytes), &roles)

		return roles, err
	}

	return nil, nil
}

func (c *CacheCloneRedis) Set(ctx context.Context, key any, value any) error {
	jsonData, err := json.Marshal(value)

	if err != nil {
		panic(err)
	}

	return c.CacheManager.Set(ctx, key, jsonData)
}

func (c *CacheCloneRedis) Delete(ctx context.Context, key any) error {
	return c.CacheManager.Delete(ctx, key)
}

func (c *CacheCloneRedis) Clear(ctx context.Context) error {
	return c.CacheManager.Clear(ctx)
}
