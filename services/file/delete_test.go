package file_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestFileServiceDeleteErrorNotFound(t *testing.T) {
	ms, server := createFileService(t)
	userModel := utils.Must(ms.DB().Model("user"))
	_ = utils.Must(userModel.CreateFromJSON(context.Background(), `{
		"username": "test",
		"password": "test",
		"provider": "local"
	}`))

	fileModel := utils.Must(ms.DB().Model("file"))
	fileID := utils.Must(fileModel.CreateFromJSON(context.Background(), `{
		"disk": "local_test",
		"path": "some/path/test.txt",
		"name": "test.txt",
		"size": 1,
		"type": "text/plain",
		"user_id": 1
	}`))

	// Case 1: success
	req := httptest.NewRequest("DELETE", "/file", bytes.NewReader([]byte(fmt.Sprintf(`[%d]`, fileID))))
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
}

func TestFileServiceDelete(t *testing.T) {
	ms, server := createFileService(t)
	userModel := utils.Must(ms.DB().Model("user"))
	_ = utils.Must(userModel.CreateFromJSON(context.Background(), `{
		"username": "test",
		"password": "test",
		"provider": "local"
	}`))
	fileModel := utils.Must(ms.DB().Model("file"))
	fileID := utils.Must(fileModel.CreateFromJSON(context.Background(), `{
		"disk": "local_test",
		"path": "some/path/test.txt",
		"name": "test.txt",
		"size": 1,
		"type": "text/plain",
		"user_id": 1
	}`))

	testFile := &fs.File{
		Disk:   "local_test",
		Path:   "some/path/test.txt",
		Name:   "test.txt",
		Size:   4,
		Type:   "text/plain",
		Reader: bytes.NewReader([]byte("test")),
	}

	result := utils.Must(ms.Disk().Put(context.Background(), testFile))
	assert.Equal(t, "http://localhost:3000/files/some/path/test.txt", result.URL)

	// Case 1: success
	req := httptest.NewRequest("DELETE", "/file", bytes.NewReader([]byte(fmt.Sprintf(`[%d]`, fileID))))
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
}
