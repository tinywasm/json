package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestParseIntoFielderNotObject(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`[1,2,3]`, m); err == nil {
		t.Fatal("expected error when input is not an object")
	}
}

func TestSkipWhitespace(t *testing.T) {
	var n int64
	m := &mockFielder{
		schema: []fmt.Field{{Name: "n", Type: fmt.FieldInt}},
		pointers: []any{&n},
	}
	input := " \t\r\n{ \t\r\n\"n\" \t\r\n: \t\r\n42 \t\r\n} \t\r\n"
	if err := json.Decode(input, m); err != nil {
		t.Fatal(err)
	}
	if n != 42 {
		t.Errorf("expected 42, got %d", n)
	}
}

func TestPeekNextEmpty(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(``, m); err == nil {
		t.Fatal("expected error for empty data")
	}
}
