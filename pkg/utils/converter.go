package utils

import (
	"errors"
	"fmt"
	"math"
)

func IntToUint[
	IN ~int | ~int8 | ~int16 | ~int32 | ~int64,
	OUT ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64,
](value IN) (OUT, error) {
	if value < 0 {
		return 0, errors.New("negative value cannot be converted to uint")
	}

	return OUT(value), nil
}

func IntPointerToUint[
	IN ~int | ~int8 | ~int16 | ~int32 | ~int64,
	OUT ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64,
](value *IN) (OUT, error) {
	if value == nil || *value < 0 {
		return 0, errors.New("nil pointer or negative value cannot be converted to uint")
	}

	return OUT(*value), nil
}

func UintPointerToUint[
	IN ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64,
	OUT ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64,
](value *IN) (OUT, error) {
	if value == nil {
		return 0, errors.New("nil pointer cannot be converted to uint")
	}

	return OUT(*value), nil
}

func FloatPointerToUint[
	IN ~float32 | ~float64,
	OUT ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64,
](value *IN) (OUT, error) {
	if value == nil || *value < 0 {
		return 0, errors.New("nil pointer or negative value cannot be converted to uint")
	}

	return OUT(*value), nil
}

func FloatToUint[
	IN ~float32 | ~float64,
	OUT ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64,
](value IN) (OUT, error) {
	if value < 0 {
		return 0, errors.New("negative value cannot be converted to uint")
	}

	// Only allow conversion if the value is a whole number.
	if value != IN(int64(value)) {
		return 0, errors.New("float value must be a whole number")
	}

	return OUT(value), nil
}

func AnyToUint[
	OUT ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64,
](value any) (out OUT, err error) {
	switch v := value.(type) {
	case int:
		return IntToUint[int, OUT](v)
	case int8:
		return IntToUint[int8, OUT](v)
	case int16:
		return IntToUint[int16, OUT](v)
	case int32:
		return IntToUint[int32, OUT](v)
	case int64:
		return IntToUint[int64, OUT](v)
	case uint:
		return OUT(v), nil
	case uint8:
		return OUT(v), nil
	case uint16:
		return OUT(v), nil
	case uint32:
		return OUT(v), nil
	case uint64:
		return OUT(v), nil
	case float32:
		return FloatToUint[float32, OUT](v)
	case float64:
		return FloatToUint[float64, OUT](v)
	case *int:
		return IntPointerToUint[int, OUT](v)
	case *int8:
		return IntPointerToUint[int8, OUT](v)
	case *int16:
		return IntPointerToUint[int16, OUT](v)
	case *int32:
		return IntPointerToUint[int32, OUT](v)
	case *int64:
		return IntPointerToUint[int64, OUT](v)
	case *uint:
		return UintPointerToUint[uint, OUT](v)
	case *uint8:
		return UintPointerToUint[uint8, OUT](v)
	case *uint16:
		return UintPointerToUint[uint16, OUT](v)
	case *uint32:
		return UintPointerToUint[uint32, OUT](v)
	case *uint64:
		return UintPointerToUint[uint64, OUT](v)
	case *float32:
		return FloatPointerToUint[float32, OUT](v)
	case *float64:
		return FloatPointerToUint[float64, OUT](v)
	default:
		return out, fmt.Errorf("unsupported type when converting to uint: %T", value)
	}
}

func AnyToInt[
	OUT ~int | ~int8 | ~int16 | ~int32 | ~int64,
](value any) (out OUT, err error) {
	var convert = func(v int64) (OUT, error) {
		out := OUT(v)
		if int64(out) != v {
			return 0, fmt.Errorf("value %d out of range for target type", v)
		}
		return out, nil
	}

	switch v := value.(type) {
	case int:
		return convert(int64(v))
	case int8:
		return convert(int64(v))
	case int16:
		return convert(int64(v))
	case int32:
		return convert(int64(v))
	case int64:
		return convert(v)
	case uint:
		return convert(int64(v))
	case uint8:
		return convert(int64(v))
	case uint16:
		return convert(int64(v))
	case uint32:
		return convert(int64(v))
	case uint64:
		if v > math.MaxInt64 {
			return out, fmt.Errorf("value %d out of range for signed integer", v)
		}
		return convert(int64(v))
	case float32:
		return floatToInt(float64(v), convert)
	case float64:
		return floatToInt(v, convert)
	case *int:
		if v == nil {
			return out, fmt.Errorf("unsupported type when converting to int: %T", value)
		}
		return AnyToInt[OUT](*v)
	case *int8:
		if v == nil {
			return out, fmt.Errorf("unsupported type when converting to int: %T", value)
		}
		return AnyToInt[OUT](*v)
	case *int16:
		if v == nil {
			return out, fmt.Errorf("unsupported type when converting to int: %T", value)
		}
		return AnyToInt[OUT](*v)
	case *int32:
		if v == nil {
			return out, fmt.Errorf("unsupported type when converting to int: %T", value)
		}
		return AnyToInt[OUT](*v)
	case *int64:
		if v == nil {
			return out, fmt.Errorf("unsupported type when converting to int: %T", value)
		}
		return AnyToInt[OUT](*v)
	case *uint:
		if v == nil {
			return out, fmt.Errorf("unsupported type when converting to int: %T", value)
		}
		return AnyToInt[OUT](*v)
	case *uint8:
		if v == nil {
			return out, fmt.Errorf("unsupported type when converting to int: %T", value)
		}
		return AnyToInt[OUT](*v)
	case *uint16:
		if v == nil {
			return out, fmt.Errorf("unsupported type when converting to int: %T", value)
		}
		return AnyToInt[OUT](*v)
	case *uint32:
		if v == nil {
			return out, fmt.Errorf("unsupported type when converting to int: %T", value)
		}
		return AnyToInt[OUT](*v)
	case *uint64:
		if v == nil {
			return out, fmt.Errorf("unsupported type when converting to int: %T", value)
		}
		return AnyToInt[OUT](*v)
	case *float32:
		if v == nil {
			return out, fmt.Errorf("unsupported type when converting to int: %T", value)
		}
		return AnyToInt[OUT](*v)
	case *float64:
		if v == nil {
			return out, fmt.Errorf("unsupported type when converting to int: %T", value)
		}
		return AnyToInt[OUT](*v)
	default:
		return out, fmt.Errorf("unsupported type when converting to int: %T", value)
	}
}

func floatToInt[
	OUT ~int | ~int8 | ~int16 | ~int32 | ~int64,
](value float64, convert func(int64) (OUT, error)) (OUT, error) {
	if value < math.MinInt64 || value > math.MaxInt64 {
		return 0, fmt.Errorf("float value %f out of range for signed integer", value)
	}
	if value != math.Trunc(value) {
		return 0, fmt.Errorf("float value must be a whole number")
	}
	return convert(int64(value))
}
