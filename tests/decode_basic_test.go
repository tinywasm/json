package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestDecodeSimple(t *testing.T) {
	var name string
	var age int64
	var active bool
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Name", Type: fmt.FieldText, JSON: "name"},
			{Name: "Age", Type: fmt.FieldInt, JSON: "age"},
			{Name: "Active", Type: fmt.FieldBool, JSON: "active"},
		},
		pointers: []any{&name, &age, &active},
	}
	input := `{"name":"Alice","age":30,"active":true}`
	if err := json.Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if name != "Alice" || age != 30 || active != true {
		t.Errorf("got %s, %d, %v", name, age, active)
	}
}

func TestDecodeNested(t *testing.T) {
	var city string
	inner := &mockFielder{
		schema: []fmt.Field{
			{Name: "City", Type: fmt.FieldText, JSON: "city"},
		},
		pointers: []any{&city},
	}
	var user string
	outer := &mockFielder{
		schema: []fmt.Field{
			{Name: "User", Type: fmt.FieldText, JSON: "user"},
			{Name: "Address", Type: fmt.FieldStruct, JSON: "address"},
		},
		pointers: []any{&user, inner},
	}
	input := `{"user":"Alice","address":{"city":"Paris"}}`
	if err := json.Decode(input, outer); err != nil {
		t.Fatal(err)
	}
	if user != "Alice" || city != "Paris" {
		t.Errorf("got %s, %s", user, city)
	}
}

func TestDecodeStringEscapes(t *testing.T) {
	var msg string
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Msg", Type: fmt.FieldText, JSON: "msg"},
		},
		pointers: []any{&msg},
	}
	input := `{"msg":"hello \"world\"\n\r\t\\"}`
	if err := json.Decode(input, m); err != nil {
		t.Fatal(err)
	}
	expected := "hello \"world\"\n\r\t\\"
	if msg != expected {
		t.Errorf("expected %q, got %q", expected, msg)
	}
}

func TestDecodeNull(t *testing.T) {
	name := "Alice"
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Name", Type: fmt.FieldText, JSON: "name"},
		},
		pointers: []any{&name},
	}
	input := `{"name":null}`
	if err := json.Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if name != "Alice" {
		t.Errorf("name changed to %s", name)
	}
}

func TestDecodeBytes(t *testing.T) {
	var data []byte
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Data", Type: fmt.FieldBlob, JSON: "data"},
		},
		pointers: []any{&data},
	}
	input := `{"data":"hello"}`
	if err := json.Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("got %s", string(data))
	}
}
