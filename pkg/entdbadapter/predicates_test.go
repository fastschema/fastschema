package entdbadapter

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	dialectSql "entgo.io/ent/dialect/sql"
	entSchema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var groupSchemaJSON = `{
	"name": "group",
	"namespace": "groups",
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
				"field": "group",
				"owner": true
			}
		},
		{
			"name": "parent",
			"label": "Parent",
			"type": "relation",
			"relation": {
				"type": "o2m",
				"schema": "group",
				"field": "children"
			}
		},
		{
			"name": "children",
			"label": "Children",
			"type": "relation",
			"relation": {
				"type": "o2m",
				"schema": "group",
				"field": "parent",
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
      "name": "year",
      "label": "Year",
      "type": "uint"
    },
		{
			"name": "group",
			"label": "Group",
			"type": "relation",
			"relation": {
				"type": "o2m",
				"schema": "group",
				"field": "cars"
			}
		}
	]
}`

var postSchemaJSON = `{
	"name": "post",
	"namespace": "posts",
	"label_field": "title",
	"fields": [
		{
			"name": "title",
			"label": "Title",
			"type": "string"
		},
		{
			"name": "tags",
			"label": "Tags",
			"type": "relation",
			"relation": {
				"type": "m2m",
				"schema": "tag",
				"field": "posts",
				"owner": true
			}
		},
		{
			"name": "legacy_tags",
			"label": "Legacy Tags",
			"type": "relation",
			"relation": {
				"type": "m2m",
				"schema": "tag",
				"field": "legacy_posts",
				"owner": true,
				"source_column": "legacy_tag_ref",
				"target_column": "legacy_post_ref",
				"junction_table": "post_legacy_tags"
			}
		}
	]
}`

var tagSchemaJSON = `{
	"name": "tag",
	"namespace": "tags",
	"label_field": "label",
	"fields": [
		{
			"name": "label",
			"label": "Label",
			"type": "string"
		},
		{
			"name": "posts",
			"label": "Posts",
			"type": "relation",
			"relation": {
				"type": "m2m",
				"schema": "post",
				"field": "tags"
			}
		},
		{
			"name": "legacy_posts",
			"label": "Legacy Posts",
			"type": "relation",
			"relation": {
				"type": "m2m",
				"schema": "post",
				"field": "legacy_tags",
				"source_column": "legacy_post_ref",
				"target_column": "legacy_tag_ref",
				"junction_table": "post_legacy_tags"
			}
		}
	]
}`

func TestCreateFieldPredicate(t *testing.T) {
	type args struct {
		name               string
		predicate          *db.Predicate
		expectSQLPredicate *dialectSql.Predicate
		expectError        error
	}

	tests := []args{
		{
			name:               "EQ",
			predicate:          db.EQ("name", "John"),
			expectSQLPredicate: dialectSql.EQ("name", "John"),
		},
		{
			name:               "NEQ",
			predicate:          db.NEQ("name", "John"),
			expectSQLPredicate: dialectSql.NEQ("name", "John"),
		},
		{
			name:               "GT",
			predicate:          db.GT("age", 5),
			expectSQLPredicate: dialectSql.GT("age", 5),
		},
		{
			name:               "GTE",
			predicate:          db.GTE("age", 5),
			expectSQLPredicate: dialectSql.GTE("age", 5),
		},
		{
			name:               "LT",
			predicate:          db.LT("age", 5),
			expectSQLPredicate: dialectSql.LT("age", 5),
		},
		{
			name:               "LTE",
			predicate:          db.LTE("age", 5),
			expectSQLPredicate: dialectSql.LTE("age", 5),
		},
		{
			name:               "Like",
			predicate:          db.Like("name", "%John%"),
			expectSQLPredicate: dialectSql.Like("name", "%John%"),
		},
		{
			name: "LikeInvalid",
			predicate: &db.Predicate{
				Field:    "name",
				Operator: db.OpLike,
				Value:    1,
			},
			expectError: errors.New("value of field name.$like = 1 (int) must be string"),
		},
		{
			name:               "In",
			predicate:          db.In("name", []any{"John", "Doe"}),
			expectSQLPredicate: dialectSql.In("name", []any{"John", "Doe"}...),
		},
		{
			name: "InInvalid",
			predicate: &db.Predicate{
				Field:    "name",
				Operator: db.OpIN,
				Value:    1,
			},
			expectError: errors.New("value of field name.$in = 1 (int) must be an array"),
		},
		{
			name:               "NotIn",
			predicate:          db.NotIn("name", []any{"John", "Doe"}),
			expectSQLPredicate: dialectSql.NotIn("name", []any{"John", "Doe"}...),
		},
		{
			name:               "Null",
			predicate:          db.Null("name", true),
			expectSQLPredicate: dialectSql.IsNull("name"),
		},
		{
			name:               "NotNull",
			predicate:          db.Null("name", false),
			expectSQLPredicate: dialectSql.NotNull("name"),
		},
		{
			name: "Invalid",
			predicate: &db.Predicate{
				Field:    "name",
				Operator: db.OpInvalid,
			},
			expectError: errors.New("operator invalid not supported"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFn, err := CreateFieldPredicate(tt.predicate)
			assert.Equal(t, tt.expectError, err)

			if tt.expectError == nil {
				selector := dialectSql.Select("*").From(dialectSql.Table("users"))
				got := gotFn(selector)

				if tt.expectSQLPredicate != nil {
					expectQuery, expectArgs := tt.expectSQLPredicate.Query()
					gotQuery, gotArgs := got.Query()

					assert.Contains(t, gotQuery, expectQuery)
					assert.Equal(t, expectArgs, gotArgs)
				} else {
					assert.Nil(t, got)
				}
			}
		})
	}
}

