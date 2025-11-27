package fs

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

type StaticConfig struct {
	Compress      bool          `json:"compress"`
	ByteRange     bool          `json:"byte_range"`
	Browse        bool          `json:"browse"`
	Download      bool          `json:"download"`
	Index         string        `json:"index"`
	CacheDuration time.Duration `json:"cache_duration"` // Default value 10 * time.Second.
	MaxAge        int           `json:"max_age"`        // Default value 0
}

type StaticFs struct {
	// Config is the static file server configuration.
	Config *StaticConfig

	// BasePath is the base url path to serve the static files, e.g. "/static".
	// Default is "/"
	BasePath string

	// RootDir is the root directory of the static files in the FileSystem.
	RootDir string

	// RootFS is the FileSystem that provides access to a file system.
	RootFS http.FileSystem

	// FSPrefix is a prefix to be added to the path when reading from the FileSystem.
	// Use when using embed.FS, default ""
	FSPrefix string
}

// HandlerFn is a function that generates a resolver function
type HandlerFn[Input, Output any] func(c Context, input Input) (Output, error)

// Middleware is a function that can be used to add middleware to a resource
type Middleware func(c Context) error

// ResourcesManager is a resource manager that can be used to manage resources
type ResourcesManager struct {
	*Resource

	Middlewares []Middleware
	Hooks       func() *Hooks
}

// Clone clones the resource manager and all sub resources
func (rs *ResourcesManager) Clone() *ResourcesManager {
	clone := &ResourcesManager{
		Resource:    &Resource{},
		Middlewares: make([]Middleware, len(rs.Middlewares)),
		Hooks:       rs.Hooks,
	}

	if rs.Resource != nil {
		clone.Resource = rs.Resource.Clone()
	}

	copy(clone.Middlewares, rs.Middlewares)

	return clone
}

// Init validates the resource and all sub resources
func (rs *ResourcesManager) Init() error {
	// check current resource and all sub resources to prevent duplicate resource id
	resourceIds := make(map[string]bool)
	for _, resource := range rs.resources {
		if _, ok := resourceIds[resource.id]; ok {
			return errors.New("Duplicate resource id: " + resource.id)
		}

		resourceIds[resource.id] = true
	}

	for _, resource := range rs.resources {
		if err := resource.Init(); err != nil {
			return err
		}
	}

	return nil
}

// Resource is a resource that can be used to define a resource tree
type Resource struct {
	id         string
	group      bool
	resources  []*Resource
	name       string
	handler    Handler
	meta       *Meta
	signatures Signatures
}

// NewResourcesManager creates a new resources manager
func NewResourcesManager() *ResourcesManager {
	return &ResourcesManager{
		Resource: &Resource{
			name:      "",
			group:     true,
			resources: make([]*Resource, 0),
		},
		Middlewares: make([]Middleware, 0),
	}
}

// NewResource creates a new resource with the given name, handler and meta
//
//	handler is a function that takes a context and an input and returns an output and an error
//	If the solver input type is not "any", the input will be parsed from the context
func NewResource[Input, Output any](
	name string,
	handler HandlerFn[Input, Output],
	metas ...*Meta,
) *Resource {
	var inputValue Input
	var outputValue Output

	parseInput := utils.IsNotAny(inputValue)
	resource := &Resource{
		name:       name,
		signatures: []any{inputValue, outputValue},
		resources:  make([]*Resource, 0),
		handler: func(ctx Context) (any, error) {
			var input Input

			if parseInput {
				if err := ctx.Bind(&input); err != nil {
					return nil, errors.BadRequest(err.Error())
				}
			}

			return handler(ctx, input)
		},
	}

	if len(metas) > 0 {
		resource.meta = metas[0]
	}

	resource.generateID("")

	return resource
}

func createResourceWithMethod[Input, Output any](
	name string,
	method string,
	resolver HandlerFn[Input, Output],
	metas ...*Meta,
) *Resource {
	if len(metas) == 0 {
		metas = append(metas, &Meta{})
	}

	switch method {
	case "GET":
		if metas[0].Get == "" {
			metas[0].Get = name
		}
	case "HEAD":
		if metas[0].Head == "" {
			metas[0].Head = name
		}
	case "POST":
		if metas[0].Post == "" {
			metas[0].Post = name
		}
	case "PUT":
		if metas[0].Put == "" {
			metas[0].Put = name
		}
	case "DELETE":
		if metas[0].Delete == "" {
			metas[0].Delete = name
		}
	case "CONNECT":
		if metas[0].Connect == "" {
			metas[0].Connect = name
		}
	case "OPTIONS":
		if metas[0].Options == "" {
			metas[0].Options = name
		}
	case "TRACE":
		if metas[0].Trace == "" {
			metas[0].Trace = name
		}
	case "PATCH":
		if metas[0].Patch == "" {
			metas[0].Patch = name
		}
	case "WS":
		if metas[0].WS == "" {
			metas[0].WS = name
		}
	}

	return NewResource(name, resolver, metas...)
}

