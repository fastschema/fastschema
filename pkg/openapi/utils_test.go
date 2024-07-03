package openapi_test

import (
	"sort"
	"testing"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/openapi"
	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
)

func TestCreateParameters(t *testing.T) {
	args := fs.Args{
		"param1": {
			Description: "Parameter 1",
			Required:    true,
			Type:        fs.TypeString,
			Example:     "example1",
		},
		"param2": {
			Description: "Parameter 2",
			Required:    false,
			Type:        fs.TypeInt32,
			Example:     "123",
		},
	}

	params := []string{"param1"}

	expected := []*ogen.Parameter{
		{
			Name:        "param1",
			In:          "path",
			Description: "Parameter 1",
			Required:    true,
			Schema: &ogen.Schema{
				Type:   "string",
				Format: "string",
			},
			Examples: map[string]*ogen.Example{
				"param1": {
					Summary: "param1 example",
					Value:   []byte(`"example1"`),
				},
			},
		},
		{
			Name:        "param2",
			In:          "query",
			Description: "Parameter 2",
			Required:    false,
			Schema: &ogen.Schema{
				Type:   "integer",
				Format: "int32",
			},
			Examples: map[string]*ogen.Example{
				"param2": {
					Summary: "param2 example",
					Value:   []byte(`"123"`),
				},
			},
		},
	}

	result, err := openapi.CreateParameters(args, params)
	assert.NoError(t, err)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	assert.Equal(t, expected, result)
}

