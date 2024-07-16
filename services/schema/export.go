package schemaservice

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

func (ss *SchemaService) Export(c fs.Context, _ any) (any, error) {
	// Get the schemas
	// split by name with comma
	schemaNames := strings.Split(c.Arg("names"), ",")

	// Create a buffer to write our archive to
	buffer := new(bytes.Buffer)

	// Create a new zip archive
	zipWriter := zip.NewWriter(buffer)

	for _, schema := range schemaNames {
		_, err := ss.app.SchemaBuilder().Schema(schema)
		if err != nil {
			return nil, errors.NotFound(err.Error())
		}
		schemaFile := fmt.Sprintf("%s/%s.json", ss.app.SchemaBuilder().Dir(), schema)
		// Read the file content
		data, err := ioutil.ReadFile(schemaFile)
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
