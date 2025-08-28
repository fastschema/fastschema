package entdbadapter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"entgo.io/ent/dialect"
	dialectSql "entgo.io/ent/dialect/sql"
	entSchema "entgo.io/ent/dialect/sql/schema"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestCreateError(t *testing.T) {
	mut := &Mutation{
		model: &Model{name: "user"},
	}
	_, err := mut.Create(context.Background(), nil)
	assert.Equal(t, errors.New("model or schema user not found"), err)
}

func TestCreateClientIsNotEntClient(t *testing.T) {
	mut := &Mutation{
		model: &Model{
			name:        "user",
			schema:      &schema.Schema{},
			entIDColumn: &entSchema.Column{},
		},
		client: nil,
	}
	_, err := mut.Create(context.Background(), entity.New())
	assert.Equal(t, errors.New("client is not an ent adapter"), err)
}

// Test cases copied from: ent/dialect/sql/sqlgraph/graph_test.go#TestCreateNode
// Skipped these tests:
// - modifiers: Custom modifiers are currently not supported by fastschema
// - edges/m2m/fields: "Ent: Edge Schema" is currently not supported by fastschema (https://entgo.io/docs/schema-edges#edge-schema)
// - edges/m2m/bidi/fields: "Ent: Edge Schema" is currently not supported by fastschema (https://entgo.io/docs/schema-edges#edge-schema)
// - schema: Custom db schema name is currently not supported by fastschema

