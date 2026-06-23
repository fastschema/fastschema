package entdbadapter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/expr"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

// perParentLimitConfig holds configuration for per-parent limit/offset using window functions.
type perParentLimitConfig struct {
	partitionColumn string // Column to partition by (e.g., FK column)
	limit           uint
	offset          uint
}

type Query struct {
	limit           uint
	offset          uint
	fields          []string
	order           []string
	entities        []*entity.Entity
	predicates      []*db.Predicate
	relationOptions db.RelationOptions
	client          db.Client
	model           *Model
	querySpec       *sqlgraph.QuerySpec
	perParentLimit  *perParentLimitConfig // For per-parent limit/offset in edge queries
}

func (q *Query) WithTrashed() db.Querier {
	if !q.client.Config().UseSoftDeletes {
		return q
	}

	// if soft deletes are enabled, predicates[0] is always "deleted_at IS NULL"
	// we need to remove it to allow querying trashed entities
	if len(q.predicates) > 0 && q.predicates[0].Field == "deleted_at" {
		q.predicates = q.predicates[1:]
	}

	return q
}

func (q *Query) OnlyTrashed() db.Querier {
	if !q.client.Config().UseSoftDeletes {
		return q
	}

	// if soft deletes are enabled, predicates[0] is always "deleted_at IS NULL"
	// we need to replace it with "deleted_at IS NOT NULL"
	if len(q.predicates) > 0 && q.predicates[0].Field == "deleted_at" {
		q.predicates[0] = db.Null("deleted_at", false)
	} else {
		q.predicates = append([]*db.Predicate{db.Null("deleted_at", true)}, q.predicates...)
	}

	return q
}

func (q *Query) Options() *db.QueryOption {
	return &db.QueryOption{
		Limit:      q.limit,
		Offset:     q.offset,
		Columns:    &q.fields,
		Order:      q.order,
		Predicates: &q.predicates,
		Schema:     q.model.schema,
	}
}

// Limit sets the limit of the query.
func (q *Query) Limit(limit uint) db.Querier {
	q.limit = limit
	return q
}

// Offset sets the offset of the query.
func (q *Query) Offset(offset uint) db.Querier {
	q.offset = offset
	return q
}

// Order sets the order of the query.
func (q *Query) Order(order ...string) db.Querier {
	q.order = append(q.order, order...)
	return q
}

// Select sets the columns of the query.
func (q *Query) Select(fields ...string) db.Querier {
	q.fields = append(q.fields, fields...)
	return q
}

// Where adds the given predicates to the query.
func (q *Query) Where(predicates ...*db.Predicate) db.Querier {
	q.predicates = append(q.predicates, predicates...)
	return q
}

// WithRelationOptions sets options for loading relation records.
func (q *Query) WithRelationOptions(options db.RelationOptions) db.Querier {
	q.relationOptions = options
	return q
}

// Count returns the number of entities that match the query.
func (q *Query) Count(ctx context.Context, options ...*db.QueryOption) (int, error) {
	option := append(options, &db.QueryOption{})[0]
	if option == nil {
		option = &db.QueryOption{}
	}

	entAdapter, ok := q.client.(EntAdapter)
	if !ok {
		return 0, errors.New("client is not an ent adapter")
	}

	opts := q.Options()
	opts.Column = option.Column
	opts.Unique = option.Unique

	if err := runPreDBQueryHooks(ctx, q.client, opts); err != nil {
		return 0, err
	}

	if opts != nil {
		q.querySpec.Unique = opts.Unique
		if opts.Column != "" {
			q.querySpec.Node.Columns = []string{opts.Column}
		}
	}

	if len(q.predicates) > 0 {
		sqlPredicatesFn, err := createEntPredicates(entAdapter, q.model, q.predicates)
		if err != nil {
			return 0, err
		}
		q.querySpec.Predicate = func(s *sql.Selector) {
			s.Where(sql.And(sqlPredicatesFn(s)...))
		}
	}

	count, err := sqlgraph.CountNodes(ctx, entAdapter.Driver(), q.querySpec)
	if err != nil {
		return 0, err
	}

	_, err = runPostDBQueryHooks(ctx, q.client, opts, []*entity.Entity{
		entity.New().Set("count", count),
	})

	return count, err
}

