package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

type validatingFielder struct {
	mockFielder
	valid bool
}

func (v *validatingFielder) Validate() error {
	if !v.valid {
		return fmt.Err("validation", "failed")
	}
	return nil
}

func TestDecodeValidation(t *testing.T) {
	var name string
	m := &validatingFielder{
		mockFielder: mockFielder{
			schema: []fmt.Field{{Name: "name", Type: fmt.FieldText}},
			pointers: []any{&name},
		},
		valid: false,
	}

	input := `{"name":"Alice"}`
	err := json.Decode(input, m)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if name != "Alice" {
		t.Errorf("expected name to be Alice, got %s", name)
	}

	m.valid = true
	err = json.Decode(input, m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDecodeRawNoValidation(t *testing.T) {
	m := &validatingFielder{
		mockFielder: mockFielder{
			schema: []fmt.Field{{Name: "name", Type: fmt.FieldText}},
			pointers: []any{new(string)},
		},
		valid: false,
	}

	input := `{"name":"Alice"}`
	err := json.DecodeRaw(input, m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
