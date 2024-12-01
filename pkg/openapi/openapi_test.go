package openapi_test

import (
	"testing"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/openapi"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	contentservice "github.com/fastschema/fastschema/services/content"
	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
)

func handler[T any](c fs.Context, input T) (T, error) {
	return input, nil
}

func meta(name string) *fs.Meta {
	return &fs.Meta{Post: "/" + name}
}

func TestResourceInfoClone(t *testing.T) {
	r := &openapi.ResourceInfo{
		ID:         "testresource",
		Path:       "/testresource",
		Method:     "GET",
		Signatures: fs.Signatures{testStruct{}, testStruct{}},
		Args:       fs.Args{},
		Public:     true,
	}

	clone := r.Clone()
	assert.Equal(t, r.ID, clone.ID)
	assert.Equal(t, r.Path, clone.Path)
	assert.Equal(t, r.Method, clone.Method)
	assert.Equal(t, r.Signatures, clone.Signatures)
	assert.Equal(t, r.Args, clone.Args)
	assert.Equal(t, r.Public, clone.Public)
}

func TestNewSpec(t *testing.T) {
	// Case 1: Error nil config
	oas, err := openapi.NewSpec(nil)
	assert.Error(t, err)
	assert.Nil(t, oas)

	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir()))

	// Case 2: Error create duplicate argument
	resources := fs.NewResourcesManager()
	group := resources.Group("group", &fs.Meta{
		Args: fs.Args{
			"arg": {
				Type:     fs.TypeInt,
				Required: true,
			},
		},
	})
	group.Add(
		fs.NewResource("bool", handler[bool], &fs.Meta{
			Post: "/bool",
			Args: fs.Args{
				"arg": {
					Type:     fs.TypeInt,
					Required: true,
				},
			},
		}),
	)
	config := &openapi.OpenAPISpecConfig{
		Resources:     resources,
		SchemaBuilder: sb,
		BaseURL:       "http://localhost:8080",
	}
	oas, err = openapi.NewSpec(config)
	assert.Nil(t, oas)
	assert.Contains(t, err.Error(), "duplicate key arg in args")

	// Case 3: Error create missing params
	resources = fs.NewResourcesManager()
	resources.Add(
		fs.NewResource("bool", handler[bool], &fs.Meta{
			Post: "/bool/:param",
			Args: fs.Args{},
		}),
	)
	config = &openapi.OpenAPISpecConfig{
		Resources:     resources,
		SchemaBuilder: sb,
		BaseURL:       "http://localhost:8080",
	}
	oas, err = openapi.NewSpec(config)
	assert.Nil(t, oas)
	assert.Contains(t, err.Error(), "missing param param in args")

	// Case 4: Success
	resources = fs.NewResourcesManager()
	resources.Add(
		fs.NewResource("bool", handler[bool], meta("bool")),
	)
	config = &openapi.OpenAPISpecConfig{
		Resources:     resources,
		SchemaBuilder: sb,
		BaseURL:       "http://localhost:8080",
	}
	oas, err = openapi.NewSpec(config)
	assert.NoError(t, err)
	assert.NotNil(t, oas)
	assert.NoError(t, oas.Create())
	assert.NotNil(t, oas.Spec())
}

func TestCreatePathItem(t *testing.T) {
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir()))
	resources := fs.NewResourcesManager()
	resources.Add(
		fs.NewResource("bool", handler[bool], meta("bool")),
	)

	config := &openapi.OpenAPISpecConfig{
		Resources:     resources,
		SchemaBuilder: sb,
		BaseURL:       "http://localhost:8080",
	}

	oas, err := openapi.NewSpec(config)
	assert.NoError(t, err)
	assert.NotNil(t, oas)

	resourceGet := &openapi.ResourceInfo{
		ID:     "root.testresource",
		Path:   "/testresource",
		Method: "GET",
		Signatures: fs.Signatures{
			&fs.Signature{
				Type: &struct{ Name string }{},
				Name: "TestStruct",
			},
			&fs.Signature{
				Type: &struct{ Name string }{},
				Name: "TestStruct",
			},
		},
		Args:   fs.Args{},
		Public: true,
	}

	var getResourceAndExpect = func(method string) (*openapi.ResourceInfo, map[string]*ogen.PathItem) {
		resource := resourceGet.Clone()
		resource.Method = method

		operation := &ogen.Operation{
			Description: "",
			OperationID: "root.testresource",
			Tags:        []string{"Root"},
			Parameters:  []*ogen.Parameter{},
			Responses: ogen.Responses{
				"200": &ogen.Response{
					Description: "Successful response",
					Content: map[string]ogen.Media{
						"application/json": {
							Schema: ogen.NewSchema().AddRequiredProperties(
								openapi.RefSchema("TestStruct").ToProperty("data"),
							),
						},
					},
				},
			},
		}
		requestBody := &ogen.RequestBody{
			Content: map[string]ogen.Media{
				"application/json": {
					Schema: openapi.RefSchema("TestStruct"),
				},
			},
		}

		expected := map[string]*ogen.PathItem{
			"/testresource": {
				Description: "root.testresource",
			},
		}

		switch method {
		case "GET":
			expected["/testresource"].Get = operation
		case "HEAD":
			expected["/testresource"].Head = operation
		case "POST":
			expected["/testresource"].Post = operation
			expected["/testresource"].Post.RequestBody = requestBody
		case "PUT":
			expected["/testresource"].Put = operation
			expected["/testresource"].Put.RequestBody = requestBody
		case "DELETE":
			expected["/testresource"].Delete = operation
		case "TRACE":
			expected["/testresource"].Trace = operation
		case "PATCH":
			expected["/testresource"].Patch = operation
			expected["/testresource"].Patch.RequestBody = requestBody
		case "OPTIONS":
			expected["/testresource"].Options = operation
		}

		return resource, expected
	}

	for _, method := range []string{"GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "PATCH", "OPTIONS"} {
		t.Logf("Testing method %s", method)
		resource, expected := getResourceAndExpect(method)
		pathItem := utils.Must(oas.CreatePathItem(resource))
		assert.Equal(t, expected, pathItem)
	}
}

