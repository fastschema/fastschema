package roleservice

import (
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
)

var ignoreContentSchemas = []string{
	"user",
	"role",
	"permission",
	"roles_users",
}

func (rs *RoleService) ResourcesList(c fs.Context, _ any) ([]*fs.Resource, error) {
	// Override the resources to remove the content resource
	// Add the content resource with the schemas
	resources := rs.Resources().Clone()

	apiResources := []*fs.Resource{}
	apiGroup := resources.Find("api")
	if apiGroup != nil {
		apiResources = apiGroup.Resources()
	}

	schemas := rs.DB().SchemaBuilder().Schemas()
	for _, r := range apiResources {
		if r.Name() == "content" || r.Name() == "realtime" {
			apiGroup.Remove(r)
		}
	}

	contentGroup := apiGroup.Group("content")
	realtimeContentGroup := apiGroup.Group("realtime").Group("content")
	for _, schema := range schemas {
		if utils.Contains(ignoreContentSchemas, schema.Name) || schema.IsJunctionSchema {
			continue
		}

		contentSchemaGroup := contentGroup.Group(schema.Name)
		contentSchemaGroup.AddResource("list", nil, &fs.Meta{
			Get: "/",
		})
		contentSchemaGroup.AddResource("detail", nil, &fs.Meta{
			Get: "/:id",
		})
		contentSchemaGroup.AddResource("create", nil, &fs.Meta{
			Post: "/",
		})
		contentSchemaGroup.AddResource("update", nil, &fs.Meta{
			Put: "/:id",
		})
		contentSchemaGroup.AddResource("delete", nil, &fs.Meta{
			Delete: "/:id",
		})
		contentSchemaGroup.AddResource("bulk-update", nil, &fs.Meta{
			Put: "/update",
		})
		contentSchemaGroup.AddResource("bulk-delete", nil, &fs.Meta{
			Delete: "/delete",
		})

		realtimeSchemaGroup := realtimeContentGroup.Group(schema.Name)
		realtimeSchemaGroup.AddResource("*", nil, &fs.Meta{})
		realtimeSchemaGroup.AddResource("create", nil, &fs.Meta{})
		realtimeSchemaGroup.AddResource("update", nil, &fs.Meta{})
		realtimeSchemaGroup.AddResource("delete", nil, &fs.Meta{})
	}

	return resources.Resources(), nil
}
