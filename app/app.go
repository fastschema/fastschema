package app

import (
	"github.com/fastschema/fastschema/schema"
)

// App is the interface that defines the methods that an app must implement
type App interface {
	Key() string
	SchemaBuilder() *schema.Builder
	DB() DBClient
	Resources() *ResourcesManager
	Reload(migration *Migration) (err error)
	Logger() Logger
	UpdateCache() error
	Roles() []*Role
	Disk(names ...string) Disk

	AddResource(resource *Resource)
	AddMiddlewares(hooks ...Middleware)
	Hooks() *Hooks
	OnBeforeResolve(hooks ...Middleware)
	OnAfterResolve(hooks ...Middleware)
	OnAfterDBContentList(AfterDBContentListHook)
}

// ResolveHook is a function that can be used to add hooks to a resource
type ResolveHook = Middleware

// Hooks is a struct that contains app hooks
type Hooks struct {
	BeforeResolve      []ResolveHook
	AfterResolve       []ResolveHook
	AfterDBContentList []AfterDBContentListHook
}
