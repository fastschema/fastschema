package entdbadapter

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"entgo.io/ent/dialect"
	dialectSql "entgo.io/ent/dialect/sql"
	entSchema "entgo.io/ent/dialect/sql/schema"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
)

func TestDeleteNodes(t *testing.T) {
	tests := []MockTestDeleteData{
		{
			Name:       "delete",
			Schema:     "user",
			Predicates: []*db.Predicate{db.EQ("id", 1)},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `users` WHERE `id` = ?")).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			Name:   "delete/multiple_conditions",
			Schema: "user",
			Predicates: []*db.Predicate{
				db.And(db.GT("id", 1), db.LT("id", 10)),
				db.Like("name", "%test%"),
			},
			Expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(utils.EscapeQuery("DELETE FROM `users` WHERE (`id` > ? AND `id` < ?) AND `name` LIKE ?")).
					WithArgs(1, 10, "%test%").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
	}

	sb := createSchemaBuilder()
	MockRunDeleteTests(func(d *sql.DB) db.Client {
		driver := utils.Must(NewEntClient(&db.Config{
			Driver: "sqlmock",
		}, sb, dialectSql.OpenDB(dialect.MySQL, d)))
		return driver
	}, sb, t, tests)
}

func TestDeleteClientIsNotEntClient(t *testing.T) {
	mut := &Mutation{
		model: &Model{
			name:        "user",
			schema:      &schema.Schema{},
			entIDColumn: &entSchema.Column{},
		},
		client: nil,
	}
	_, err := mut.Delete(context.Background())
	assert.Equal(t, errors.New("client is not an ent adapter"), err)
}

func TestDeleteInvalidOperator(t *testing.T) {
	mut := &Mutation{
		model: &Model{
			name:        "user",
			schema:      &schema.Schema{},
			entIDColumn: &entSchema.Column{},
		},
		client: createMockAdapter(t),
		predicates: []*db.Predicate{
			{
				Field:    "id",
				Operator: db.OpInvalid,
			},
		},
	}
	_, err := mut.Delete(context.Background())
	assert.Equal(t, errors.New("operator invalid not supported"), err)
}
