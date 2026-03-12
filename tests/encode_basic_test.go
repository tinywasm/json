package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

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
	if err := json.Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"name":"Alice","age":30,"active":true}`
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
	if err := json.Encode(m, &out); err != nil {
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
	if err := json.Encode(m, &out); err != nil {
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
	if err := json.Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"data":"hello"}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

// TestEncodeStructNotFielder — FieldStruct cuyo value no implementa Fielder → omitido
func TestEncodeStructNotFielder(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "User", Type: fmt.FieldStruct, JSON: "user"},
		},
		values: []any{"not-a-fielder"},
	}
	var out string
	if err := json.Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"user":"not-a-fielder"}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

// TestEncodeControlChars — chars < 0x20 → \u00XX
func TestEncodeControlChars(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Msg", Type: fmt.FieldText, JSON: "msg"},
		},
		values: []any{"\x01\x1f"},
	}
	var out string
	if err := json.Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	// \x01 -> \u0001, \x1f -> \u001f
	expected := `{"msg":"\u0001\u001f"}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}
