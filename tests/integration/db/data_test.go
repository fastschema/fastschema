package db

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DBTestCreateData struct {
	Name        string
	Schema      string
	InputJSON   string
	ClearTables []string
	Run         func(model app.Model, entity *schema.Entity) (*schema.Entity, error)
	Prepare     func(t *testing.T)
	Expect      func(t *testing.T, m app.Model, e *schema.Entity)
	WantErr     bool
	ExpectError error
}

type DBTestUpdateData struct {
	Name         string
	Schema       string
	InputJSON    string
	ClearTables  []string
	Run          func(model app.Model, entity *schema.Entity) (int, error)
	Expect       func(t *testing.T, m app.Model)
	Prepare      func(t *testing.T, m app.Model)
	Predicates   []*app.Predicate
	WantErr      bool
	WantAffected int
	Transaction  bool
}

type DBTestDeleteData struct {
	Name         string
	Schema       string
	ClearTables  []string
	Run          func(model app.Model) (int, error)
	Expect       func(t *testing.T, m app.Model)
	Prepare      func(t *testing.T, m app.Model)
	Predicates   []*app.Predicate
	WantErr      bool
	ExpectError  error
	WantAffected int
	Transaction  bool
}

type DBTestQueryData struct {
	Name        string
	Schema      string
	Filter      string
	Limit       uint
	Offset      uint
	Columns     []string
	Order       []string
	ClearTables []string
	Run         func(
		model app.Model,
		predicates []*app.Predicate,
		limit, offset uint,
		order []string,
		columns ...string,
	) ([]*schema.Entity, error)
	Prepare func(t *testing.T, client app.DBClient, m app.Model) []*schema.Entity
	Expect  func(
		t *testing.T,
		m app.Model,
		preparedEntities []*schema.Entity,
		results []*schema.Entity,
	)
	ExpectError string
}

type DBTestCountData struct {
	Name        string
	Schema      string
	Filter      string
	Column      string
	Unique      bool
	ClearTables []string
	Run         func(
		model app.Model,
		predicates []*app.Predicate,
		unique bool,
		column string,
	) (int, error)
	Prepare func(t *testing.T, client app.DBClient, m app.Model) int
	Expect  func(
		t *testing.T,
		m app.Model,
		preparedCount int,
		results int,
	)
	ExpectError string
}

func ClearDBData(client app.DBClient, tables ...string) {
	sqls := []string{}

	if client.Dialect() == dialect.MySQL {
		sqls = append(sqls, "SET FOREIGN_KEY_CHECKS=0")
	}

	if client.Dialect() == dialect.SQLite {
		sqls = append(sqls, "PRAGMA foreign_keys = OFF;")
	}

	if client.Dialect() == dialect.MySQL {
		sqls = append(sqls, strings.Join(utils.Map(tables, func(table string) string {
			return fmt.Sprintf("TRUNCATE TABLE `%s`", table)
		}), ";"))
	}

	if client.Dialect() == dialect.SQLite {
		sqls = append(sqls, strings.Join(utils.Map(tables, func(table string) string {
			return fmt.Sprintf(
				"DELETE FROM %s; DELETE FROM SQLITE_SEQUENCE WHERE name='%s'",
				table,
				table,
			)
		}), ";"))
	}

	if client.Dialect() == dialect.Postgres {
		sqls = append(sqls, fmt.Sprintf(
			"TRUNCATE TABLE %s CASCADE",
			strings.Join(tables, ", "),
		))
		sqls = append(sqls, utils.Map(tables, func(table string) string {
			return fmt.Sprintf(
				"ALTER SEQUENCE IF EXISTS %s_id_seq RESTART WITH 1",
				table,
			)
		})...)
	}

	if client.Dialect() == dialect.MySQL {
		sqls = append(sqls, "SET FOREIGN_KEY_CHECKS=1")
	}

	if client.Dialect() == dialect.SQLite {
		sqls = append(sqls, "PRAGMA foreign_keys = ON;")
	}

	sqls = utils.Filter(sqls, func(sql string) bool {
		return strings.TrimSpace(sql) != ""
	})

	if err := client.Exec(
		context.Background(),
		strings.Join(sqls, "; "),
		[]any{},
		nil,
	); err != nil {
		panic(err)
	}
	fmt.Printf("\n")
}

func DBRunCreateTests(client app.DBClient, t *testing.T, tests []DBTestCreateData) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			fmt.Printf("Running test: %s\n", tt.Name)
			entity, err := schema.NewEntityFromJSON(tt.InputJSON)
			require.NoError(t, err)

			model, err := client.Model(tt.Schema)
			require.NoError(t, err)

			ClearDBData(client, tt.ClearTables...)

			if tt.Prepare != nil {
				tt.Prepare(t)
				fmt.Printf("\n")
			}

			runFn := tt.Run
			if runFn == nil {
				runFn = func(model app.Model, entity *schema.Entity) (*schema.Entity, error) {
					createdEntityID := utils.Must(model.Create(entity))
					return model.Query(app.EQ("id", createdEntityID)).First()
				}
			}

			entity, err = runFn(model, entity)
			require.Equal(t, tt.WantErr, err != nil, err)
			if tt.WantErr && tt.ExpectError != nil {
				require.Equal(t, tt.ExpectError, err)
			}

			if err == nil {
				tt.Expect(t, model, entity)
			}

			fmt.Printf("\n\n\n")
		})
	}
}

