package schemaservice

import (
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/schema"
)

func (ss *SchemaService) Detail(c app.Context, _ *any) (*schema.Schema, error) {
	s, err := ss.app.SchemaBuilder().Schema(c.Arg("name"))
	if err != nil {
		return nil, errors.NotFound(err.Error())
	}

	return s, nil
}
