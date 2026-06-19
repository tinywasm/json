package tests

import (
	"github.com/tinywasm/json"
	"testing"
)

func TestParseArrayMissingComma(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"name":[1 2]}`, m); err == nil {
		t.Fatal("expected error for missing comma in array")
	}
}

func TestParseIntoFielderMissingQuote(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{name:1}`, m); err == nil {
		t.Fatal("expected error for missing quote in object key")
	}
}

func TestParseIntoFielderMissingColon(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"name" 1}`, m); err == nil {
		t.Fatal("expected error for missing colon in object")
	}
}

func TestParseIntoFielderMissingComma(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"name":"a" "age":2}`, m); err == nil {
		t.Fatal("expected error for missing comma in object")
	}
}

func TestParseObjectMissingQuote(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"unknown":{a:1}}`, m); err == nil {
		t.Fatal("expected error for missing quote in nested object key")
	}
}

func TestParseObjectMissingComma(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"unknown":{"a":1 "b":2}}`, m); err == nil {
		t.Fatal("expected error for missing comma in nested object")
	}
}

func TestParseValueUnexpected(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"name":!}`, m); err == nil {
		t.Fatal("expected error for unexpected character")
	}
}
