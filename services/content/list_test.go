package contentservice_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	contentservice "github.com/fastschema/fastschema/services/content"
	"github.com/stretchr/testify/assert"
)

type TestListItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TestPagination struct {
	Total       uint            `json:"total"`
	PerPage     uint            `json:"per_page"`
	CurrentPage uint            `json:"current_page"`
	LastPage    uint            `json:"last_page"`
	Items       []*TestListItem `json:"items"`
}

type TestResponse struct {
	Data *TestPagination `json:"data"`
}

func TestNewPagination(t *testing.T) {
	total := uint(100)
	perPage := uint(10)
	currentPage := uint(1)
	data := []*entity.Entity{
		entity.New(1),
		entity.New(2),
		entity.New(3),
		entity.New(4),
		entity.New(5),
	}

	pagination := contentservice.NewPagination(total, perPage, currentPage, data)

	assert.NotNil(t, pagination)
	assert.Equal(t, total, pagination.Total)
	assert.Equal(t, perPage, pagination.PerPage)
	assert.Equal(t, currentPage, pagination.CurrentPage)
	assert.Equal(t, uint(math.Ceil(float64(total)/float64(perPage))), pagination.LastPage)
	assert.Equal(t, data, pagination.Items)
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
		utils.Must(blogModel.CreateFromJSON(context.Background(), fmt.Sprintf(`{"name": "test blog %d"}`, i+1)))
	}

	// create 10 blog users
	userModel := utils.Must(cs.DB().Model("user"))
	for i := 0; i < 10; i++ {
		utils.Must(userModel.CreateFromJSON(context.Background(), fmt.Sprintf(`{"username": "user%d", "password": "123", "provider": "local"}`, i+1)))
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
	assert.Equal(t, uint(10), paginatedData.Total)
	assert.Equal(t, uint(3), paginatedData.PerPage)
	assert.Equal(t, uint(1), paginatedData.CurrentPage)
	assert.Equal(t, uint(4), paginatedData.LastPage)
	assert.Len(t, paginatedData.Items, 3)
	assert.Equal(t, "test blog 10", paginatedData.Items[0].Name)
	assert.Equal(t, 10, paginatedData.Items[0].ID)

	// Case 4: list success with filter id less than 5
	req = httptest.NewRequest("GET", `/content/blog?filter={"id":{"$lt":5}}&sort=id`, nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))

	assert.NoError(t, json.Unmarshal([]byte(response), &data))
	paginatedData = data.Data
	assert.Equal(t, uint(4), paginatedData.Total)
	assert.Equal(t, uint(10), paginatedData.PerPage)
	assert.Equal(t, uint(1), paginatedData.CurrentPage)
	assert.Equal(t, uint(1), paginatedData.LastPage)
	assert.Len(t, paginatedData.Items, 4)
	assert.Equal(t, "test blog 1", paginatedData.Items[0].Name)
	assert.Equal(t, 1, paginatedData.Items[0].ID)

	// Case 5: list user with default select fields
	req = httptest.NewRequest("GET", "/content/user", nil)
	resp = utils.Must(server.Test(req))
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 200, resp.StatusCode)
	response = utils.Must(utils.ReadCloserToString(resp.Body))

	assert.NoError(t, json.Unmarshal([]byte(response), &data))
	paginatedData = data.Data
	assert.Equal(t, uint(10), paginatedData.Total)
	assert.Equal(t, uint(10), paginatedData.PerPage)
	assert.Equal(t, uint(1), paginatedData.CurrentPage)
	assert.Equal(t, uint(1), paginatedData.LastPage)
	assert.Contains(t, response, `"id":`)
	assert.Contains(t, response, `"username":`)
	assert.Contains(t, response, `"roles":`)
	assert.Contains(t, response, `"created_at":`)
}
