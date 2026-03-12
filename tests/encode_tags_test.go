package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestEncodeNested(t *testing.T) {
	inner := &mockFielder{
		schema: []fmt.Field{
			{Name: "City", Type: fmt.FieldText, JSON: "city"},
		},
		values: []any{"Paris"},
	}
	outer := &mockFielder{
		schema: []fmt.Field{
			{Name: "User", Type: fmt.FieldText, JSON: "user"},
			{Name: "Address", Type: fmt.FieldStruct, JSON: "address"},
		},
		values: []any{"Alice", inner},
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

func TestEncodeJSONKeyOmitEmpty(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "FirstName", Type: fmt.FieldText, JSON: ",omitempty"},
		},
		values: []any{""},
	}
	var out string
	if err := json.Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeOmitEmpty(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Name", Type: fmt.FieldText, JSON: "name"},
			{Name: "Age", Type: fmt.FieldInt, JSON: "age,omitempty"},
		},
		values: []any{"Alice", int64(0)},
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

func TestEncodeJSONExclude(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Name", Type: fmt.FieldText, JSON: "name"},
			{Name: "Secret", Type: fmt.FieldText, JSON: "-"},
		},
		values: []any{"Alice", "shhh"},
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

func TestEncodeJSONKey(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "FirstName", Type: fmt.FieldText, JSON: "first_name"},
		},
		values: []any{"Alice"},
	}
	var out string
	if err := json.Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"first_name":"Alice"}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeJSONKeyFallback(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "FirstName", Type: fmt.FieldText, JSON: ""},
		},
		values: []any{"Alice"},
	}
	var out string
	if err := json.Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"FirstName":"Alice"}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}
