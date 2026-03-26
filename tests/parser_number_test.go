package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestParseNumberNegative(t *testing.T) {
	var n int64
	m := &mockFielder{
		schema: []fmt.Field{{Name: "n", Type: fmt.FieldInt}},
		pointers: []any{&n},
	}
	input := `{"n":-42}`
	if err := json.Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if n != -42 {
		t.Errorf("expected -42, got %d", n)
	}
}

func TestParseNumberScientific(t *testing.T) {
	var f float64
	m := &mockFielder{
		schema: []fmt.Field{{Name: "f", Type: fmt.FieldFloat}},
		pointers: []any{&f},
	}
	input := `{"f":1e2}`
	if err := json.Decode(input, m); err != nil {
		t.Logf("Decode error for 1e2: %v", err)
	} else if f != 100.0 {
		t.Errorf("expected 100.0, got %f", f)
	}

	input = `{"f":1.5}`
	if err := json.Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if f != 1.5 {
		t.Errorf("expected 1.5, got %f", f)
	}
}
