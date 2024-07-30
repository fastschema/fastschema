package caches

import (
	"context"

	"github.com/fastschema/fastschema/fs"
)

type CacheCloneLocal struct {
	Driver string `json:"driver"`
	roles  any
}

func NewLocal(config *CacheCloneLocal) (fs.Cache, error) {
	return &CacheCloneLocal{}, nil
}

func (c *CacheCloneLocal) Name() string {
	return "local"
}

func (c *CacheCloneLocal) Get(ctx context.Context, key any) (any, error) {
	if key == "roles" {
		return c.roles, nil
	}

	return nil, nil
}

func (c *CacheCloneLocal) Set(ctx context.Context, key any, value any) error {
	if key == "roles" {
		c.roles = value
	}

	return nil
}

func (c *CacheCloneLocal) Delete(ctx context.Context, key any) error {
	c.roles = nil

	return nil
}

func (c *CacheCloneLocal) Clear(ctx context.Context) error {
	c.roles = nil

	return nil
}
