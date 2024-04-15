package entdbadapter

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"entgo.io/ent/dialect"
	dialectSql "entgo.io/ent/dialect/sql"
	entSchema "entgo.io/ent/dialect/sql/schema"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestUpdateError(t *testing.T) {
	mut := &Mutation{
		model: &Model{name: "user"},
	}
	_, err := mut.Update(nil)
	assert.Equal(t, errors.New("model or schema user not found"), err)
}

func TestUpdateClientIsNotEntClient(t *testing.T) {
	mut := &Mutation{
		model: &Model{
			name:        "user",
			schema:      &schema.Schema{},
			entIDColumn: &entSchema.Column{},
		},
		client: nil,
	}
	_, err := mut.Update(schema.NewEntity())
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
			Predicates: []*app.Predicate{app.EQ("id", 1)},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `name` = ?, `age` = ?, `updated_at` = NOW() WHERE `id` = ?")).
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
			Predicates: []*app.Predicate{app.EQ("id", 1)},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `name` = ?, `bio` = LOWER(`bio`), `updated_at` = NOW() WHERE `id` = ?")).
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
			Predicates: []*app.Predicate{
				app.EQ("id", 1),
				app.IsFalse("deleted"),
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `name` = ?, `deleted` = ?, `age` = COALESCE(`users`.`age`, 0) + ?, `updated_at` = NOW() WHERE `id` = ? AND NOT `deleted`")).
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
			Predicates: []*app.Predicate{
				app.EQ("id", 1),
				app.IsFalse("deleted"),
				app.LTE("age", 30),
			},
			Transaction: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users` WHERE `id` = ? AND NOT `deleted` AND `age` <= ?")).
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
			Predicates: []*app.Predicate{
				app.EQ("id", 1),
				app.IsFalse("deleted"),
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `bio` = NULL, `name` = ?, `deleted` = ?, `updated_at` = NOW() WHERE `id` = ? AND NOT `deleted`")).
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
			Predicates:  []*app.Predicate{app.EQ("id", 1)},
			Transaction: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users` WHERE `id` = ?")).
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
			Predicates:  []*app.Predicate{app.EQ("id", 1)},
			Transaction: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users` WHERE `id` = ?")).
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
					WithArgs(1, 2).
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
					"parent": { "id": 2 }
				}
			}`,
			Predicates:  []*app.Predicate{app.EQ("id", 1)},
			Transaction: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users` WHERE `id` = ?")).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `workplace_id` = NULL, `parent_id` = ?, `room_id` = ?, `updated_at` = NOW() WHERE `id` = ?")).
					WithArgs(2, 2, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `cars` SET `owner_id` = NULL WHERE `owner_id` = ?")).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:   "edges/o2o_bidi",
			Schema: "user",
			InputJSON: `{
				"$clear": {
					"partner": true,
					"spouse": { "id": 2 }
				},
				"$add": {
					"spouse": { "id": 3 }
				}
			}`,
			Predicates:  []*app.Predicate{app.EQ("id", 1)},
			Transaction: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users` WHERE `id` = ?")).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(1))
				// Clear the "partner" from 1's column, and set "spouse 3".
				// "spouse 2" is implicitly removed when setting a different foreign-key.
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `partner_id` = NULL, `spouse_id` = ?, `updated_at` = NOW() WHERE `id` = ?")).
					WithArgs(3, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Clear the "partner_id" column from previous 1's partner.
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `partner_id` = NULL WHERE `partner_id` = ?")).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Clear "spouse 1" from 3's column.
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `spouse_id` = NULL WHERE `id` = ? AND `spouse_id` = ?")).
					WithArgs(2, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Set 3's column to point "spouse 1".
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `spouse_id` = ? WHERE `id` = ? AND `spouse_id` IS NULL")).
					WithArgs(1, 3).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:   "edges/clear_add_m2m",
			Schema: "user",
			InputJSON: `{
				"$clear": {
					"blocking": true,
					"friends": { "id": 2 },
					"groups": [ { "id": 3 }, { "id": 7 } ],
					"following": true,
					"comments": true
				},
				"$add": {
					"friends": [ { "id": 3 }, { "id": 4 } ],
					"groups": [ { "id": 5 }, { "id": 6 }, { "id": 7 } ]
				}
			}`,
			Predicates:  []*app.Predicate{app.EQ("id", 1)},
			Transaction: true,
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users` WHERE `id` = ?")).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(1))
				// Clear all blocked users.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `blockers_blocking` WHERE `blockers` = ?")).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Clear comment responders.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `comments_responder` WHERE `responder` = ?")).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Clear all user following.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `followers_following` WHERE `followers` = ?")).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 2))
				// Clear user friends.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `friends_user` WHERE (`friends` = ? AND `user` = ?) OR (`user` = ? AND `friends` = ?)")).
					WithArgs(1, 2, 1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Remove user groups.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `groups_users` WHERE `users` = ? AND `groups` IN (?, ?)")).
					WithArgs(1, 3, 7).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Add new friends.
				mock.ExpectExec(utils.EscapeQuery("INSERT INTO `friends_user` (`friends`, `user`) VALUES (?, ?), (?, ?), (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `friends` = `friends_user`.`friends`, `user` = `friends_user`.`user`")).
					WithArgs(1, 3, 3, 1, 1, 4, 4, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				// Add new groups.
				mock.ExpectExec(utils.EscapeQuery("INSERT INTO `groups_users` (`groups`, `users`) VALUES (?, ?), (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `groups` = `groups_users`.`groups`, `users` = `groups_users`.`users`")).
					WithArgs(5, 1, 6, 1, 7, 1).
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
			Predicates: []*app.Predicate{
				app.EQ("id", 1),
				app.IsFalse("deleted"),
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `bio` = NULL, `name` = ?, `deleted` = ?, `age` = COALESCE(`users`.`age`, 0) + ?, `updated_at` = NOW() WHERE `id` = ? AND NOT `deleted`")).
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
			Predicates: []*app.Predicate{
				app.EQ("id", 1),
				app.IsFalse("deleted"),
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `bio` = NULL, `deleted` = ?, `age` = COALESCE(`users`.`age`, 0) + ?, `updated_at` = NOW() WHERE `id` = ? AND NOT `deleted`")).
					WithArgs(true, float64(1), 1).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
	}

	sb := createSchemaBuilder()
	MockRunUpdateTests(func(d *sql.DB) app.DBClient {
		driver := utils.Must(NewEntClient(&app.DBConfig{
			Driver:     "sqlmock",
			LogQueries: false,
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
			Predicates: []*app.Predicate{app.EQ("name", "User 1")},
			Expect: func(mock sqlmock.Sqlmock) {
				// Clear fields.
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `name` = NULL, `age` = NULL, `updated_at` = NOW() WHERE `name` = ?")).
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
						"id": 4
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
						AddRow(1))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `users` SET `workplace_id` = NULL, `parent_id` = ?, `room_id` = ?, `updated_at` = NOW() WHERE `id` = ?")).
					WithArgs(4, 5, 1).
					WillReturnResult(sqlmock.NewResult(0, 2))
				mock.ExpectExec(utils.EscapeQuery("UPDATE `cars` SET `owner_id` = NULL WHERE `owner_id` = ?")).
					WithArgs(1).
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
					"followers": [ { "id": 5 }, { "id": 6 } ],
					"friends": [ { "id": 7 }, { "id": 8 } ]
				},
				"$add": {
					"groups": [ { "id": 4 }, { "id": 5 } ],
					"followers": [ { "id": 7 }, { "id": 8 } ],
				}
			}`,
			Expect: func(mock sqlmock.Sqlmock) {
				// Get all node ids first.
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users`")).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(1))
					// Clear user's followers.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `followers_following` WHERE `following` = ? AND `followers` IN (?, ?)")).
					WithArgs(1, 5, 6).
					WillReturnResult(sqlmock.NewResult(0, 2))
					// Clear user's friends.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `friends_user` WHERE (`friends` = ? AND `user` IN (?, ?)) OR (`user` = ? AND `friends` IN (?, ?))")).
					WithArgs(1, 7, 8, 1, 7, 8).
					WillReturnResult(sqlmock.NewResult(0, 2))
				// Clear user's groups.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `groups_users` WHERE `users` = ? AND `groups` IN (?, ?)")).
					WithArgs(1, 2, 3).
					WillReturnResult(sqlmock.NewResult(0, 2))
					// Attach new friends to user.
				mock.ExpectExec(utils.EscapeQuery("INSERT INTO `followers_following` (`followers`, `following`) VALUES (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `followers` = `followers_following`.`followers`, `following` = `followers_following`.`following`")).
					WithArgs(7, 1, 8, 1).
					WillReturnResult(sqlmock.NewResult(0, 2))
				// Attach new groups to user.
				mock.ExpectExec(utils.EscapeQuery("INSERT INTO `groups_users` (`groups`, `users`) VALUES (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `groups` = `groups_users`.`groups`, `users` = `groups_users`.`users`")).
					WithArgs(4, 1, 5, 1).
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
					"followers": [ { "id": 5 }, { "id": 6 } ],
					"friends": [ { "id": 7 }, { "id": 8 } ],
					"groups": [ { "id": 2 }, { "id": 3 } ]
				},
				"$add": {
					"groups": [ { "id": 4 }, { "id": 5 } ],
					"followers": [ { "id": 7 }, { "id": 8 } ]
				}
			}`,
			Expect: func(mock sqlmock.Sqlmock) {
				// Get all node ids first.
				mock.ExpectQuery(utils.EscapeQuery("SELECT `id` FROM `users`")).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).
						AddRow(10).
						AddRow(20))
				// Clear user's followers.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `followers_following` WHERE `following` IN (?, ?) AND `followers` IN (?, ?)")).
					WithArgs(10, 20, 5, 6).
					WillReturnResult(sqlmock.NewResult(0, 2))
					// Clear user's friends.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `friends_user` WHERE (`friends` IN (?, ?) AND `user` IN (?, ?)) OR (`user` IN (?, ?) AND `friends` IN (?, ?))")).
					WithArgs(10, 20, 7, 8, 10, 20, 7, 8).
					WillReturnResult(sqlmock.NewResult(0, 2))
				// Clear user's groups.
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `groups_users` WHERE `users` IN (?, ?) AND `groups` IN (?, ?)")).
					WithArgs(10, 20, 2, 3).
					WillReturnResult(sqlmock.NewResult(0, 2))
					// Attach new friends to user.
				mock.ExpectExec(utils.EscapeQuery("INSERT INTO `followers_following` (`followers`, `following`) VALUES (?, ?), (?, ?), (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `followers` = `followers_following`.`followers`, `following` = `followers_following`.`following`")).
					WithArgs(7, 10, 7, 20, 8, 10, 8, 20).
					WillReturnResult(sqlmock.NewResult(0, 4))
				// Attach new groups to user.
				mock.ExpectExec(utils.EscapeQuery("INSERT INTO `groups_users` (`groups`, `users`) VALUES (?, ?), (?, ?), (?, ?), (?, ?) ON DUPLICATE KEY UPDATE `groups` = `groups_users`.`groups`, `users` = `groups_users`.`users`")).
					WithArgs(4, 10, 4, 20, 5, 10, 5, 20).
					WillReturnResult(sqlmock.NewResult(0, 4))
			},
			WantAffected: 2,
		},
	}

	sb := createSchemaBuilder()
	MockRunUpdateTests(func(d *sql.DB) app.DBClient {
		driver := utils.Must(NewEntClient(&app.DBConfig{
			Driver:     "sqlmock",
			LogQueries: false,
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, sb, t, tests, true)
}
