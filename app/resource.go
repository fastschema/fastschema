package app

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

var resourceNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_:/]+$`)

type StaticFs struct {
	Root       http.FileSystem
	BasePath   string
	PathPrefix string
}

// ResolverFn is a function that generates a resolver function
type ResolverFn[Input, Output any] func(c Context, input Input) (Output, error)

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
	resolver   Resolver
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

// NewResource creates a new resource with the given name, solverGenerator and meta
//
//	solverGenerator is a function that generates a resolver function.
//	If the solver input type is not "any", the input will be parsed from the context
func NewResource[Input, Output any](
	name string,
	resolver ResolverFn[Input, Output],
	metas ...*Meta,
) *Resource {
	var inputValue Input
	var outputValue Output

	parseInput := utils.IsNotAny(inputValue)
	resource := &Resource{
		name:       name,
		signatures: []any{inputValue, outputValue},
		resources:  make([]*Resource, 0),
		resolver: func(ctx Context) (any, error) {
			var input Input
			if parseInput {
				if err := ctx.Parse(&input); err != nil {
					return nil, err
				}
			}

			return resolver(ctx, input)
		},
	}

	if len(metas) > 0 {
		resource.meta = metas[0]
	}

	return resource
}

func (r *Resource) generateID(parentID string) string {
	if parentID == "" {
		return r.name
	}

	return parentID + "." + r.name
}

func (r *Resource) add(resource *Resource) (self *Resource) {
	resource.id = resource.generateID(r.id)
	r.resources = append(r.resources, resource)
	return r
}

func (r *Resource) Remove(resource *Resource) (self *Resource) {
	for i, res := range r.resources {
		if res == resource {
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
		resolver:   r.resolver,
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
func (r *Resource) AddResource(name string, resolver Resolver, metas ...*Meta) (self *Resource) {
	resource := &Resource{
		name:       name,
		resolver:   resolver,
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
	var matchedResource *Resource = nil
	parts := strings.Split(resourceID, ".")
	currentResource := r

	for _, part := range parts {
		for _, resource := range currentResource.resources {
			if resource.name == part {
				matchedResource = resource
				currentResource = resource
				break
			}
		}
	}

	return matchedResource
}

// ID returns the id of the resource
func (r *Resource) ID() string {
	return r.id
}

// Name returns the name of the resource
func (r *Resource) Name() string {
	return r.name
}

// Resolver returns the resolver of the resource
func (r *Resource) Resolver() Resolver {
	return r.resolver
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

	groupResource.generateID(r.id)
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

	for i := 0; i < level; i++ {
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

	// check resource name to match resourceNameRegexp
	if !resourceNameRegexp.MatchString(r.name) {
		return fmt.Errorf("Resource name must match %s, got %s", resourceNameRegexp.String(), r.name)
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
	entity := schema.NewEntity().
		Set("id", r.id).
		Set("name", r.name)

	if r.group {
		entity.Set("group", r.group)
	}

	if r.meta != nil {
		entity.Set("meta", r.meta)
	}

	// if len(r.signature) > 0 {
	// 	entity.Set("signature", r.signature)
	// }

	if len(r.resources) > 0 {
		entity.Set("resources", r.resources)
	}

	return entity.MarshalJSON()
}
