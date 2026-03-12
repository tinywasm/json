package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

// TestEncodeNumericTypes — int, int32, int64, uint, uint64, float32, float64
func TestEncodeNumericTypes(t *testing.T) {
    cases := []struct {
        name     string
        ptr      any
        ft       fmt.FieldType
        expected string
    }{
        {"int", ptrInt(5), fmt.FieldInt, `{"v":5}`},
        {"int32", ptrInt32(5), fmt.FieldInt, `{"v":5}`},
        {"int64", ptrInt64(5), fmt.FieldInt, `{"v":5}`},
        {"float32", ptrFloat32(1.5), fmt.FieldFloat, `{"v":1.5}`},
        {"float64", ptrFloat64(1.5), fmt.FieldFloat, `{"v":1.5}`},
        {"uint", ptrUint(5), fmt.FieldInt, `{"v":5}`},
        {"uint64", ptrUint64(5), fmt.FieldInt, `{"v":5}`},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            m := &mockFielder{
                schema: []fmt.Field{{Name: "V", Type: c.ft, JSON: "v"}},
                pointers: []any{c.ptr},
            }
            var out string
            if err := json.Encode(m, &out); err != nil {
                t.Fatal(err)
            }
            if out != c.expected {
                t.Errorf("expected %s, got %s", c.expected, out)
            }
        })
    }
}