func TestCreateEntPredicates(t *testing.T) {
	sb := &schema.Builder{}

	groupSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(groupSchemaJSON), groupSchema))
	assert.NoError(t, groupSchema.Init(false))

	carSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(carSchemaJSON), carSchema))
	assert.NoError(t, carSchema.Init(false))

	postSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(postSchemaJSON), postSchema))
	assert.NoError(t, postSchema.Init(false))

	tagSchema := &schema.Schema{}
	assert.Nil(t, json.Unmarshal([]byte(tagSchemaJSON), tagSchema))
	assert.NoError(t, tagSchema.Init(false))

	sb.AddSchema(groupSchema)
	sb.AddSchema(carSchema)
	sb.AddSchema(postSchema)
	sb.AddSchema(tagSchema)
	assert.NoError(t, sb.Init())

	client, err := NewMockExpectClient(func(d *sql.DB) db.Client {
		driver := utils.Must(NewEntClient(&db.Config{
			Driver: "sqlmock",
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, sb, func(m sqlmock.Sqlmock) {
		m.ExpectBegin()
		m.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(1, 1))
	}, false)
	require.NoError(t, err)
	assert.NotNil(t, client)

	entAdapter, ok := client.(EntAdapter)
	require.True(t, ok)

	carModel := &Model{
		client:  entAdapter,
		columns: []*Column{},
		schema:  carSchema,
		entPrimaryColumn: &entSchema.Column{
			Name:      "id",
			Type:      field.TypeUint64,
			Increment: true,
			Unique:    true,
		},
		entTable: &entSchema.Table{
			Name:        "cars",
			Columns:     []*entSchema.Column{},
			PrimaryKey:  []*entSchema.Column{},
			ForeignKeys: []*entSchema.ForeignKey{},
			Annotation: &entsql.Annotation{
				Charset:   "utf8mb4",
				Collation: "utf8mb4_unicode_ci",
			},
		},
	}

	postModelIface, err := entAdapter.Model("post")
	require.NoError(t, err)
	postModel, ok := postModelIface.(*Model)
	require.True(t, ok)

	type args struct {
		model       *Model
		name        string
		predicates  []*db.Predicate
		expectQuery string
		expectArgs  []any
	}

	tests := []args{
		{
			name:       "Nil",
			model:      carModel,
			predicates: []*db.Predicate{nil},
		},
		{
			name:  "And",
			model: carModel,
			predicates: []*db.Predicate{
				db.Like("name", "%car%"),
				db.GT("year", 2000),
			},
			expectQuery: "`cars`.`name` LIKE ? AND `cars`.`year` > ?",
			expectArgs:  []any{"%car%", 2000},
		},
		{
			name:  "Or",
			model: carModel,
			predicates: []*db.Predicate{
				db.Or(
					db.Like("name", "%car%"),
					db.GT("year", 2000),
				),
			},
			expectQuery: "`cars`.`name` LIKE ? OR `cars`.`year` > ?",
			expectArgs:  []any{"%car%", 2000},
		},
		{
			name:  "Relation",
			model: carModel,
			predicates: []*db.Predicate{
				db.GT("year", 2000),
				db.Like("group.parent.name", "%group%"),
			},
			expectQuery: "`cars`.`year` > ? AND EXISTS (SELECT `groups`.`id` FROM `groups` WHERE `cars`.`group_id` = `groups`.`id` AND EXISTS (SELECT `groups_edge`.`id` FROM `groups` AS `groups_edge` WHERE `groups`.`parent_id` = `groups_edge`.`id` AND `groups_edge`.`name` LIKE ?))",
			expectArgs:  []any{2000, "%group%"},
		},
		{
			name:  "RelationM2MNEQ",
			model: postModel,
			predicates: []*db.Predicate{
				db.NEQ("tags.id", uint64(10001)),
			},
			expectQuery: "NOT (`posts`.`id` IN (SELECT `posts_tags`.`posts` FROM `posts_tags` JOIN `tags` AS `t1` ON `posts_tags`.`tags` = `t1`.`id` WHERE `t1`.`id` = ?))",
			expectArgs:  []any{uint64(10001)},
		},
		{
			name:  "RelationM2MNEQCustomColumns",
			model: postModel,
			predicates: []*db.Predicate{
				db.NEQ("legacy_tags.id", uint64(42)),
			},
			expectQuery: "NOT (`posts`.`id` IN (SELECT `post_legacy_tags`.`legacy_post_ref` FROM `post_legacy_tags` JOIN `tags` AS `t1` ON `post_legacy_tags`.`legacy_tag_ref` = `t1`.`id` WHERE `t1`.`id` = ?))",
			expectArgs:  []any{uint64(42)},
		},
		// Implicit PK filter tests
		{
			name:  "ImplicitPKFilterEQ",
			model: postModel,
			predicates: []*db.Predicate{
				db.EQ("tags", uint64(1)),
			},
			expectQuery: "`posts`.`id` IN (SELECT `posts_tags`.`posts` FROM `posts_tags` JOIN `tags` AS `t1` ON `posts_tags`.`tags` = `t1`.`id` WHERE `t1`.`id` = ?)",
			expectArgs:  []any{uint64(1)},
		},
		{
			name:  "ImplicitPKFilterIn",
			model: postModel,
			predicates: []*db.Predicate{
				db.In("tags", []any{uint64(1), uint64(2)}),
			},
			expectQuery: "`posts`.`id` IN (SELECT `posts_tags`.`posts` FROM `posts_tags` JOIN `tags` AS `t1` ON `posts_tags`.`tags` = `t1`.`id` WHERE `t1`.`id` IN (?, ?))",
			expectArgs:  []any{uint64(1), uint64(2)},
		},
		{
			name:  "ImplicitPKFilterNEQ",
			model: postModel,
			predicates: []*db.Predicate{
				db.NEQ("tags", uint64(10)),
			},
			expectQuery: "NOT (`posts`.`id` IN (SELECT `posts_tags`.`posts` FROM `posts_tags` JOIN `tags` AS `t1` ON `posts_tags`.`tags` = `t1`.`id` WHERE `t1`.`id` = ?))",
			expectArgs:  []any{uint64(10)},
		},
		{
			name:  "ImplicitPKFilterNIN",
			model: postModel,
			predicates: []*db.Predicate{
				db.NotIn("tags", []any{uint64(1), uint64(2)}),
			},
			expectQuery: "NOT (`posts`.`id` IN (SELECT `posts_tags`.`posts` FROM `posts_tags` JOIN `tags` AS `t1` ON `posts_tags`.`tags` = `t1`.`id` WHERE `t1`.`id` IN (?, ?)))",
			expectArgs:  []any{uint64(1), uint64(2)},
		},
		{
			name:  "ImplicitPKFilterGT",
			model: postModel,
			predicates: []*db.Predicate{
				db.GT("tags", uint64(5)),
			},
			expectQuery: "`posts`.`id` IN (SELECT `posts_tags`.`posts` FROM `posts_tags` JOIN `tags` AS `t1` ON `posts_tags`.`tags` = `t1`.`id` WHERE `t1`.`id` > ?)",
			expectArgs:  []any{uint64(5)},
		},
		{
			name:  "ImplicitPKFilterCustomColumns",
			model: postModel,
			predicates: []*db.Predicate{
				db.EQ("legacy_tags", uint64(100)),
			},
			expectQuery: "`posts`.`id` IN (SELECT `post_legacy_tags`.`legacy_post_ref` FROM `post_legacy_tags` JOIN `tags` AS `t1` ON `post_legacy_tags`.`legacy_tag_ref` = `t1`.`id` WHERE `t1`.`id` = ?)",
			expectArgs:  []any{uint64(100)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.model)
			selector := dialectSql.Select("*").From(dialectSql.Table(tt.model.schema.Namespace))
			gotFn, err := createEntPredicates(entAdapter, tt.model, tt.predicates)
			assert.NoError(t, err)
			got := gotFn(selector)

			gotQuery, gotArgs := dialectSql.And(got...).Query()
			assert.Equal(t, tt.expectQuery, gotQuery)
			assert.Equal(t, tt.expectArgs, gotArgs)
		})
	}
}
