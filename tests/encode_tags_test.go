package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestEncodeNested(t *testing.T) {
	inner := &mockFielder{
		schema: []fmt.Field{
			{Name: "city", Type: fmt.FieldText},
		},
		pointers: []any{ptrString("Paris")},
	}
	outer := &mockFielder{
		schema: []fmt.Field{
			{Name: "user", Type: fmt.FieldText},
			{Name: "address", Type: fmt.FieldStruct},
		},
		pointers: []any{ptrString("Alice"), inner},
	}
	var out string
	if err := json.Encode(outer, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"user":"Alice","address":{"city":"Paris"}}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeOmitEmpty(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "name", Type: fmt.FieldText},
			{Name: "age", Type: fmt.FieldInt, OmitEmpty: true},
		},
		pointers: []any{ptrString("Alice"), ptrInt64(0)},
	}
	var out string
	if err := json.Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"name":"Alice"}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}
