package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestParseBoolFalse(t *testing.T) {
	var b bool = true
	m := &mockFielder{
		schema: []fmt.Field{{Name: "b", Type: fmt.FieldBool}},
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
	m := &simpleModel{}
	if err := json.Decode(`{"name":tru}`, m); err == nil {
		t.Fatal("expected error for invalid bool true")
	}
}

func TestParseBoolFalseInvalid(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"name":fals}`, m); err == nil {
		t.Fatal("expected error for invalid bool false")
	}
}

func TestParseNullInvalid(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"name":nul}`, m); err == nil {
		t.Fatal("expected error for invalid null")
	}
}
