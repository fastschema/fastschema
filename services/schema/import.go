package schemaservice

import (
	"fmt"
	"os"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

func (ss *SchemaService) Import(c fs.Context, _ any) (fs.Map, error) {
	// upload to tmp dir
	files, err := c.Files()
	if err != nil {
		return nil, err
	}

	// check if total files size > 5MB
	totalSize := 0
	for _, file := range files {
		totalSize += int(file.Size)
	}
	if totalSize > 5*1024*1024 {
		return nil, errors.BadRequest("total files size should be less than 5MB")
	}

	// upload files to tmp dir
	randomTpmSchemaDir := utils.RandomString(16)
	tmpDir := fmt.Sprintf("%s/%s", ss.app.Disk().Root(), randomTpmSchemaDir)
	defer os.RemoveAll(tmpDir)
	for _, file := range files {
		filePath := fmt.Sprintf("%s/%s", randomTpmSchemaDir, file.Name)
		file.Path = filePath
		if _, err := ss.app.Disk().Put(c.Context(), file); err != nil {
			c.Logger().Errorf("Error uploading file: %s", err)
			return nil, err
		}
	}

	// get all schemas from directory
	schemas, err := schema.GetSchemasFromDir(fmt.Sprintf("%s/", tmpDir))
	if err != nil {
		return nil, err
	}

	// validate one by one schema with the following rules
	for _, sc := range schemas {
		currentSchemaFile := fmt.Sprintf("%s/%s.json", ss.app.SchemaBuilder().Dir(), sc.Name)
		if utils.IsFileExists(currentSchemaFile) {
			return nil, errors.BadRequest("schema already exists in current system")
		}
	}

	_, err = schema.NewBuilderFromDir(tmpDir)
	if err != nil {
		return nil, err
	}

	for _, sc := range schemas {
		schemaFile := fmt.Sprintf("%s/%s.json", ss.app.SchemaBuilder().Dir(), sc.Name)
		if err := sc.SaveToFile(schemaFile); err != nil {
			return nil, errors.InternalServerError("could not save schema")
		}
	}

	if err := ss.app.Reload(c.Context(), nil); err != nil {
		c.Logger().Errorf("could not reload app: %s", err.Error())
		return nil, errors.InternalServerError("could not reload app: %s", err.Error())
	}

	return fs.Map{"message": "Schema imported"}, nil
}
