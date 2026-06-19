package tests

import (
	"github.com/tinywasm/json"
	"testing"
)

func TestParseArrayEmpty(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"name":[]}`, m); err != nil {
		t.Fatal(err)
	}
}

func TestParseArrayMissingBracket(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"name":[1,2}`, m); err == nil {
		t.Fatal("expected error for missing bracket")
	}
}

func TestParseArrayBadSeparator(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"name":[1;2]}`, m); err == nil {
		t.Fatal("expected error for bad separator in array")
	}
}

func TestParseValueArray(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"name":[1,2,3]}`, m); err != nil {
		t.Fatal(err)
	}
}

func TestParseValueUnknownChar(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"name":@invalid}`, m); err == nil {
		t.Fatal("expected error for unknown character")
	}
}
