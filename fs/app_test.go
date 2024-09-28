package fs_test

import (
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/stretchr/testify/assert"
)

func TestHooks_Clone(t *testing.T) {
	h := &fs.Hooks{
		DBHooks: &db.Hooks{
			PostDBQuery:  []db.PostDBQuery{},
			PostDBCreate: []db.PostDBCreate{},
			PostDBUpdate: []db.PostDBUpdate{},
			PostDBDelete: []db.PostDBDelete{},
			PreDBQuery:   []db.PreDBQuery{},
			PreDBCreate:  []db.PreDBCreate{},
			PreDBUpdate:  []db.PreDBUpdate{},
			PreDBDelete:  []db.PreDBDelete{},
		},
		PreResolve:  []fs.ResolveHook{func(ctx fs.Context) error { return nil }},
		PostResolve: []fs.ResolveHook{func(ctx fs.Context) error { return nil }},
	}

	clone := h.Clone()

	assert.EqualValues(t, h.DBHooks, clone.DBHooks)
	assert.Equal(t, len(h.PreResolve), len(clone.PreResolve))
	assert.Equal(t, len(h.PostResolve), len(clone.PostResolve))
}
