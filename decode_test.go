package json

import (
	"bytes"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestDecodeFielderSimple(t *testing.T) {
	var name string
	var age int64
	var active bool
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Name", Type: fmt.FieldText},
			{Name: "Age", Type: fmt.FieldInt},
			{Name: "Active", Type: fmt.FieldBool},
		},
		pointers: []any{&name, &age, &active},
	}

	js := `{"Name":"John Doe","Age":30,"Active":true}`
	if err := Decode(js, m); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if name != "John Doe" {
		t.Errorf("expected name John Doe, got %s", name)
	}
	if age != 30 {
		t.Errorf("expected age 30, got %d", age)
	}
	if active != true {
		t.Errorf("expected active true, got %v", active)
	}
}

func TestDecodeFielderNested(t *testing.T) {
	var city string
	inner := &mockFielder{
		schema: []fmt.Field{
			{Name: "City", Type: fmt.FieldText},
		},
		pointers: []any{&city},
	}

	var user string
	outer := &mockFielder{
		schema: []fmt.Field{
			{Name: "User", Type: fmt.FieldText},
			{Name: "Address", Type: fmt.FieldStruct},
		},
		pointers: []any{&user, inner},
	}

	js := `{"User":"Alice","Address":{"City":"New York"}}`
	if err := Decode(js, outer); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if user != "Alice" {
		t.Errorf("expected user Alice, got %s", user)
	}
	if city != "New York" {
		t.Errorf("expected city New York, got %s", city)
	}
}

func TestDecodeFielderJSONKey(t *testing.T) {
	var fullName string
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "FullName", Type: fmt.FieldText, JSON: "full_name"},
		},
		pointers: []any{&fullName},
	}

	js := `{"full_name":"John Doe"}`
	if err := Decode(js, m); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if fullName != "John Doe" {
		t.Errorf("expected FullName John Doe, got %s", fullName)
	}
}

func TestDecodePrimitives(t *testing.T) {
	// String
	var s string
	if err := Decode(`"hello"`, &s); err != nil || s != "hello" {
		t.Errorf("Decode string failed: %v, %s", err, s)
	}

	// Int64
	var i int64
	if err := Decode(`123`, &i); err != nil || i != 123 {
		t.Errorf("Decode int64 failed: %v, %d", err, i)
	}

	// Float64
	var f float64
	if err := Decode(`3.14`, &f); err != nil || (f < 3.139 || f > 3.141) {
		t.Errorf("Decode float64 failed: %v, %f", err, f)
	}

	// Bool
	var b bool
	if err := Decode(`true`, &b); err != nil || b != true {
		t.Errorf("Decode bool failed: %v, %v", err, b)
	}
}

func TestDecodeArrays(t *testing.T) {
	var arr []any
	if err := Decode(`["a",1,true]`, &arr); err != nil {
		t.Fatalf("Decode array failed: %v", err)
	}
	if len(arr) != 3 {
		t.Fatalf("expected length 3, got %d", len(arr))
	}
	if arr[0] != "a" || arr[1] != int64(1) || arr[2] != true {
		t.Errorf("unexpected array content: %v", arr)
	}
}

func TestDecodeObjects(t *testing.T) {
	var m map[string]any
	if err := Decode(`{"a":1,"b":"c"}`, &m); err != nil {
		t.Fatalf("Decode object failed: %v", err)
	}
	if m["a"] != int64(1) || m["b"] != "c" {
		t.Errorf("unexpected map content: %v", m)
	}
}

func TestDecodeInputs(t *testing.T) {
	js := `"hello"`

	// From string
	var s1 string
	if err := Decode(js, &s1); err != nil || s1 != "hello" {
		t.Errorf("Decode from string failed: %v", err)
	}

	// From bytes
	var s2 string
	if err := Decode([]byte(js), &s2); err != nil || s2 != "hello" {
		t.Errorf("Decode from bytes failed: %v", err)
	}

	// From reader
	var s3 string
	if err := Decode(bytes.NewReader([]byte(js)), &s3); err != nil || s3 != "hello" {
		t.Errorf("Decode from reader failed: %v", err)
	}
}

func TestDecodeStringEscapes(t *testing.T) {
	js := `"quote: \", backslash: \\, newline: \n, unicode: \u0041"`
	var s string
	if err := Decode(js, &s); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	expected := "quote: \", backslash: \\, newline: \n, unicode: A"
	if s != expected {
		t.Errorf("expected %q, got %q", expected, s)
	}
}
