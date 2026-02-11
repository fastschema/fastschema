package entdbadapter

import (
	"context"
	"database/sql/driver"
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

type Query struct {
	limit            uint
	offset           uint
	fields           []string
	order            []string
	entities         []*entity.Entity
	predicates       []*db.Predicate
	withEdgesFields  []*schema.Field
	edgeDirectSelect map[string]bool
	client           db.Client
	model            *Model
	querySpec        *sqlgraph.QuerySpec
}

func collectEntityIDs(schemaName string, idField *schema.Field, entities []*entity.Entity) ([]driver.Value, map[string]*entity.Entity, error) {
	ids := make([]driver.Value, 0, len(entities))
	byKey := make(map[string]*entity.Entity, len(entities))
	for _, node := range entities {
		var idValue any
		if idField != nil {
			idValue = node.Get(idField.Name)
		}
		if isZeroValue(idValue) {
			idValue = node.ID()
		}
		if isZeroValue(idValue) {
			return nil, nil, fmt.Errorf("entity %s has invalid id", schemaName)
		}
		normalized, err := normalizeIDValue(idField, idValue)
		if err != nil {
			return nil, nil, fmt.Errorf("entity %s has invalid id: %w", schemaName, err)
		}
		key := valueKey(normalized)
		ids = append(ids, normalized)
		byKey[key] = node
	}
	return ids, byKey, nil
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

// Get returns the list of entities that match the query.
func (q *Query) Get(ctx context.Context) (_ []*entity.Entity, err error) {
	var selectFieldNames []string
	directColumnNames := []string{q.model.entIDColumn.Name}
	fkColumns := []string{}
	edgeColumns := map[string][]string{}
	allSelectsAreEdges := true
	option := q.Options()

	if err := runPreDBQueryHooks(ctx, q.client, option); err != nil {
		return nil, err
	}

	if len(q.fields) > 0 {
		var directSelections map[string]bool
		if selectFieldNames, edgeColumns, directSelections, err = q.parseNestedFields(q.fields); err != nil {
			return nil, err
		}
		q.edgeDirectSelect = directSelections

		for _, fieldName := range selectFieldNames {
			column, err := q.model.Column(fieldName)
			if err != nil {
				return nil, err
			}

			if column.field.Type.IsRelationType() {
				relation := column.field.Relation
				q.withEdgesFields = append(q.withEdgesFields, column.field)
				if relation.Type != schema.M2M && !relation.Owner {
					fkColumns = append(fkColumns, relation.SourceColumn)
				}

				if relation.Type != schema.M2M && relation.Owner && relation.BackRef != nil {
					targetColumn := relation.BackRef.TargetColumn
					if targetColumn != "" && targetColumn != q.model.entIDColumn.Name {
						fkColumns = append(fkColumns, targetColumn)
					}
				}
			} else if fieldName != q.model.entIDColumn.Name {
				directColumnNames = append(directColumnNames, fieldName)
				allSelectsAreEdges = false
			}
		}
	} else {
		q.edgeDirectSelect = nil
	}

	directColumnNames = append(directColumnNames, fkColumns...)
	directColumnNames = utils.Unique(directColumnNames)
	entAdapter, ok := q.client.(EntAdapter)
	if !ok {
		return nil, errors.New("client is not an ent adapter")
	}

	builder := sql.Dialect(entAdapter.Driver().Dialect())
	if !allSelectsAreEdges {
		q.querySpec.Node.Columns = directColumnNames
	}
	q.querySpec.From = builder.
		Select(directColumnNames...).
		From(builder.Table(q.model.schema.Namespace))

	if len(q.predicates) > 0 {
		sqlPredicatesFn, err := createEntPredicates(entAdapter, q.model, q.predicates)
		if err != nil {
			return nil, err
		}
		currentPredicate := q.querySpec.Predicate
		q.querySpec.Predicate = func(s *sql.Selector) {
			if currentPredicate != nil {
				currentPredicate(s)
			}

			s.Where(sql.And(sqlPredicatesFn(s)...))
		}
	}

	if len(q.order) > 0 {
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
				return nil, err
			}

			if !column.field.Sortable {
				return nil, fmt.Errorf(`column %q is not sortable`, columnName)
			}

			orderSelectors = append(orderSelectors, func(s *sql.Selector) {
				s.OrderBy(orderFn(s.C(columnName)))
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
	edgeIDs, byID, err := collectEntityIDs(q.model.name, q.model.schema.IDField(), q.entities)
	if err != nil {
		return err
	}

	for _, node := range q.entities {
		node.Set(field.Name, make([]*entity.Entity, 0))
	}

	nids := make(map[string]map[*entity.Entity]struct{})

	edgeQuery := edgeModel.Query()
	entEdgeQuery, ok := edgeQuery.(*Query)
	if !ok {
		return fmt.Errorf(`unexpected edge query type %T`, edgeQuery)
	}

	relation := field.Relation
	conditionColumn := utils.If(relation.IsBidi(), relation.SourceSchemaName, relation.BackRef.SourceFieldName)
	if !relation.IsBidi() && relation.TargetColumn != "" {
		conditionColumn = relation.TargetColumn
	}
	joinColumn := utils.If(relation.IsBidi(), relation.SourceSchemaName, relation.SourceFieldName)
	if relation.SourceColumn != "" {
		joinColumn = relation.SourceColumn
	}

	// Separate direct columns from nested field paths and relation fields
	directColumns := []string{}
	nestedFields := []string{}
	relationFields := []string{}

	for _, col := range edgeColumns {
		if strings.Contains(col, ".") {
			nestedFields = append(nestedFields, col)
			continue
		}

		column, err := edgeModel.Column(col)
		if err != nil {
			return fmt.Errorf("invalid column %q for model %s: %w", col, edgeModel.name, err)
		}

		if column.field.Type.IsRelationType() {
			relationFields = append(relationFields, col)
		} else {
			directColumns = append(directColumns, col)
		}
	}

	entEdgeQuery.querySpec.Predicate = func(s *sql.Selector) {
		joinJuntion := sql.Table(relation.JunctionTable)
		s.Join(joinJuntion).On(joinJuntion.C(joinColumn), s.C(edgeModel.entIDColumn.Name))
		s.Where(sql.InValues(joinJuntion.C(conditionColumn), edgeIDs...))
		selectColumn := relation.BackRef.SourceFieldName
		if !relation.IsBidi() && relation.TargetColumn != "" {
			selectColumn = relation.TargetColumn
		}
		s.Select(joinJuntion.C(selectColumn) + " AS " + selectColumn + "_id")

		// Need complete entity data for recursive loading
		if len(directColumns) == 0 || len(relationFields) > 0 || len(nestedFields) > 0 {
			directColumns = edgeModel.DBColumns()
		}

		if !utils.Contains(directColumns, edgeModel.entIDColumn.Name) {
			directColumns = append([]string{edgeModel.entIDColumn.Name}, directColumns...)
		}

		s.AppendSelect(utils.Map(directColumns, func(c string) string {
			return s.C(c)
		})...)
		s.SetDistinct(false)
	}

	assignFn := entEdgeQuery.querySpec.Assign
	valuesFn := entEdgeQuery.querySpec.ScanValues
	selectColumn := relation.BackRef.SourceFieldName
	if !relation.IsBidi() && relation.TargetColumn != "" {
		selectColumn = relation.TargetColumn
	}

	junctionSchema := relation.JunctionSchema
	if junctionSchema == nil {
		return fmt.Errorf("relation %s.%s missing junction schema", relation.SourceSchemaName, relation.SourceFieldName)
	}
	selectField := junctionSchema.Field(selectColumn)
	if selectField == nil {
		return fmt.Errorf("junction column %s not found for relation %s.%s", selectColumn, relation.SourceSchemaName, relation.SourceFieldName)
	}

	entEdgeQuery.querySpec.ScanValues = func(columns []string) ([]any, error) {
		values, err := valuesFn(columns[1:])
		if err != nil {
			return nil, err
		}
		aliasScan := columnScanValue(selectField.Type)
		return append([]any{aliasScan}, values...), nil
	}

	entEdgeQuery.querySpec.Assign = func(columns []string, values []any) error {
		aliasValue, err := columnAssignValue(selectColumn, selectField.Type, values[0], entity.New())
		if err != nil {
			return err
		}
		if err := assignFn(columns[1:], values[1:]); err != nil {
			return err
		}
		if aliasValue == nil {
			return fmt.Errorf("junction column %s returned nil", selectColumn)
		}
		baseEntity, ok := byID[valueKey(aliasValue)]
		if !ok {
			return fmt.Errorf("no base entity found for junction value %v", aliasValue)
		}
		if len(entEdgeQuery.entities) == 0 {
			return fmt.Errorf("edge assignment missing neighbor entity for %v", aliasValue)
		}
		neighbor := entEdgeQuery.entities[len(entEdgeQuery.entities)-1]
		inKey := valueKey(neighbor.ID())
		if nids[inKey] == nil {
			nids[inKey] = map[*entity.Entity]struct{}{}
		}
		nids[inKey][baseEntity] = struct{}{}
		return nil
	}

	entEdgeQuery.order = []string{edgeModel.entIDColumn.Name}

	// Add nested fields and relation fields to the edge query for recursive processing
	entEdgeQuery.fields = append(entEdgeQuery.fields, nestedFields...)
	entEdgeQuery.fields = append(entEdgeQuery.fields, relationFields...)
	neighbors, err := entEdgeQuery.Get(ctx)
	if err != nil {
		return err
	}

	for _, n := range neighbors {
		key := valueKey(n.ID())
		nodes, ok := nids[key]
		if !ok {
			continue
		}
		for kn := range nodes {
			kn.Set(field.Name, append(kn.Get(field.Name).([]*entity.Entity), n))
		}
		delete(nids, key)
	}

	return nil
}

// loadNonOwnerEdges loads the non-owner edges for the given field.
func (q *Query) loadNonOwnerEdges(ctx context.Context, field *schema.Field, edgeModel *Model, edgeColumns []string) error {
	selectFullEdge := q.edgeDirectSelect != nil && q.edgeDirectSelect[field.Name]
	relation := field.Relation
	edgeSchemaName := relation.TargetSchemaName
	ids := make([]any, 0, len(q.entities))
	nodeids := make(map[string][]*entity.Entity)
	fkColumn := relation.SourceColumn
	targetColumn := relation.TargetColumn
	if targetColumn == "" {
		targetColumn = edgeModel.entIDColumn.Name
	}

	builder := q.client.SchemaBuilder()
	if builder == nil {
		return fmt.Errorf("schema builder is not initialized")
	}

	targetField, err := getRelationTargetField(builder, relation)
	if err != nil {
		return err
	}

	for _, parent := range q.entities {
		fkValue := parent.Get(fkColumn)
		if isZeroValue(fkValue) {
			continue
		}

		normalized, err := normalizeIDValue(targetField, fkValue)
		if err != nil {
			return err
		}

		key := valueKey(normalized)
		if _, ok := nodeids[key]; !ok {
			ids = append(ids, normalized)
		}
		nodeids[key] = append(nodeids[key], parent)
	}

	if len(ids) == 0 {
		return nil
	}

	// Separate direct columns from nested field paths and relation fields
	directColumns := []string{}
	nestedFields := []string{}
	relationFields := []string{}

	for _, col := range edgeColumns {
		if strings.Contains(col, ".") {
			nestedFields = append(nestedFields, col)
			continue
		}

		column, err := edgeModel.Column(col)
		if err != nil {
			return fmt.Errorf("invalid column %q for model %s: %w", col, edgeModel.name, err)
		}

		if column.field.Type.IsRelationType() {
			relationFields = append(relationFields, col)
		} else {
			directColumns = append(directColumns, col)
		}
	}

	if selectFullEdge {
		directColumns = nil
	} else {
		if !utils.Contains(directColumns, edgeModel.entIDColumn.Name) {
			directColumns = append(directColumns, edgeModel.entIDColumn.Name)
		}
		if targetColumn != edgeModel.entIDColumn.Name && !utils.Contains(directColumns, targetColumn) {
			directColumns = append(directColumns, targetColumn)
		}
		directColumns = utils.Unique(directColumns)
	}

	edgeQuery := edgeModel.Query()
	if len(directColumns) > 0 {
		edgeQuery = edgeQuery.Select(directColumns...)
	}

	// Add nested fields and relation fields for recursive processing
	if len(nestedFields) > 0 || len(relationFields) > 0 {
		entEdgeQuery, ok := edgeQuery.(*Query)
		if ok {
			entEdgeQuery.fields = append(entEdgeQuery.fields, nestedFields...)
			entEdgeQuery.fields = append(entEdgeQuery.fields, relationFields...)
		}
	}

	edgeQuery = edgeQuery.Where(db.In(targetColumn, ids))
	neighbors, err := edgeQuery.Get(ctx)
	if err != nil {
		return err
	}

	for _, n := range neighbors {
		var refValue any
		if targetColumn == edgeModel.entIDColumn.Name {
			refValue = n.ID()
		} else {
			refValue = n.Get(targetColumn)
		}

		if isZeroValue(refValue) {
			return invalidFKError(edgeSchemaName, targetColumn, n.ID(), fmt.Errorf("empty reference value"))
		}

		normalized, err := normalizeIDValue(targetField, refValue)
		if err != nil {
			return err
		}

		nodes, ok := nodeids[valueKey(normalized)]
		if !ok {
			return noFKNodeError(q.model.name, edgeSchemaName, fkColumn, n.ID(), refValue)
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
	selectFullEdge := q.edgeDirectSelect != nil && q.edgeDirectSelect[field.Name]
	relation := field.Relation
	edgeSchemaName := relation.TargetSchemaName
	fks := make([]any, 0, len(q.entities))
	nodeids := make(map[string]*entity.Entity)
	fkColumn := relation.BackRef.SourceColumn
	refColumn := relation.BackRef.TargetColumn
	useTargetColumn := refColumn != "" && refColumn != q.model.entIDColumn.Name
	parentField := q.model.schema.IDField()
	if useTargetColumn {
		parentField = q.model.schema.Field(refColumn)
		if parentField == nil {
			return fmt.Errorf("field %s.%s not found", q.model.name, refColumn)
		}
	}
	if parentField == nil {
		return fmt.Errorf("schema %s is missing an id field definition", q.model.name)
	}

	for _, entity := range q.entities {
		var refValue any
		if useTargetColumn {
			refValue = entity.Get(refColumn)
			if isZeroValue(refValue) {
				return invalidFKError(q.model.name, refColumn, entity.ID(), fmt.Errorf("empty reference value"))
			}
		} else {
			refValue = entity.ID()
		}

		normalized, err := normalizeIDValue(parentField, refValue)
		if err != nil {
			return err
		}

		fks = append(fks, normalized)
		nodeids[valueKey(normalized)] = entity
	}

	// Separate direct columns from nested field paths and relation fields
	directColumns := []string{}
	nestedFields := []string{}
	relationFields := []string{}

	for _, col := range edgeColumns {
		if strings.Contains(col, ".") {
			nestedFields = append(nestedFields, col)
			continue
		}

		column, err := edgeModel.Column(col)
		if err != nil {
			return fmt.Errorf("invalid column %q for model %s: %w", col, edgeModel.name, err)
		}

		if column.field.Type.IsRelationType() {
			relationFields = append(relationFields, col)
		} else {
			directColumns = append(directColumns, col)
		}
	}

	if selectFullEdge {
		directColumns = nil
	} else {
		if !utils.Contains(directColumns, edgeModel.entIDColumn.Name) {
			directColumns = append(directColumns, edgeModel.entIDColumn.Name)
		}
		if fkColumn != "" && !utils.Contains(directColumns, fkColumn) {
			directColumns = append(directColumns, fkColumn)
		}
		directColumns = utils.Unique(directColumns)
	}

	edgeQuery := edgeModel.Query()
	if len(directColumns) > 0 {
		edgeQuery = edgeQuery.Select(directColumns...)
	}

	// Add nested fields and relation fields for recursive processing
	if len(nestedFields) > 0 || len(relationFields) > 0 {
		entEdgeQuery, ok := edgeQuery.(*Query)
		if ok {
			entEdgeQuery.fields = append(entEdgeQuery.fields, nestedFields...)
			entEdgeQuery.fields = append(entEdgeQuery.fields, relationFields...)
		}
	}

	neighbors, err := edgeQuery.Where(db.In(fkColumn, fks)).Get(ctx)
	if err != nil {
		return err
	}

	for _, n := range neighbors {
		fkValue := n.Get(fkColumn)
		if isZeroValue(fkValue) {
			return invalidFKError(edgeSchemaName, fkColumn, n.ID(), fmt.Errorf("empty reference value"))
		}

		normalized, err := normalizeIDValue(parentField, fkValue)
		if err != nil {
			return err
		}

		node, ok := nodeids[valueKey(normalized)]
		if !ok {
			return noFKNodeError(q.model.name, edgeSchemaName, fkColumn, n.ID(), fkValue)
		}

		if err := assignFn(node, n); err != nil {
			return err
		}
	}

	return nil
}

func fieldTypeError(fieldType string, fieldValue any) error {
	return fmt.Errorf("expected value of type '%s', got '%T'", fieldType, fieldValue)
}

func invalidFKError(edgeSchemaName, fkColumn string, id any, err error) error {
	return fmt.Errorf(
		`invalid FK value %s.%s for node id=%v: %w`,
		edgeSchemaName, fkColumn, id, err,
	)
}

func noFKNodeError(schemaName, edgeSchemaName, fkColumn string, id, fk any) error {
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
