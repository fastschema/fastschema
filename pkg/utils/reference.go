package utils

import (
	"reflect"
)

// GetDereferencedType returns the dereferenced type of a value 'v'
func GetDereferencedType(v any) reflect.Type {
	if reflect.TypeOf(v) == nil {
		return nil
	}

	// v can be a pointer that points to its own address,
	// so we need to check if v is a pointer to itself to avoid infinite loops.
	var originalAddress uintptr
	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		originalAddress = reflect.ValueOf(v).Pointer()
	}

	for reflect.TypeOf(v).Kind() == reflect.Ptr {
		if reflect.ValueOf(v).IsZero() {
			v = CreateZeroValue(reflect.TypeOf(v).Elem())
			continue
		}

		v = reflect.ValueOf(v).Elem().Interface()

		if reflect.TypeOf(v).Kind() == reflect.Ptr {
			newAddress := reflect.ValueOf(v).Pointer()
			if originalAddress == newAddress {
				break
			}
		}
	}

	return reflect.Indirect(reflect.ValueOf(v)).Type()
}

// CreateZeroValue creates a zero value of a type.
func CreateZeroValue(t reflect.Type) any {
	return reflect.New(t).Elem().Interface()
}

// GeneratePointerChain creates a chain of pointers by taking the address of the original value
// 'v' 'times' number of times. It returns the final value in the pointer chain.
func GeneratePointerChain(v any, times int) any {
	var stack = []any{v}
	for i := 0; i < times; i++ {
		stack = append(stack, &stack[len(stack)-1])
	}

	return stack[len(stack)-1]
}

// Dereferenceable returns true if the value is dereferenceable.
func Dereferenceable(v any) bool {
	for reflect.TypeOf(v) != nil && reflect.TypeOf(v).Kind() == reflect.Ptr {
		if reflect.ValueOf(v).IsZero() {
			v = reflect.New(reflect.TypeOf(v).Elem()).Elem().Interface()
			continue
		}

		v = reflect.ValueOf(v).Elem().Interface()
	}

	// v can not be dereferenced if all of the following conditions are met:
	// - vType = (interface {}) <nil>
	// - vValue = (string) (len=15) "<invalid Value>"
	if reflect.TypeOf(v) == nil && !reflect.ValueOf(v).IsValid() {
		return false
	}

	return true
}

func IsNotAny(v any) bool {
	if !Dereferenceable(v) {
		return false
	}

	return reflect.TypeOf(GetDereferencedType(v)).Kind() != reflect.Interface
}
