package caches

import (
	"context"
	"errors"
	"sync"

	"github.com/fastschema/fastschema/fs"
)

type LocalCache struct {
	sync.Mutex
	name   string
	config *LocalCache
	cache  map[any]any
}

func NewLocal(name string, config *LocalCache) (fs.Cache, error) {
	return &LocalCache{
		name:   name,
		config: config,
		cache:  make(map[any]any),
	}, nil
}

func (c *LocalCache) Name() string {
	return c.name
}

func (c *LocalCache) Driver() string {
	return "local"
}

func (c *LocalCache) Get(ctx context.Context, key any, binds ...any) (any, error) {
	c.Lock()
	defer c.Unlock()

	data, ok := c.cache[key.(string)]
	if !ok {
		return nil, errors.New("cache key not found")
	}

	return data, nil
}

func (c *LocalCache) Set(ctx context.Context, key any, value any) error {
	c.Lock()
	defer c.Unlock()

	if c.cache == nil {
		c.cache = make(map[any]any)
	}

	c.cache[key] = value

	return nil
}

func (c *LocalCache) Delete(ctx context.Context, key any) error {
	c.Lock()
	defer c.Unlock()

	delete(c.cache, key)

	return nil
}

func (c *LocalCache) Clear(ctx context.Context) error {
	c.Lock()
	defer c.Unlock()

	c.cache = make(map[any]any)

	return nil
}
