package schemaservice

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

func (ss *SchemaService) Download(c fs.Context, _ any) (any, error) {
	s, err := ss.app.SchemaBuilder().Schema(c.Arg("name"))
	if err != nil {
		return nil, errors.NotFound(err.Error())
	}

	schemaFile := fmt.Sprintf("%s/%s.json", ss.app.SchemaBuilder().Dir(), s.Name)

	openFile, err := os.Open(schemaFile)
	defer func() {
		_ = openFile.Close()
	}()

	if err != nil {
		return nil, err
	}

	tempBuffer := make([]byte, 512)    // Create a byte array to read the file later
	_, err = openFile.Read(tempBuffer) // Read the file into  byte
	if err != nil {
		return nil, err
	}
	fileContentType := http.DetectContentType(tempBuffer) // Get file header

	fileStat, _ := openFile.Stat()                     // Get info from file
	fileSize := strconv.FormatInt(fileStat.Size(), 10) // Get file size as a string

	filename := fmt.Sprintf("%s.json", s.Name)

	// Set the headers
	header := make(http.Header)
	header.Set("Content-Type", fileContentType+";"+filename)
	header.Set("Content-Length", fileSize)

	_, err = openFile.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	return &fs.HTTPResponse{
		StatusCode: http.StatusOK,
		Header:     header,
		File:       schemaFile,
	}, nil
}
