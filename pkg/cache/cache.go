package caches

import (
	"github.com/fastschema/fastschema/fs"
)

func NewFromConfig(cacheDriverConfigs []*fs.CacheDriverConfig) ([]fs.Cache, error) {
	var caches []fs.Cache

	for _, cacheDriverConfig := range cacheDriverConfigs {
		switch cacheDriverConfig.Driver {
		case "redis":
			redisCache, err := NewRedis(&CacheCloneRedisConfig{
				Driver:   cacheDriverConfig.Driver,
				Address:  cacheDriverConfig.Address,
				Password: cacheDriverConfig.Password,
				Database: cacheDriverConfig.Database,
			})

			if err != nil {
				return nil, err
			}

			caches = append(caches, redisCache)
		default:
			local, err := NewLocal(&CacheCloneLocal{
				Driver: cacheDriverConfig.Driver,
			})

			if err != nil {
				return nil, err
			}

			caches = append(caches, local)
		}
	}

	return caches, nil
}
