package app

import (
	"bytes"
	"encoding/json"
)

// OperatorType is the type of the operator
type OperatorType int

const (
	OpInvalid OperatorType = iota
	OpEQ
	OpNEQ
	OpGT
	OpGTE
	OpLT
	OpLTE
	OpLIKE
	OpIN
	OpNIN
	OpNULL
	endOperatorTypes
)

var (
	operatorTypeToStrings = [...]string{
		OpInvalid: "invalid",
		OpEQ:      "$eq",
		OpNEQ:     "$neq",
		OpGT:      "$gt",
		OpGTE:     "$gte",
		OpLT:      "$lt",
		OpLTE:     "$lte",
		OpLIKE:    "$like",
		OpIN:      "$in",
		OpNIN:     "$nin",
		OpNULL:    "$null",
	}

	stringToOperatorTypes = map[string]OperatorType{
		"invalid": OpInvalid,
		"$eq":     OpEQ,
		"$neq":    OpNEQ,
		"$gt":     OpGT,
		"$gte":    OpGTE,
		"$lt":     OpLT,
		"$lte":    OpLTE,
		"$like":   OpLIKE,
		"$in":     OpIN,
		"$nin":    OpNIN,
		"$null":   OpNULL,
	}
)

// String returns the string representation of a type.
func (t OperatorType) String() string {
	if t < endOperatorTypes {
		return operatorTypeToStrings[t]
	}
	return operatorTypeToStrings[OpInvalid]
}

// Valid reports if the given type if known type.
func (t OperatorType) Valid() bool {
	return t > OpInvalid && t < endOperatorTypes
}

// MarshalJSON marshal an enum value to the quoted json string value
func (t OperatorType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(operatorTypeToStrings[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (t *OperatorType) UnmarshalJSON(b []byte) error {
	var j string
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	*t = stringToOperatorTypes[j] // If the string can't be found, it will be set to the zero value: 'invalid'
	return nil
}

func And(predicates ...*Predicate) *Predicate {
	return &Predicate{
		And: predicates,
	}
}

func Or(predicates ...*Predicate) *Predicate {
	return &Predicate{
		Or: predicates,
	}
}

func EQ(field string, value any, relationFields ...string) *Predicate {
	return &Predicate{
		Field:              field,
		Operator:           OpEQ,
		Value:              value,
		RelationFieldNames: relationFields,
	}
}

func NEQ(field string, value any, relationFields ...string) *Predicate {
	return &Predicate{
		Field:              field,
		Operator:           OpNEQ,
		Value:              value,
		RelationFieldNames: relationFields,
	}
}

func GT(field string, value any, relationFields ...string) *Predicate {
	return &Predicate{
		Field:              field,
		Operator:           OpGT,
		Value:              value,
		RelationFieldNames: relationFields,
	}
}

func GTE(field string, value any, relationFields ...string) *Predicate {
	return &Predicate{
		Field:              field,
		Operator:           OpGTE,
		Value:              value,
		RelationFieldNames: relationFields,
	}
}

func LT(field string, value any, relationFields ...string) *Predicate {
	return &Predicate{
		Field:              field,
		Operator:           OpLT,
		Value:              value,
		RelationFieldNames: relationFields,
	}
}

func LTE(field string, value any, relationFields ...string) *Predicate {
	return &Predicate{
		Field:              field,
		Operator:           OpLTE,
		Value:              value,
		RelationFieldNames: relationFields,
	}
}

func Like(field string, value string, relationFields ...string) *Predicate {
	return &Predicate{
		Field:              field,
		Operator:           OpLIKE,
		Value:              value,
		RelationFieldNames: relationFields,
	}
}

func In(field string, values []any, relationFields ...string) *Predicate {
	return &Predicate{
		Field:              field,
		Operator:           OpIN,
		Value:              values,
		RelationFieldNames: relationFields,
	}
}

func NotIn(field string, values []any, relationFields ...string) *Predicate {
	return &Predicate{
		Field:              field,
		Operator:           OpNIN,
		Value:              values,
		RelationFieldNames: relationFields,
	}
}

func Null(field string, value bool, relationFields ...string) *Predicate {
	return &Predicate{
		Field:              field,
		Operator:           OpNULL,
		Value:              value,
		RelationFieldNames: relationFields,
	}
}

func IsFalse(field string, relationFields ...string) *Predicate {
	return &Predicate{
		Field:              field,
		Operator:           OpEQ,
		Value:              false,
		RelationFieldNames: relationFields,
	}
}

func IsTrue(field string, relationFields ...string) *Predicate {
	return &Predicate{
		Field:              field,
		Operator:           OpEQ,
		Value:              true,
		RelationFieldNames: relationFields,
	}
}