// Get is a shortcut to create a new resource with rest method GET and using name as the get path
func Get[Input, Output any](name string, handler HandlerFn[Input, Output], metas ...*Meta) *Resource {
	return createResourceWithMethod(name, "GET", handler, metas...)
}

// Head is a shortcut to create a new resource with rest method HEAD and using name as the head path
func Head[Input, Output any](name string, handler HandlerFn[Input, Output], metas ...*Meta) *Resource {
	return createResourceWithMethod(name, "HEAD", handler, metas...)
}

// Post is a shortcut to create a new resource with rest method POST and using name as the post path
func Post[Input, Output any](name string, handler HandlerFn[Input, Output], metas ...*Meta) *Resource {
	return createResourceWithMethod(name, "POST", handler, metas...)
}

// Put is a shortcut to create a new resource with rest method PUT and using name as the put path
func Put[Input, Output any](name string, handler HandlerFn[Input, Output], metas ...*Meta) *Resource {
	return createResourceWithMethod(name, "PUT", handler, metas...)
}

// Delete is a shortcut to create a new resource with rest method DELETE and using name as the delete path
func Delete[Input, Output any](name string, handler HandlerFn[Input, Output], metas ...*Meta) *Resource {
	return createResourceWithMethod(name, "DELETE", handler, metas...)
}

// Connect is a shortcut to create a new resource with rest method CONNECT and using name as the connect path
func Connect[Input, Output any](name string, handler HandlerFn[Input, Output], metas ...*Meta) *Resource {
	return createResourceWithMethod(name, "CONNECT", handler, metas...)
}

// Options is a shortcut to create a new resource with rest method OPTIONS and using name as the options path
func Options[Input, Output any](name string, handler HandlerFn[Input, Output], metas ...*Meta) *Resource {
	return createResourceWithMethod(name, "OPTIONS", handler, metas...)
}

// Trace is a shortcut to create a new resource with rest method TRACE and using name as the trace path
func Trace[Input, Output any](name string, handler HandlerFn[Input, Output], metas ...*Meta) *Resource {
	return createResourceWithMethod(name, "TRACE", handler, metas...)
}

// Patch is a shortcut to create a new resource with rest method PATCH and using name as the patch path
func Patch[Input, Output any](name string, handler HandlerFn[Input, Output], metas ...*Meta) *Resource {
	return createResourceWithMethod(name, "PATCH", handler, metas...)
}

// WS is a shortcut to create a new resource with rest method WS and using name as the ws path
func WS[Input, Output any](name string, handler HandlerFn[Input, Output], metas ...*Meta) *Resource {
	return createResourceWithMethod(name, "WS", handler, metas...)
}

func (r *Resource) generateID(parentID string) {
	if parentID == "" {
		r.id = r.name
		return
	}

	r.id = parentID + "." + r.name
}

func (r *Resource) add(resource *Resource) (self *Resource) {
	resource.generateID(r.id)
	r.resources = append(r.resources, resource)
	return r
}

func (r *Resource) Remove(name string) (self *Resource) {
	for i, res := range r.resources {
		if res.name == name {
			r.resources = append(r.resources[:i], r.resources[i+1:]...)
			break
		}
	}

	return r
}

// Clone clones the resource and all sub resources
func (r *Resource) Clone() *Resource {
	clone := &Resource{
		id:         r.id,
		name:       r.name,
		handler:    r.handler,
		signatures: r.signatures,
		group:      r.group,
	}

	if r.meta != nil {
		clone.meta = r.meta.Clone()
	}

	for _, resource := range r.resources {
		clone.add(resource.Clone())
	}

	return clone
}

// AddResource adds a new resource to the current resource as a child and returns the current resource
// extras can be used to pass additional information to the resource. Currently supported extras are:
//   - *Meta: used to pass meta information to the resource, example: &Meta{"rest.POST": "/login"}
//   - *Signature: used to pass input and output signatures to the resource, example: &Signature{Input: LoginData{}, Output: LoginResponse{}}
func (r *Resource) AddResource(name string, handler Handler, metas ...*Meta) (self *Resource) {
	resource := &Resource{
		name:       name,
		handler:    handler,
		signatures: []any{},
	}

	if len(metas) > 0 {
		resource.meta = metas[0]

		if metas[0].Signatures != nil {
			resource.signatures = metas[0].Signatures
		}
	}

	return r.add(resource)
}

// Add adds a new resource to the current resource as a child and returns the current resource
func (r *Resource) Add(resources ...*Resource) (self *Resource) {
	for _, resource := range resources {
		r.add(resource)
	}

	return r
}

