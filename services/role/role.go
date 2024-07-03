package roleservice

import (
	"context"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
)

type AppLike interface {
	DB() db.Client
	Key() string
	UpdateCache(ctx context.Context) error
	Resources() *fs.ResourcesManager
}

type RoleService struct {
	DB          func() db.Client
	AppKey      func() string
	UpdateCache func(context.Context) error
	Resources   func() *fs.ResourcesManager
}

func New(app AppLike) *RoleService {
	return &RoleService{
		DB:          app.DB,
		AppKey:      app.Key,
		UpdateCache: app.UpdateCache,
		Resources:   app.Resources,
	}
}
