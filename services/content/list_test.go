package contentservice_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

type TestListItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TestPagination struct {
	Pagination *app.PaginationInfo `json:"pagination"`
	Data       []*TestListItem     `json:"data"`
}

type TestResponse struct {
	Data *TestPagination `json:"data"`
}

func TestContentServiceList(t *testing.T) {
	cs, server := createContentService(t)

	// Case 1: schema not found
	req := httptest.NewRequest("GET", "/content/test", nil)
	resp := utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"message":"model test not found"`)

	// create 10 blog posts
	blogModel := utils.Must(cs.DB().Model("blog"))
	for i := 0; i < 10; i++ {
		utils.Must(blogModel.CreateFromJSON(fmt.Sprintf(`{"name": "test blog %d"}`, i+1)))
	}

	// create 10 blog users
	userModel := utils.Must(cs.DB().Model("user"))
	for i := 0; i < 10; i++ {
		utils.Must(userModel.CreateFromJSON(fmt.Sprintf(`{"username": "user%d", "password": "123"}`, i+1)))
	}

	// Case 2: invalid predicate
	req = httptest.NewRequest("GET", "/content/blog?filter=invalid", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 400, resp.StatusCode)

	// Case 3: list success with limit, sort, select
	req = httptest.NewRequest("GET", "/content/blog?limit=3&sort=-id&select=name", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response := utils.Must(utils.ReadCloserToString(resp.Body))

	var data TestResponse
	assert.NoError(t, json.Unmarshal([]byte(response), &data))
	paginatedData := data.Data
	assert.Equal(t, uint(10), paginatedData.Pagination.Total)
	assert.Equal(t, uint(3), paginatedData.Pagination.PerPage)
	assert.Equal(t, uint(1), paginatedData.Pagination.CurrentPage)
	assert.Equal(t, uint(4), paginatedData.Pagination.LastPage)
	assert.Len(t, paginatedData.Data, 3)
	assert.Equal(t, "test blog 10", paginatedData.Data[0].Name)
	assert.Equal(t, 10, paginatedData.Data[0].ID)

	// Case 4: list success with filter id less than 5
	req = httptest.NewRequest("GET", `/content/blog?filter={"id":{"$lt":5}}&sort=id`, nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))

	assert.NoError(t, json.Unmarshal([]byte(response), &data))
	paginatedData = data.Data
	assert.Equal(t, uint(4), paginatedData.Pagination.Total)
	assert.Equal(t, uint(10), paginatedData.Pagination.PerPage)
	assert.Equal(t, uint(1), paginatedData.Pagination.CurrentPage)
	assert.Equal(t, uint(1), paginatedData.Pagination.LastPage)
	assert.Len(t, paginatedData.Data, 4)
	assert.Equal(t, "test blog 1", paginatedData.Data[0].Name)
	assert.Equal(t, 1, paginatedData.Data[0].ID)

	// Case 5: list user with default select fields
	req = httptest.NewRequest("GET", "/content/user", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))

	assert.NoError(t, json.Unmarshal([]byte(response), &data))
	paginatedData = data.Data
	assert.Equal(t, uint(10), paginatedData.Pagination.Total)
	assert.Equal(t, uint(10), paginatedData.Pagination.PerPage)
	assert.Equal(t, uint(1), paginatedData.Pagination.CurrentPage)
	assert.Equal(t, uint(1), paginatedData.Pagination.LastPage)
	assert.Contains(t, response, `"id":`)
	assert.Contains(t, response, `"username":`)
	assert.Contains(t, response, `"roles":`)
	assert.Contains(t, response, `"created_at":`)
}
