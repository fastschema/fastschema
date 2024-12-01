package schema

import (
	"context"
	"testing"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/expr"
	"github.com/stretchr/testify/assert"
)

func createSchema(getter, setter string) *Schema {
	return &Schema{
		Name:           "test",
		Namespace:      "tests",
		LabelFieldName: "field1",
		Fields: []*Field{
			{
				Name:   "field1",
				Type:   TypeString,
				Getter: getter,
				Setter: setter,
			},
		},
	}
}

func TestApplyGetters(t *testing.T) {
	ctx := context.WithValue(context.Background(), "counter", 5)

	// Getter success
	{
		s := createSchema("$context.Value('counter')", "")
		assert.NoError(t, s.Init(false))
		e := entity.New().Set("field1", 0)
		err := s.ApplyGetters(ctx, e, expr.Config{})
		assert.NoError(t, err)
		assert.Equal(t, 5, e.Get("field1"))
	}

	// Getter run error
	{
		s := createSchema("$context.Value('invalid') + 5", "")
		assert.NoError(t, s.Init(false))
		e := entity.New().Set("field1", 0)
		err := s.ApplyGetters(ctx, e, expr.Config{})
		assert.Error(t, err)
	}

	// Getter undefined
	{
		s := createSchema("$undefined", "")
		assert.NoError(t, s.Init(false))
		e := entity.New().Set("field1", 0)
		err := s.ApplyGetters(ctx, e, expr.Config{})
		assert.NoError(t, err)
		val, exist := e.Data().Get("field1")
		assert.False(t, exist)
		assert.Nil(t, val)
	}
}

func TestApplySetters(t *testing.T) {
	ctx := context.WithValue(context.Background(), "counter", 5)

	// Setter success
	{
		s := createSchema("", "$context.Value('counter') *5")
		assert.NoError(t, s.Init(false))
		e := entity.New().Set("field1", 0)
		err := s.ApplySetters(ctx, e, expr.Config{})
		assert.NoError(t, err)
		assert.Equal(t, 25, e.Get("field1"))
	}

	// Setter undefined
	{
		s := createSchema("", "$undefined")
		assert.NoError(t, s.Init(false))
		e := entity.New().Set("field1", 0)
		err := s.ApplySetters(ctx, e, expr.Config{})
		assert.NoError(t, err)
		val, exist := e.Data().Get("field1")
		assert.False(t, exist)
		assert.Nil(t, val)
	}

	// Setter run error
	{
		s := createSchema("", "$args.Value + 'string'")
		assert.NoError(t, s.Init(false))
		e := entity.New().Set("field1", 0)
		err := s.ApplySetters(ctx, e, expr.Config{})
		assert.Error(t, err)
	}

	// Setter relation singlge value
	{
		s := createSchema("", "")
		s.Fields = append(s.Fields, &Field{
			Name:   "relm2m",
			Type:   TypeRelation,
			Setter: "5",
			Relation: &Relation{
				TargetSchemaName: "test2",
				TargetFieldName:  "test2field",
				Type:             O2M,
				Owner:            false,
			},
		})

		assert.NoError(t, s.Init(false))
		e := entity.New()
		err := s.ApplySetters(ctx, e, expr.Config{})
		assert.NoError(t, err)
		relm2m := e.Get("relm2m").(*entity.Entity)
		assert.Equal(t, `{"id":5}`, relm2m.String())
	}

	// Setter relation invalid id
	{
		s := createSchema("", "")
		s.Fields = append(s.Fields, &Field{
			Name:   "relm2m",
			Type:   TypeRelation,
			Setter: "'string'",
			Relation: &Relation{
				TargetSchemaName: "test2",
				TargetFieldName:  "test2field",
				Type:             O2M,
				Owner:            false,
			},
		})

		assert.NoError(t, s.Init(false))
		e := entity.New()
		err := s.ApplySetters(ctx, e, expr.Config{})
		assert.Error(t, err)
	}

	// Setter relation M2M
	{
		s := createSchema("", "")
		s.Fields = append(s.Fields, &Field{
			Name:   "relm2m",
			Type:   TypeRelation,
			Setter: "5",
			Relation: &Relation{
				TargetSchemaName: "test2",
				TargetFieldName:  "test2field",
				Type:             M2M,
			},
		})

		assert.NoError(t, s.Init(false))
		e := entity.New().Set("field1", 0)
		err := s.ApplySetters(ctx, e, expr.Config{})
		assert.NoError(t, err)
		relm2m := e.Get("relm2m").([]*entity.Entity)
		assert.Equal(t, `{"id":5}`, relm2m[0].String())
	}

	// Setter relation O2M owner
	{
		s := createSchema("", "")
		s.Fields = append(s.Fields, &Field{
			Name:   "relm2m",
			Type:   TypeRelation,
			Setter: "5",
			Relation: &Relation{
				TargetSchemaName: "test2",
				TargetFieldName:  "test2field",
				Type:             O2M,
				Owner:            true,
			},
		})

		assert.NoError(t, s.Init(false))
		e := entity.New().Set("field1", 0)
		err := s.ApplySetters(ctx, e, expr.Config{})
		assert.NoError(t, err)
		relm2m := e.Get("relm2m").([]*entity.Entity)
		assert.Equal(t, `{"id":5}`, relm2m[0].String())
	}
}
