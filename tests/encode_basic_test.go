package tests

import "github.com/tinywasm/model"

import (
	"github.com/tinywasm/json"
	"testing"
)

func TestEncodeSimple(t *testing.T) {
	m := &mockFielder{
		schema: []model.Field{
			{Name: "name", Type: model.Text()},
			{Name: "age", Type: model.Int()},
			{Name: "active", Type: model.Bool()},
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
		schema:   []model.Field{{Name: "v", Type: model.Blob()}},
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
		schema: []model.Field{
			{Name: "msg", Type: model.Text()},
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
		schema: []model.Field{
			{Name: "val", Type: model.Text()},
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
		schema: []model.Field{
			{Name: "data", Type: model.Blob()},
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
		schema: []model.Field{
			{Name: "user", Type: model.Struct(nil)},
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
		schema: []model.Field{
			{Name: "msg", Type: model.Text()},
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
