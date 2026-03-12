package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

// TestDecodeStructNotFielder  — ptr no implementa Fielder → campo descartado
func TestDecodeStructNotFielder(t *testing.T) {
	var s string = "initial"
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Nested", Type: fmt.FieldStruct, JSON: "nested"},
		},
		pointers: []any{&s}, // string doesn't implement Fielder
	}
	// It should parse and discard the object
	if err := json.Decode(`{"nested":{"a":1}}`, m); err != nil {
		t.Fatal(err)
	}
	if s != "initial" {
		t.Errorf("expected initial, got %s", s)
	}
}

// TestDecodeExtraNestedObject — campo desconocido objeto  → descartado
func TestDecodeExtraNestedObject(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"extra":{"a":1}}`, m); err != nil {
		t.Fatal(err)
	}
}

// TestDecodeExtraArray        — campo desconocido array   → descartado
func TestDecodeExtraArray(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"extra":[1,2,3]}`, m); err != nil {
		t.Fatal(err)
	}
}