func TestCreatePathItemErrorMissingParam(t *testing.T) {
	config := &openapi.OpenAPISpecConfig{
		Resources: fs.NewResourcesManager(),
	}

	oas := utils.Must(openapi.NewSpec(config))
	resource := &openapi.ResourceInfo{
		ID:     "root.testresource",
		Path:   "/testresource/:param/missing",
		Method: "GET",
		Signatures: fs.Signatures{
			&fs.Signature{
				Type: &struct{ Name string }{},
				Name: "TestStruct",
			},
			&fs.Signature{
				Type: &struct{ Name string }{},
				Name: "TestStruct",
			},
		},
		Args:   fs.Args{},
		Public: true,
	}

	_, err := oas.CreatePathItem(resource)
	assert.Error(t, err)
}

func TestCreatePathItemErrorInvalidArgExample(t *testing.T) {
	config := &openapi.OpenAPISpecConfig{
		Resources: fs.NewResourcesManager(),
	}

	oas := utils.Must(openapi.NewSpec(config))
	resource := &openapi.ResourceInfo{
		ID:     "root.testresource",
		Path:   "/testresource",
		Method: "GET",
		Args: fs.Args{
			"arg": {
				Type:        fs.TypeInt,
				Required:    true,
				Description: "The arg",
				Example:     make(chan int),
			},
		},
		Public: true,
	}

	_, err := oas.CreatePathItem(resource)
	assert.Error(t, err)
}

func TestCreateResourcesForSchemas(t *testing.T) {
	resources := fs.NewResourcesManager()
	api := resources.Group("api")
	var createArg = func(t fs.ArgType, desc string) fs.Arg {
		return fs.Arg{Type: t, Required: true, Description: desc}
	}

	api.Group("content", &fs.Meta{
		Prefix: "/content/:schema",
		Args: fs.Args{
			"schema": {
				Required:    true,
				Type:        fs.TypeString,
				Description: "The schema name",
			},
		},
	}).
		Add(fs.NewResource("list", func(c fs.Context, _ any) (*contentservice.Pagination, error) {
			return nil, nil
		}, &fs.Meta{Get: "/"})).
		Add(fs.NewResource("detail", func(c fs.Context, _ any) (*entity.Entity, error) {
			return nil, nil
		}, &fs.Meta{
			Get:  "/:id",
			Args: fs.Args{"id": createArg(fs.TypeUint64, "The content ID")},
		})).
		Add(fs.NewResource("create", func(c fs.Context, _ any) (*entity.Entity, error) {
			return nil, nil
		}, &fs.Meta{Post: "/"})).
		Add(fs.NewResource("update", func(c fs.Context, _ any) (*entity.Entity, error) {
			return nil, nil
		}, &fs.Meta{
			Put:  "/:id",
			Args: fs.Args{"id": createArg(fs.TypeUint64, "The content ID")},
		})).
		Add(fs.NewResource("delete", func(c fs.Context, _ any) (any, error) {
			return nil, nil
		}, &fs.Meta{
			Delete: "/:id",
			Args:   fs.Args{"id": createArg(fs.TypeUint64, "The content ID")},
		}))

	schemaBuilder := utils.Must(schema.NewBuilderFromSchemas(
		t.TempDir(),
		map[string]*schema.Schema{
			"category": {
				Name:           "category",
				Namespace:      "Schema",
				LabelFieldName: "name",
				Fields: []*schema.Field{
					{Name: "name", Type: schema.TypeString},
				},
			},
		},
		fs.SystemSchemaTypes...,
	),
	)
	baseURL := "http://localhost:8080"
	config := &openapi.OpenAPISpecConfig{
		Resources:     resources,
		SchemaBuilder: schemaBuilder,
		BaseURL:       baseURL,
	}

	oas := utils.Must(openapi.NewSpec(config))
	assert.NotNil(t, oas)

	oas.CreateResourcesForSchemas()
	ogenUserSchema := oas.Schema("Schema.Category")
	assert.NotNil(t, ogenUserSchema)
}
