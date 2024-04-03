package db

import (
	"encoding/json"
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

var modelSchemaJSON = `{
	"name": "model",
	"namespace": "models",
	"label_field": "name",
	"fields": [
		{
      "name": "name",
      "label": "Name",
      "type": "string",
      "unique": true
    },
		{
			"name": "cars",
			"label": "Cars",
			"type": "relation",
			"relation": {
				"type": "o2m",
				"schema": "car",
				"field": "model",
				"owner": true
			}
		}
	]
}`

var carSchemaJSON = `{
	"name": "car",
	"namespace": "cars",
	"label_field": "name",
	"fields": [
		{
      "name": "name",
      "label": "Name",
      "type": "string",
      "unique": true
    },
		{
			"name": "model",
			"label": "Model",
			"type": "relation",
			"relation": {
				"type": "o2m",
				"schema": "model",
				"field": "cars"
			}
		},
		{
			"name": "owner",
			"label": "Owner",
			"type": "relation",
			"relation": {
				"type": "o2m",
				"schema": "user",
				"field": "cars"
			}
		}
	]
}`

var userSchemaJSON = `{
  "name": "user",
  "namespace": "users",
  "label_field": "name",
  "fields": [
    {
      "name": "name",
      "label": "Name",
      "type": "string",
      "unique": true
    },
    {
      "name": "age",
      "label": "Age",
      "type": "uint"
    },
		{
			"name": "cars",
			"label": "Cars",
			"type": "relation",
			"relation": {
				"type": "o2m",
				"schema": "car",
				"field": "owner",
				"owner": true
			}
		}
  ]
}`

type filterArgs struct {
	name         string
	filter       string
	expectError  string
	expectResult []*Predicate
}

func TestClonePredicate(t *testing.T) {
	p := Or(
		EQ("name", "test"),
		And(
			EQ("age", 10),
			Or(
				EQ("name", "test"),
				EQ("name", "test2"),
			),
		),
	)
	p2 := p.Clone()
	assert.Equal(t, p, p2)
}

func TestCreateFieldPredicate(t *testing.T) {
	userSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(userSchemaJSON), userSchema))

	type args struct {
		field        string
		value        any
		expectError  string
		expectResult []*Predicate
	}

	tests := []args{
		{
			field:        "name",
			value:        "test",
			expectResult: []*Predicate{EQ("name", "test")},
		},
		{
			field:       "name",
			value:       utils.Must(schema.NewEntityFromJSON(`{"$eq": false}`)),
			expectError: "filter error: invalid value for field name.$eq (string) = false (bool)",
		},
		{
			field:        "name",
			value:        utils.Must(schema.NewEntityFromJSON(`{"$eq": "test"}`)),
			expectResult: []*Predicate{EQ("name", "test")},
		},
		{
			field:        "name",
			value:        utils.Must(schema.NewEntityFromJSON(`{"$neq": "test"}`)),
			expectResult: []*Predicate{NEQ("name", "test")},
		},
		{
			field:        "age",
			value:        utils.Must(schema.NewEntityFromJSON(`{"$gt": 1}`)),
			expectResult: []*Predicate{GT("age", float64(1))},
		},
		{
			field:        "age",
			value:        utils.Must(schema.NewEntityFromJSON(`{"$gte": 5}`)),
			expectResult: []*Predicate{GTE("age", float64(5))},
		},
		{
			field:        "age",
			value:        utils.Must(schema.NewEntityFromJSON(`{"$lt": 5}`)),
			expectResult: []*Predicate{LT("age", float64(5))},
		},
		{
			field:        "age",
			value:        utils.Must(schema.NewEntityFromJSON(`{"$lte": 5}`)),
			expectResult: []*Predicate{LTE("age", float64(5))},
		},
		{
			field:        "name",
			value:        utils.Must(schema.NewEntityFromJSON(`{"$like": "test%"}`)),
			expectResult: []*Predicate{Like("name", "test%")},
		},
		{
			field:       "name",
			value:       utils.Must(schema.NewEntityFromJSON(`{"$like": 1}`)),
			expectError: "filter error: invalid value for field name.$like (string) = 1 (float64)",
		},
		{
			field:        "age",
			value:        utils.Must(schema.NewEntityFromJSON(`{"$in": [1, 2, 3]}`)),
			expectResult: []*Predicate{In("age", []any{float64(1), float64(2), float64(3)})},
		},
		{
			field:       "age",
			value:       utils.Must(schema.NewEntityFromJSON(`{"$in": 2}`)),
			expectError: "filter error: $in operator must be an array",
		},
		{
			field: "age",
			value: utils.Must(schema.NewEntityFromJSON(`{"$nin": [1, 2, 3]}`)),
			expectResult: []*Predicate{
				NotIn("age", []any{float64(1), float64(2), float64(3)}),
			},
		},
		{
			field:       "age",
			value:       utils.Must(schema.NewEntityFromJSON(`{"$nin": 2}`)),
			expectError: "filter error: $nin operator must be an array",
		},
		{
			field:        "age",
			value:        utils.Must(schema.NewEntityFromJSON(`{"$null": true}`)),
			expectResult: []*Predicate{Null("age", true)},
		},
		{
			field:        "age",
			value:        utils.Must(schema.NewEntityFromJSON(`{"$null": false}`)),
			expectResult: []*Predicate{Null("age", false)},
		},
		{
			field:       "age",
			value:       utils.Must(schema.NewEntityFromJSON(`{"$null": "ok"}`)),
			expectError: "filter error: $null operator must be a boolean",
		},
		{
			field:        "age",
			value:        5,
			expectResult: []*Predicate{EQ("age", 5)},
		},
		{
			field:       "age",
			value:       "five",
			expectError: "filter error: invalid value for field age (uint) = five (string)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			f, err := userSchema.Field(tt.field)
			assert.Nil(t, err)
			assert.NotNil(t, f)
			result, err := createFieldPredicate(f, tt.value)
			if err != nil {
				assert.Equal(t, tt.expectError, err.Error())
			}
			assert.Equal(t, tt.expectResult, result)
		})
	}
}

