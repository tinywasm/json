package tests

import "github.com/tinywasm/fmt"

type mockFielder struct {
	schema   []fmt.Field
	pointers []any
	err      error
}

func (m *mockFielder) Schema() []fmt.Field { return m.schema }
func (m *mockFielder) Pointers() []any {
	if m.err != nil {
		return nil
	}
	return m.pointers
}

func ptrString(s string) *string { return &s }
func ptrInt(i int) *int { return &i }
func ptrInt8(i int8) *int8 { return &i }
func ptrInt16(i int16) *int16 { return &i }
func ptrInt32(i int32) *int32 { return &i }
func ptrInt64(i int64) *int64 { return &i }
func ptrUint(i uint) *uint { return &i }
func ptrUint8(i uint8) *uint8 { return &i }
func ptrUint16(i uint16) *uint16 { return &i }
func ptrUint32(i uint32) *uint32 { return &i }
func ptrUint64(i uint64) *uint64 { return &i }
func ptrFloat32(f float32) *float32 { return &f }
func ptrFloat64(f float64) *float64 { return &f }
func ptrBool(b bool) *bool { return &b }
func ptrBytes(b []byte) *[]byte { return &b }
