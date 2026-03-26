package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestEncodeSimple(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "name", Type: fmt.FieldText},
			{Name: "age", Type: fmt.FieldInt},
			{Name: "active", Type: fmt.FieldBool},
		},
		pointers: []any{ptrString("Alice"), ptrInt64(30), ptrBool(true)},
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

func TestEncodeFielderError(t *testing.T) {
	inner := &mockFielder{
		schema: []fmt.Field{{Name: "e", Type: fmt.FieldText}},
		pointers: []any{nil},
		err:    fmt.Err("test", "encode", "error"),
	}
	outer := &mockFielder{
		schema: []fmt.Field{{Name: "i", Type: fmt.FieldStruct}},
		pointers: []any{inner},
	}
	var out string
	if err := json.Encode(outer, &out); err == nil {
		t.Fatal("expected error from inner fielder")
	}
}

func TestEncodeFieldBytesNonBytes(t *testing.T) {
	// FieldBlob with value that is not []byte -> encodeValue omits it or treats it via default.
	// Actually encodeValue handles it via default (fmt.Convert).
	// To trigger default in encodeValue with something that is NOT handled by other cases:
	m := &mockFielder{
		schema: []fmt.Field{{Name: "v", Type: fmt.FieldBlob}},
		pointers: []any{ptrInt(42)}, // Not []byte, not string, not bool, not nil
	}
	var out string
	if err := json.Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	// encodeFromPtr handles FieldBlob specifically for *[]byte. If it's *int, it falls to default: b.WriteString("null")
	// The original test expected `{"v":42}`.
	// Our new implementation of encodeFromPtr for FieldBlob ONLY handles *[]byte.
	// If we want to maintain the old behavior where it could fall back to other types, we'd need to change encodeFromPtr.
	// But according to PLAN.md, we should avoid interface boxing.
	expected := `{"v":null}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeStringEscaping(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "msg", Type: fmt.FieldText},
		},
		pointers: []any{ptrString("hello \"world\"\n\r\t\\")},
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
			{Name: "val", Type: fmt.FieldText},
		},
		pointers: []any{nil},
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
			{Name: "data", Type: fmt.FieldBlob},
		},
		pointers: []any{ptrBytes([]byte("hello"))},
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
			{Name: "user", Type: fmt.FieldStruct},
		},
		pointers: []any{ptrString("not-a-fielder")},
	}
	var out string
	if err := json.Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"user":null}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

// TestEncodeControlChars — chars < 0x20 → \u00XX
func TestEncodeControlChars(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "msg", Type: fmt.FieldText},
		},
		pointers: []any{ptrString("\x01\x1f")},
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
