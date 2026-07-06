package tests

import "github.com/tinywasm/model"

import (
	"github.com/tinywasm/json"
	"testing"
)

func TestDecodeMissingField(t *testing.T) {
	age := int64(20)
	m := &mockFielder{
		schema: []model.Field{
			{Name: "age", Type: model.FieldInt},
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
		schema: []model.Field{
			{Name: "name", Type: model.FieldText},
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
