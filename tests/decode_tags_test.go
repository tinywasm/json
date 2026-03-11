package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestDecodeJSONKey(t *testing.T) {
	var firstName string
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "FirstName", Type: fmt.FieldText, JSON: "first_name"},
		},
		pointers: []any{&firstName},
	}
	input := `{"first_name":"Alice"}`
	if err := json.Decode(input, m); err != nil {
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
	if err := json.Decode(input, m); err != nil {
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
	if err := json.Decode(input, m); err != nil {
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
	if err := json.Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if name != "Alice" {
		t.Errorf("got %s", name)
	}
}
