package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestDecodeFloatFromInt(t *testing.T) {
	var price float64
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "price", Type: fmt.FieldFloat},
		},
		pointers: []any{&price},
	}
	input := `{"price":100}`
	if err := json.Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if price != 100.0 {
		t.Errorf("got %f", price)
	}
}

// TestDecodeInt        — writeValue with *int
func TestDecodeInt(t *testing.T) {
	var v int
	m := &mockFielder{
		schema: []fmt.Field{{Name: "v", Type: fmt.FieldInt}},
		pointers: []any{&v},
	}
	if err := json.Decode(`{"v":42}`, m); err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

func TestDecodeFieldRaw(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"object", `{"v":{"a":1}}`, `{"a":1}`},
		{"array", `{"v":[1,2,3]}`, `[1,2,3]`},
		{"string", `{"v":"hello"}`, `"hello"`},
		{"null", `{"v":null}`, `null`},
		{"number", `{"v":123.45}`, `123.45`},
		{"bool", `{"v":true}`, `true`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var raw string
			m := &mockFielder{
				schema:   []fmt.Field{{Name: "v", Type: fmt.FieldRaw}},
				pointers: []any{&raw},
			}
			if err := json.Decode(c.input, m); err != nil {
				t.Fatal(err)
			}
			if raw != c.expected {
				t.Errorf("expected %s, got %s", c.expected, raw)
			}
		})
	}
}

// TestDecodeInt32      — writeValue with *int32
func TestDecodeInt32(t *testing.T) {
	var v int32
	m := &mockFielder{
		schema: []fmt.Field{{Name: "v", Type: fmt.FieldInt}},
		pointers: []any{&v},
	}
	if err := json.Decode(`{"v":42}`, m); err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

// TestDecodeFloat32    — writeValue with *float32
func TestDecodeFloat32(t *testing.T) {
	var v float32
	m := &mockFielder{
		schema: []fmt.Field{{Name: "v", Type: fmt.FieldFloat}},
		pointers: []any{&v},
	}
	if err := json.Decode(`{"v":1.5}`, m); err != nil {
		t.Fatal(err)
	}
	if v != 1.5 {
		t.Errorf("expected 1.5, got %f", v)
	}
}

// TestDecodeInt32FromFloat  — parser returns float64 → *int32
func TestDecodeInt32FromFloat(t *testing.T) {
	var v int32
	m := &mockFielder{
		schema: []fmt.Field{{Name: "v", Type: fmt.FieldInt}},
		pointers: []any{&v},
	}
	if err := json.Decode(`{"v":42.0}`, m); err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

// TestDecodeIntFromFloat    — parser returns float64 → *int
func TestDecodeIntFromFloat(t *testing.T) {
	var v int
	m := &mockFielder{
		schema: []fmt.Field{{Name: "v", Type: fmt.FieldInt}},
		pointers: []any{&v},
	}
	if err := json.Decode(`{"v":42.0}`, m); err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

// TestDecodeFloat32FromInt  — parser returns int64  → *float32
func TestDecodeFloat32FromInt(t *testing.T) {
	var v float32
	m := &mockFielder{
		schema: []fmt.Field{{Name: "v", Type: fmt.FieldFloat}},
		pointers: []any{&v},
	}
	if err := json.Decode(`{"v":42}`, m); err != nil {
		t.Fatal(err)
	}
	if v != 42.0 {
		t.Errorf("expected 42.0, got %f", v)
	}
}

// TestDecodeInt64Ptr — writeValue with *int64
func TestDecodeInt64Ptr(t *testing.T) {
	var v int64
	m := &mockFielder{
		schema:   []fmt.Field{{Name: "v", Type: fmt.FieldInt}},
		pointers: []any{&v},
	}
	if err := json.Decode(`{"v":42}`, m); err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

// TestDecodeInt64FromFloat — parser returns float64 → *int64
func TestDecodeInt64FromFloat(t *testing.T) {
	var v int64
	m := &mockFielder{
		schema:   []fmt.Field{{Name: "v", Type: fmt.FieldInt}},
		pointers: []any{&v},
	}
	if err := json.Decode(`{"v":42.0}`, m); err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}
