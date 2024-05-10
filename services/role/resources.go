package roleservice

import "github.com/fastschema/fastschema/app"

func (rs *RoleService) ResourcesList(c app.Context, _ any) ([]*app.Resource, error) {
	// Override the resources to remove the content resource
	// Add the content resource with the schemas
	resources := rs.Resources().Clone()
	schemas := rs.DB().SchemaBuilder().Schemas()
	for _, r := range resources.Resources() {
		if r.Name() == "content" {
			resources.Remove(r)
		}
	}

	contentGroup := resources.Group("content")
	for _, schema := range schemas {
		if schema.IsSystemSchema {
			continue
		}

		schemaGroup := contentGroup.Group(schema.Name)
		schemaGroup.AddResource("list", nil, &app.Meta{
			Get: "/",
		})
		schemaGroup.AddResource("detail", nil, &app.Meta{
			Get: "/:id",
		})
		schemaGroup.AddResource("create", nil, &app.Meta{
			Post: "/",
		})
		schemaGroup.AddResource("update", nil, &app.Meta{
			Put: "/:id",
		})
		schemaGroup.AddResource("delete", nil, &app.Meta{
			Delete: "/:id",
		})
	}

	return resources.Resources(), nil
}
