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
			{Name: "Price", Type: fmt.FieldFloat, JSON: "price"},
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

// TestDecodeInt        — writeValue con *int
func TestDecodeInt(t *testing.T) {
	var v int
	m := &mockFielder{
		schema: []fmt.Field{{Name: "V", Type: fmt.FieldInt, JSON: "v"}},
		pointers: []any{&v},
	}
	if err := json.Decode(`{"v":42}`, m); err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

// TestDecodeInt32      — writeValue con *int32
func TestDecodeInt32(t *testing.T) {
	var v int32
	m := &mockFielder{
		schema: []fmt.Field{{Name: "V", Type: fmt.FieldInt, JSON: "v"}},
		pointers: []any{&v},
	}
	if err := json.Decode(`{"v":42}`, m); err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

// TestDecodeFloat32    — writeValue con *float32
func TestDecodeFloat32(t *testing.T) {
	var v float32
	m := &mockFielder{
		schema: []fmt.Field{{Name: "V", Type: fmt.FieldFloat, JSON: "v"}},
		pointers: []any{&v},
	}
	if err := json.Decode(`{"v":1.5}`, m); err != nil {
		t.Fatal(err)
	}
	if v != 1.5 {
		t.Errorf("expected 1.5, got %f", v)
	}
}

// TestDecodeInt32FromFloat  — parser retorna float64 → *int32
func TestDecodeInt32FromFloat(t *testing.T) {
	var v int32
	m := &mockFielder{
		schema: []fmt.Field{{Name: "V", Type: fmt.FieldInt, JSON: "v"}},
		pointers: []any{&v},
	}
	if err := json.Decode(`{"v":42.0}`, m); err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

// TestDecodeIntFromFloat    — parser retorna float64 → *int
func TestDecodeIntFromFloat(t *testing.T) {
	var v int
	m := &mockFielder{
		schema: []fmt.Field{{Name: "V", Type: fmt.FieldInt, JSON: "v"}},
		pointers: []any{&v},
	}
	if err := json.Decode(`{"v":42.0}`, m); err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

// TestDecodeFloat32FromInt  — parser retorna int64  → *float32
func TestDecodeFloat32FromInt(t *testing.T) {
	var v float32
	m := &mockFielder{
		schema: []fmt.Field{{Name: "V", Type: fmt.FieldFloat, JSON: "v"}},
		pointers: []any{&v},
	}
	if err := json.Decode(`{"v":42}`, m); err != nil {
		t.Fatal(err)
	}
	if v != 42.0 {
		t.Errorf("expected 42.0, got %f", v)
	}
}
