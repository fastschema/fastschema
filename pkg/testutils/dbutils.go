package testutils

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"database/sql"

	"entgo.io/ent/dialect"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func CreateSchemaBuilder(schemaDir string) *schema.Builder {
	var err error
	var sb *schema.Builder

	if sb, err = schema.NewBuilderFromDir(schemaDir); err != nil {
		panic(err)
	}

	return sb
}

func ClearDBData(client db.Client, tables ...string) {
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

// NewMockClient creates a new mock db Client.
func NewMockClient(
	createMockClient func(db *sql.DB) db.Client,
	s *schema.Builder,
	beforeCreateClient func(m sqlmock.Sqlmock),
	expectTransaction bool,
) (db.Client, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	if expectTransaction {
		mock.ExpectBegin()
	}

	if beforeCreateClient != nil {
		beforeCreateClient(mock)
	}

	if expectTransaction {
		mock.ExpectCommit()
	}

	driver := createMockClient(db)

	return driver, nil
}

func MockRunCreateTests(createMockClient func(db *sql.DB) db.Client, sb *schema.Builder, t *testing.T, tests []MockTestCreateData) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			fmt.Printf("Running test: %s\n", tt.Name)
			entity, err := schema.NewEntityFromJSON(tt.InputJSON)
			require.NoError(t, err)

			client, err := NewMockClient(createMockClient, sb, tt.Expect, tt.Transaction)
			require.NoError(t, err)

			model, err := client.Model(tt.Schema)
			require.NoError(t, err)

			runFn := tt.Run
			if runFn == nil {
				runFn = func(model db.Model, entity *schema.Entity) error {
					_, err := utils.Must(model.Mutation()).Create(entity)
					return err
				}
			}

			err = runFn(model, entity)
			if err != nil {
				assert.Equal(t, tt.ExpectError, err.Error())
			}

			fmt.Printf("\n\n\n")
		})
	}
}

func MockRunUpdateTests(createMockClient func(db *sql.DB) db.Client, sb *schema.Builder, t *testing.T, tests []MockTestUpdateData, extended ...bool) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			fmt.Printf("Running test: %s\n", tt.Name)
			client, err := NewMockClient(createMockClient, sb, tt.Expect, tt.Transaction)
			require.NoError(t, err)
			entity, err := schema.NewEntityFromJSON(tt.InputJSON)
			require.NoError(t, err)

			model, err := client.Model(tt.Schema)
			require.NoError(t, err)
			runFn := tt.Run
			if runFn == nil {
				runFn = func(model db.Model, entity *schema.Entity) (int, error) {
					mut := utils.Must(model.Mutation())
					if len(tt.Predicates) > 0 {
						mut = mut.Where(tt.Predicates...)
					}
					return mut.Update(entity)
				}
			}

			affected, err := runFn(model, entity)
			require.Equal(t, tt.WantErr, err != nil, err)
			if len(extended) > 0 && extended[0] {
				require.Equal(t, tt.WantAffected, affected)
			}
			fmt.Printf("\n\n\n")
		})
	}
}

func MockRunDeleteTests(createMockClient func(db *sql.DB) db.Client, sb *schema.Builder, t *testing.T, tests []MockTestDeleteData, extended ...bool) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			fmt.Printf("Running test: %s\n", tt.Name)
			client, err := NewMockClient(createMockClient, sb, tt.Expect, tt.Transaction)
			require.NoError(t, err)

			model, err := client.Model(tt.Schema)
			require.NoError(t, err)
			runFn := tt.Run
			if runFn == nil {
				runFn = func(model db.Model) (int, error) {
					mut := utils.Must(model.Mutation())
					if len(tt.Predicates) > 0 {
						mut = mut.Where(tt.Predicates...)
					}
					return mut.Delete()
				}
			}

			affected, err := runFn(model)
			require.Equal(t, tt.WantErr, err != nil, err)
			if len(extended) > 0 && extended[0] {
				require.Equal(t, tt.WantAffected, affected)
			}
			fmt.Printf("\n\n\n")
		})
	}
}

