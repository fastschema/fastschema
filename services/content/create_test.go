package contentservice_test

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestContentServiceCreate(t *testing.T) {
	_, server := createContentService(t)
	req := httptest.NewRequest("POST", "/content/blog", bytes.NewReader([]byte(`{"name": "test blog"}`)))
	resp, err := server.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	fmt.Println(utils.Must(utils.ReadCloserToString(resp.Body)))
}
