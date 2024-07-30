package fs

import "context"

// Cache is the interface that defines the methods that a cache must implement
type Cache interface {
	Name() string
	Set(ctx context.Context, key any, value any) error
	Get(ctx context.Context, key any) (any, error)
	Delete(ctx context.Context, key any) error
	Clear(ctx context.Context) error
}

// Cache driver config holds the driver configuration
type CacheDriverConfig struct {
	Driver   string `json:"driver"`
	Address  string `json:"address"`
	Password string `json:"password"`
	Database int    `json:"database"`
}

// Clone returns a clone of the cache driver configuration
func (dc *CacheDriverConfig) Clone() *CacheDriverConfig {
	return &CacheDriverConfig{
		Address:  dc.Address,
		Password: dc.Password,
		Database: dc.Database,
	}
}

// CacheConfig holds the cache configuration
type CacheConfig struct {
	DefaultCache      string               `json:"default_driver"`
	CacheDriverConfig []*CacheDriverConfig `json:"caches"`
}

// Clone returns a clone of the cache configuration
func (dc *CacheConfig) Clone() *CacheConfig {
	if dc == nil {
		return nil
	}

	clone := &CacheConfig{
		DefaultCache:      dc.DefaultCache,
		CacheDriverConfig: make([]*CacheDriverConfig, len(dc.CacheDriverConfig)),
	}

	for i, dc := range dc.CacheDriverConfig {
		clone.CacheDriverConfig[i] = dc.Clone()
	}

	return clone
}
