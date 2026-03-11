package tests

import (
	"bytes"
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestEncodeToBytes(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{{Name: "A", Type: fmt.FieldInt, JSON: "a"}},
		values: []any{int64(1)},
	}
	var out []byte
	if err := json.Encode(m, &out); err != nil {
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
	if err := json.Encode(m, &buf); err != nil {
		t.Fatal(err)
	}
	expected := `{"a":1}`
	if buf.String() != expected {
		t.Errorf("expected %s, got %s", expected, buf.String())
	}
}

// TestEncodeToString — output *string (ausente en tests originales)
func TestEncodeToString(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{{Name: "A", Type: fmt.FieldInt, JSON: "a"}},
		values: []any{int64(1)},
	}
	var out string
	if err := json.Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"a":1}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

// TestEncodeInvalidOutput — output desconocido → error
func TestEncodeInvalidOutput(t *testing.T) {
	m := &mockFielder{
		schema: []fmt.Field{{Name: "A", Type: fmt.FieldInt, JSON: "a"}},
		values: []any{int64(1)},
	}
	if err := json.Encode(m, 123); err == nil {
		t.Fatal("expected error for invalid output type")
	}
}
