package schemaservice

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/fastschema/fastschema/fs"
)

func (ss *SchemaService) Export(c fs.Context, _ any) (any, error) {
	// // Get the schemas
	// // split by name with comma
	// schemaNames := strings.Split(c.Arg("names"), ",")

	// randomTpmSchemaDir := utils.RandomString(16)
	// tmpDir := fmt.Sprintf("%s/%s", ss.app.Disk().Root(), randomTpmSchemaDir)
	// err := os.Mkdir(tmpDir, os.ModePerm)
	// if err != nil {
	// 	return nil, err
	// }
	// // defer os.RemoveAll(tmpDir)
	// archiveFilePath := fmt.Sprintf("%s/schemas.zip", tmpDir)
	// archive, err := os.Create(archiveFilePath)
	// if err != nil {
	// 	return nil, err
	// }

	// for i := 0; i < len(schemaNames); i++ {
	// 	s, err := ss.app.SchemaBuilder().Schema(schemaNames[i])
	// 	if err != nil {
	// 		return nil, errors.NotFound(err.Error())
	// 	}

	// 	schemaFile := fmt.Sprintf("%s/%s.json", ss.app.SchemaBuilder().Dir(), s.Name)

	// 	zipWriter := zip.NewWriter(archive)
	// 	openFile, err := os.Open(schemaFile)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	w1, err := zipWriter.Create(fmt.Sprintf("%s.json", s.Name))
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if _, err := io.Copy(w1, openFile); err != nil {
	// 		return nil, err
	// 	}

	// 	openFile.Close()
	// 	zipWriter.Close()
	// }

	// archive.Close()

	// filename := "schemas.zip"

	// // Set the headers
	// header := make(http.Header)
	// header.Set("Content-Type", "application/zip")
	// header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// return &fs.HTTPResponse{
	// 	StatusCode: http.StatusOK,
	// 	Header:     header,
	// 	File:       archiveFilePath,
	// }, nil
	////////////////////////////////////////////////////////////////////////////////////////////////
	// List of JSON files to include in the ZIP

	schemaFile := fmt.Sprintf("%s/%s.json", ss.app.SchemaBuilder().Dir(), "tag")
	fmt.Println(schemaFile)
	files := []string{schemaFile} // Update with your actual file paths

	// Create a buffer to write our archive to
	buffer := new(bytes.Buffer)
	fmt.Println("lllll", buffer)

	// Create a new zip archive
	zipWriter := zip.NewWriter(buffer)

	for _, file := range files {
		// Read the file content
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		fmt.Println(data)

		// Create a zip entry for the file
		zipFile, err := zipWriter.Create("tag.json")
		if err != nil {
			return nil, err
		}

		// Write the file content to the zip entry
		_, err = zipFile.Write(data)
		if err != nil {
			return nil, err
		}
	}

	fmt.Println("hhhhh", buffer)

	// Make sure to check the error on Close
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	// Set the relevant headers
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=files.zip")

	// Write the buffer to the response
	if _, err := io.Copy(c.Response().BodyWriter(), buffer); err != nil {
		return nil, err
	}

	return nil, nil
}
