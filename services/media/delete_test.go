package mediaservice_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestMediaServiceDeleteErrorNotFound(t *testing.T) {
	ms, server := createMediaService(t)
	mediaModel := utils.Must(ms.DB().Model("media"))
	mediaID := utils.Must(mediaModel.CreateFromJSON(`{
		"disk": "local_test",
		"path": "some/path/test.txt",
		"name": "test.txt",
		"size": 1,
		"type": "text/plain"
	}`))

	// Case 1: success
	req := httptest.NewRequest("DELETE", "/media", bytes.NewReader([]byte(fmt.Sprintf(`[%d]`, mediaID))))
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 500, resp.StatusCode)
}

func TestMediaServiceDelete(t *testing.T) {
	ms, server := createMediaService(t)
	mediaModel := utils.Must(ms.DB().Model("media"))
	mediaID := utils.Must(mediaModel.CreateFromJSON(`{
		"disk": "local_test",
		"path": "some/path/test.txt",
		"name": "test.txt",
		"size": 1,
		"type": "text/plain"
	}`))

	testFile := &app.File{
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
	req := httptest.NewRequest("DELETE", "/media", bytes.NewReader([]byte(fmt.Sprintf(`[%d]`, mediaID))))
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
}
