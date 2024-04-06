package app

import (
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
	Reload(migration *db.Migration) (err error)
	Logger() logger.Logger
	UpdateCache() error
	Roles() []*Role
	GetRolesFromIDs(ids []uint64) []*Role
	GetRoleDetail(roleID uint64) *Role
	GetRolePermission(roleID uint64, action string) *Permission
	Disk(names ...string) Disk

	AddResource(resource *Resource)
	AddMiddlewares(hooks ...Middleware)
	Hooks() *Hooks
	OnBeforeResolve(hooks ...Middleware)
	OnAfterResolve(hooks ...Middleware)
	OnAfterDBContentList(db.AfterDBContentListHook)
}

// ResolveHook is a function that can be used to add hooks to a resource
type ResolveHook = Middleware

// Hooks is a struct that contains app hooks
type Hooks struct {
	BeforeResolve []ResolveHook
	AfterResolve  []ResolveHook
	ContentList   []db.AfterDBContentListHook
}
