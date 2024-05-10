package openapi_test

import (
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/openapi"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	contentservice "github.com/fastschema/fastschema/services/content"
	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
)

func handler[T any](c app.Context, input T) (T, error) {
	return input, nil
}

func meta(name string) *app.Meta {
	return &app.Meta{Post: "/" + name}
}

func TestResourceInfoClone(t *testing.T) {
	r := &openapi.ResourceInfo{
		ID:         "testresource",
		Path:       "/testresource",
		Method:     "GET",
		Signatures: app.Signatures{testStruct{}, testStruct{}},
		Args:       app.Args{},
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
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir()))
	resources := app.NewResourcesManager()
	resources.Add(
		app.NewResource("bool", handler[bool], meta("bool")),
	)

	config := &openapi.OpenAPISpecConfig{
		Resources:     resources,
		SchemaBuilder: sb,
		BaseURL:       "http://localhost:8080",
	}

	oas, err := openapi.NewSpec(config)
	assert.NoError(t, err)
	assert.NotNil(t, oas)
	assert.NoError(t, oas.Create())
	assert.NotNil(t, oas.Spec())
}

func TestNewSpecError(t *testing.T) {
	oas, err := openapi.NewSpec(nil)
	assert.Error(t, err)
	assert.Nil(t, oas)
}

func TestCreatePathItem(t *testing.T) {
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir()))
	resources := app.NewResourcesManager()
	resources.Add(
		app.NewResource("bool", handler[bool], meta("bool")),
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
		Signatures: app.Signatures{
			&app.Signature{
				Type: &struct{ Name string }{},
				Name: "TestStruct",
			},
			&app.Signature{
				Type: &struct{ Name string }{},
				Name: "TestStruct",
			},
		},
		Args:   app.Args{},
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
							Schema: openapi.RefSchema("TestStruct"),
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
		Resources: app.NewResourcesManager(),
	}

	oas := utils.Must(openapi.NewSpec(config))
	resource := &openapi.ResourceInfo{
		ID:     "root.testresource",
		Path:   "/testresource/:param/missing",
		Method: "GET",
		Signatures: app.Signatures{
			&app.Signature{
				Type: &struct{ Name string }{},
				Name: "TestStruct",
			},
			&app.Signature{
				Type: &struct{ Name string }{},
				Name: "TestStruct",
			},
		},
		Args:   app.Args{},
		Public: true,
	}

	_, err := oas.CreatePathItem(resource)
	assert.Error(t, err)
}

func TestCreateResourcesForSchemas(t *testing.T) {
	resources := app.NewResourcesManager()
	api := resources.Group("api")
	var createArg = func(t app.ArgType, desc string) app.Arg {
		return app.Arg{Type: t, Required: true, Description: desc}
	}

	api.Group("content", &app.Meta{
		Prefix: "/content/:schema",
		Args: app.Args{
			"schema": {
				Required:    true,
				Type:        app.TypeString,
				Description: "The schema name",
			},
		},
	}).
		Add(app.NewResource("list", func(c app.Context, _ any) (*contentservice.Pagination, error) {
			return nil, nil
		}, &app.Meta{Get: "/"})).
		Add(app.NewResource("detail", func(c app.Context, _ any) (*schema.Entity, error) {
			return nil, nil
		}, &app.Meta{
			Get:  "/:id",
			Args: app.Args{"id": createArg(app.TypeUint64, "The content ID")},
		})).
		Add(app.NewResource("create", func(c app.Context, _ any) (*schema.Entity, error) {
			return nil, nil
		}, &app.Meta{Post: "/"})).
		Add(app.NewResource("update", func(c app.Context, _ any) (*schema.Entity, error) {
			return nil, nil
		}, &app.Meta{
			Put:  "/:id",
			Args: app.Args{"id": createArg(app.TypeUint64, "The content ID")},
		})).
		Add(app.NewResource("delete", func(c app.Context, _ any) (any, error) {
			return nil, nil
		}, &app.Meta{
			Delete: "/:id",
			Args:   app.Args{"id": createArg(app.TypeUint64, "The content ID")},
		}))

	schemaBuilder := utils.Must(schema.NewBuilderFromSchemas(t.TempDir(), map[string]*schema.Schema{
		"category": {
			Name:           "category",
			Namespace:      "Schema",
			LabelFieldName: "name",
			Fields: []*schema.Field{
				{Name: "name", Type: schema.TypeString},
			},
		},
	}))
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
