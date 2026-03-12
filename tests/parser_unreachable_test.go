package tests

import (
	"github.com/tinywasm/json"
	"testing"
)

func TestParseArrayMissingComma(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"a":[1 2]}`, m); err == nil {
		t.Fatal("expected error for missing comma in array")
	}
}

func TestParseIntoFielderMissingQuote(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{a:1}`, m); err == nil {
		t.Fatal("expected error for missing quote in object key")
	}
}

func TestParseIntoFielderMissingColon(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"a" 1}`, m); err == nil {
		t.Fatal("expected error for missing colon in object")
	}
}

func TestParseIntoFielderMissingComma(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"a":1 "b":2}`, m); err == nil {
		t.Fatal("expected error for missing comma in object")
	}
}

func TestParseObjectMissingQuote(t *testing.T) {
	// To trigger parseObject we need a field that is NOT in the schema
	// and its value is an object.
	m := &mockFielder{}
	if err := json.Decode(`{"unknown":{a:1}}`, m); err == nil {
		t.Fatal("expected error for missing quote in nested object key")
	}
}

func TestParseObjectMissingComma(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"unknown":{"a":1 "b":2}}`, m); err == nil {
		t.Fatal("expected error for missing comma in nested object")
	}
}

func TestParseValueUnexpected(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"a":!}`, m); err == nil {
		t.Fatal("expected error for unexpected character")
	}
}