// First returns the first entity that matches the query.
// Returns NotFoundError if no entity was found.
func (q *Query) First(ctx context.Context) (*entity.Entity, error) {
	q.Limit(1)
	entities, err := q.Get(ctx)

	if err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		return nil, &db.NotFoundError{Message: "no entities found"}
	}

	return entities[0], nil
}

// Only returns the only entity that matches the query.
// Returns NotFoundError if no or more than one entity was found.
func (q *Query) Only(ctx context.Context) (*entity.Entity, error) {
	entities, err := q.Get(ctx)

	if err != nil {
		return nil, err
	}

	if len(entities) > 1 {
		return nil, errors.New("more than one entity found")
	}

	if len(entities) == 0 {
		return nil, &db.NotFoundError{Message: "no entities found"}
	}

	return entities[0], nil
}

func (q *Query) parseNestedFields(fields []string) ([]string, map[string][]string, map[string]bool, error) {
	edgeColumns := map[string][]string{}
	processedFields := []string{}
	directSelections := map[string]bool{}

	for _, originalField := range fields {
		if !strings.Contains(originalField, ".") {
			processedFields = append(processedFields, originalField)
			directSelections[originalField] = true
			continue
		}

		// Get the first part of the field path and the remaining path
		dotIndex := strings.Index(originalField, ".")
		if dotIndex == 0 || dotIndex == len(originalField)-1 {
			return nil, nil, nil, fmt.Errorf(`invalid column name %q`, originalField)
		}

		// The first part is the edge name, and the remaining path is the nested field
		firstField := originalField[:dotIndex]
		remainingPath := originalField[dotIndex+1:]
		processField := firstField
		// The remaining path will be processed recursively by the edge loader's Get() call
		edgeColumns[firstField] = utils.Unique(append(edgeColumns[firstField], remainingPath))

		processedFields = append(processedFields, processField)
	}

	return utils.Unique(processedFields), edgeColumns, directSelections, nil
}

// edgeSelection holds both the edge field and its nested columns.
// If columns is nil, all columns are selected.
type edgeSelection struct {
	field   *schema.Field
	columns []string // nested columns to select; nil means select all
}

// queryBuildResult holds the result of building query columns.
type queryBuildResult struct {
	directColumnNames  []string
	fkColumns          []string
	edges              map[string]*edgeSelection
	allSelectsAreEdges bool
}

// buildQueryColumns processes the selected fields and separates them into direct columns,
// FK columns, and edge columns for relation loading.
func (q *Query) buildQueryColumns() (*queryBuildResult, error) {
	result := &queryBuildResult{
		directColumnNames:  []string{q.model.entPrimaryColumn.Name},
		fkColumns:          []string{},
		edges:              map[string]*edgeSelection{},
		allSelectsAreEdges: true,
	}

	if len(q.fields) == 0 {
		return result, nil
	}

	selectFieldNames, edgeColumns, directSelections, err := q.parseNestedFields(q.fields)
	if err != nil {
		return nil, err
	}

	for _, fieldName := range selectFieldNames {
		column, err := q.model.Column(fieldName)
		if err != nil {
			return nil, err
		}

		if column.field.Type.IsRelationType() {
			relation := column.field.Relation
			// If edge was directly selected (e.g., "tags"), columns is nil (select all)
			// Otherwise, columns contains specific nested fields (e.g., from "tags.name")
			var columns []string
			if !directSelections[fieldName] {
				columns = edgeColumns[fieldName]
			}
			result.edges[fieldName] = &edgeSelection{
				field:   column.field,
				columns: columns,
			}
			if relation.Type != schema.M2M && !relation.Owner {
				result.fkColumns = append(result.fkColumns, relation.SourceColumn)
			}

			if relation.Type != schema.M2M && relation.Owner && relation.BackRef != nil {
				targetColumn := relation.BackRef.TargetColumn
				if targetColumn != "" && targetColumn != q.model.entPrimaryColumn.Name {
					result.fkColumns = append(result.fkColumns, targetColumn)
				}
			}
		} else if fieldName != q.model.entPrimaryColumn.Name {
			result.directColumnNames = append(result.directColumnNames, fieldName)
			result.allSelectsAreEdges = false
		}
	}

	return result, nil
}

