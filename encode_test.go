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

func TestEncodeSimple(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Name", Type: fmt.FieldText, JSON: "name"},
			{Name: "Age", Type: fmt.FieldInt, JSON: "age"},
			{Name: "Active", Type: fmt.FieldBool, JSON: "active"},
		},
		values: []any{"Alice", int64(30), true},
	}
	var out string
	if err := Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"name":"Alice","age":30,"active":true}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

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
	if err := Encode(outer, &out); err != nil {
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
			{Name: "Name", Type: fmt.FieldText, JSON: "name"},
			{Name: "Age", Type: fmt.FieldInt, JSON: "age,omitempty"},
		},
		values: []any{"Alice", int64(0)},
	}
	var out string
	if err := Encode(m, &out); err != nil {
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
	if err := Encode(m, &out); err != nil {
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
	if err := Encode(m, &out); err != nil {
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
	if err := Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"FirstName":"Alice"}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeStringEscaping(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Msg", Type: fmt.FieldText, JSON: "msg"},
		},
		values: []any{"hello \"world\"\n\r\t\\"},
	}
	var out string
	if err := Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"msg":"hello \"world\"\n\r\t\\"}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeNilField(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Val", Type: fmt.FieldText, JSON: "val"},
		},
		values: []any{nil},
	}
	var out string
	if err := Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"val":null}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeBytes(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Data", Type: fmt.FieldBlob, JSON: "data"},
		},
		values: []any{[]byte("hello")},
	}
	var out string
	if err := Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"data":"hello"}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeToBytes(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{{Name: "A", Type: fmt.FieldInt, JSON: "a"}},
		values: []any{int64(1)},
	}
	var out []byte
	if err := Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"a":1}`
	if string(out) != expected {
		t.Errorf("expected %s, got %s", expected, string(out))
	}
}

func TestEncodeToWriter(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{{Name: "A", Type: fmt.FieldInt, JSON: "a"}},
		values: []any{int64(1)},
	}
	var buf bytes.Buffer
	if err := Encode(m, &buf); err != nil {
		t.Fatal(err)
	}
	expected := `{"a":1}`
	if buf.String() != expected {
		t.Errorf("expected %s, got %s", expected, buf.String())
	}
}
