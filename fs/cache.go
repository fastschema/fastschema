package fs

import (
	"context"
)

// Cache is the interface that defines the methods that a cache must implement
type Cache interface {
	Driver() string
	Name() string
	Set(ctx context.Context, key any, value any) error
	Get(ctx context.Context, key any, binds ...any) (any, error)
	Delete(ctx context.Context, key any) error
	Clear(ctx context.Context) error
}

// CacheConfig holds the cache configuration
type CacheConfig struct {
	DefaultCache      string `json:"default_driver"`
	CacheDriverConfig []Map  `json:"caches"`
}

// Clone returns a clone of the cache configuration
func (dc *CacheConfig) Clone() *CacheConfig {
	if dc == nil {
		return nil
	}

	clone := &CacheConfig{
		DefaultCache:      dc.DefaultCache,
		CacheDriverConfig: make([]Map, len(dc.CacheDriverConfig)),
	}

	for i, dc := range dc.CacheDriverConfig {
		clonedMap := Map{}
		for k, v := range dc {
			clonedMap[k] = v
		}

		clone.CacheDriverConfig[i] = clonedMap
	}

	return clone
}
