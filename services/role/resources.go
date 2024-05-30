package roleservice

import "github.com/fastschema/fastschema/fs"

func (rs *RoleService) ResourcesList(c fs.Context, _ any) ([]*fs.Resource, error) {
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
		schemaGroup.AddResource("list", nil, &fs.Meta{
			Get: "/",
		})
		schemaGroup.AddResource("detail", nil, &fs.Meta{
			Get: "/:id",
		})
		schemaGroup.AddResource("create", nil, &fs.Meta{
			Post: "/",
		})
		schemaGroup.AddResource("update", nil, &fs.Meta{
			Put: "/:id",
		})
		schemaGroup.AddResource("delete", nil, &fs.Meta{
			Delete: "/:id",
		})
	}

	return resources.Resources(), nil
}
