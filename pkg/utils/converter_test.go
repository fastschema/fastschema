package utils_test

import (
	"testing"

	"github.com/fastschema/fastschema/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestIntToUint(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    uint64
		wantErr bool
	}{
		{"int to uint", int(42), 42, false},
		{"int8 to uint", int8(42), 42, false},
		{"int16 to uint", int16(42), 42, false},
		{"int32 to uint", int32(42), 42, false},
		{"int64 to uint", int64(42), 42, false},
		{"negative int to uint", int(-42), 0, true},
		{"negative int8 to uint", int8(-42), 0, true},
		{"negative int16 to uint", int16(-42), 0, true},
		{"negative int32 to uint", int32(-42), 0, true},
		{"negative int64 to uint", int64(-42), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got uint64
			var err error

			switch v := tt.input.(type) {
			case int:
				got, err = utils.IntToUint[int, uint64](v)
			case int8:
				got, err = utils.IntToUint[int8, uint64](v)
			case int16:
				got, err = utils.IntToUint[int16, uint64](v)
			case int32:
				got, err = utils.IntToUint[int32, uint64](v)
			case int64:
				got, err = utils.IntToUint[int64, uint64](v)
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIntPointerToUint(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    uint64
		wantErr bool
	}{
		{"*int to uint", func() *int { v := 42; return &v }(), 42, false},
		{"*int8 to uint", func() *int8 { v := int8(42); return &v }(), 42, false},
		{"*int16 to uint", func() *int16 { v := int16(42); return &v }(), 42, false},
		{"*int32 to uint", func() *int32 { v := int32(42); return &v }(), 42, false},
		{"*int64 to uint", func() *int64 { v := int64(42); return &v }(), 42, false},
		{"nil *int to uint", (*int)(nil), 0, true},
		{"nil *int8 to uint", (*int8)(nil), 0, true},
		{"nil *int16 to uint", (*int16)(nil), 0, true},
		{"nil *int32 to uint", (*int32)(nil), 0, true},
		{"nil *int64 to uint", (*int64)(nil), 0, true},
		{"negative *int to uint", func() *int { v := -42; return &v }(), 0, true},
		{"negative *int8 to uint", func() *int8 { v := int8(-42); return &v }(), 0, true},
		{"negative *int16 to uint", func() *int16 { v := int16(-42); return &v }(), 0, true},
		{"negative *int32 to uint", func() *int32 { v := int32(-42); return &v }(), 0, true},
		{"negative *int64 to uint", func() *int64 { v := int64(-42); return &v }(), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got uint64
			var err error

			switch v := tt.input.(type) {
			case *int:
				got, err = utils.IntPointerToUint[int, uint64](v)
			case *int8:
				got, err = utils.IntPointerToUint[int8, uint64](v)
			case *int16:
				got, err = utils.IntPointerToUint[int16, uint64](v)
			case *int32:
				got, err = utils.IntPointerToUint[int32, uint64](v)
			case *int64:
				got, err = utils.IntPointerToUint[int64, uint64](v)
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUintPointerToUint(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    uint64
		wantErr bool
	}{
		{"*uint to uint", func() *uint { v := uint(42); return &v }(), 42, false},
		{"*uint8 to uint", func() *uint8 { v := uint8(42); return &v }(), 42, false},
		{"*uint16 to uint", func() *uint16 { v := uint16(42); return &v }(), 42, false},
		{"*uint32 to uint", func() *uint32 { v := uint32(42); return &v }(), 42, false},
		{"*uint64 to uint", func() *uint64 { v := uint64(42); return &v }(), 42, false},
		{"nil *uint to uint", (*uint)(nil), 0, true},
		{"nil *uint8 to uint", (*uint8)(nil), 0, true},
		{"nil *uint16 to uint", (*uint16)(nil), 0, true},
		{"nil *uint32 to uint", (*uint32)(nil), 0, true},
		{"nil *uint64 to uint", (*uint64)(nil), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got uint64
			var err error

			switch v := tt.input.(type) {
			case *uint:
				got, err = utils.UintPointerToUint[uint, uint64](v)
			case *uint8:
				got, err = utils.UintPointerToUint[uint8, uint64](v)
			case *uint16:
				got, err = utils.UintPointerToUint[uint16, uint64](v)
			case *uint32:
				got, err = utils.UintPointerToUint[uint32, uint64](v)
			case *uint64:
				got, err = utils.UintPointerToUint[uint64, uint64](v)
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestFloatPointerToUint(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    uint64
		wantErr bool
	}{
		{"*float32 to uint", func() *float32 { v := float32(42); return &v }(), 42, false},
		{"*float64 to uint", func() *float64 { v := float64(42); return &v }(), 42, false},
		{"nil *float32 to uint", (*float32)(nil), 0, true},
		{"nil *float64 to uint", (*float64)(nil), 0, true},
		{"negative *float32 to uint", func() *float32 { v := float32(-42); return &v }(), 0, true},
		{"negative *float64 to uint", func() *float64 { v := float64(-42); return &v }(), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got uint64
			var err error

			switch v := tt.input.(type) {
			case *float32:
				got, err = utils.FloatPointerToUint[float32, uint64](v)
			case *float64:
				got, err = utils.FloatPointerToUint[float64, uint64](v)
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestFloatToUint(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    uint64
		wantErr bool
	}{
		{"float32 to uint", float32(42), 42, false},
		{"float64 to uint", float64(42), 42, false},
		{"negative float32 to uint", float32(-42), 0, true},
		{"negative float64 to uint", float64(-42), 0, true},
		{"float32 with decimal to uint", float32(42.5), 0, true},
		{"float64 with decimal to uint", float64(42.5), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got uint64
			var err error

			switch v := tt.input.(type) {
			case float32:
				got, err = utils.FloatToUint[float32, uint64](v)
			case float64:
				got, err = utils.FloatToUint[float64, uint64](v)
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAnyToUint(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    uint64
		wantErr bool
	}{
		{"int to uint", int(42), 42, false},
		{"int8 to uint", int8(42), 42, false},
		{"int16 to uint", int16(42), 42, false},
		{"int32 to uint", int32(42), 42, false},
		{"int64 to uint", int64(42), 42, false},
		{"uint to uint", uint(42), 42, false},
		{"uint8 to uint", uint8(42), 42, false},
		{"uint16 to uint", uint16(42), 42, false},
		{"uint32 to uint", uint32(42), 42, false},
		{"uint64 to uint", uint64(42), 42, false},
		{"float32 to uint", float32(42), 42, false},
		{"float64 to uint", float64(42), 42, false},
		{"*int to uint", func() *int { v := 42; return &v }(), 42, false},
		{"*int8 to uint", func() *int8 { v := int8(42); return &v }(), 42, false},
		{"*int16 to uint", func() *int16 { v := int16(42); return &v }(), 42, false},
		{"*int32 to uint", func() *int32 { v := int32(42); return &v }(), 42, false},
		{"*int64 to uint", func() *int64 { v := int64(42); return &v }(), 42, false},
		{"*uint to uint", func() *uint { v := uint(42); return &v }(), 42, false},
		{"*uint8 to uint", func() *uint8 { v := uint8(42); return &v }(), 42, false},
		{"*uint16 to uint", func() *uint16 { v := uint16(42); return &v }(), 42, false},
		{"*uint32 to uint", func() *uint32 { v := uint32(42); return &v }(), 42, false},
		{"*uint64 to uint", func() *uint64 { v := uint64(42); return &v }(), 42, false},
		{"*float32 to uint", func() *float32 { v := float32(42); return &v }(), 42, false},
		{"*float64 to uint", func() *float64 { v := float64(42); return &v }(), 42, false},
		{"negative int to uint", int(-42), 0, true},
		{"negative float to uint", float32(-42), 0, true},
		{"nil pointer to uint", (*int)(nil), 0, true},
		{"unsupported type", "string", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.AnyToUint[uint64](tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
