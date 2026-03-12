package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestParseBoolFalse(t *testing.T) {
	var b bool = true
	m := &mockFielder{
		schema: []fmt.Field{{Name: "B", Type: fmt.FieldBool, JSON: "b"}},
		pointers: []any{&b},
	}
	if err := json.Decode(`{"b":false}`, m); err != nil {
		t.Fatal(err)
	}
	if b != false {
		t.Errorf("expected false, got %v", b)
	}
}

func TestParseBoolInvalid(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"b":tru}`, m); err == nil {
		t.Fatal("expected error for invalid bool true")
	}
}

func TestParseBoolFalseInvalid(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"b":fals}`, m); err == nil {
		t.Fatal("expected error for invalid bool false")
	}
}

func TestParseNullInvalid(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"b":nul}`, m); err == nil {
		t.Fatal("expected error for invalid null")
	}
}
