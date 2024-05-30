package file_test

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func createTestImage(t *testing.T) string {
	tmpFilePath := t.TempDir() + "/image.png"
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	c := color.RGBA{255, 255, 255, 255}
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, c)
		}
	}

	f, err := os.Create(tmpFilePath)
	assert.NoError(t, err)
	defer f.Close()

	assert.NoError(t, png.Encode(f, img))
	return tmpFilePath
}

type uploadResponse struct {
	Data map[string][]*fs.File `json:"data"`
}

func createFileBody(t *testing.T) (*multipart.Writer, *bytes.Buffer) {
	filePath := createTestImage(t)
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	file, err := os.Open(filePath)
	assert.NoError(t, err)

	w, err := mw.CreateFormFile("field", filePath)
	assert.NoError(t, err)
	_, err = io.Copy(w, file)
	assert.NoError(t, err)
	return mw, body
}

func TestFileServiceUpload(t *testing.T) {
	ms, server := createFileService(t)

	// Case 1: Error due to contraint violation
	mw, body := createFileBody(t)
	mw.Close()
	req := httptest.NewRequest("POST", "/file/upload", body)
	req.Header.Add("Content-Type", mw.FormDataContentType())
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)

	fileUploadResponse := uploadResponse{}
	response := utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NoError(t, json.Unmarshal([]byte(response), &fileUploadResponse))
	assert.Len(t, fileUploadResponse.Data["error"], 1)

	// Case 2: Error due to invalid file
	userModel := utils.Must(ms.DB().Model("user"))
	assert.True(t, utils.Must(userModel.CreateFromJSON(context.Background(), `{"username": "test", "password": "123"}`)) > 0)

	req = httptest.NewRequest("POST", "/file/upload", body)
	req.Header.Add("Content-Type", mw.FormDataContentType())
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)

	// Case 3: Success
	mw, body = createFileBody(t)
	mw.Close()
	req = httptest.NewRequest("POST", "/file/upload", body)
	req.Header.Add("Content-Type", mw.FormDataContentType())
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))
	assert.NoError(t, json.Unmarshal([]byte(response), &fileUploadResponse))
	assert.Len(t, fileUploadResponse.Data["success"], 1)
}
