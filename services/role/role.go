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
}

type RoleService struct {
	DB          func() db.Client
	AppKey      func() string
	UpdateCache func(context.Context) error
}

func New(app AppLike) *RoleService {
	return &RoleService{
		DB:          app.DB,
		AppKey:      app.Key,
		UpdateCache: app.UpdateCache,
	}
}

func (rs *RoleService) CreateResource(api *fs.Resource) {
	api.Group("role").
		Add(fs.NewResource("list", rs.List, &fs.Meta{Get: "/"})).
		Add(fs.NewResource("detail", rs.Detail, &fs.Meta{
			Get:  "/:id",
			Args: fs.Args{"id": fs.CreateArg(fs.TypeUint64, "The role ID")},
		})).
		Add(fs.NewResource("create", rs.Create, &fs.Meta{Post: "/"})).
		Add(fs.NewResource("update", rs.Update, &fs.Meta{
			Put:  "/:id",
			Args: fs.Args{"id": fs.CreateArg(fs.TypeUint64, "The role ID")},
		})).
		Add(fs.NewResource("delete", rs.Delete, &fs.Meta{
			Delete: "/:id",
			Args:   fs.Args{"id": fs.CreateArg(fs.TypeUint64, "The role ID")},
		}))
}
