package schemaservice

import (
	"context"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/schema"
)

type AppLike interface {
	DB() db.Client
	Reload(ctx context.Context, migration *db.Migration) error
	SchemaBuilder() *schema.Builder
}

type SchemaService struct {
	app AppLike
}

func New(app AppLike) *SchemaService {
	ss := &SchemaService{app}

	return ss
}
