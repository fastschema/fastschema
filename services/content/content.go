package contentservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
)

type AppLike interface {
	DB() db.Client
}

type ContentService struct {
	DB func() db.Client
}

func New(app AppLike) *ContentService {
	return &ContentService{
		DB: app.DB,
	}
}

var schemaArgs = fs.Args{
	"schema": {
		Required:    true,
		Type:        fs.TypeString,
		Description: "The schema name",
	},
}

func (cs *ContentService) CreateResource(api *fs.Resource) {
	api.Group("content", &fs.Meta{Prefix: "/content/:schema", Args: schemaArgs}).
		Add(fs.NewResource("list", cs.List, &fs.Meta{
			Get: "/",
		})).
		Add(fs.NewResource("detail", cs.Detail, &fs.Meta{
			Get:  "/:id",
			Args: fs.Args{"id": fs.CreateArg(fs.TypeUint64, "The content ID")},
		})).
		Add(fs.NewResource("create", cs.Create, &fs.Meta{Post: "/"})).
		Add(fs.NewResource("bulk-update", cs.BulkUpdate, &fs.Meta{
			Put: "/update",
		})).
		Add(fs.NewResource("update", cs.Update, &fs.Meta{
			Put:  "/:id",
			Args: fs.Args{"id": fs.CreateArg(fs.TypeUint64, "The content ID")},
		})).
		Add(fs.NewResource("bulk-delete", cs.BulkDelete, &fs.Meta{
			Delete: "/delete",
		})).
		Add(fs.NewResource("delete", cs.Delete, &fs.Meta{
			Delete: "/:id",
			Args:   fs.Args{"id": fs.CreateArg(fs.TypeUint64, "The content ID")},
		}))
}
