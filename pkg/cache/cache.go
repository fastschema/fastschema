package caches

import (
	"errors"

	"github.com/fastschema/fastschema/fs"
	"github.com/redis/go-redis/v9"
)

func NewFromConfig(cacheConfigs []fs.Map) ([]fs.Cache, error) {
	local, err := NewLocal("local", &LocalCache{})
	if err != nil {
		return nil, err
	}

	caches := []fs.Cache{local}
	for _, cacheConfig := range cacheConfigs {
		if cacheConfig == nil {
			continue
		}

		driver, ok := cacheConfig["driver"].(string)
		if driver == "" || !ok {
			return nil, errors.New("cache driver is required")
		}

		name, ok := cacheConfig["name"].(string)
		if name == "" || !ok {
			return nil, errors.New("cache name is required")
		}

		switch driver {
		case "redis":
			redisCache, err := NewRedis(name, &redis.Options{
				Network:               fs.MapValue(cacheConfig, "network", ""),
				Addr:                  fs.MapValue(cacheConfig, "address", ""),
				ClientName:            fs.MapValue(cacheConfig, "client_name", ""),
				Username:              fs.MapValue(cacheConfig, "username", ""),
				Password:              fs.MapValue(cacheConfig, "password", ""),
				DB:                    fs.MapValue(cacheConfig, "db", 0),
				MaxRetries:            fs.MapValue(cacheConfig, "max_retries", 0),
				ContextTimeoutEnabled: fs.MapValue(cacheConfig, "context_timeout_enabled", false),
				PoolFIFO:              fs.MapValue(cacheConfig, "pool_fifo", false),
				PoolSize:              fs.MapValue(cacheConfig, "pool_size", 0),
				MinIdleConns:          fs.MapValue(cacheConfig, "min_idle_conns", 0),
				MaxIdleConns:          fs.MapValue(cacheConfig, "max_idle_conns", 0),
			})

			if err != nil {
				return nil, err
			}

			caches = append(caches, redisCache)
		default:
			// do nothing
		}
	}

	return caches, nil
}
