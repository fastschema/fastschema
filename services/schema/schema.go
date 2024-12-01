package schemaservice

import (
	"context"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/schema"
)

type AppLike interface {
	DB() db.Client
	Reload(ctx context.Context, migration *db.Migration) error
	SchemaBuilder() *schema.Builder
	Disk(names ...string) fs.Disk
}

type SchemaService struct {
	app AppLike
}

func New(app AppLike) *SchemaService {
	ss := &SchemaService{app}

	return ss
}

func (ss *SchemaService) CreateResource(api *fs.Resource) {
	api.Group("schema").
		Add(fs.NewResource("list", ss.List, &fs.Meta{Get: "/"})).
		Add(fs.NewResource("create", ss.Create, &fs.Meta{Post: "/"})).
		Add(fs.NewResource("detail", ss.Detail, &fs.Meta{
			Get:  "/:name",
			Args: fs.Args{"name": fs.CreateArg(fs.TypeString, "The schema name")},
		})).
		Add(fs.NewResource("update", ss.Update, &fs.Meta{
			Put:  "/:name",
			Args: fs.Args{"name": fs.CreateArg(fs.TypeString, "The schema name")},
		})).
		Add(fs.NewResource("delete", ss.Delete, &fs.Meta{
			Delete: "/:name",
			Args:   fs.Args{"name": fs.CreateArg(fs.TypeString, "The schema name")},
		})).
		Add(fs.NewResource("import", ss.Import, &fs.Meta{Post: "/import"})).
		Add(fs.NewResource("export", ss.Export, &fs.Meta{Post: "/export"}))
}
