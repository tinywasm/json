package tests

import (
	"github.com/tinywasm/json"
	"testing"
)

func TestParseObjectBadSeparator(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"a":1;}`, m); err == nil {
		t.Fatal("expected error for bad separator in object")
	}
}

func TestParseObjectMissingColon(t *testing.T) {
	m := &mockFielder{}
	if err := json.Decode(`{"a" 1}`, m); err == nil {
		t.Fatal("expected error for missing colon in object")
	}
}