// Find returns the resource with the given id
// The id is in the format of "group1.group2.group3.resource"
// While group1, group2 and group3 are name of the groups and resource is the name of the resource
func (r *Resource) Find(resourceID string) *Resource {
	if r.id == resourceID {
		return r
	}

	for _, resource := range r.resources {
		if found := resource.Find(resourceID); found != nil {
			return found
		}
	}

	return nil
}

// ID returns the id of the resource
func (r *Resource) ID() string {
	return r.id
}

// Name returns the name of the resource
func (r *Resource) Name() string {
	return r.name
}

// Handler returns the resolver of the resource
func (r *Resource) Handler() Handler {
	return r.handler
}

// Meta returns the meta of the resource
func (r *Resource) Meta() *Meta {
	return r.meta
}

// Signature returns the signature of the resource
func (r *Resource) Signature() Signatures {
	return r.signatures
}

// Resources returns the sub resources of the resource
func (r *Resource) Resources() []*Resource {
	return r.resources
}

// IsGroup returns true if the resource is a group
func (r *Resource) IsGroup() bool {
	return r.group
}

// IsPublic returns true if the resource is white listed
func (r *Resource) IsPublic() bool {
	if r.meta == nil {
		return false
	}
	return r.meta.Public
}

// Group creates a new resource group and adds it to the current resource as a child and returns the group resource
func (r *Resource) Group(name string, metas ...*Meta) (group *Resource) {
	groupResource := &Resource{
		group:     true,
		resources: make([]*Resource, 0),
		name:      name,
	}

	if len(metas) > 0 {
		groupResource.meta = metas[0]
	}

	r.add(groupResource)

	return groupResource
}

// String returns a string representation of the resource
func (r *Resource) String() string {
	if r.group {
		prefix := "/" + r.name
		if r.meta != nil && r.meta.Prefix != "" {
			prefix = r.meta.Prefix
		}
		name := r.name
		if name == "" {
			name = "root"
		}

		return fmt.Sprintf("[%s] %s", name, prefix)
	}

	printFormat := "- %s"
	printArgs := []any{r.name}

	if r.meta != nil {
		methods := make([]string, 0)
		if r.meta.Get != "" {
			methods = append(methods, "GET: "+r.meta.Get)
		}

		if r.meta.Head != "" {
			methods = append(methods, "HEAD: "+r.meta.Head)
		}

		if r.meta.Post != "" {
			methods = append(methods, "POST: "+r.meta.Post)
		}

		if r.meta.Put != "" {
			methods = append(methods, "PUT: "+r.meta.Put)
		}

		if r.meta.Delete != "" {
			methods = append(methods, "DELETE: "+r.meta.Delete)
		}

		if r.meta.Connect != "" {
			methods = append(methods, "CONNECT: "+r.meta.Connect)
		}

		if r.meta.Options != "" {
			methods = append(methods, "OPTIONS: "+r.meta.Options)
		}

		if r.meta.Trace != "" {
			methods = append(methods, "TRACE: "+r.meta.Trace)
		}

		if r.meta.Patch != "" {
			methods = append(methods, "PATCH: "+r.meta.Patch)
		}

		if len(methods) > 0 {
			printArgs = append(printArgs, strings.Join(methods, ", "))
			printFormat += " - %s"
		}
	}

	return fmt.Sprintf(printFormat, printArgs...)
}

// Print prints the resource and all sub resources
func (r *Resource) Print() {
	level := 0
	if r.id != "" {
		level++
	}

	for _, char := range r.id {
		if char == '.' {
			level++
		}
	}

	for i := 1; i < level; i++ {
		fmt.Print("  ")
	}

	if !r.group {
		fmt.Print("")
	}

	fmt.Println(r.String())

	for _, resource := range r.resources {
		resource.Print()
	}
}

// Init validates the resource and all sub resources
func (r *Resource) Init() error {
	// check empty resource name
	if r.name == "" {
		return errors.New("Resource name cannot be empty")
	}

	// check current resource and all sub resources to prevent duplicate resource id
	resourceIds := make(map[string]bool)
	for _, resource := range r.resources {
		if _, ok := resourceIds[resource.id]; ok {
			return errors.New("Duplicate resource id: " + resource.id)
		}

		resourceIds[resource.id] = true
	}

	for _, resource := range r.resources {
		if err := resource.Init(); err != nil {
			return err
		}
	}

	return nil
}

// MarshalJSON marshals the resource to json
func (r *Resource) MarshalJSON() ([]byte, error) {
	e := entity.New().
		Set("id", r.id).
		Set("name", r.name)

	if r.group {
		e.Set("group", r.group)
	}

	if r.meta != nil {
		e.Set("meta", r.meta)
	}

	// if len(r.signature) > 0 {
	// 	entity.Set("signature", r.signature)
	// }

	if len(r.resources) > 0 {
		e.Set("resources", r.resources)
	}

	return e.MarshalJSON()
}
