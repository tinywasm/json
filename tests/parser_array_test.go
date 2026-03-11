package tests

import (
	"github.com/tinywasm/json"
	"testing"
)

func TestParseArrayEmpty(t *testing.T) {
	m := &mockFielder{}
	// Discarding empty array
	if err := json.Decode(`{"a":[]}`, m); err != nil {
		t.Fatal(err)
	}
}

func TestParseArrayMissingBracket(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"a":[1,2}`, m); err == nil {
		t.Fatal("expected error for missing bracket")
	}
}

func TestParseArrayBadSeparator(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"a":[1;2]}`, m); err == nil {
		t.Fatal("expected error for bad separator in array")
	}
}

func TestParseValueArray(t *testing.T) {
	m := &mockFielder{}
	// Test parseValue with array (discarded)
	if err := json.Decode(`{"a":[1,2,3]}`, m); err != nil {
		t.Fatal(err)
	}
}

func TestParseValueUnknownChar(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"a":@invalid}`, m); err == nil {
		t.Fatal("expected error for unknown character")
	}
}
