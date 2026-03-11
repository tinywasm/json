package json

import (
	"bytes"
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
	if err := Decode(input, m); err != nil {
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
	if err := Decode(input, outer); err != nil {
		t.Fatal(err)
	}
	if user != "Alice" || city != "Paris" {
		t.Errorf("got %s, %s", user, city)
	}
}

func TestDecodeJSONKey(t *testing.T) {
	var firstName string
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "FirstName", Type: fmt.FieldText, JSON: "first_name"},
		},
		pointers: []any{&firstName},
	}
	input := `{"first_name":"Alice"}`
	if err := Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if firstName != "Alice" {
		t.Errorf("got %s", firstName)
	}
}

func TestDecodeJSONExclude(t *testing.T) {
	secret := "initial"
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Secret", Type: fmt.FieldText, JSON: "-"},
		},
		pointers: []any{&secret},
	}
	input := `{"Secret":"new"}`
	if err := Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if secret != "initial" {
		t.Errorf("secret changed to %s", secret)
	}
}

func TestDecodeMissingField(t *testing.T) {
	age := int64(20)
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Age", Type: fmt.FieldInt, JSON: "age"},
		},
		pointers: []any{&age},
	}
	input := `{}`
	if err := Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if age != 20 {
		t.Errorf("age changed to %d", age)
	}
}

func TestDecodeExtraField(t *testing.T) {
	var name string
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Name", Type: fmt.FieldText, JSON: "name"},
		},
		pointers: []any{&name},
	}
	input := `{"name":"Alice","extra":"ignore"}`
	if err := Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if name != "Alice" {
		t.Errorf("got %s", name)
	}
}

func TestDecodeIntFromFloat(t *testing.T) {
	var age int64
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Age", Type: fmt.FieldInt, JSON: "age"},
		},
		pointers: []any{&age},
	}
	input := `{"age":30.0}`
	if err := Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if age != 30 {
		t.Errorf("got %d", age)
	}
}

func TestDecodeFloatFromInt(t *testing.T) {
	var price float64
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Price", Type: fmt.FieldFloat, JSON: "price"},
		},
		pointers: []any{&price},
	}
	input := `{"price":100}`
	if err := Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if price != 100.0 {
		t.Errorf("got %f", price)
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
	if err := Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("got %s", string(data))
	}
}

func TestDecodeFromReader(t *testing.T) {
	var name string
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "Name", Type: fmt.FieldText, JSON: "name"},
		},
		pointers: []any{&name},
	}
	input := bytes.NewReader([]byte(`{"name":"Alice"}`))
	if err := Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if name != "Alice" {
		t.Errorf("got %s", name)
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
	if err := Decode(input, m); err != nil {
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
	if err := Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if name != "Alice" {
		t.Errorf("name changed to %s", name)
	}
}

func TestDecodeInvalidJSON(t *testing.T) {
	m := &mockFielder{}
	input := `{"name":`
	if err := Decode(input, m); err == nil {
		t.Error("expected error for invalid JSON")
	}
}