// buildQueryPredicates builds the predicate function for the query spec.
func (q *Query) buildQueryPredicates(entAdapter EntAdapter) error {
	if len(q.predicates) == 0 {
		return nil
	}

	sqlPredicatesFn, err := createEntPredicates(entAdapter, q.model, q.predicates)
	if err != nil {
		return err
	}

	currentPredicate := q.querySpec.Predicate
	q.querySpec.Predicate = func(s *sql.Selector) {
		if currentPredicate != nil {
			currentPredicate(s)
		}
		s.Where(sql.And(sqlPredicatesFn(s)...))
	}

	return nil
}

// buildQueryOrder builds the order function for the query spec.
func (q *Query) buildQueryOrder() error {
	if len(q.order) == 0 {
		return nil
	}

	orderSelectors := []func(*sql.Selector){}

	for _, order := range q.order {
		orderFn := sql.Asc
		columnName := order

		if after, ok := strings.CutPrefix(order, "-"); ok {
			columnName = after
			orderFn = sql.Desc
		}

		column, err := q.model.Column(columnName)
		if err != nil {
			return err
		}

		if !column.field.Sortable {
			return fmt.Errorf(`column %q is not sortable`, columnName)
		}

		// Capture columnName and orderFn for closure
		colName, ordFn := columnName, orderFn
		orderSelectors = append(orderSelectors, func(s *sql.Selector) {
			s.OrderBy(ordFn(s.C(colName)))
		})
	}

	q.querySpec.Order = func(s *sql.Selector) {
		for _, orderSelector := range orderSelectors {
			orderSelector(s)
		}
	}

	return nil
}

// Get returns the list of entities that match the query.
func (q *Query) Get(ctx context.Context) (_ []*entity.Entity, err error) {
	option := q.Options()

	if err := runPreDBQueryHooks(ctx, q.client, option); err != nil {
		return nil, err
	}

	// Build query columns
	buildResult, err := q.buildQueryColumns()
	if err != nil {
		return nil, err
	}

	// Combine direct and FK columns
	buildResult.directColumnNames = append(buildResult.directColumnNames, buildResult.fkColumns...)
	allColumns := utils.Unique(buildResult.directColumnNames)

	entAdapter, ok := q.client.(EntAdapter)
	if !ok {
		return nil, errors.New("client is not an ent adapter")
	}

	// Use window function query for per-parent limit/offset
	if q.perParentLimit != nil {
		// For window function queries, we need explicit columns.
		// If no specific columns were selected (SELECT *), use all model columns.
		if len(q.fields) == 0 {
			allColumns = q.model.DBColumns()
		}
		return q.getWithPerParentLimit(ctx, entAdapter, allColumns, buildResult)
	}

	// Build query spec
	builder := sql.Dialect(entAdapter.Driver().Dialect())
	if !buildResult.allSelectsAreEdges {
		q.querySpec.Node.Columns = allColumns
	}
	q.querySpec.From = builder.
		Select(allColumns...).
		From(builder.Table(q.model.schema.Namespace))

	// Build predicates
	if err := q.buildQueryPredicates(entAdapter); err != nil {
		return nil, err
	}

	// Build order
	if err := q.buildQueryOrder(); err != nil {
		return nil, err
	}

	// Apply limit and offset
	if q.limit > 0 {
		q.querySpec.Limit = int(q.limit)
	}
	if q.offset > 0 {
		q.querySpec.Offset = int(q.offset)
	}

	// Execute query
	if err := sqlgraph.QueryNodes(ctx, entAdapter.Driver(), q.querySpec); err != nil {
		return nil, err
	}

	// Load edges
	if err := q.loadEdges(ctx, buildResult.edges); err != nil {
		return nil, err
	}

	// Apply getters
	for _, entity := range q.entities {
		if err := q.model.schema.ApplyGetters(ctx, entity, expr.Config{
			DB: func() expr.DBLike {
				return entAdapter
			},
		}); err != nil {
			return nil, err
		}
	}

	return runPostDBQueryHooks(ctx, q.client, option, q.entities)
}

