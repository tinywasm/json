package json

import (
	"bytes"
	"github.com/tinywasm/fmt"
	"testing"
)

type mockFielder struct {
	schema   []fmt.Field
	values   []any
	pointers []any
}

func (m *mockFielder) Schema() []fmt.Field { return m.schema }
func (m *mockFielder) Values() []any       { return m.values }
func (m *mockFielder) Pointers() []any     { return m.pointers }

func TestEncodeFielderSimple(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Name", Type: fmt.FieldText},
			{Name: "Age", Type: fmt.FieldInt},
			{Name: "Active", Type: fmt.FieldBool},
		},
		values: []any{"John Doe", 30, true},
	}

	var out string
	if err := Encode(m, &out); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	expected := `{"Name":"John Doe","Age":30,"Active":true}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeFielderNested(t *testing.T) {
	inner := &mockFielder{
		schema: []fmt.Field{
			{Name: "City", Type: fmt.FieldText},
		},
		values: []any{"New York"},
	}

	outer := &mockFielder{
		schema: []fmt.Field{
			{Name: "User", Type: fmt.FieldText},
			{Name: "Address", Type: fmt.FieldStruct},
		},
		values: []any{"Alice", inner},
	}

	var out string
	if err := Encode(outer, &out); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	expected := `{"User":"Alice","Address":{"City":"New York"}}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeFielderOmitEmpty(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Name", Type: fmt.FieldText},
			{Name: "Age", Type: fmt.FieldInt, JSON: "age,omitempty"},
			{Name: "Email", Type: fmt.FieldText, JSON: "email,omitempty"},
		},
		values: []any{"John", 0, ""},
	}

	var out string
	if err := Encode(m, &out); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	expected := `{"Name":"John"}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeFielderJSONExclude(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Name", Type: fmt.FieldText},
			{Name: "Secret", Type: fmt.FieldText, JSON: "-"},
		},
		values: []any{"John", "password"},
	}

	var out string
	if err := Encode(m, &out); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	expected := `{"Name":"John"}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeFielderJSONKey(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "FullName", Type: fmt.FieldText, JSON: "full_name"},
		},
		values: []any{"John Doe"},
	}

	var out string
	if err := Encode(m, &out); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	expected := `{"full_name":"John Doe"}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodePrimitives(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{"hello", `"hello"`},
		{true, `true`},
		{false, `false`},
		{123, `123`},
		{int64(1234567890), `1234567890`},
		{3.14, `3.14`},
		{nil, `null`},
		{[]byte("base64"), `"base64"`},
	}

	for _, tc := range tests {
		var out string
		if err := Encode(tc.input, &out); err != nil {
			t.Errorf("Encode(%v) failed: %v", tc.input, err)
			continue
		}
		if out != tc.expected {
			t.Errorf("Encode(%v) = %s, expected %s", tc.input, out, tc.expected)
		}
	}
}

func TestEncodeSlices(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{[]string{"a", "b"}, `["a","b"]`},
		{[]int{1, 2}, `[1,2]`},
		{[]any{"a", 1, true}, `["a",1,true]`},
	}

	for _, tc := range tests {
		var out string
		if err := Encode(tc.input, &out); err != nil {
			t.Errorf("Encode(%v) failed: %v", tc.input, err)
			continue
		}
		if out != tc.expected {
			t.Errorf("Encode(%v) = %s, expected %s", tc.input, out, tc.expected)
		}
	}
}

func TestEncodeMaps(t *testing.T) {
	m := map[string]any{"a": 1}
	var out string
	if err := Encode(m, &out); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	expected := `{"a":1}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeStringEscaping(t *testing.T) {
	s := "quote: \", backslash: \\, newline: \n, tab: \t, control: \x01"
	var out string
	if err := Encode(s, &out); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	expected := `"quote: \", backslash: \\, newline: \n, tab: \t, control: \u0001"`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeOutputs(t *testing.T) {
	input := "hello"

	// To string
	var s string
	if err := Encode(input, &s); err != nil || s != `"hello"` {
		t.Errorf("Encode to string failed: %v, %s", err, s)
	}

	// To bytes
	var b []byte
	if err := Encode(input, &b); err != nil || string(b) != `"hello"` {
		t.Errorf("Encode to bytes failed: %v, %s", err, string(b))
	}

	// To writer
	var buf bytes.Buffer
	if err := Encode(input, &buf); err != nil || buf.String() != `"hello"` {
		t.Errorf("Encode to writer failed: %v, %s", err, buf.String())
	}
}

func TestEncodeUnsupportedType(t *testing.T) {
	type unsupported struct{ A int }
	var out string
	err := Encode(unsupported{A: 1}, &out)
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}
