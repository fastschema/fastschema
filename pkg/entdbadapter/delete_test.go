package entdbadapter

import (
	"database/sql"
	"testing"

	"entgo.io/ent/dialect"
	dialectSql "entgo.io/ent/dialect/sql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/testutils"
	"github.com/fastschema/fastschema/pkg/utils"
)

var sbd = testutils.CreateSchemaBuilder("../../tests/data/schemas")

func TestDeleteNodes(t *testing.T) {
	tests := []testutils.MockTestDeleteData{
		{
			Name:       "delete",
			Schema:     "user",
			Predicates: []*app.Predicate{app.EQ("id", 1)},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `users` WHERE `id` = ?")).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:   "delete/multiple_conditions",
			Schema: "user",
			Predicates: []*app.Predicate{
				app.And(app.GT("id", 1), app.LT("id", 10)),
				app.Like("name", "%test%"),
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `users` WHERE (`id` > ? AND `id` < ?) AND `name` LIKE ?")).
					WithArgs(1, 10, "%test%").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
	}

	testutils.MockRunDeleteTests(func(d *sql.DB) app.DBClient {
		driver := utils.Must(NewEntClient(&app.DBConfig{
			Driver: "sqlmock",
		}, sbd, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, sbd, t, tests)
}
