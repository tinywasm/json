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
}

func TestEncodeFieldBytesNonBytes(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{{Name: "v", Type: fmt.FieldBlob}},
		pointers: []any{ptrInt(42)},
	}
	var out string
	if err := json.Encode(m, &out); err != nil {
		t.Fatal(err)
	}
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
	expected := `{"msg":"\u0001\u001f"}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}
