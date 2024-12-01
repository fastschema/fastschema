package schema

import (
	"context"
	"fmt"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/expr"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
)

type Undefined struct{}

type GetterArgs struct {
	Schema *Schema
	Entity *entity.Entity
	Value  any
	Exist  bool
}

type SetterArgs struct {
	Schema *Schema
	Entity *entity.Entity
	Value  any
	Exist  bool
}

type SetterProgram = expr.Program[SetterArgs, any]
type GetterProgram = expr.Program[GetterArgs, any]

// ApplyGetters applies the getter programs to the given entity.
func (s *Schema) ApplyGetters(ctx context.Context, e *entity.Entity, configs ...expr.Config) error {
	for _, field := range s.Fields {
		if field.getterProgram == nil {
			continue
		}

		value, exist := e.Data().Get(field.Name)
		getter, err := field.getterProgram.Run(ctx, GetterArgs{
			Schema: s,
			Entity: e,
			Value:  value,
			Exist:  exist,
		}, configs...)
		if err != nil {
			return errors.BadRequest(err.Error())
		}

		if getter.IsUndefined() {
			e.Delete(field.Name)
			continue
		}

		e.Set(field.Name, getter.Raw())
	}

	return nil
}

// ApplySetters applies the setter programs to the given entity.
func (s *Schema) ApplySetters(ctx context.Context, e *entity.Entity, configs ...expr.Config) error {
	for _, field := range s.Fields {
		if field.setterProgram == nil {
			continue
		}

		var setterValue any
		value, exist := e.Data().Get(field.Name)
		setter, err := field.setterProgram.Run(ctx, SetterArgs{
			Schema: s,
			Entity: e,
			Value:  value,
			Exist:  exist,
		}, configs...)

		if err != nil {
			return errors.BadRequest(err.Error())
		}

		if setter.IsUndefined() {
			e.Delete(field.Name)
			continue
		}

		if field.Relation != nil {
			isM2M := field.Relation.Type == M2M
			isO2MOwner := field.Relation.Type == O2M && field.Relation.Owner

			idValue, err := utils.AnyToUint[uint64](setter.Raw())
			if err != nil {
				return fmt.Errorf("invalid setter value %v for field %s: %w", setter, field.Name, err)
			}

			if isM2M || isO2MOwner {
				setterValue = []*entity.Entity{entity.New(idValue)}
			} else {
				setterValue = entity.New(idValue)
			}
		} else {
			setterValue = setter.Raw()
		}

		e.Set(field.Name, setterValue)
	}

	return nil
}
