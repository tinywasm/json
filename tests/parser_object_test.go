package tests

import (
	"github.com/tinywasm/json"
	"testing"
)

func TestParseObjectBadSeparator(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"name":1;}`, m); err == nil {
		t.Fatal("expected error for bad separator in object")
	}
}

func TestParseObjectMissingColon(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"name" 1}`, m); err == nil {
		t.Fatal("expected error for missing colon in object")
	}
}

func TestParseObjectEmpty(t *testing.T) {
	m := &simpleModel{}
	if err := json.Decode(`{"name":{}}`, m); err != nil {
		t.Fatal(err)
	}
}