func defaultRunFn(
	model db.Model,
	predicates []*db.Predicate,
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

func MockRunQueryTests(
	createMockClient func(db *sql.DB) db.Client,
	sb *schema.Builder,
	t *testing.T,
	tests []MockTestQueryData,
	extended ...bool,
) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			fmt.Printf("Running test: %s\n", tt.Name)
			client, err := NewMockClient(createMockClient, sb, tt.Expect, false)
			require.NoError(t, err)

			model, err := client.Model(tt.Schema)
			require.NoError(t, err)

			runFn := tt.Run
			if runFn == nil {
				runFn = defaultRunFn
			}

			var predicates []*db.Predicate
			if tt.Filter != "" {
				predicates, err = db.CreatePredicatesFromFilterObject(sb, model.Schema(), tt.Filter)
				require.NoError(t, err)
			}

			entities, err := runFn(model, predicates, tt.Limit, tt.Offset, tt.Order, tt.Columns...)
			if tt.ExpectError == "" {
				assert.NoError(t, err)
				expectEntititiesJSONs := make([]string, len(tt.ExpectEntities))
				entitiesJSONs := make([]string, len(entities))
				for i, e := range tt.ExpectEntities {
					expectEntitityJSON, err := e.ToJSON()
					require.NoError(t, err)
					expectEntititiesJSONs[i] = expectEntitityJSON
				}

				for i, e := range entities {
					expectJSON, err := e.ToJSON()
					require.NoError(t, err)
					entitiesJSONs[i] = expectJSON
				}

				if !assert.Equal(t, expectEntititiesJSONs, entitiesJSONs) {
					fmt.Println("------------WANT-----------")
					for _, we := range expectEntititiesJSONs {
						fmt.Println(we)
					}
					fmt.Println("------------GOT-----------")
					for _, e := range entitiesJSONs {
						fmt.Println(e)
					}
				}
			} else {
				assert.Error(t, err)
				require.Equal(t, tt.ExpectError, err.Error(), err)
			}

			fmt.Printf("\n\n\n")
		})
	}
}

func DBRunCreateTests(client db.Client, t *testing.T, tests []DBTestCreateData) {
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
				runFn = func(model db.Model, entity *schema.Entity) (*schema.Entity, error) {
					createdEntityID := utils.Must(model.Create(entity))
					return model.Query(db.EQ("id", createdEntityID)).First()
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

func DBRunUpdateTests(client db.Client, t *testing.T, tests []DBTestUpdateData) {
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
				runFn = func(model db.Model, entity *schema.Entity) (int, error) {
					mut := utils.Must(model.Mutation())
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

func DBRunDeleteTests(client db.Client, t *testing.T, tests []DBTestDeleteData) {
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
				runFn = func(model db.Model) (int, error) {
					mut := utils.Must(model.Mutation())
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

func DBRunQueryTests(client db.Client, t *testing.T, tests []DBTestQueryData) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			fmt.Printf("Running test: %s\n", tt.Name)

			model, err := client.Model(tt.Schema)
			require.NoError(t, err)

			runFn := tt.Run
			if runFn == nil {
				runFn = defaultRunFn
			}

			var predicates []*db.Predicate
			if tt.Filter != "" {
				predicates, err = db.CreatePredicatesFromFilterObject(client.SchemaBuilder(), model.Schema(), tt.Filter)
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

func MockRunCountTests(
	createMockClient func(db *sql.DB) db.Client,
	sb *schema.Builder,
	t *testing.T,
	tests []MockTestCountData,
) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			fmt.Printf("Running test: %s\n", tt.Name)
			client, err := NewMockClient(createMockClient, sb, tt.Expect, false)
			require.NoError(t, err)

			model, err := client.Model(tt.Schema)
			require.NoError(t, err)

			runFn := tt.Run
			if runFn == nil {
				runFn = func(
					model db.Model,
					predicates []*db.Predicate,
					unique bool,
					column string,
				) (int, error) {
					query := model.Query()
					if len(predicates) > 0 {
						query = query.Where(predicates...)
					}
					countOpts := &db.CountOption{
						Unique: unique,
						Column: column,
					}

					return query.Count(countOpts)
				}
			}

			var predicates []*db.Predicate
			if tt.Filter != "" {
				predicates, err = db.CreatePredicatesFromFilterObject(sb, model.Schema(), tt.Filter)
				require.NoError(t, err)
			}

			count, err := runFn(model, predicates, tt.Unique, tt.Column)
			if tt.ExpectError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.ExpectCount, count)
			} else {
				assert.Error(t, err)
				require.Equal(t, tt.ExpectError, err.Error(), err)
			}

			fmt.Printf("\n\n\n")
		})
	}
}

func DBRunCountTests(client db.Client, t *testing.T, tests []DBTestCountData) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			fmt.Printf("Running test: %s\n", tt.Name)

			model, err := client.Model(tt.Schema)
			require.NoError(t, err)

			runFn := tt.Run
			if runFn == nil {
				runFn = func(
					model db.Model,
					predicates []*db.Predicate,
					unique bool,
					column string,
				) (int, error) {
					query := model.Query()
					if len(predicates) > 0 {
						query = query.Where(predicates...)
					}

					countOptions := &db.CountOption{
						Unique: unique,
						Column: column,
					}

					return query.Count(countOptions)
				}
			}

			var predicates []*db.Predicate
			if tt.Filter != "" {
				predicates, err = db.CreatePredicatesFromFilterObject(client.SchemaBuilder(), model.Schema(), tt.Filter)
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
