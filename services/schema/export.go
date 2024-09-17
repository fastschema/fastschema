package schemaservice

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

type SchemasExport struct {
	Schemas *[]string `json:"schemas"`
}

func (ss *SchemaService) Export(c fs.Context, schemasExport *SchemasExport) (any, error) {
	if len(*schemasExport.Schemas) == 0 {
		return nil, errors.BadRequest("schemas is required")
	}

	schemasIsNotExist := make([]string, 0)
	for _, sc := range *schemasExport.Schemas {
		currentSchemaFile := fmt.Sprintf("%s/%s.json", ss.app.SchemaBuilder().Dir(), sc)
		if !utils.IsFileExists(currentSchemaFile) {
			schemasIsNotExist = append(schemasIsNotExist, sc)
		}
	}
	if len(schemasIsNotExist) > 0 {
		return nil, errors.NotFound("schemas %s is not exist", strings.Join(schemasIsNotExist, ", "))
	}

	// Create a buffer to write our archive to
	buffer := new(bytes.Buffer)
	// Create a new zip archive
	zipWriter := zip.NewWriter(buffer)
	for _, schema := range *schemasExport.Schemas {
		schemaFile := fmt.Sprintf("%s/%s.json", ss.app.SchemaBuilder().Dir(), schema)
		// Read the file content
		data, err := os.ReadFile(schemaFile)
		if err != nil {
			return nil, err
		}

		// Create a zip entry for the file
		zipFile, err := zipWriter.Create(fmt.Sprintf("%s.json", schema))
		if err != nil {
			return nil, err
		}

		// Write the file content to the zip entry
		_, err = zipFile.Write(data)
		if err != nil {
			return nil, err
		}
	}
	// Make sure to check the error on Close
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	// Set the headers
	header := make(http.Header)
	header.Set("Content-Type", "application/zip")
	header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", "schemas.zip"))

	return &fs.HTTPResponse{
		StatusCode: http.StatusOK,
		Header:     header,
		Stream:     buffer,
	}, nil
}
