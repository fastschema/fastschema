package db

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/schema"
)

// RelationOption defines options for loading relation records.
// It supports limit, offset, sort, filter, and select for controlling
// how relation records are loaded per entity.
type RelationOption struct {
	// Limit specifies the maximum number of relation records to load per entity.
	// A value of 0 means no limit.
	Limit uint `json:"limit,omitempty"`

	// Offset specifies the number of relation records to skip per entity.
	// Used for pagination of relation records.
	Offset uint `json:"offset,omitempty"`

	// Sort specifies the sort order for relation records.
	// Supports the same format as the main query sort (e.g., "name", "-created_at").
	Sort string `json:"sort,omitempty"`

	// Filter specifies the filter for relation records.
	// Supports the same filter format as the main query filter.
	Filter map[string]any `json:"filter,omitempty"`

	// Select specifies which fields to select from relation records.
	// If empty, all fields are selected.
	Select []string `json:"select,omitempty"`
}

// RelationOptions is a map of relation field name to RelationOption.
// It is used to configure how each relation is loaded.
type RelationOptions map[string]*RelationOption

// Clone creates a deep copy of the RelationOption.
func (ro *RelationOption) Clone() *RelationOption {
	if ro == nil {
		return nil
	}

	cloned := &RelationOption{
		Limit:  ro.Limit,
		Offset: ro.Offset,
		Sort:   ro.Sort,
	}

	if ro.Filter != nil {
		cloned.Filter = make(map[string]any, len(ro.Filter))
		maps.Copy(cloned.Filter, ro.Filter)
	}

	if ro.Select != nil {
		cloned.Select = make([]string, len(ro.Select))
		copy(cloned.Select, ro.Select)
	}

	return cloned
}

// Clone creates a deep copy of the RelationOptions.
func (ros RelationOptions) Clone() RelationOptions {
	if ros == nil {
		return nil
	}

	cloned := make(RelationOptions, len(ros))
	for k, v := range ros {
		cloned[k] = v.Clone()
	}

	return cloned
}

// ParseRelationOptions parses a JSON string into RelationOptions.
// The JSON should be an object where keys are relation field names
// and values are RelationOption objects.
//
// Example:
//
//	{
//	  "categories": {
//	    "limit": 5,
//	    "offset": 0,
//	    "sort": "name",
//	    "filter": {"status": "active"}
//	  },
//	  "tags": {
//	    "limit": 3,
//	    "sort": "-created_at"
//	  }
//	}
func ParseRelationOptions(jsonStr string) (RelationOptions, error) {
	if jsonStr == "" {
		return nil, nil
	}

	var options RelationOptions
	if err := json.Unmarshal([]byte(jsonStr), &options); err != nil {
		return nil, fmt.Errorf("invalid relation options format: %w", err)
	}

	return options, nil
}

// GetNestedOptions extracts options for nested relations.
// For example, if we have options for "author.country" and we're loading "author",
// this method returns options for "country" to be used when loading nested relations.
func (ros RelationOptions) GetNestedOptions(parentField string) RelationOptions {
	if ros == nil {
		return nil
	}

	prefix := parentField + "."
	nested := make(RelationOptions)

	for key, opt := range ros {
		if after, ok := strings.CutPrefix(key, prefix); ok {
			nestedKey := after
			nested[nestedKey] = opt
		}
	}

	if len(nested) == 0 {
		return nil
	}

	return nested
}

// Get returns the RelationOption for the given field name.
// Returns nil if no options are defined for the field.
func (ros RelationOptions) Get(fieldName string) *RelationOption {
	if ros == nil {
		return nil
	}
	return ros[fieldName]
}

// CreatePredicatesFromRelationFilter creates predicates from a RelationOption filter.
// It uses the schema builder to validate the filter fields.
func CreatePredicatesFromRelationFilter(
	sb *schema.Builder,
	s *schema.Schema,
	filter map[string]any,
) ([]*Predicate, error) {
	if filter == nil {
		return nil, nil
	}

	filterEntity := entity.NewEntityFromMap(filter)
	return createObjectPredicates(sb, s, filterEntity)
}
