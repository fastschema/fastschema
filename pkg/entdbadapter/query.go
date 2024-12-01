package entdbadapter

import (
	"context"
	"database/sql/driver"
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

type Query struct {
	limit           uint
	offset          uint
	fields          []string
	order           []string
	entities        []*entity.Entity
	predicates      []*db.Predicate
	withEdgesFields []*schema.Field
	client          db.Client
	model           *Model
	querySpec       *sqlgraph.QuerySpec
}

func (q *Query) Options() *db.QueryOption {
	return &db.QueryOption{
		Limit:      q.limit,
		Offset:     q.offset,
		Columns:    q.fields,
		Order:      q.order,
		Predicates: q.predicates,
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

// Count returns the number of entities that match the query.
func (q *Query) Count(ctx context.Context, options ...*db.QueryOption) (int, error) {
	option := append(options, &db.QueryOption{})[0]
	if option == nil {
		option = &db.QueryOption{}
	}

	entAdapter, ok := q.client.(EntAdapter)
	if !ok {
		return 0, fmt.Errorf("client is not an ent adapter")
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

// Only returns the matched entity or an error if there is more than one.
func (q *Query) Only(ctx context.Context) (*entity.Entity, error) {
	entities, err := q.Get(ctx)

	if err != nil {
		return nil, err
	}

	if len(entities) > 1 {
		return nil, fmt.Errorf("more than one entity found")
	}

	if len(entities) == 0 {
		return nil, &db.NotFoundError{Message: "no entities found"}
	}

	return entities[0], nil
}

// Get returns the list of entities that match the query.
func (q *Query) Get(ctx context.Context) ([]*entity.Entity, error) {
	columnNames := []string{}
	edgeColumns := map[string][]string{}
	allSelectsAreEdges := true
	option := q.Options()

	if err := runPreDBQueryHooks(ctx, q.client, option); err != nil {
		return nil, err
	}

	if len(q.fields) > 0 {
		columnNames = append(columnNames, q.model.entIDColumn.Name)

		for _, columnName := range q.fields {
			if strings.Contains(columnName, ".") {
				parts := strings.Split(columnName, ".")
				if len(parts) != 2 {
					return nil, fmt.Errorf(`invalid column name %q`, columnName)
				}

				columnName = parts[0]
				edgeColumns[columnName] = append(edgeColumns[columnName], parts[1])
			}

			column, err := q.model.Column(columnName)
			if err != nil {
				return nil, err
			}

			if column.field.Type.IsRelationType() {
				existedEdgeFields := utils.Filter(q.withEdgesFields, func(f *schema.Field) bool {
					return f.Name == column.field.Name
				})

				if len(existedEdgeFields) > 0 {
					continue
				}

				relation := column.field.Relation
				q.withEdgesFields = append(q.withEdgesFields, column.field)
				if relation.Type != schema.M2M && !relation.Owner {
					columnNames = append(columnNames, relation.GetTargetFKColumn())
				}
				continue
			}

			if columnName != q.model.entIDColumn.Name {
				columnNames = append(columnNames, columnName)
				allSelectsAreEdges = false
			}
		}
	}

	entAdapter, ok := q.client.(EntAdapter)
	if !ok {
		return nil, fmt.Errorf("client is not an ent adapter")
	}

	builder := sql.Dialect(entAdapter.Driver().Dialect())
	if !allSelectsAreEdges {
		q.querySpec.Node.Columns = columnNames
	}
	q.querySpec.From = builder.
		Select(columnNames...).
		From(builder.Table(q.model.schema.Namespace))

	if len(q.predicates) > 0 {
		sqlPredicatesFn, err := createEntPredicates(entAdapter, q.model, q.predicates)
		if err != nil {
			return nil, err
		}
		q.querySpec.Predicate = func(s *sql.Selector) {
			s.Where(sql.And(sqlPredicatesFn(s)...))
		}
	}

	if len(q.order) > 0 {
		orderSelectors := []func(*sql.Selector){}

		for _, order := range q.order {
			orderFn := sql.Asc
			columnName := order

			if strings.HasPrefix(order, "-") {
				columnName = strings.TrimPrefix(order, "-")
				orderFn = sql.Desc
			}

			column, err := q.model.Column(columnName)
			if err != nil {
				return nil, err
			}

			if !column.field.Sortable {
				return nil, fmt.Errorf(`column %q is not sortable`, columnName)
			}

			orderSelectors = append(orderSelectors, func(s *sql.Selector) {
				s.OrderBy(orderFn(columnName))
			})
		}

		q.querySpec.Order = func(s *sql.Selector) {
			for _, orderSelector := range orderSelectors {
				orderSelector(s)
			}
		}
	}

	if q.limit > 0 {
		q.querySpec.Limit = int(q.limit)
	}

	if q.offset > 0 {
		q.querySpec.Offset = int(q.offset)
	}

	if err := sqlgraph.QueryNodes(ctx, entAdapter.Driver(), q.querySpec); err != nil {
		return nil, err
	}

	if err := q.loadEdges(ctx, edgeColumns); err != nil {
		return nil, err
	}

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

// loadEdges loads the edges for the given fields.
func (q *Query) loadEdges(ctx context.Context, edgesColumns map[string][]string) error {
	for _, edgeField := range q.withEdgesFields {
		relation := edgeField.Relation
		edgeModel, err := q.client.Model(relation.TargetSchemaName)
		if err != nil {
			return err
		}

		edgeEntModel, ok := edgeModel.(*Model)
		if !ok {
			return fmt.Errorf(`unexpected model type %T`, edgeModel)
		}

		edgeColumns := edgesColumns[edgeField.Name]

		if relation.Type == schema.O2M {
			if err := q.loadEdgesO2M(ctx, edgeField, edgeEntModel, edgeColumns); err != nil {
				return err
			}
		}

		if relation.Type == schema.O2O {
			if err := q.loadEdgesO2O(ctx, edgeField, edgeEntModel, edgeColumns); err != nil {
				return err
			}
		}

		if relation.Type == schema.M2M {
			if err := q.loadEdgesM2M(ctx, edgeField, edgeEntModel, edgeColumns); err != nil {
				return err
			}
		}
	}
	return nil
}

// loadEdgesO2M loads the one-to-many edges for the given field.
func (q *Query) loadEdgesO2M(
	ctx context.Context,
	field *schema.Field,
	edgeModel *Model,
	edgeColumns []string,
) error {
	if !field.Relation.Owner {
		return q.loadNonOwnerEdges(ctx, field, edgeModel, edgeColumns)
	}

	return q.loadOwnerEdges(
		ctx, field, edgeModel, edgeColumns,
		func(node *entity.Entity, neighbor *entity.Entity) error {
			edgeValues := node.Get(field.Name)
			if edgeValues == nil {
				node.Set(field.Name, []*entity.Entity{neighbor})
				return nil
			}

			edgeEntities, ok := edgeValues.([]*entity.Entity)
			if !ok {
				return invalidEntityArrayError(q.model.name, field.Name, edgeValues)
			}

			edgeEntities = append(edgeEntities, neighbor)
			node.Set(field.Name, edgeEntities)
			return nil
		},
	)
}

// loadEdgesO2O loads the one-to-one edges for the given field.
func (q *Query) loadEdgesO2O(
	ctx context.Context,
	field *schema.Field,
	edgeModel *Model,
	edgeColumns []string,
) error {
	if !field.Relation.Owner {
		return q.loadNonOwnerEdges(ctx, field, edgeModel, edgeColumns)
	}

	return q.loadOwnerEdges(
		ctx, field, edgeModel, edgeColumns,
		func(node *entity.Entity, neighbor *entity.Entity) error {
			node.Set(field.Name, neighbor)
			return nil
		},
	)
}

// loadEdgesM2M loads the many-to-many edges for the given field.
func (q *Query) loadEdgesM2M(
	ctx context.Context,
	field *schema.Field,
	edgeModel *Model,
	edgeColumns []string,
) error {
	edgeIDs := make([]driver.Value, len(q.entities))
	byID := make(map[uint64]*entity.Entity)
	nids := make(map[uint64]map[*entity.Entity]struct{})
	for i, node := range q.entities {
		edgeIDs[i] = node.ID()
		byID[node.ID()] = node
		node.Set(field.Name, make([]*entity.Entity, 0))
	}

	edgeQuery := edgeModel.Query()
	entEdgeQuery, ok := edgeQuery.(*Query)
	if !ok {
		return fmt.Errorf(`unexpected edge query type %T`, edgeQuery)
	}

	relation := field.Relation
	conditionColumn := utils.If(relation.IsBidi(), relation.SchemaName, relation.BackRef.FieldName)
	joinColumn := utils.If(relation.IsBidi(), relation.SchemaName, relation.FieldName)
	entEdgeQuery.querySpec.Predicate = func(s *sql.Selector) {
		joinJuntion := sql.Table(relation.JunctionTable)
		s.Join(joinJuntion).On(joinJuntion.C(joinColumn), s.C(edgeModel.entIDColumn.Name))
		s.Where(sql.InValues(joinJuntion.C(conditionColumn), edgeIDs...))
		s.Select(joinJuntion.C(relation.BackRef.FieldName) + " AS " + relation.BackRef.FieldName + "_id")

		if len(edgeColumns) == 0 {
			edgeColumns = edgeModel.DBColumns()
		}
		if !utils.Contains(edgeColumns, edgeModel.entIDColumn.Name) {
			edgeColumns = append([]string{edgeModel.entIDColumn.Name}, edgeColumns...)
		}

		s.AppendSelect(utils.Map(edgeColumns, func(c string) string {
			return s.C(c)
		})...)
		s.SetDistinct(false)
	}

	assignFn := entEdgeQuery.querySpec.Assign
	valuesFn := entEdgeQuery.querySpec.ScanValues
	entEdgeQuery.querySpec.ScanValues = func(columns []string) ([]any, error) {
		values, err := valuesFn(columns[1:])
		if err != nil {
			return nil, err
		}
		return append([]any{new(sql.NullInt64)}, values...), nil
	}

	entEdgeQuery.querySpec.Assign = func(columns []string, values []any) error {
		outValue := uint64(values[0].(*sql.NullInt64).Int64)
		inValue := uint64(values[1].(*sql.NullInt64).Int64)
		if nids[inValue] == nil {
			nids[inValue] = map[*entity.Entity]struct{}{byID[outValue]: {}}
			return assignFn(columns[1:], values[1:])
		}
		nids[inValue][byID[outValue]] = struct{}{}
		return nil
	}

	entEdgeQuery.order = []string{edgeModel.entIDColumn.Name}
	neighbors, err := entEdgeQuery.Get(ctx)
	if err != nil {
		return err
	}

	for _, n := range neighbors {
		nodes, ok := nids[n.ID()]
		if !ok {
			continue
		}
		for kn := range nodes {
			kn.Set(field.Name, append(kn.Get(field.Name).([]*entity.Entity), n))
		}
	}

	return nil
}

// loadNonOwnerEdges loads the non-owner edges for the given field.
func (q *Query) loadNonOwnerEdges(ctx context.Context, field *schema.Field, edgeModel *Model, edgeColumns []string) error {
	relation := field.Relation
	edgeSchemaName := relation.TargetSchemaName
	ids := make([]any, 0, len(q.entities))
	nodeids := make(map[uint64][]*entity.Entity)
	fkColumn := relation.GetTargetFKColumn()

	for _, entity := range q.entities {
		fkUint64, err := entity.GetUint64(fkColumn, true)
		if err != nil {
			return invalidFKError(edgeSchemaName, fkColumn, entity.ID(), err)
		}

		if fkUint64 == 0 {
			continue
		}

		if _, ok := nodeids[fkUint64]; !ok {
			ids = append(ids, fkUint64)
		}
		nodeids[fkUint64] = append(nodeids[fkUint64], entity)
	}

	if len(ids) == 0 {
		return nil
	}

	edgeQuery := edgeModel.Query().Select(edgeColumns...).Where(db.In(edgeModel.entIDColumn.Name, ids))
	neighbors, err := edgeQuery.Get(ctx)
	if err != nil {
		return err
	}

	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID()]
		if !ok {
			return noFKNodeError(q.model.name, edgeSchemaName, fkColumn, n.ID(), 0)
		}

		for i := range nodes {
			nodes[i].Set(field.Name, n)
		}
	}

	return nil
}

// loadOwnerEdges loads the owner edges for the given field.
func (q *Query) loadOwnerEdges(
	ctx context.Context,
	field *schema.Field,
	edgeModel *Model,
	edgeColumns []string,
	assignFn func(node, neighbor *entity.Entity) error,
) error {
	relation := field.Relation
	edgeSchemaName := relation.TargetSchemaName
	fks := make([]any, 0, len(q.entities))
	nodeids := make(map[uint64]*entity.Entity)
	fkColumn := relation.BackRef.GetTargetFKColumn()

	for _, entity := range q.entities {
		entityID := entity.ID()
		fks = append(fks, entityID)
		nodeids[entityID] = entity
	}

	if len(edgeColumns) > 0 && !utils.Contains(edgeColumns, fkColumn) {
		edgeColumns = append(edgeColumns, fkColumn)
	}

	neighbors, err := edgeModel.Query().Select(edgeColumns...).Where(db.In(fkColumn, fks)).Get(ctx)
	if err != nil {
		return err
	}

	for _, n := range neighbors {
		fkUint64, err := n.GetUint64(fkColumn, false)
		if err != nil {
			return invalidFKError(edgeSchemaName, fkColumn, n.ID(), err)
		}

		node, ok := nodeids[fkUint64]
		if !ok {
			return noFKNodeError(q.model.name, edgeSchemaName, fkColumn, n.ID(), fkUint64)
		}

		if err := assignFn(node, n); err != nil {
			return err
		}
	}

	return nil
}

func fieldTypeError(fieldType string, fieldValue any) error {
	return fmt.Errorf("unexpected type %T for field type %s", fieldValue, fieldType)
}

func invalidFKError(edgeSchemaName, fkColumn string, id uint64, err error) error {
	return fmt.Errorf(
		`invalid FK value %s.%s for node id=%d: %w`,
		edgeSchemaName, fkColumn, id, err,
	)
}

func noFKNodeError(schemaName, edgeSchemaName, fkColumn string, id, fk uint64) error {
	return fmt.Errorf(
		`no FK node (%s) found for (%s=%v).%s=%v`,
		schemaName, edgeSchemaName, id, fkColumn, fk,
	)
}

func invalidEntityArrayError(schemaName, fieldName string, edgeValues any) error {
	return fmt.Errorf(
		`edge values %s.%s=%v (%T) is not []*entity.Entity`,
		schemaName, fieldName, edgeValues, edgeValues,
	)
}
