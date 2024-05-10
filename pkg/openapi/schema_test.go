package openapi_test

import (
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/openapi"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
)

func TestGetSchemaRefName(t *testing.T) {
	refName := openapi.GetSchemaRefName("user")
	expectedRefName := "Schema.User"
	assert.Equal(t, expectedRefName, refName)

	refNameWithTypes := openapi.GetSchemaRefName("user", "Create")
	expectedRefNameWithTypes := "Schema.User.Create"
	assert.Equal(t, expectedRefNameWithTypes, refNameWithTypes)
}

func TestContentListResponseSchema(t *testing.T) {
	s := &schema.Schema{
		Name: "test",
		Fields: []*schema.Field{
			{
				Name: "field1",
				Type: schema.TypeString,
			},
			{
				Name: "field2",
				Type: schema.TypeInt,
			},
		},
	}

	responseSchema := openapi.ContentListResponseSchema(s)

	assert.NotNil(t, responseSchema)
	assert.Contains(t, responseSchema.Required, "total")
	assert.Contains(t, responseSchema.Required, "per_page")
	assert.Contains(t, responseSchema.Required, "current_page")
	assert.Contains(t, responseSchema.Required, "last_page")

	itemsProperties := utils.Filter(responseSchema.Properties, func(p ogen.Property) bool {
		return p.Name == "items"
	})

	assert.Len(t, itemsProperties, 1)
	assert.Equal(t, "items", itemsProperties[0].Name)
	assert.Equal(t, "array", itemsProperties[0].Schema.Type)
}

func TestContentDetailSchema(t *testing.T) {
	s := &schema.Schema{
		Name:             "user",
		Namespace:        "schema",
		LabelFieldName:   "name",
		DisableTimestamp: false,
		DBColumns:        []string{"column1", "column2"},
		IsSystemSchema:   true,
		IsJunctionSchema: false,
		Fields: []*schema.Field{
			{
				Name:  "name",
				Type:  schema.TypeString,
				Label: "Name",
			},
			{
				Name:  "field2",
				Type:  schema.TypeString,
				Label: "label2",
			},
		},
	}

	expectedSchema := openapi.RefSchema("Schema.User")
	result := openapi.ContentDetailSchema(s)
	assert.Equal(t, expectedSchema, result)
}

func TestContentCreateSchema(t *testing.T) {
	s := &schema.Schema{
		Name:             "user",
		Namespace:        "schema",
		LabelFieldName:   "name",
		DisableTimestamp: false,
		DBColumns:        []string{"column1", "column2"},
		IsSystemSchema:   true,
		IsJunctionSchema: false,
		Fields: []*schema.Field{
			{
				Name:  "field1",
				Type:  schema.TypeString,
				Label: "label1",
			},
			{
				Name:  "field2",
				Type:  schema.TypeString,
				Label: "label2",
			},
		},
	}

	expectedSchema := ogen.NewSchema()
	expectedSchema.Ref = "#/components/schemas/Schema.User.Create"
	result := openapi.ContentCreateSchema(s)
	assert.Equal(t, expectedSchema, result)
}

func TestSchemasToOGenSchemas(t *testing.T) {
	resources := app.NewResourcesManager()
	resources.Add(app.NewResource("test", func(c app.Context, _ any) (any, error) {
		return nil, nil
	}))

	// Case 1: SchemaBuilder is nil
	oas := utils.Must(openapi.NewSpec(&openapi.OpenAPISpecConfig{
		Resources: resources,
	}))
	assert.NoError(t, oas.SchemasToOGenSchemas())

	// Case 2: SchemaBuilder is not nil
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir()))
	roleSchema := utils.Must(sb.Schema("role"))
	assert.NotNil(t, roleSchema)
	oas = utils.Must(openapi.NewSpec(&openapi.OpenAPISpecConfig{
		SchemaBuilder: sb,
		Resources:     resources,
	}))

	assert.NoError(t, oas.SchemasToOGenSchemas())
	ogenSchemaRole := oas.Schema("Schema.Role")
	ogenSchemaPermission := oas.Schema("Schema.Permission")
	assert.NotNil(t, ogenSchemaRole)
	assert.NotNil(t, oas.Schema("Schema.User"))
	assert.NotNil(t, ogenSchemaPermission)

	roleUsersProperty := utils.Filter(ogenSchemaRole.Properties, func(p ogen.Property) bool {
		return p.Name == "users"
	})
	assert.Len(t, roleUsersProperty, 1)
	assert.Equal(t, "array", roleUsersProperty[0].Schema.Type)

	rolePermissionsProperty := utils.Filter(ogenSchemaRole.Properties, func(p ogen.Property) bool {
		return p.Name == "permissions"
	})
	assert.Len(t, rolePermissionsProperty, 1)

	permissionRoleProperty := utils.Filter(ogenSchemaPermission.Properties, func(p ogen.Property) bool {
		return p.Name == "role"
	})
	assert.Len(t, permissionRoleProperty, 1)
	assert.Equal(t, "#/components/schemas/Schema.Role", permissionRoleProperty[0].Schema.Ref)
}
