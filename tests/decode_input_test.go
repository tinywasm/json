package tests

import (
	"bytes"
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestDecodeFromReader(t *testing.T) {
	var name string
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "name", Type: fmt.FieldText},
		},
		pointers: []any{&name},
	}
	input := bytes.NewReader([]byte(`{"name":"Alice"}`))
	if err := json.Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if name != "Alice" {
		t.Errorf("got %s", name)
	}
}

func TestDecodeInvalidJSON(t *testing.T) {
	m := &simpleModel{}
	input := `{"name":`
	if err := json.Decode(input, m); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDecodeInvalidInput(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(123, m); err == nil {
		t.Fatal("expected error for invalid input type")
	}
}

func TestDecodeNotObject(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`"string"`, m); err == nil {
		t.Fatal("expected error when JSON is not an object")
	}
}

func TestDecodeFromBytes(t *testing.T) {
	var name string
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "name", Type: fmt.FieldText},
		},
		pointers: []any{&name},
	}
	if err := json.Decode([]byte(`{"name":"Alice"}`), m); err != nil {
		t.Fatal(err)
	}
	if name != "Alice" {
		t.Errorf("got %s", name)
	}
}