// getWithPerParentLimit executes a query with per-parent limit/offset using window functions.
// It generates SQL like:
//
//	SELECT * FROM (
//	  SELECT *, ROW_NUMBER() OVER (PARTITION BY partition_col ORDER BY order_col) AS row_num
//	  FROM table WHERE ...
//	) AS ranked
//	WHERE row_num > offset AND row_num <= offset + limit
func (q *Query) getWithPerParentLimit(
	ctx context.Context,
	entAdapter EntAdapter,
	allColumns []string,
	buildResult *queryBuildResult,
) ([]*entity.Entity, error) {
	cfg := q.perParentLimit
	builder := sql.Dialect(entAdapter.Driver().Dialect())
	table := builder.Table(q.model.schema.Namespace)

	// Build the inner query with all columns plus ROW_NUMBER()
	inner := builder.Select(allColumns...).From(table)

	// Build window function ORDER BY
	orderCols := q.order
	if len(orderCols) == 0 {
		orderCols = []string{q.model.entPrimaryColumn.Name}
	}

	windowFn := sql.RowNumber().PartitionBy(cfg.partitionColumn)
	for _, col := range orderCols {
		if after, ok := strings.CutPrefix(col, "-"); ok {
			windowFn = windowFn.OrderBy(after + " DESC")
		} else {
			windowFn = windowFn.OrderBy(col)
		}
	}
	inner.AppendSelectExprAs(windowFn, "row_num")

	// Apply predicates to inner query
	if len(q.predicates) > 0 {
		sqlPredicatesFn, err := createEntPredicates(entAdapter, q.model, q.predicates)
		if err != nil {
			return nil, err
		}
		inner.Where(sql.And(sqlPredicatesFn(inner)...))
	}

	// Alias the inner query
	inner.As("ranked")

	// Build outer query with row_num filter
	outer := builder.Select(allColumns...).From(inner)

	// Apply row_num conditions for per-parent limit/offset
	if cfg.offset > 0 {
		outer.Where(sql.GT("row_num", cfg.offset))
	}
	if cfg.limit > 0 {
		maxRowNum := cfg.offset + cfg.limit
		outer.Where(sql.LTE("row_num", maxRowNum))
	}

	// Apply ordering to outer query (maintain order within each partition)
	for _, col := range orderCols {
		if after, ok := strings.CutPrefix(col, "-"); ok {
			outer.OrderBy(sql.Desc(after))
		} else {
			outer.OrderBy(sql.Asc(col))
		}
	}

	// Execute query
	query, args := outer.Query()
	entities, err := driverQuery(entAdapter.Driver(), ctx, query, args)
	if err != nil {
		return nil, err
	}

	q.entities = entities

	// Load edges
	if err := q.loadEdges(ctx, buildResult.edges); err != nil {
		return nil, err
	}

	// Apply getters
	for _, entity := range q.entities {
		if err := q.model.schema.ApplyGetters(ctx, entity, expr.Config{
			DB: func() expr.DBLike {
				return entAdapter
			},
		}); err != nil {
			return nil, err
		}
	}

	option := q.Options()
	return runPostDBQueryHooks(ctx, q.client, option, q.entities)
}
