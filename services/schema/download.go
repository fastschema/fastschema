package schemaservice

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

func (ss *SchemaService) Download(c fs.Context, _ any) (any, error) {
	// Get the schemas
	// split by name with comma
	schemaNames := strings.Split(c.Arg("names"), ",")

	randomTpmSchemaDir := utils.RandomString(16)
	tmpDir := fmt.Sprintf("%s/%s", ss.app.Disk().Root(), randomTpmSchemaDir)
	err := os.Mkdir(tmpDir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	// defer os.RemoveAll(tmpDir)
	archiveFilePath := fmt.Sprintf("%s/schemas.zip", tmpDir)
	archive, err := os.Create(archiveFilePath)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(schemaNames); i++ {
		s, err := ss.app.SchemaBuilder().Schema(schemaNames[i])
		if err != nil {
			return nil, errors.NotFound(err.Error())
		}

		schemaFile := fmt.Sprintf("%s/%s.json", ss.app.SchemaBuilder().Dir(), s.Name)

		zipWriter := zip.NewWriter(archive)
		openFile, err := os.Open(schemaFile)
		if err != nil {
			return nil, err
		}

		w1, err := zipWriter.Create(fmt.Sprintf("%s.json", s.Name))
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(w1, openFile); err != nil {
			return nil, err
		}

		openFile.Close()
		zipWriter.Close()
	}

	archive.Close()

	filename := "schemas.zip"

	// Set the headers
	header := make(http.Header)
	header.Set("Content-Type", "application/zip")
	header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	return &fs.HTTPResponse{
		StatusCode: http.StatusOK,
		Header:     header,
		File:       archiveFilePath,
	}, nil
}
