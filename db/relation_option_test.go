package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRelationOptions(t *testing.T) {
	t.Run("empty string returns nil", func(t *testing.T) {
		opts, err := ParseRelationOptions("")
		require.NoError(t, err)
		assert.Nil(t, opts)
	})

	t.Run("valid options", func(t *testing.T) {
		jsonStr := `{
			"categories": {
				"limit": 5,
				"offset": 10,
				"sort": "name",
				"filter": {"status": "active"},
				"select": ["id", "name"]
			},
			"tags": {
				"limit": 3,
				"sort": "-created_at"
			}
		}`
		opts, err := ParseRelationOptions(jsonStr)
		require.NoError(t, err)
		require.NotNil(t, opts)

		// Check categories options
		catOpts := opts.Get("categories")
		require.NotNil(t, catOpts)
		assert.Equal(t, uint(5), catOpts.Limit)
		assert.Equal(t, uint(10), catOpts.Offset)
		assert.Equal(t, "name", catOpts.Sort)
		assert.Equal(t, map[string]any{"status": "active"}, catOpts.Filter)
		assert.Equal(t, []string{"id", "name"}, catOpts.Select)

		// Check tags options
		tagOpts := opts.Get("tags")
		require.NotNil(t, tagOpts)
		assert.Equal(t, uint(3), tagOpts.Limit)
		assert.Equal(t, "-created_at", tagOpts.Sort)
		assert.Equal(t, uint(0), tagOpts.Offset)
		assert.Nil(t, tagOpts.Filter)
		assert.Nil(t, tagOpts.Select)
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		_, err := ParseRelationOptions(`{"invalid`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid relation options format")
	})

	t.Run("non-existent field returns nil", func(t *testing.T) {
		opts, err := ParseRelationOptions(`{"tags": {"limit": 5}}`)
		require.NoError(t, err)
		assert.Nil(t, opts.Get("categories"))
	})
}

func TestRelationOptionClone(t *testing.T) {
	t.Run("clone nil returns nil", func(t *testing.T) {
		var opt *RelationOption
		assert.Nil(t, opt.Clone())
	})

	t.Run("clone with all fields", func(t *testing.T) {
		opt := &RelationOption{
			Limit:  10,
			Offset: 5,
			Sort:   "-name",
			Filter: map[string]any{"status": "active"},
			Select: []string{"id", "name"},
		}
		cloned := opt.Clone()
		require.NotNil(t, cloned)
		assert.Equal(t, opt.Limit, cloned.Limit)
		assert.Equal(t, opt.Offset, cloned.Offset)
		assert.Equal(t, opt.Sort, cloned.Sort)
		assert.Equal(t, opt.Filter, cloned.Filter)
		assert.Equal(t, opt.Select, cloned.Select)

		// Verify deep copy
		cloned.Filter["status"] = "inactive"
		assert.Equal(t, "active", opt.Filter["status"])
		cloned.Select[0] = "changed"
		assert.Equal(t, "id", opt.Select[0])
	})
}

func TestRelationOptionsClone(t *testing.T) {
	t.Run("clone nil returns nil", func(t *testing.T) {
		var opts RelationOptions
		assert.Nil(t, opts.Clone())
	})

	t.Run("clone with multiple options", func(t *testing.T) {
		opts := RelationOptions{
			"tags":       {Limit: 5, Sort: "name"},
			"categories": {Limit: 10, Offset: 2},
		}
		cloned := opts.Clone()
		require.NotNil(t, cloned)
		assert.Len(t, cloned, 2)
		assert.Equal(t, uint(5), cloned["tags"].Limit)
		assert.Equal(t, uint(10), cloned["categories"].Limit)

		// Verify deep copy
		cloned["tags"].Limit = 100
		assert.Equal(t, uint(5), opts["tags"].Limit)
	})
}

func TestRelationOptionsGetNestedOptions(t *testing.T) {
	t.Run("nil options returns nil", func(t *testing.T) {
		var opts RelationOptions
		assert.Nil(t, opts.GetNestedOptions("author"))
	})

	t.Run("no matching nested options returns nil", func(t *testing.T) {
		opts := RelationOptions{
			"tags": {Limit: 5},
		}
		assert.Nil(t, opts.GetNestedOptions("author"))
	})

	t.Run("get nested options", func(t *testing.T) {
		opts := RelationOptions{
			"author":                 {Limit: 1},
			"author.country":         {Limit: 1},
			"author.country.region":  {Limit: 5, Sort: "name"},
			"tags":                   {Limit: 10},
			"categories.subcategory": {Limit: 3},
		}

		// Get nested options for author
		authorNested := opts.GetNestedOptions("author")
		require.NotNil(t, authorNested)
		assert.Len(t, authorNested, 2)
		assert.NotNil(t, authorNested["country"])
		assert.NotNil(t, authorNested["country.region"])

		// Get nested options for author.country
		countryNested := opts.GetNestedOptions("author.country")
		require.NotNil(t, countryNested)
		assert.Len(t, countryNested, 1)
		assert.NotNil(t, countryNested["region"])
		assert.Equal(t, uint(5), countryNested["region"].Limit)

		// No nested for tags
		tagsNested := opts.GetNestedOptions("tags")
		assert.Nil(t, tagsNested)
	})
}

func TestRelationOptionsGet(t *testing.T) {
	t.Run("nil options returns nil", func(t *testing.T) {
		var opts RelationOptions
		assert.Nil(t, opts.Get("tags"))
	})

	t.Run("existing key returns option", func(t *testing.T) {
		opts := RelationOptions{
			"tags": {Limit: 5},
		}
		assert.NotNil(t, opts.Get("tags"))
		assert.Equal(t, uint(5), opts.Get("tags").Limit)
	})

	t.Run("non-existing key returns nil", func(t *testing.T) {
		opts := RelationOptions{
			"tags": {Limit: 5},
		}
		assert.Nil(t, opts.Get("categories"))
	})
}
