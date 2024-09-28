package fs

import (
	"context"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/schema"
)

// App is the interface that defines the methods that an app must implement
type App interface {
	Key() string
	SchemaBuilder() *schema.Builder
	DB() db.Client
	Resources() *ResourcesManager
	Reload(ctx context.Context, migration *db.Migration) (err error)
	Logger() logger.Logger
	UpdateCache(ctx context.Context) error
	Roles() ([]*Role, error)
	Disk(names ...string) Disk
	Disks() []Disk
	Cache(names ...string) Cache
	Caches() []Cache

	AddResource(resource *Resource)
	AddMiddlewares(hooks ...Middleware)
	Hooks() *Hooks
	OnPreResolve(hooks ...Middleware)
	OnPostResolve(hooks ...Middleware)
	OnPostDBGet(db.PostDBGet)
}

// ResolveHook is a function that can be used to add hooks to a resource
type ResolveHook = Middleware

// Hooks is a struct that contains app hooks
type Hooks struct {
	DBHooks     *db.Hooks
	PreResolve  []ResolveHook
	PostResolve []ResolveHook
}