func DBRunUpdateTests(client app.DBClient, t *testing.T, tests []DBTestUpdateData) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			fmt.Printf("Running test: %s\n", tt.Name)
			entity, err := schema.NewEntityFromJSON(tt.InputJSON)
			require.NoError(t, err)

			model, err := client.Model(tt.Schema)
			assert.NoError(t, err)
			require.NotNil(t, model)

			ClearDBData(client, tt.ClearTables...)

			if tt.Prepare != nil {
				tt.Prepare(t, model)
				fmt.Printf("\n")
			}

			runFn := tt.Run
			if runFn == nil {
				runFn = func(model app.Model, entity *schema.Entity) (int, error) {
					mut := model.Mutation()
					if len(tt.Predicates) > 0 {
						mut = mut.Where(tt.Predicates...)
					}
					return mut.Update(entity)
				}
			}

			affected, err := runFn(model, entity)
			require.Equal(t, tt.WantErr, err != nil, err)
			require.Equal(t, tt.WantAffected, affected)

			if err == nil {
				tt.Expect(t, model)
			}

			fmt.Printf("\n\n\n")
		})
	}
}

func DBRunDeleteTests(client app.DBClient, t *testing.T, tests []DBTestDeleteData) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			fmt.Printf("Running test: %s\n", tt.Name)

			model, err := client.Model(tt.Schema)
			assert.NoError(t, err)
			require.NotNil(t, model)
			ClearDBData(client, tt.ClearTables...)

			if tt.Prepare != nil {
				tt.Prepare(t, model)
				fmt.Printf("\n")
			}

			runFn := tt.Run
			if runFn == nil {
				runFn = func(model app.Model) (int, error) {
					mut := model.Mutation()
					if len(tt.Predicates) > 0 {
						mut = mut.Where(tt.Predicates...)
					}
					return mut.Delete()
				}
			}

			affected, err := runFn(model)
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}
			require.Equal(t, tt.WantErr, err != nil, "wantErr != err: "+errMsg)
			require.Equal(t, tt.WantAffected, affected, "affected != wantAffected")

			if err == nil {
				tt.Expect(t, model)
			}

			fmt.Printf("\n\n\n")
		})
	}
}

func MockDefaultQueryRunFn(
	model app.Model,
	predicates []*app.Predicate,
	limit, offset uint,
	order []string,
	columns ...string,
) ([]*schema.Entity, error) {
	query := model.Query()
	if len(predicates) > 0 {
		query = query.Where(predicates...)
	}

	if len(order) > 0 {
		query = query.Order(order...)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if len(columns) > 0 {
		query = query.Select(columns...)
	}

	return query.Get()
}

func DBRunQueryTests(client app.DBClient, t *testing.T, tests []DBTestQueryData) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			fmt.Printf("Running test: %s\n", tt.Name)

			model, err := client.Model(tt.Schema)
			require.NoError(t, err)

			runFn := tt.Run
			if runFn == nil {
				runFn = MockDefaultQueryRunFn
			}

			var predicates []*app.Predicate
			if tt.Filter != "" {
				predicates, err = app.CreatePredicatesFromFilterObject(client.SchemaBuilder(), model.Schema(), tt.Filter)
				require.NoError(t, err)
			}

			ClearDBData(client, tt.ClearTables...)
			var preparedEntities []*schema.Entity
			if tt.Prepare != nil {
				preparedEntities = tt.Prepare(t, client, model)
			}

			entities, err := runFn(model, predicates, tt.Limit, tt.Offset, tt.Order, tt.Columns...)
			if tt.ExpectError == "" {
				assert.NoError(t, err)
				preparedJSONs := utils.Map(preparedEntities, func(e *schema.Entity) map[string]any {
					return e.ToMap()
				})
				entityJSONs := utils.Map(entities, func(e *schema.Entity) map[string]any {
					return e.ToMap()
				})

				if !assert.Equal(t, preparedJSONs, entityJSONs) {
					fmt.Println("------------WANT-----------")
					for _, we := range preparedJSONs {
						fmt.Println(we)
					}
					fmt.Println("------------GOT-----------")
					for _, e := range entityJSONs {
						fmt.Println(e)
					}
				}
				if tt.Expect != nil {
					tt.Expect(t, model, preparedEntities, entities)
				}
			} else {
				assert.Error(t, err)
				require.Equal(t, tt.ExpectError, err.Error(), err)
			}

			fmt.Printf("\n\n\n")
		})
	}
}

func DBRunCountTests(client app.DBClient, t *testing.T, tests []DBTestCountData) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			fmt.Printf("Running test: %s\n", tt.Name)

			model, err := client.Model(tt.Schema)
			require.NoError(t, err)

			runFn := tt.Run
			if runFn == nil {
				runFn = func(
					model app.Model,
					predicates []*app.Predicate,
					unique bool,
					column string,
				) (int, error) {
					query := model.Query()
					if len(predicates) > 0 {
						query = query.Where(predicates...)
					}

					countOptions := &app.CountOption{
						Unique: unique,
						Column: column,
					}

					return query.Count(countOptions)
				}
			}

			var predicates []*app.Predicate
			if tt.Filter != "" {
				predicates, err = app.CreatePredicatesFromFilterObject(client.SchemaBuilder(), model.Schema(), tt.Filter)
				require.NoError(t, err)
			}

			ClearDBData(client, tt.ClearTables...)
			var preparedCount int
			if tt.Prepare != nil {
				preparedCount = tt.Prepare(t, client, model)
			}

			results, err := runFn(model, predicates, tt.Unique, tt.Column)
			if tt.ExpectError == "" {
				assert.NoError(t, err)

				assert.Equal(t, preparedCount, results)

				if tt.Expect != nil {
					tt.Expect(t, model, preparedCount, results)
				}
			} else {
				assert.Error(t, err)
				require.Equal(t, tt.ExpectError, err.Error(), err)
			}

			fmt.Printf("\n\n\n")
		})
	}
}
