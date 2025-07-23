package openapi

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen/ir"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ResourceInfo struct {
	ID         string
	Path       string
	Method     string
	Signatures fs.Signatures
	Args       fs.Args
	Public     bool
}

type ReferenceSchemaType struct {
	Name string
	Type reflect.Type
}

type OpenAPISpec struct {
	config             *OpenAPISpecConfig
	spec               []byte
	ogenSpec           *ogen.Spec
	referenceSchemas   map[string]ReferenceSchemaType
	unnamedSchemaIndex int
	mu                 sync.Mutex
}

type OpenAPISpecConfig struct {
	Resources     *fs.ResourcesManager
	SchemaBuilder *schema.Builder
	BaseURL       string
}

// Clone creates a copy of the ResourceInfo object.
func (r *ResourceInfo) Clone() *ResourceInfo {
	return &ResourceInfo{
		ID:         r.ID,
		Path:       r.Path,
		Method:     r.Method,
		Signatures: r.Signatures,
		Args:       r.Args,
		Public:     r.Public,
	}
}

// NewSpec creates a new OpenAPISpec instance based on system resources and schema builder.
func NewSpec(config *OpenAPISpecConfig) (*OpenAPISpec, error) {
	if config == nil || config.Resources == nil {
		return nil, errors.InternalServerError("resources is required")
	}

	oas := &OpenAPISpec{
		config: &OpenAPISpecConfig{
			Resources:     config.Resources.Clone(),
			SchemaBuilder: config.SchemaBuilder,
			BaseURL:       config.BaseURL,
		},
		referenceSchemas: map[string]ReferenceSchemaType{},
		ogenSpec: &ogen.Spec{
			OpenAPI: "3.1.0",
			Info:    CreateInfoObject(),
			Servers: []ogen.Server{
				{
					URL: config.BaseURL,
				},
			},
			Paths: map[string]*ogen.PathItem{},
			Components: &ogen.Components{
				Schemas: map[string]*ogen.Schema{
					SchemaIDOnlyName: IDOnlySchema,
				},
				Parameters: CreateParamsObject(),
				Responses: ogen.Responses{
					"401": unAuthorizedResponse,
				},
			},
			Extensions: ogen.Extensions(nil),
		},
	}

	if err := oas.Create(); err != nil {
		return nil, err
	}

	return oas, nil
}

func (oas *OpenAPISpec) Create() (err error) {
	oas.SchemasToOGenSchemas()
	oas.CreateResourcesForSchemas()
	flattenedResources, err := FlattenResources(oas.config.Resources.Resources(), "/", fs.Args{})
	if err != nil {
		return err
	}

	for _, resource := range flattenedResources {
		pathItems, err := oas.CreatePathItem(resource)
		if err != nil {
			return err
		}

		oas.ogenSpec.Paths = MergePathItems(oas.ogenSpec.Paths, pathItems)
	}

	oas.ResolveSchemaReferences()

	if oas.spec, err = json.MarshalIndent(oas.ogenSpec, "", "  "); err != nil {
		return err
	}

	return nil
}

func (oas *OpenAPISpec) Spec() []byte {
	return oas.spec
}

func (oas *OpenAPISpec) Schema(name string) *ogen.Schema {
	return oas.ogenSpec.Components.Schemas[name]
}

// CreateResourcesForSchemas creates resources for the system schemas.
func (oas *OpenAPISpec) CreateResourcesForSchemas() {
	if oas.config.SchemaBuilder == nil {
		return
	}

	schemas := oas.config.SchemaBuilder.Schemas()
	api := oas.config.Resources.Find("api")
	if api == nil {
		return // no api resource
	}

	contentGroup := api.Group("content")
	contentIDArg := fs.Arg{
		Required:    true,
		Type:        fs.TypeUint64,
		Description: "The content ID",
	}
	contentFilterArg := fs.Arg{
		Type:        fs.TypeJSON,
		Description: "Filter the results by a field",
		Example:     `{"name":{"$like":"%test%"}}`,
	}

	listArgs := fs.Args{
		"sort": {
			Type:        fs.TypeString,
			Description: "Sort the results by a field",
			Example:     "-id",
		},
		"filter": contentFilterArg,
		"page": {
			Type:        fs.TypeUint,
			Description: "The page number",
		},
		"limit": {
			Type:        fs.TypeUint,
			Description: "The number of items per page",
		},
	}

	for _, s := range schemas {
		if s.IsJunctionSchema {
			continue
		}

		schemaGroup := contentGroup.Group(s.Name)
		contentCreateSchema := ContentCreateSchema(s)
		contentDetailSchema := ContentDetailSchema(s)

		schemaGroup.AddResource("list", nil, &fs.Meta{
			Get:        "/",
			Signatures: []any{nil, ContentListResponseSchema(s)},
			Args:       listArgs,
		})
		schemaGroup.AddResource("detail", nil, &fs.Meta{
			Get:        "/:id",
			Signatures: []any{nil, contentDetailSchema},
			Args: fs.Args{
				"id": contentIDArg,
				"select": {
					Type:        fs.TypeString,
					Description: "Select the fields to return",
					Example:     "id,name",
				},
			},
		})
		schemaGroup.AddResource("create", nil, &fs.Meta{
			Post:       "/",
			Signatures: []any{contentCreateSchema, contentDetailSchema},
		})
		schemaGroup.AddResource("update", nil, &fs.Meta{
			Put:        "/:id",
			Args:       fs.Args{"id": contentIDArg},
			Signatures: []any{contentCreateSchema, contentDetailSchema},
		})
		schemaGroup.AddResource("delete", nil, &fs.Meta{
			Delete:     "/:id",
			Args:       fs.Args{"id": contentIDArg},
			Signatures: []any{nil, IDOnlySchema},
		})
		schemaGroup.AddResource("bulk-update", nil, &fs.Meta{
			Put:        "/update",
			Args:       fs.Args{"filter": contentFilterArg},
			Signatures: []any{contentCreateSchema, 0},
		})
		schemaGroup.AddResource("bulk-delete", nil, &fs.Meta{
			Delete:     "/delete",
			Args:       fs.Args{"filter": contentFilterArg},
			Signatures: []any{nil, 0},
		})
	}
}

