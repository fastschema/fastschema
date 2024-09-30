package fs

import (
	"context"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/schema"
)

type Hookable interface {
	Hooks() *Hooks

	OnPreResolve(hooks ...Middleware)
	OnPostResolve(hooks ...Middleware)

	OnPreDBQuery(hooks ...db.PreDBQuery)
	OnPostDBQuery(hooks ...db.PostDBQuery)

	OnPreDBExec(hooks ...db.PreDBExec)
	OnPostDBExec(hooks ...db.PostDBExec)

	OnPreDBCreate(hooks ...db.PreDBCreate)
	OnPostDBCreate(hooks ...db.PostDBCreate)

	OnPreDBUpdate(hooks ...db.PreDBUpdate)
	OnPostDBUpdate(hooks ...db.PostDBUpdate)

	OnPreDBDelete(hooks ...db.PreDBDelete)
	OnPostDBDelete(hooks ...db.PostDBDelete)
}

// App is the interface that defines the methods that an app must implement
type App interface {
	Hookable
	Key() string
	Config() *Config
	SchemaBuilder() *schema.Builder
	DB() db.Client
	Resources() *ResourcesManager
	Reload(ctx context.Context, migration *db.Migration) (err error)
	Logger() logger.Logger
	UpdateCache(ctx context.Context, keys ...string) error
	Roles() ([]*Role, error)
	Disk(names ...string) Disk
	Disks() []Disk
	Cache(names ...string) Cache
	Caches() []Cache

	AddResource(resource *Resource)
	AddMiddlewares(hooks ...Middleware)
}

// ResolveHook is a function that can be used to add hooks to a resource
type ResolveHook = Middleware

// Hooks is a struct that contains app hooks
type Hooks struct {
	DBHooks     *db.Hooks
	PreResolve  []ResolveHook
	PostResolve []ResolveHook
}

func (h *Hooks) Clone() *Hooks {
	return &Hooks{
		DBHooks:     h.DBHooks.Clone(),
		PreResolve:  append([]ResolveHook{}, h.PreResolve...),
		PostResolve: append([]ResolveHook{}, h.PostResolve...),
	}
}
