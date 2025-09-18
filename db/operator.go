package db

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
	OpLike
	OpNotLike
	OpContains
	OpNotContains
	OpContainsFold
	OpNotContainsFold
	OpIN
	OpNIN
	OpNULL
	endOperatorTypes
)

var (
	operatorTypeToStrings = [...]string{
		OpInvalid:         "invalid",
		OpEQ:              "$eq",
		OpNEQ:             "$neq",
		OpGT:              "$gt",
		OpGTE:             "$gte",
		OpLT:              "$lt",
		OpLTE:             "$lte",
		OpLike:            "$like",
		OpNotLike:         "$notlike",
		OpContains:        "$contains",
		OpNotContains:     "$notcontains",
		OpContainsFold:    "$containsfold",
		OpNotContainsFold: "$notcontainsfold",
		OpIN:              "$in",
		OpNIN:             "$nin",
		OpNULL:            "$null",
	}

	stringToOperatorTypes = map[string]OperatorType{
		"invalid":          OpInvalid,
		"$eq":              OpEQ,
		"$neq":             OpNEQ,
		"$gt":              OpGT,
		"$gte":             OpGTE,
		"$lt":              OpLT,
		"$lte":             OpLTE,
		"$like":            OpLike,
		"$notlike":         OpNotLike,
		"$contains":        OpContains,
		"$notcontains":     OpNotContains,
		"$containsfold":    OpContainsFold,
		"$notcontainsfold": OpNotContainsFold,
		"$in":              OpIN,
		"$nin":             OpNIN,
		"$null":            OpNULL,
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
	return &Predicate{And: predicates}
}

func Or(predicates ...*Predicate) *Predicate {
	return &Predicate{Or: predicates}
}

func EQ(field string, value any, relationFields ...string) *Predicate {
	return &Predicate{field, OpEQ, value, relationFields, nil, nil}
}

func NEQ(field string, value any, relationFields ...string) *Predicate {
	return &Predicate{field, OpNEQ, value, relationFields, nil, nil}
}

func GT(field string, value any, relationFields ...string) *Predicate {
	return &Predicate{field, OpGT, value, relationFields, nil, nil}
}

func GTE(field string, value any, relationFields ...string) *Predicate {
	return &Predicate{field, OpGTE, value, relationFields, nil, nil}
}

func LT(field string, value any, relationFields ...string) *Predicate {
	return &Predicate{field, OpLT, value, relationFields, nil, nil}
}

func LTE(field string, value any, relationFields ...string) *Predicate {
	return &Predicate{field, OpLTE, value, relationFields, nil, nil}
}

func Like(field string, value string, relationFields ...string) *Predicate {
	return &Predicate{field, OpLike, value, relationFields, nil, nil}
}

func NotLike(field string, value string, relationFields ...string) *Predicate {
	return &Predicate{field, OpNotLike, value, relationFields, nil, nil}
}

func Contains(field string, value string, relationFields ...string) *Predicate {
	return &Predicate{field, OpContains, value, relationFields, nil, nil}
}

func NotContains(field string, value string, relationFields ...string) *Predicate {
	return &Predicate{field, OpNotContains, value, relationFields, nil, nil}
}

func ContainsFold(field string, value string, relationFields ...string) *Predicate {
	return &Predicate{field, OpContainsFold, value, relationFields, nil, nil}
}

func NotContainsFold(field string, value string, relationFields ...string) *Predicate {
	return &Predicate{field, OpNotContainsFold, value, relationFields, nil, nil}
}

func In[T any](field string, values []T, relationFields ...string) *Predicate {
	return &Predicate{field, OpIN, values, relationFields, nil, nil}
}

func NotIn[T any](field string, values []T, relationFields ...string) *Predicate {
	return &Predicate{field, OpNIN, values, relationFields, nil, nil}
}

func Null(field string, value bool, relationFields ...string) *Predicate {
	return &Predicate{field, OpNULL, value, relationFields, nil, nil}
}

func IsFalse(field string, relationFields ...string) *Predicate {
	return &Predicate{field, OpEQ, false, relationFields, nil, nil}
}

func IsTrue(field string, relationFields ...string) *Predicate {
	return &Predicate{field, OpEQ, true, relationFields, nil, nil}
}
