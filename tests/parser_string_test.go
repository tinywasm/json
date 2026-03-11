package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestParseStringEscapeBF(t *testing.T) {
	var s string
	m := &mockFielder{
		schema: []fmt.Field{{Name: "S", Type: fmt.FieldText, JSON: "s"}},
		pointers: []any{&s},
	}
	input := `{"s":"\b\f\/"}`
	if err := json.Decode(input, m); err != nil {
		t.Fatal(err)
	}
	expected := "\b\f/"
	if s != expected {
		t.Errorf("expected %q, got %q", expected, s)
	}
}

func TestParseStringUnicode(t *testing.T) {
	var s string
	m := &mockFielder{
		schema: []fmt.Field{{Name: "S", Type: fmt.FieldText, JSON: "s"}},
		pointers: []any{&s},
	}
	input := `{"s":"\u0041"}`
	if err := json.Decode(input, m); err != nil {
		t.Fatal(err)
	}
	expected := "A"
	if s != expected {
		t.Errorf("expected %q, got %q", expected, s)
	}
}

func TestParseStringUnicodeShort(t *testing.T) {
	m := &mockFielder{}
	input := `{"s":"\u004"}`
	if err := json.Decode(input, m); err == nil {
		t.Fatal("expected error for short unicode escape")
	}
}

func TestParseStringInvalidEscape(t *testing.T) {
	m := &mockFielder{}
	input := `{"s":"\q"}`
	if err := json.Decode(input, m); err == nil {
		t.Fatal("expected error for invalid escape sequence")
	}
}

func TestParseStringUnexpectedEOF(t *testing.T) {
	m := &mockFielder{}
	input := `{"s":"abc`
	if err := json.Decode(input, m); err == nil {
		t.Fatal("expected error for unexpected EOF")
	}
}

func TestParseStringNotQuote(t *testing.T) {
	m := &mockFielder{}
	input := `{key:1}`
	if err := json.Decode(input, m); err == nil {
		t.Fatal("expected error for key not starting with quote")
	}
}