func TestMergeParameters(t *testing.T) {
	first := []*ogen.Parameter{
		{
			Name:        "param1",
			In:          "path",
			Description: "Parameter 1",
			Required:    true,
			Schema: &ogen.Schema{
				Type:   "string",
				Format: "string",
			},
			Examples: map[string]*ogen.Example{
				"param1": {
					Summary: "param1 example",
					Value:   []byte(`"example1"`),
				},
			},
		},
	}

	second := []*ogen.Parameter{
		{
			Name:        "param2",
			In:          "query",
			Description: "Parameter 2",
			Required:    false,
			Schema: &ogen.Schema{
				Type:   "integer",
				Format: "int32",
			},
			Examples: map[string]*ogen.Example{
				"param2": {
					Summary: "param2 example",
					Value:   []byte(`"123"`),
				},
			},
		},
	}

	expected := []*ogen.Parameter{
		{
			Name:        "param1",
			In:          "path",
			Description: "Parameter 1",
			Required:    true,
			Schema: &ogen.Schema{
				Type:   "string",
				Format: "string",
			},
			Examples: map[string]*ogen.Example{
				"param1": {
					Summary: "param1 example",
					Value:   []byte(`"example1"`),
				},
			},
		},
		{
			Name:        "param2",
			In:          "query",
			Description: "Parameter 2",
			Required:    false,
			Schema: &ogen.Schema{
				Type:   "integer",
				Format: "int32",
			},
			Examples: map[string]*ogen.Example{
				"param2": {
					Summary: "param2 example",
					Value:   []byte(`"123"`),
				},
			},
		},
	}

	result := openapi.MergeParameters(first, second)
	assert.Equal(t, expected, result)
}
func TestMergeArgs(t *testing.T) {
	first := fs.Args{
		"param1": {
			Description: "Parameter 1",
			Required:    true,
			Type:        fs.TypeString,
			Example:     "example1",
		},
	}

	second := fs.Args{
		"param2": {
			Description: "Parameter 2",
			Required:    false,
			Type:        fs.TypeInt32,
			Example:     "123",
		},
	}

	expected := fs.Args{
		"param1": {
			Description: "Parameter 1",
			Required:    true,
			Type:        fs.TypeString,
			Example:     "example1",
		},
		"param2": {
			Description: "Parameter 2",
			Required:    false,
			Type:        fs.TypeInt32,
			Example:     "123",
		},
	}

	result, err := openapi.MergeArgs(first, second)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestMergeArgsDuplicateKey(t *testing.T) {
	first := fs.Args{
		"param1": {
			Description: "Parameter 1",
			Required:    true,
			Type:        fs.TypeString,
			Example:     "example1",
		},
	}

	second := fs.Args{
		"param1": {
			Description: "Parameter 2",
			Required:    false,
			Type:        fs.TypeInt32,
			Example:     "123",
		},
	}

	_, err := openapi.MergeArgs(first, second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key param1 in args")
}
func TestMergePathItems(t *testing.T) {
	op := &ogen.Operation{}
	first := map[string]*ogen.PathItem{
		"/path1": {
			Get:     op,
			Put:     op,
			Post:    op,
			Delete:  op,
			Options: op,
			Head:    op,
			Patch:   op,
			Trace:   op,
		},
	}

	second := map[string]*ogen.PathItem{
		"/path1": {
			Get:     op,
			Put:     op,
			Post:    op,
			Delete:  op,
			Options: op,
			Head:    op,
			Patch:   op,
			Trace:   op,
		},
		"/path2": {
			Post: op,
		},
	}

	expected := map[string]*ogen.PathItem{
		"/path1": {
			Get:     op,
			Put:     op,
			Post:    op,
			Delete:  op,
			Options: op,
			Head:    op,
			Patch:   op,
			Trace:   op,
		},
		"/path2": {
			Post: op,
		},
	}

	result := openapi.MergePathItems(first, second)
	assert.Equal(t, expected, result)
}

func TestJoinPaths(t *testing.T) {
	// Test case 1
	paths1 := []string{"", "/api/", "v1", "users"}
	expected1 := "/api/v1/users"
	result1 := openapi.JoinPaths(paths1...)
	assert.Equal(t, expected1, result1)

	// Test case 2
	paths2 := []string{"", "/api", "", "/v1", "users"}
	expected2 := "/api/v1/users"
	result2 := openapi.JoinPaths(paths2...)
	assert.Equal(t, expected2, result2)

	// Test case 3
	paths3 := []string{"", "/api", "v1", ""}
	expected3 := "/api/v1"
	result3 := openapi.JoinPaths(paths3...)
	assert.Equal(t, expected3, result3)

	// Test case 4
	paths4 := []string{"/api", "v1", "users"}
	expected4 := "/api/v1/users"
	result4 := openapi.JoinPaths(paths4...)
	assert.Equal(t, expected4, result4)

	// Test case 5
	paths5 := []string{"/api", "", "v1", "users"}
	expected5 := "/api/v1/users"
	result5 := openapi.JoinPaths(paths5...)
	assert.Equal(t, expected5, result5)

	// Test case 6
	paths6 := []string{"/api", "v1", ""}
	expected6 := "/api/v1"
	result6 := openapi.JoinPaths(paths6...)
	assert.Equal(t, expected6, result6)
}

func TestNormalizePath(t *testing.T) {
	// Test case 1
	path := "/path/:id/other"
	expectedPath := "/path/{id}/other"
	expectedParams := []string{"id"}

	result1, params1 := openapi.NormalizePath(path)
	assert.Equal(t, expectedPath, result1)
	assert.Equal(t, expectedParams, params1)

	// Test case 2
	path = "/path/:id/:name"
	expectedPath = "/path/{id}/{name}"
	expectedParams = []string{"id", "name"}

	result2, params2 := openapi.NormalizePath(path)
	assert.Equal(t, expectedPath, result2)
	assert.Equal(t, expectedParams, params2)

	// Test case 3
	path = "/path"
	expectedPath = "/path"
	expectedParams = ([]string)(nil)

	result3, params3 := openapi.NormalizePath(path)
	assert.Equal(t, expectedPath, result3)
	assert.Equal(t, expectedParams, params3)
}

func TestExtractPathParams(t *testing.T) {
	// Test case 1
	s1 := "/users/:id/posts/:postID"
	expected1 := []string{"id", "postID"}
	result1 := openapi.ExtractPathParams(s1)
	assert.Equal(t, expected1, result1)

	// Test case 2
	s2 := "/products/:productID"
	expected2 := []string{"productID"}
	result2 := openapi.ExtractPathParams(s2)
	assert.Equal(t, expected2, result2)

	// Test case 3
	s3 := "/categories"
	expected3 := ([]string)(nil)
	result3 := openapi.ExtractPathParams(s3)
	assert.Equal(t, expected3, result3)
}

func TestFlattenResources(t *testing.T) {
	root := fs.NewResourcesManager().Group("root")
	r1 := fs.NewResource("r1", func(c fs.Context, _ any) (any, error) {
		return nil, nil
	}, &fs.Meta{
		Prefix: "/groupr1",

		Get:     "/get",
		Head:    "/head",
		Post:    "/post",
		Put:     "/put",
		Delete:  "/delete",
		Connect: "/connect",
		Options: "/options",
		Trace:   "/trace",
		Patch:   "/patch",
		WS:      "/ws",
	})

	r2 := fs.NewResource("r2", func(c fs.Context, _ any) (any, error) {
		return nil, nil
	})

	r3 := root.Group("r3")
	r3.Add(fs.NewResource("sub1", func(c fs.Context, _ any) (any, error) {
		return nil, nil
	}))

	r4 := r3.Group("sub2", &fs.Meta{
		Prefix: "/groupr4",
	})
	r4.Add(fs.NewResource("sub3", func(c fs.Context, _ any) (any, error) {
		return nil, nil
	}))

	root.Add(r1)
	root.Add(r2)

	expected := []*openapi.ResourceInfo{
		{
			ID:         "root.r3.sub1",
			Path:       "/r3/sub1",
			Method:     "GET",
			Signatures: []any{nil, nil},
			Args:       fs.Args{},
			Public:     false,
		},
		{
			ID:         "root.r3.sub2.sub3",
			Path:       "/r3/groupr4/sub3",
			Method:     "GET",
			Signatures: []any{nil, nil},
			Args:       fs.Args{},
			Public:     false,
		},
		{
			ID:         "root.r1",
			Path:       "/get",
			Method:     "GET",
			Signatures: []any{nil, nil},
			Args:       fs.Args{},
			Public:     false,
		},
		{
			ID:         "root.r1",
			Path:       "/head",
			Method:     "HEAD",
			Signatures: []any{nil, nil},
			Args:       fs.Args{},
			Public:     false,
		},
		{
			ID:         "root.r1",
			Path:       "/post",
			Method:     "POST",
			Signatures: []any{nil, nil},
			Args:       fs.Args{},
			Public:     false,
		},
		{
			ID:         "root.r1",
			Path:       "/put",
			Method:     "PUT",
			Signatures: []any{nil, nil},
			Args:       fs.Args{},
			Public:     false,
		},
		{
			ID:         "root.r1",
			Path:       "/delete",
			Method:     "DELETE",
			Signatures: []any{nil, nil},
			Args:       fs.Args{},
			Public:     false,
		},
		{
			ID:         "root.r1",
			Path:       "/connect",
			Method:     "CONNECT",
			Signatures: []any{nil, nil},
			Args:       fs.Args{},
			Public:     false,
		},
		{
			ID:         "root.r1",
			Path:       "/options",
			Method:     "OPTIONS",
			Signatures: []any{nil, nil},
			Args:       fs.Args{},
			Public:     false,
		},
		{
			ID:         "root.r1",
			Path:       "/trace",
			Method:     "TRACE",
			Signatures: []any{nil, nil},
			Args:       fs.Args{},
			Public:     false,
		},
		{
			ID:         "root.r1",
			Path:       "/patch",
			Method:     "PATCH",
			Signatures: []any{nil, nil},
			Args:       fs.Args{},
			Public:     false,
		},
		{
			ID:         "root.r1",
			Path:       "/ws",
			Method:     "GET",
			Signatures: []any{nil, nil},
			Args:       fs.Args{},
			Public:     false,
		},
		{
			ID:         "root.r2",
			Path:       "/r2",
			Method:     "GET",
			Signatures: []any{nil, nil},
			Args:       fs.Args{},
			Public:     false,
		},
	}

	result, err := openapi.FlattenResources(root.Resources(), "", fs.Args{})
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestRefSchema(t *testing.T) {
	name := "testSchema"
	expected := &ogen.Schema{Ref: "#/components/schemas/testSchema"}
	result := openapi.RefSchema(name)
	assert.Equal(t, expected, result)
}
