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