func TestCreateObjectPredicates(t *testing.T) {
	b := &schema.Builder{}

	modelSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(modelSchemaJSON), modelSchema))
	assert.NoError(t, modelSchema.Init(false))
	b.AddSchema(modelSchema)

	carSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(carSchemaJSON), carSchema))
	assert.NoError(t, carSchema.Init(false))
	b.AddSchema(carSchema)

	userSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(userSchemaJSON), userSchema))
	assert.NoError(t, userSchema.Init(false))
	b.AddSchema(userSchema)

	assert.NoError(t, b.Init())

	tests := []filterArgs{
		{
			name:         "empty",
			filter:       "{}",
			expectResult: []*Predicate{},
		},
		{
			name:        "invalid field",
			filter:      `{"invalid_field": "test"}`,
			expectError: "filter error: field user.invalid_field not found",
		},
		{
			name:         "single_field",
			filter:       `{"name": "test"}`,
			expectResult: []*Predicate{EQ("name", "test")},
		},
		{
			name: "single_field_multiple_condition",
			filter: `{
				"name": {
					"$neq": "test",
					"$like": "test%"
				}
			}`,
			expectResult: []*Predicate{And(
				NEQ("name", "test"),
				Like("name", "test%"),
			)},
		},
		{
			name: "multiple_field",
			filter: `{
				"name": "test",
				"age": {
					"$gt": 5
				}
			}`,
			expectResult: []*Predicate{
				EQ("name", "test"),
				GT("age", float64(5)),
			},
		},
		{
			name: "or",
			filter: `{
				"$or": [
					{"name": "test"},
					{"age": 5}
				]
			}`,
			expectResult: []*Predicate{
				Or(
					EQ("name", "test"),
					EQ("age", float64(5)),
				),
			},
		},
		{
			name: "or_field_multiple_field",
			filter: `{
				"$or": [
					{
						"name": {
							"$neq": "test",
							"$like": "test%"
						},
						"age": {
							"$lt": 10
						}
					},
					{"age": 5}
				]
			}`,
			expectResult: []*Predicate{
				Or(
					And(
						And(
							NEQ("name", "test"),
							Like("name", "test%"),
						),
						LT("age", float64(10)),
					),
					EQ("age", float64(5)),
				),
			},
		},
		{
			name: "and",
			filter: `{
				"name": {
					"$like": "test%",
					"$neq": "test2"
				},
				"$and": [
					{
						"name": {
							"$neq": "test2"
						}
					},
					{"age": 5}
				]
			}`,
			expectResult: []*Predicate{
				And(
					Like("name", "test%"),
					NEQ("name", "test2"),
				),
				And(
					NEQ("name", "test2"),
					EQ("age", float64(5)),
				),
			},
		},
		{
			name:        "or_invalid",
			filter:      `{"$or": {}}`,
			expectError: "invalid $or/$and value",
		},
		{
			name:        "invalid_value",
			filter:      `{"$or": [{"age": "five"}]}`,
			expectError: "filter error: invalid value for field age (uint) = five (string)",
		},
		{
			name: "relation_invalid_format",
			filter: `{
				"cars.": {
					"$like": "carmodel",
				}
			}`,
			expectError: "filter error: cars. is not a valid relation field",
		},
		{
			name: "relation_invalid_field",
			filter: `{
				"groups.name": {
					"$like": "group1",
				}
			}`,
			expectError: "filter error: field user.groups not found",
		},
		{
			name: "relation_invalid_field_type",
			filter: `{
				"age.name": {
					"$like": "group1",
				}
			}`,
			expectError: "filter error: age.name is not a relation field",
		},
		{
			name: "relation_invalid_last_field_type",
			filter: `{
				"cars.name2": {
					"$like": "group1",
				}
			}`,
			expectError: "filter error: field car.name2 not found",
		},
		{
			name: "relation",
			filter: `{
				"cars.model.name": {
					"$like": "carmodel",
				}
			}`,
			expectResult: []*Predicate{
				Like("name", "carmodel", []string{"cars", "model"}...),
			},
		},
	}

	for _, tt := range tests {
		t.Run("predicates/"+tt.name, func(t *testing.T) {
			actual, err := createObjectPredicates(
				b,
				userSchema,
				utils.Must(schema.NewEntityFromJSON(tt.filter)),
			)
			if err != nil {
				assert.Equal(t, tt.expectError, err.Error())
			}
			assert.Equal(t, tt.expectResult, actual)
		})
	}
}

func TestCreatePredicatesFromFilterObject(t *testing.T) {
	testSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(userSchemaJSON), testSchema))
	tests := []filterArgs{
		{
			filter:       "{}",
			expectResult: []*Predicate{},
		},
		{
			filter:      "_",
			expectError: "filter error: Value looks like object, but can't find closing '}' symbol",
		},
	}

	for _, tt := range tests {
		t.Run("filter", func(t *testing.T) {
			actual, err := CreatePredicatesFromFilterObject(nil, testSchema, tt.filter)
			if err != nil {
				assert.Equal(t, tt.expectError, err.Error())
			}
			assert.Equal(t, tt.expectResult, actual)
		})
	}
}
