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
		pointers: []any{ptrString("Paris")},
	}
	outer := &mockFielder{
		schema: []fmt.Field{
			{Name: "User", Type: fmt.FieldText, JSON: "user"},
			{Name: "Address", Type: fmt.FieldStruct, JSON: "address"},
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

func TestEncodeJSONKeyOmitEmpty(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "FirstName", Type: fmt.FieldText, JSON: ",omitempty"},
		},
		pointers: []any{ptrString("")},
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

func TestEncodeJSONExclude(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Name", Type: fmt.FieldText, JSON: "name"},
			{Name: "Secret", Type: fmt.FieldText, JSON: "-"},
		},
		pointers: []any{ptrString("Alice"), ptrString("shhh")},
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
		pointers: []any{ptrString("Alice")},
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
		pointers: []any{ptrString("Alice")},
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