// CreatePathItem creates a PathItem object for the given ResourceInfo.
//
//	Returns a map of path strings to PathItem objects.
func (oas *OpenAPISpec) CreatePathItem(r *ResourceInfo) (map[string]*ogen.PathItem, error) {
	paths := make(map[string]*ogen.PathItem)
	pathItem := ogen.NewPathItem().SetDescription(r.ID)
	tags := []string{"Main"}
	groups := strings.Split(r.ID, ".")
	path, pathParams := NormalizePath(r.Path)

	// if params is not existed in the args, return error
	for _, param := range pathParams {
		if _, ok := r.Args[param]; !ok {
			return nil, errors.InternalServerError("resource: %s missing param %s in args", r.ID, param)
		}
	}

	// use the nearest group as tag
	if len(groups) > 1 {
		capitalized := cases.Title(language.English).String(groups[len(groups)-2])
		tags = []string{capitalized}
	}

	operation := &ogen.Operation{
		Tags:        tags,
		OperationID: r.ID,
		Responses:   ogen.Responses{},
	}

	if len(r.Signatures) > 0 && utils.Contains([]string{"POST", "PUT", "PATCH"}, r.Method) {
		bodyType := r.Signatures[0]
		bodyTypeName := ""
		signature, ok := bodyType.(*fs.Signature)
		if ok {
			bodyType = signature.Type
			bodyTypeName = signature.Name
		}

		bodySchema := oas.TypeToOgenSchema(bodyType, &TypeToOgenConfig{structName: bodyTypeName})
		if bodySchema != nil {
			operation.RequestBody = &ogen.RequestBody{
				Content: map[string]ogen.Media{
					ir.EncodingJSON.String(): {Schema: bodySchema},
				},
			}
		}
	}

	if len(r.Signatures) > 1 {
		responseType := r.Signatures[1]
		responseTypeName := ""
		signature, ok := responseType.(*fs.Signature)
		if ok {
			responseType = signature.Type
			responseTypeName = signature.Name
		}

		responseSchema := oas.TypeToOgenSchema(responseType, &TypeToOgenConfig{structName: responseTypeName})
		if responseSchema != nil {
			dataSchema := ogen.NewSchema()
			dataSchema.AddRequiredProperties(responseSchema.ToProperty("data"))
			operation.Responses["200"] = &ogen.Response{
				Description: "Successful response",
				Content: map[string]ogen.Media{
					ir.EncodingJSON.String(): {
						Schema: dataSchema,
					},
				},
			}
		}
	}

	if !r.Public {
		operation.Responses["401"] = &ogen.Response{Ref: "#/components/responses/401"}
		operation.Parameters = append(
			operation.Parameters,
			&ogen.Parameter{Ref: "#/components/parameters/authBearerHeader"},
		)
	}

	parameters, err := CreateParameters(r.Args, pathParams)
	if err != nil {
		return nil, err
	}

	operation.Parameters = MergeParameters(operation.Parameters, parameters)

	switch r.Method {
	case "GET":
		pathItem.Get = operation
	case "HEAD":
		pathItem.Head = operation
	case "POST":
		pathItem.Post = operation
	case "PUT":
		pathItem.Put = operation
	case "DELETE":
		pathItem.Delete = operation
	case "TRACE":
		pathItem.Trace = operation
	case "PATCH":
		pathItem.Patch = operation
	case "OPTIONS":
		pathItem.Options = operation
	}

	paths[path] = pathItem

	return paths, nil
}

func (oas *OpenAPISpec) GetSchemaIndex() int {
	oas.mu.Lock()
	defer oas.mu.Unlock()
	oas.unnamedSchemaIndex++
	return oas.unnamedSchemaIndex
}