func TestMockCreateNode(t *testing.T) {
	tests := []MockTestCreateData{
		{
			Name:      "fields",
			Schema:    "user",
			InputJSON: `{ "name": "User 1", "age": 10 }`,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `users` (`name`, `age`) VALUES (?, ?)")).
					WithArgs("User 1", float64(10)).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:      "fields/json",
			Schema:    "user",
			InputJSON: `{ "name": "User 1", "json": {} }`,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `users` (`name`, `json`) VALUES (?, ?)")).
					WithArgs("User 1", []byte("{}")).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:        "fields/error",
			Schema:      "user",
			InputJSON:   `{ "invalid": 1 }`,
			ExpectError: "column error: column user.invalid not found",
		},
		{
			Name:        "fields/error",
			Schema:      "user",
			InputJSON:   `{ "groups": [1] }`,
			ExpectError: "relation value for user.groups is invalid",
		},
	}

	sb := createSchemaBuilder()
	MockRunCreateTests(func(d *sql.DB) db.Client {
		driver := utils.Must(NewEntClient(&db.Config{
			Driver:     "sqlmock",
			LogQueries: true,
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, sb, t, tests)
}

func TestMockCreateNodeHookError(t *testing.T) {
	tests := []MockTestCreateData{
		{
			Name:        "fields",
			Schema:      "user",
			InputJSON:   `{ "name": "User 1", "age": 10 }`,
			ExpectError: "post create hook: hook error",
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `users` (`name`, `age`) VALUES (?, ?)")).
					WithArgs("User 1", float64(10)).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
	}

	sb := createSchemaBuilder()
	MockRunCreateTests(func(d *sql.DB) db.Client {
		driver := utils.Must(NewEntClient(&db.Config{
			Driver: "sqlmock",
			Hooks: func() *db.Hooks {
				return &db.Hooks{
					PostDBCreate: []db.PostDBCreate{
						func(
							ctx context.Context,
							schema *schema.Schema,
							dataCreate *entity.Entity,
							id uint64,
						) error {
							assert.Greater(t, id, uint64(0))
							return errors.New("hook error")
						},
					},
				}
			},
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, sb, t, tests)
}

func TestMockCreateNodePreHookError(t *testing.T) {
	tests := []MockTestCreateData{
		{
			Name:        "fields",
			Schema:      "user",
			InputJSON:   `{ "name": "User 1", "age": 0 }`,
			ExpectError: "pre create hook: hook error",
		},
	}

	sb := createSchemaBuilder()
	MockRunCreateTests(func(d *sql.DB) db.Client {
		driver := utils.Must(NewEntClient(&db.Config{
			Driver: "sqlmock",
			Hooks: func() *db.Hooks {
				return &db.Hooks{
					PreDBCreate: []db.PreDBCreate{
						func(ctx context.Context, schema *schema.Schema, dataCreate *entity.Entity) error {
							return errors.New("hook error")
						},
					},
				}
			},
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, sb, t, tests)
}

func TestMockCreateNodeEdges(t *testing.T) {
	tests := []MockTestCreateData{
		{
			Name:        "edges/o2o_two_types",
			Schema:      "user",
			InputJSON:   `{ "name": "User 1", "sub_card": { "id": 1 } }`,
			Transaction: true,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `users` (`name`) VALUES (?)")).
					WithArgs("User 1").
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(utils.EscapeQuery("UPDATE `cards` SET `sub_owner_id` = ? WHERE `id` = ? AND `sub_owner_id` IS NULL")).
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:        "edges/o2o_two_types/inverse",
			Schema:      "card",
			InputJSON:   `{ "number": "0001", "owner": { "id": 2 } }`,
			Transaction: false,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `cards` (`number`, `owner_id`) VALUES (?, ?)")).
					WithArgs("0001", 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:        "edges/o2o_same_types/bidi",
			Schema:      "user",
			InputJSON:   `{ "name": "User 1", "spouse": { "id": 2 } }`,
			Transaction: true,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `users` (`name`, `spouse_id`) VALUES (?, ?)")).
					WithArgs("User 1", 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `spouse_id` = ? WHERE `id` = ? AND `spouse_id` IS NULL")).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:        "edges/o2o_same_types/recursive",
			Schema:      "node",
			InputJSON:   `{ "name": "Node 2", "prev": { "id": 9 } }`,
			Transaction: false,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `nodes` (`name`, `prev_id`) VALUES (?, ?)")).
					WithArgs("Node 2", 9).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(utils.EscapeQuery("UPDATE `nodes` SET `next_id` = ? WHERE `id` = ? AND `next_id` IS NULL")).
					WithArgs(1, 9).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:        "edges/o2o_same_types/recursive/inverse",
			Schema:      "node",
			InputJSON:   `{ "name": "Node 1", "next": { "id": 9 } }`,
			Transaction: true,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `nodes` (`name`) VALUES (?)")).
					WithArgs("Node 1").
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(utils.EscapeQuery("UPDATE `nodes` SET `prev_id` = ? WHERE `id` = ? AND `prev_id` IS NULL")).
					WithArgs(1, 9).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:        "edges/o2m_two_types",
			Schema:      "user",
			InputJSON:   `{ "name": "User 1", "sub_pets": [{ "id": 1 }] }`,
			Transaction: true,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `users` (`name`) VALUES (?)")).
					WithArgs("User 1").
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(utils.EscapeQuery("UPDATE `pets` SET `sub_owner_id` = ? WHERE `id` = ? AND `sub_owner_id` IS NULL")).
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:        "edges/o2m_two_types/multiple",
			Schema:      "user",
			InputJSON:   `{ "name": "User 1", "sub_pets": [{ "id": 1 }, { "id": 2 }, {"id": 3}] }`,
			Transaction: true,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `users` (`name`) VALUES (?)")).
					WithArgs("User 1").
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(utils.EscapeQuery("UPDATE `pets` SET `sub_owner_id` = ? WHERE `id` IN (?, ?, ?) AND `sub_owner_id` IS NULL")).
					WithArgs(1, 1, 2, 3).
					WillReturnResult(sqlmock.NewResult(1, 3))
			},
		},
		{
			Name:        "edges/o2m_two_types/inverse",
			Schema:      "pet",
			InputJSON:   `{ "name": "Pet 1", "owner": { "id": 2 } }`,
			Transaction: false,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `pets` (`name`, `owner_id`) VALUES (?, ?)")).
					WithArgs("Pet 1", 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:        "edges/o2m_same_types",
			Schema:      "node",
			InputJSON:   `{ "name": "Node 2", "parent": { "id": 1 } }`,
			Transaction: false,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `nodes` (`name`, `parent_id`) VALUES (?, ?)")).
					WithArgs("Node 2", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:        "edges/o2m_same_types/inverse",
			Schema:      "node",
			InputJSON:   `{ "name": "Node 1", "children": [{ "id": 2 }, { "id": 3 }, { "id": 4 }] }`,
			Transaction: true,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `nodes` (`name`) VALUES (?)")).
					WithArgs("Node 1").
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(utils.EscapeQuery("UPDATE `nodes` SET `parent_id` = ? WHERE `id` IN (?, ?, ?) AND `parent_id` IS NULL")).
					WithArgs(1, 2, 3, 4).
					WillReturnResult(sqlmock.NewResult(1, 3))
			},
		},
		{
			Name:        "edges/o2m_same_types/both",
			Schema:      "node",
			InputJSON:   `{ "name": "Node 2", "parent": { "id": 1 }, "children": [{ "id": 3 }, { "id": 4 }] }`,
			Transaction: true,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `nodes` (`name`, `parent_id`) VALUES (?, ?)")).
					WithArgs("Node 2", 1).
					WillReturnResult(sqlmock.NewResult(2, 1))
				m.ExpectExec(utils.EscapeQuery("UPDATE `nodes` SET `parent_id` = ? WHERE `id` IN (?, ?) AND `parent_id` IS NULL")).
					WithArgs(2, 3, 4).
					WillReturnResult(sqlmock.NewResult(1, 2))
			},
		},
		{
			Name:        "edges/m2m",
			Schema:      "group",
			InputJSON:   `{ "name": "GitHub", "users": [{ "id": 3 }] }`,
			Transaction: true,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `groups` (`name`) VALUES (?)")).
					WithArgs("GitHub").
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `groups_users` (`groups`, `users`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `groups` = `groups_users`.`groups`, `users` = `groups_users`.`users`")).
					WithArgs(1, 3).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:        "edges/m2m/inverse",
			Schema:      "user",
			InputJSON:   `{ "name": "user01", "groups": [{ "id": 3 }] }`,
			Transaction: true,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `users` (`name`) VALUES (?)")).
					WithArgs("user01").
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `groups_users` (`groups`, `users`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `groups` = `groups_users`.`groups`, `users` = `groups_users`.`users`")).
					WithArgs(3, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:        "edges/m2m/bidi",
			Schema:      "user",
			InputJSON:   `{ "name": "User 1", "friends": [{ "id": 3 }] }`,
			Transaction: true,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `users` (`name`) VALUES (?)")).
					WithArgs("User 1").
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `friends_user` (`friends`, `user`) VALUES (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `friends` = `friends_user`.`friends`, `user` = `friends_user`.`user`")).
					WithArgs(1, 3, 3, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:        "edges/m2m/bidi/batch",
			Schema:      "user",
			InputJSON:   `{ "name": "User 3", "friends": [{ "id": 1 }, { "id": 2 }], "groups": [{ "id": 1 }, { "id": 2 }] }`,
			Transaction: true,
			Expect: func(m sqlmock.Sqlmock) {
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `users` (`name`) VALUES (?)")).
					WithArgs("User 3").
					WillReturnResult(sqlmock.NewResult(3, 1))
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `friends_user` (`friends`, `user`) VALUES (?, ?), (?, ?), (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `friends` = `friends_user`.`friends`, `user` = `friends_user`.`user`")).
					WithArgs(3, 1, 1, 3, 3, 2, 2, 3).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(utils.EscapeQuery("INSERT INTO `groups_users` (`groups`, `users`) VALUES (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `groups` = `groups_users`.`groups`, `users` = `groups_users`.`users`")).
					WithArgs(1, 3, 2, 3).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
	}

	sb := createSchemaBuilder()
	MockRunCreateTests(func(d *sql.DB) db.Client {
		driver := utils.Must(NewEntClient(&db.Config{
			Driver:     "sqlmock",
			LogQueries: false,
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, sb, t, tests)
}

func TestMockCreateNodeWithRelationData(t *testing.T) {
	fmt.Printf("\n\n")
	tests := []MockTestCreateData{
		// {
		// 	Name:      "edges/m2o/create_relation_entity",
		// 	Schema:    "pet",
		// 	InputJSON: `{ "name": "pet 1", "owner": { "name": "User 2" } }`,
		// 	Expect: func(m sqlmock.Sqlmock) {
		// 		m.ExpectBegin()
		// 		m.ExpectExec(utils.EscapeQuery("INSERT INTO `users` (`name`) VALUES (?)")).
		// 			WithArgs("User 2").
		// 			WillReturnResult(sqlmock.NewResult(1, 1))
		// 		m.ExpectExec(utils.EscapeQuery("INSERT INTO `pets` (`name`, `owner_id`) VALUES (?, ?)")).
		// 			WithArgs("pet 1", 1).
		// 			WillReturnResult(sqlmock.NewResult(1, 1))
		// 		m.ExpectCommit()
		// 	},
		// },
		// {
		// 	Name:      "edges/o2o/inverse/create_relation_entity",
		// 	Schema:    "card",
		// 	InputJSON: `{ "number": "0001", "owner": { "name": "User 1" } }`,
		// 	Expect: func(m sqlmock.Sqlmock) {
		// 		m.ExpectBegin()
		// 		m.ExpectExec(utils.EscapeQuery("INSERT INTO `users` (`name`) VALUES (?)")).
		// 			WithArgs("User 1").
		// 			WillReturnResult(sqlmock.NewResult(1, 1))
		// 		m.ExpectExec(utils.EscapeQuery("INSERT INTO `cards` (`number`, `owner_id`) VALUES (?, ?)")).
		// 			WithArgs("0001", 1).
		// 			WillReturnResult(sqlmock.NewResult(1, 1))
		// 		m.ExpectCommit()
		// 	},
		// },
	}

	sb := createSchemaBuilder()
	MockRunCreateTests(func(d *sql.DB) db.Client {
		driver := utils.Must(NewEntClient(&db.Config{
			Driver: "sqlmock",
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, sb, t, tests)
}
