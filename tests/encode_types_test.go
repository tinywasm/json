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
        val      any
        expected string
    }{
        {"int", int(5), `{"v":5}`},
        {"int32", int32(5), `{"v":5}`},
        {"int64", int64(5), `{"v":5}`},
        {"float32", float32(1.5), `{"v":1.5}`},
        {"float64", float64(1.5), `{"v":1.5}`},
        {"uint", uint(5), `{"v":5}`},
        {"uint64", uint64(5), `{"v":5}`},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            m := &mockFielder{
                schema: []fmt.Field{{Name: "V", Type: fmt.FieldInt, JSON: "v"}},
                values: []any{c.val},
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
