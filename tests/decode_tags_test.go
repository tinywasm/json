package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestDecodeMissingField(t *testing.T) {
	age := int64(20)
	m := &mockFielder{
		schema: []fmt.Field{
			{Name: "age", Type: fmt.FieldInt},
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
			{Name: "name", Type: fmt.FieldText},
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
