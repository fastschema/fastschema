package entdbadapter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"

	"entgo.io/ent/dialect"
	dialectSql "entgo.io/ent/dialect/sql"
	entSchema "entgo.io/ent/dialect/sql/schema"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUpdateError(t *testing.T) {
	mut := &Mutation{
		model: &Model{name: "user"},
	}
	_, err := mut.Update(context.Background(), nil)
	assert.Equal(t, errors.New("model or schema user not found"), err)
}

func TestUpdateClientIsNotEntClient(t *testing.T) {
	mut := &Mutation{
		model: &Model{
			name:             "user",
			schema:           &schema.Schema{},
			entPrimaryColumn: &entSchema.Column{},
		},
		client: nil,
	}
	_, err := mut.Update(context.Background(), entity.New())
	assert.Equal(t, errors.New("client is not an ent adapter"), err)
}

func TestUpdateNodes(t *testing.T) {
	tests := []MockTestUpdateData{
		{
			Name:   "fields/set",
			Schema: "user",
			InputJSON: `{
				"name": "User 1",
				"age": 30
			}`,
			Predicates: []*db.Predicate{db.EQ("id", 1)},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `name` = ?, `age` = ?, `updated_at` = NOW() WHERE `users`.`id` = ?")).
					WithArgs("User 1", float64(30), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:   "fields/set_modifier/expr",
			Schema: "user",
			InputJSON: fmt.Sprintf(`{
				"name": "User 1 name",
				"$expr": {
					"bio": "LOWER(%s)"
				}
			}`, "`bio`"),
			Predicates: []*db.Predicate{db.EQ("id", 1)},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `name` = ?, `bio` = LOWER(`bio`), `updated_at` = NOW() WHERE `users`.`id` = ?")).
					WithArgs("User 1 name", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:   "fields/add",
			Schema: "user",
			InputJSON: `{
				"name": "User 1 updated",
				"deleted": true,
				"$add": {
					"age": 3
				}
			}`,
			Predicates: []*db.Predicate{
				db.EQ("id", 1),
				db.IsFalse("deleted"),
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `name` = ?, `deleted` = ?, `age` = COALESCE(`users`.`age`, 0) + ?, `updated_at` = NOW() WHERE `users`.`id` = ? AND NOT `users`.`deleted`")).
					WithArgs("User 1 updated", true, float64(3), 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			Name:   "fields/add_o2m_m2m",
			Schema: "user",
			InputJSON: `{
				"name": "User 1 updated",
				"deleted": true,
				"$add": {
					"sub_pets": [ { "id": 2 }, { "id": 3 } ],
					"sub_groups": [ { "id": 4 }, { "id": 5 } ]
				}
			}`,
			Predicates: []*db.Predicate{
				db.EQ("id", 1),
				db.IsFalse("deleted"),
				db.LTE("age", 30),
			},
			Transaction: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users` WHERE `users`.`id` = ? AND NOT `users`.`deleted` AND `users`.`age` <= ?")).
					WithArgs(1, 30).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `name` = ?, `deleted` = ?, `updated_at` = NOW() WHERE `id` = ?")).
					WithArgs("User 1 updated", true, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(utils.EscapeQuery("INSERT INTO `sub_groups_sub_users` (`sub_groups`, `sub_users`) VALUES (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `sub_groups` = `sub_groups_sub_users`.`sub_groups`, `sub_users` = `sub_groups_sub_users`.`sub_users`")).
					WithArgs(4, 1, 5, 1).
					WillReturnResult(sqlmock.NewResult(0, 2))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `pets` SET `sub_owner_id` = ? WHERE `id` IN (?, ?) AND `sub_owner_id` IS NULL")).
					WithArgs(1, 2, 3).
					WillReturnResult(sqlmock.NewResult(0, 2))
			},
		},
		{
			Name:   "fields/clear",
			Schema: "user",
			InputJSON: `{
				"name": "User 1 updated",
				"deleted": true,
				"$clear": {
					"bio": true
				}
			}`,
			Predicates: []*db.Predicate{
				db.EQ("id", 1),
				db.IsFalse("deleted"),
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `bio` = NULL, `name` = ?, `deleted` = ?, `updated_at` = NOW() WHERE `users`.`id` = ? AND NOT `users`.`deleted`")).
					WithArgs("User 1 updated", true, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			Name:   "fields/clear/o2o_o2m_m2m",
			Schema: "user",
			InputJSON: `{
				"name": "User 1 updated",
				"$clear": {
					"bio": true,
					"car": true,
					"sub_pets": true,
					"sub_groups": true,
					"pets": [ { "id": 2 }, { "id": 3 } ],
					"groups": [ { "id": 4 }, { "id": 5 } ]
				}
			}`,
			Predicates:  []*db.Predicate{db.EQ("id", 1)},
			Transaction: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users` WHERE `users`.`id` = ?")).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `bio` = NULL, `name` = ?, `updated_at` = NOW() WHERE `id` = ?")).
					WithArgs("User 1 updated", 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `groups_users` WHERE `users` = ? AND `groups` IN (?, ?)")).
					WithArgs(1, 4, 5).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `sub_groups_sub_users` WHERE `sub_users` = ?")).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `pets` SET `sub_owner_id` = NULL WHERE `sub_owner_id` = ?")).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `pets` SET `owner_id` = NULL WHERE `id` IN (?, ?) AND `owner_id` = ?")).
					WithArgs(2, 3, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `cars` SET `owner_id` = NULL WHERE `owner_id` = ?")).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			Name:   "fields/set/block",
			Schema: "user",
			InputJSON: `{
				"name": "User 1 updated",
				"$set": {
					"bio": "Hello World",
					"sub_card": { "id": 2 },
					"sub_pets": [ { "id": 2 }, { "id": 3 } ],
					"sub_groups": [ { "id": 4 }, { "id": 5 } ]
				}
			}`,
			Predicates:  []*db.Predicate{db.EQ("id", 1)},
			Transaction: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users` WHERE `users`.`id` = ?")).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `name` = ?, `bio` = ?, `updated_at` = NOW() WHERE `id` = ?")).
					WithArgs("User 1 updated", "Hello World", 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `sub_groups_sub_users` WHERE `sub_users` = ?")).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(utils.EscapeQuery("INSERT INTO `sub_groups_sub_users` (`sub_groups`, `sub_users`) VALUES (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `sub_groups` = `sub_groups_sub_users`.`sub_groups`, `sub_users` = `sub_groups_sub_users`.`sub_users`")).
					WithArgs(4, 1, 5, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `pets` SET `sub_owner_id` = NULL WHERE `sub_owner_id` = ?")).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `cards` SET `sub_owner_id` = NULL WHERE `sub_owner_id` = ?")).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `pets` SET `sub_owner_id` = ? WHERE `id` IN (?, ?) AND `sub_owner_id` IS NULL")).
					WithArgs(1, 2, 3).
					WillReturnResult(sqlmock.NewResult(0, 2))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `cards` SET `sub_owner_id` = ? WHERE `id` = ? AND `sub_owner_id` IS NULL")).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			Name:   "edges/o2o_non_inverse_and_m2o",
			Schema: "user",
			InputJSON: `{
				"$clear": {
					"car": true,
					"workplace": true
				},
				"$add": {
					"room": { "id": 2 },
					"parent": { "id": "00000000-0000-0000-0000-000000000002" }
				}
			}`,
			Predicates:  []*db.Predicate{db.EQ("id", testUserUUID1)},
			Transaction: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users` WHERE `users`.`id` = ?")).
					WithArgs(testUserUUID1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(testUserUUID1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `workplace_id` = NULL, `parent_id` = ?, `room_id` = ?, `updated_at` = NOW() WHERE `id` = ?")).
					WithArgs(testUserUUID2, 2, testUserUUID1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `cars` SET `owner_id` = NULL WHERE `owner_id` = ?")).
					WithArgs(testUserUUID1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:   "edges/o2o_bidi",
			Schema: "user",
			InputJSON: `{
				"$clear": {
					"partner": true,
					"spouse": { "id": "00000000-0000-0000-0000-000000000002" }
				},
				"$add": {
					"spouse": { "id": "00000000-0000-0000-0000-000000000003" }
				}
			}`,
			Predicates:  []*db.Predicate{db.EQ("id", testUserUUID1)},
			Transaction: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users` WHERE `users`.`id` = ?")).
					WithArgs(testUserUUID1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(testUserUUID1))
				// Clear the "partner" from 1's column, and set "spouse 3".
				// "spouse 2" is implicitly removed when setting a different foreign-key.
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `partner_id` = NULL, `spouse_id` = ?, `updated_at` = NOW() WHERE `id` = ?")).
					WithArgs(testUserUUID3, testUserUUID1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Clear the "partner_id" column from previous 1's partner.
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `partner_id` = NULL WHERE `partner_id` = ?")).
					WithArgs(testUserUUID1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Clear "spouse 1" from 3's column.
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `spouse_id` = NULL WHERE `id` = ? AND `spouse_id` = ?")).
					WithArgs(testUserUUID2, testUserUUID1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Set 3's column to point "spouse 1".
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `spouse_id` = ? WHERE `id` = ? AND `spouse_id` IS NULL")).
					WithArgs(testUserUUID1, testUserUUID3).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:   "edges/clear_add_m2m",
			Schema: "user",
			InputJSON: `{
				"$clear": {
					"blocking": true,
					"friends": { "id": "00000000-0000-0000-0000-000000000002" },
					"groups": [ { "id": 3 }, { "id": 7 } ],
					"following": true,
					"comments": true
				},
				"$add": {
					"friends": [ { "id": "00000000-0000-0000-0000-000000000003" }, { "id": "00000000-0000-0000-0000-000000000004" } ],
					"groups": [ { "id": 5 }, { "id": 6 }, { "id": 7 } ]
				}
			}`,
			Predicates:  []*db.Predicate{db.EQ("id", testUserUUID1)},
			Transaction: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users` WHERE `users`.`id` = ?")).
					WithArgs(testUserUUID1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(testUserUUID1))
				// Clear all blocked users.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `blockers_blocking` WHERE `blockers` = ?")).
					WithArgs(testUserUUID1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Clear comment responders.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `comments_responder` WHERE `responder` = ?")).
					WithArgs(testUserUUID1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Clear all user following.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `followers_following` WHERE `followers` = ?")).
					WithArgs(testUserUUID1).
					WillReturnResult(sqlmock.NewResult(1, 2))
				// Clear user friends.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `friends_user` WHERE (`friends` = ? AND `user` = ?) OR (`user` = ? AND `friends` = ?)")).
					WithArgs(testUserUUID1, testUserUUID2, testUserUUID1, testUserUUID2).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Remove user groups.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `groups_users` WHERE `users` = ? AND `groups` IN (?, ?)")).
					WithArgs(testUserUUID1, 3, 7).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Add new friends.
				mock.ExpectExec(utils.EscapeQuery("INSERT INTO `friends_user` (`friends`, `user`) VALUES (?, ?), (?, ?), (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `friends` = `friends_user`.`friends`, `user` = `friends_user`.`user`")).
					WithArgs(testUserUUID1, testUserUUID3, testUserUUID3, testUserUUID1, testUserUUID1, uuid.MustParse("00000000-0000-0000-0000-000000000004"), uuid.MustParse("00000000-0000-0000-0000-000000000004"), testUserUUID1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Add new groups.
				mock.ExpectExec(utils.EscapeQuery("INSERT INTO `groups_users` (`groups`, `users`) VALUES (?, ?), (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `groups` = `groups_users`.`groups`, `users` = `groups_users`.`users`")).
					WithArgs(5, testUserUUID1, 6, testUserUUID1, 7, testUserUUID1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:   "fields/add_set_clear",
			Schema: "user",
			InputJSON: `{
				"name": "User 1 updated",
				"deleted": true,
				"$add": {
					"age": 1
				},
				"$clear": {
					"bio": true
				}
			}`,
			Predicates: []*db.Predicate{
				db.EQ("id", 1),
				db.IsFalse("deleted"),
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `bio` = NULL, `name` = ?, `deleted` = ?, `age` = COALESCE(`users`.`age`, 0) + ?, `updated_at` = NOW() WHERE `users`.`id` = ? AND NOT `users`.`deleted`")).
					WithArgs("User 1 updated", true, float64(1), 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			Name:   "fields/ensure_exists",
			Schema: "user",
			InputJSON: `{
				"$add": {
					"age": 1
				},
				"$clear": {
					"bio": true
				},
				"deleted": true
			}`,
			Predicates: []*db.Predicate{
				db.EQ("id", 1),
				db.IsFalse("deleted"),
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `bio` = NULL, `deleted` = ?, `age` = COALESCE(`users`.`age`, 0) + ?, `updated_at` = NOW() WHERE `users`.`id` = ? AND NOT `users`.`deleted`")).
					WithArgs(true, float64(1), 1).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
	}

	sb := createSchemaBuilder()
	MockRunUpdateTests(func(d *sql.DB) db.Client {
		driver := utils.Must(NewEntClient(&db.Config{
			Driver:     "sqlmock",
			LogQueries: false,
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, sb, t, tests)
}

func TestUpdateNodesPreHookError(t *testing.T) {
	tests := []MockTestUpdateData{
		{
			Name:   "fields/set",
			Schema: "user",
			InputJSON: `{
				"name": "User 1",
				"age": 30
			}`,
			Predicates: []*db.Predicate{db.EQ("id", 1)},
			WantErr:    true,
		},
	}

	sb := createSchemaBuilder()
	MockRunUpdateTests(func(d *sql.DB) db.Client {
		driver := utils.Must(NewEntClient(&db.Config{
			Driver:     "sqlmock",
			LogQueries: false,
			Hooks: func() *db.Hooks {
				return &db.Hooks{
					PreDBUpdate: []db.PreDBUpdate{
						func(ctx context.Context, schema *schema.Schema, predicates *[]*db.Predicate, updateData *entity.Entity) error {
							return errors.New("hook error")
						},
					},
				}
			},
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, sb, t, tests)
}

func TestUpdateNodesHookError(t *testing.T) {
	tests := []MockTestUpdateData{
		{
			Name:   "fields/set",
			Schema: "user",
			InputJSON: `{
				"name": "User 1",
				"age": 30
			}`,
			Predicates: []*db.Predicate{db.EQ("id", 1)},
			WantErr:    true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT * FROM `users` WHERE `users`.`id` = ?")).
					WithArgs(1).
					WillReturnRows(mock.NewRows([]string{"id", "name"}).
						AddRow(1, "John"))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `name` = ?, `age` = ?, `updated_at` = NOW() WHERE `users`.`id` = ?")).
					WithArgs("User 1", float64(30), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
	}

	sb := createSchemaBuilder()
	MockRunUpdateTests(func(d *sql.DB) db.Client {
		driver := utils.Must(NewEntClient(&db.Config{
			Driver:     "sqlmock",
			LogQueries: false,
			Hooks: func() *db.Hooks {
				return &db.Hooks{
					PostDBUpdate: []db.PostDBUpdate{
						func(
							ctx context.Context,
							schema *schema.Schema,
							predicates *[]*db.Predicate,
							updateData *entity.Entity,
							originalEntities []*entity.Entity,
							affected int,
						) error {
							assert.Greater(t, len(*predicates), 0)
							assert.Greater(t, len(originalEntities), 0)
							assert.Greater(t, affected, 0)
							return errors.New("hook error")
						},
					},
				}
			},
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, sb, t, tests)
}

func TestUpdateNodesExtended(t *testing.T) {
	assert.Equal(t, 1, 1)
	tests := []MockTestUpdateData{
		{
			Name:   "without_predicate",
			Schema: "user",
			InputJSON: `{
				"name": "User 1",
				"age": 30
			}`,
			Expect: func(mock sqlmock.Sqlmock) {
				// Apply field changes.
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `name` = ?, `age` = ?, `updated_at` = NOW()")).
					WithArgs("User 1", float64(30)).
					WillReturnResult(sqlmock.NewResult(0, 2))
			},
			WantAffected: 2,
		},
		{
			Name:   "with_predicate",
			Schema: "user",
			InputJSON: `{
				"name": null,
				"age": null
			}`,
			Predicates: []*db.Predicate{db.EQ("name", "User 1")},
			Expect: func(mock sqlmock.Sqlmock) {
				// Clear fields.
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `name` = NULL, `age` = NULL, `updated_at` = NOW() WHERE `users`.`name` = ?")).
					WithArgs("User 1").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			WantAffected: 1,
		},
		{
			Name:   "with_modifier",
			Schema: "user",
			InputJSON: `{
				"$add": {
					"age": 1
				},
				"$expr": {
					"id": "id + 1"
				}
			}`,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `age` = COALESCE(`users`.`age`, 0) + ?, `id` = id + 1, `updated_at` = NOW()")).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			WantAffected: 1,
		},
		{
			Name:   "own_fks/m2o_o2o_inverse",
			Schema: "user",
			InputJSON: `{
				"$add": {
					"parent": {
						"id": "00000000-0000-0000-0000-000000000004"
					},
					"room": {
						"id": 5
					}
				},
				"$clear": {
					"car": true,
					"workplace": true
				}
			}`,
			Transaction: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users`")).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(testUserUUID1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `workplace_id` = NULL, `parent_id` = ?, `room_id` = ?, `updated_at` = NOW() WHERE `id` = ?")).
					WithArgs(uuid.MustParse("00000000-0000-0000-0000-000000000004"), 5, testUserUUID1).
					WillReturnResult(sqlmock.NewResult(0, 2))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `cars` SET `owner_id` = NULL WHERE `owner_id` = ?")).
					WithArgs(testUserUUID1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			WantAffected: 1,
		},
		{
			Name:   "o2m",
			Schema: "user",
			InputJSON: `{
				"$add": {
					"pets": [ { "id": 40 } ],
					"age": 1
				},
				"$clear": {
					"sub_pets": [ { "id": 20 }, { "id": 30 } ]
				}
			}`,
			Transaction: true,
			Expect: func(mock sqlmock.Sqlmock) {
				// Get all node ids first.
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users`")).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(10))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `age` = COALESCE(`users`.`age`, 0) + ?, `updated_at` = NOW() WHERE `id` = ?")).
					WithArgs(float64(1), 10).
					WillReturnResult(sqlmock.NewResult(0, 1))
				// Clear "sub_owner_id" column in the "pets" table.
				mock.ExpectExec(utils.EscapeQuery("UPDATE `pets` SET `sub_owner_id` = NULL WHERE `id` IN (?, ?) AND `sub_owner_id` = ?")).
					WithArgs(20, 30, 10).
					WillReturnResult(sqlmock.NewResult(0, 2))
				// Set "owner_id" column in the "pets" table.
				mock.ExpectExec(utils.EscapeQuery("UPDATE `pets` SET `owner_id` = ? WHERE `id` = ? AND `owner_id` IS NULL")).
					WithArgs(10, 40).
					WillReturnResult(sqlmock.NewResult(0, 2))
			},
			WantAffected: 1,
		},
		{
			Name:        "m2m_one",
			Schema:      "user",
			Transaction: true,
			InputJSON: `{
				"$clear": {
					"groups": [ { "id": 2 }, { "id": 3 } ],
					"followers": [ { "id": "00000000-0000-0000-0000-000000000005" }, { "id": "00000000-0000-0000-0000-000000000006" } ],
					"friends": [ { "id": "00000000-0000-0000-0000-000000000007" }, { "id": "00000000-0000-0000-0000-000000000008" } ]
				},
				"$add": {
					"groups": [ { "id": 4 }, { "id": 5 } ],
					"followers": [ { "id": "00000000-0000-0000-0000-000000000007" }, { "id": "00000000-0000-0000-0000-000000000008" } ]
				}
			}`,
			Expect: func(mock sqlmock.Sqlmock) {
				uuid1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
				uuid5 := uuid.MustParse("00000000-0000-0000-0000-000000000005")
				uuid6 := uuid.MustParse("00000000-0000-0000-0000-000000000006")
				uuid7 := uuid.MustParse("00000000-0000-0000-0000-000000000007")
				uuid8 := uuid.MustParse("00000000-0000-0000-0000-000000000008")
				// Get all node ids first.
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users`")).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(uuid1))
					// Clear user's followers.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `followers_following` WHERE `following` = ? AND `followers` IN (?, ?)")).
					WithArgs(uuid1, uuid5, uuid6).
					WillReturnResult(sqlmock.NewResult(0, 2))
					// Clear user's friends.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `friends_user` WHERE (`friends` = ? AND `user` IN (?, ?)) OR (`user` = ? AND `friends` IN (?, ?))")).
					WithArgs(uuid1, uuid7, uuid8, uuid1, uuid7, uuid8).
					WillReturnResult(sqlmock.NewResult(0, 2))
				// Clear user's groups.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `groups_users` WHERE `users` = ? AND `groups` IN (?, ?)")).
					WithArgs(uuid1, 2, 3).
					WillReturnResult(sqlmock.NewResult(0, 2))
					// Attach new friends to user.
				mock.ExpectExec(utils.EscapeQuery("INSERT INTO `followers_following` (`followers`, `following`) VALUES (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `followers` = `followers_following`.`followers`, `following` = `followers_following`.`following`")).
					WithArgs(uuid7, uuid1, uuid8, uuid1).
					WillReturnResult(sqlmock.NewResult(0, 2))
				// Attach new groups to user.
				mock.ExpectExec(utils.EscapeQuery("INSERT INTO `groups_users` (`groups`, `users`) VALUES (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `groups` = `groups_users`.`groups`, `users` = `groups_users`.`users`")).
					WithArgs(4, uuid1, 5, uuid1).
					WillReturnResult(sqlmock.NewResult(0, 2))
			},
			WantAffected: 1,
		},
		{
			Name:        "m2m_many",
			Schema:      "user",
			Transaction: true,
			InputJSON: `{
				"$clear": {
					"followers": [ { "id": "00000000-0000-0000-0000-000000000005" }, { "id": "00000000-0000-0000-0000-000000000006" } ],
					"friends": [ { "id": "00000000-0000-0000-0000-000000000007" }, { "id": "00000000-0000-0000-0000-000000000008" } ],
					"groups": [ { "id": 2 }, { "id": 3 } ]
				},
				"$add": {
					"groups": [ { "id": 4 }, { "id": 5 } ],
					"followers": [ { "id": "00000000-0000-0000-0000-000000000007" }, { "id": "00000000-0000-0000-0000-000000000008" } ]
				}
			}`,
			Expect: func(mock sqlmock.Sqlmock) {
				uuid5 := uuid.MustParse("00000000-0000-0000-0000-000000000005")
				uuid6 := uuid.MustParse("00000000-0000-0000-0000-000000000006")
				uuid7 := uuid.MustParse("00000000-0000-0000-0000-000000000007")
				uuid8 := uuid.MustParse("00000000-0000-0000-0000-000000000008")
				uuid10 := uuid.MustParse("00000000-0000-0000-0000-000000000010")
				uuid20 := uuid.MustParse("00000000-0000-0000-0000-000000000020")
				// Get all node ids first.
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users`")).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(uuid10).
						AddRow(uuid20))
				// Clear user's followers.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `followers_following` WHERE `following` IN (?, ?) AND `followers` IN (?, ?)")).
					WithArgs(uuid10, uuid20, uuid5, uuid6).
					WillReturnResult(sqlmock.NewResult(0, 2))
					// Clear user's friends.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `friends_user` WHERE (`friends` IN (?, ?) AND `user` IN (?, ?)) OR (`user` IN (?, ?) AND `friends` IN (?, ?))")).
					WithArgs(uuid10, uuid20, uuid7, uuid8, uuid10, uuid20, uuid7, uuid8).
					WillReturnResult(sqlmock.NewResult(0, 2))
				// Clear user's groups.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `groups_users` WHERE `users` IN (?, ?) AND `groups` IN (?, ?)")).
					WithArgs(uuid10, uuid20, 2, 3).
					WillReturnResult(sqlmock.NewResult(0, 2))
					// Attach new friends to user.
				mock.ExpectExec(utils.EscapeQuery("INSERT INTO `followers_following` (`followers`, `following`) VALUES (?, ?), (?, ?), (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `followers` = `followers_following`.`followers`, `following` = `followers_following`.`following`")).
					WithArgs(uuid7, uuid10, uuid7, uuid20, uuid8, uuid10, uuid8, uuid20).
					WillReturnResult(sqlmock.NewResult(0, 4))
				// Attach new groups to user.
				mock.ExpectExec(utils.EscapeQuery("INSERT INTO `groups_users` (`groups`, `users`) VALUES (?, ?), (?, ?), (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `groups` = `groups_users`.`groups`, `users` = `groups_users`.`users`")).
					WithArgs(4, uuid10, 4, uuid20, 5, uuid10, 5, uuid20).
					WillReturnResult(sqlmock.NewResult(0, 4))
			},
			WantAffected: 2,
		},
	}

	sb := createSchemaBuilder()
	MockRunUpdateTests(func(d *sql.DB) db.Client {
		driver := utils.Must(NewEntClient(&db.Config{
			Driver:     "sqlmock",
			LogQueries: false,
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, sb, t, tests, true)
}

func TestUpdateNodesExtendedError(t *testing.T) {
	sb := createSchemaBuilder()
	client := utils.Must(NewTestClient(
		utils.Must(os.MkdirTemp("", "entro_test")),
		sb,
	))

	ctx := context.Background()
	userModel := utils.Must(client.Model("user"))

	// Error predicate.
	{
		_, err := userModel.Mutation().Where(
			&db.Predicate{
				Field:    "id",
				Operator: db.OpIN,
				Value:    1,
			},
		).Update(ctx, entity.New())
		assert.Error(t, err)
	}

	// Invalid block $add
	{
		_, err := userModel.Mutation().
			Where(db.EQ("id", 1)).
			Update(ctx, entity.New().Set("$add", entity.New().Set("invalid", 1)))
		assert.Error(t, err)
	}

	// Invalid block $clear
	{
		_, err := userModel.Mutation().
			Where(db.EQ("id", 1)).
			Update(ctx, entity.New().Set("$clear", entity.New().Set("invalid", 1)))
		assert.Error(t, err)
	}

	// Invalid block $set
	{
		_, err := userModel.Mutation().
			Where(db.EQ("id", 1)).
			Update(ctx, entity.New().Set("$set", entity.New().Set("invalid", 1)))
		assert.Error(t, err)
	}
}
