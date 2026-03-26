package tests

// Realistic nesting tests using proper Fielder structs (as ormc would generate).
//
// Scenario: a tool-call request with 3 levels of nesting, unknown fields at
// every level, arrays discarded, and sibling fields after nested structs.

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/json"
	"testing"
)

// ── Structs ───────────────────────────────────────────────────────────────────

type toolInput struct {
	Query   string
	Limit   int64
	Verbose bool
}

func (t *toolInput) Schema() []fmt.Field {
	return []fmt.Field{
		{Name: "query", Type: fmt.FieldText},
		{Name: "limit", Type: fmt.FieldInt},
		{Name: "verbose", Type: fmt.FieldBool},
	}
}
func (t *toolInput) Pointers() []any {
	return []any{&t.Query, &t.Limit, &t.Verbose}
}

type toolParams struct {
	Name  string
	Input toolInput
}

func (t *toolParams) Schema() []fmt.Field {
	return []fmt.Field{
		{Name: "name", Type: fmt.FieldText},
		{Name: "input", Type: fmt.FieldStruct},
	}
}
func (t *toolParams) Pointers() []any {
	return []any{&t.Name, &t.Input}
}

type toolCall struct {
	ID     int64
	Method string
	Params toolParams
}

func (t *toolCall) Schema() []fmt.Field {
	return []fmt.Field{
		{Name: "id", Type: fmt.FieldInt},
		{Name: "method", Type: fmt.FieldText},
		{Name: "params", Type: fmt.FieldStruct},
	}
}
func (t *toolCall) Pointers() []any {
	return []any{&t.ID, &t.Method, &t.Params}
}

// ── Tests ─────────────────────────────────────────────────────────────────────

// TestDecodeNestedReal_HappyPath — three levels, all known fields, no noise.
func TestDecodeNestedReal_HappyPath(t *testing.T) {
	var req toolCall
	input := `{
		"id": 42,
		"method": "tools/call",
		"params": {
			"name": "search",
			"input": {
				"query": "tinywasm",
				"limit": 10,
				"verbose": true
			}
		}
	}`
	if err := json.Decode(input, &req); err != nil {
		t.Fatal(err)
	}
	if req.ID != 42 {
		t.Errorf("id: want 42 got %d", req.ID)
	}
	if req.Method != "tools/call" {
		t.Errorf("method: want tools/call got %s", req.Method)
	}
	if req.Params.Name != "search" {
		t.Errorf("params.name: want search got %s", req.Params.Name)
	}
	if req.Params.Input.Query != "tinywasm" {
		t.Errorf("input.query: want tinywasm got %s", req.Params.Input.Query)
	}
	if req.Params.Input.Limit != 10 {
		t.Errorf("input.limit: want 10 got %d", req.Params.Input.Limit)
	}
	if !req.Params.Input.Verbose {
		t.Error("input.verbose: want true got false")
	}
}

// TestDecodeNestedReal_UnknownFieldsAtEveryLevel — unknown fields before,
// between, and after known fields at all three levels.
func TestDecodeNestedReal_UnknownFieldsAtEveryLevel(t *testing.T) {
	var req toolCall
	input := `{
		"jsonrpc": "2.0",
		"id": 7,
		"extra_top": {"nested": "noise"},
		"method": "tools/call",
		"params": {
			"version": 1,
			"name": "echo",
			"meta": {"source": "ui", "tags": [1, 2, 3]},
			"input": {
				"trace_id": "abc-123",
				"query": "hello",
				"limit": 5,
				"filters": ["a", "b"],
				"verbose": false,
				"debug": true
			},
			"trailing": "ignored"
		},
		"timestamp": 1700000000
	}`
	if err := json.Decode(input, &req); err != nil {
		t.Fatal(err)
	}
	if req.ID != 7 {
		t.Errorf("id: want 7 got %d", req.ID)
	}
	if req.Method != "tools/call" {
		t.Errorf("method: want tools/call got %s", req.Method)
	}
	if req.Params.Name != "echo" {
		t.Errorf("params.name: want echo got %s", req.Params.Name)
	}
	if req.Params.Input.Query != "hello" {
		t.Errorf("input.query: want hello got %s", req.Params.Input.Query)
	}
	if req.Params.Input.Limit != 5 {
		t.Errorf("input.limit: want 5 got %d", req.Params.Input.Limit)
	}
	if req.Params.Input.Verbose != false {
		t.Error("input.verbose: want false got true")
	}
}

// TestDecodeNestedReal_SiblingAfterStruct — known field appears AFTER the
// nested struct in the JSON stream.
func TestDecodeNestedReal_SiblingAfterStruct(t *testing.T) {
	var req toolCall
	input := `{"params":{"input":{"query":"late","limit":3,"verbose":true},"name":"sibling_after"},"method":"tools/call","id":1}`
	if err := json.Decode(input, &req); err != nil {
		t.Fatal(err)
	}
	if req.Params.Name != "sibling_after" {
		t.Errorf("params.name: want sibling_after got %s", req.Params.Name)
	}
	if req.Params.Input.Query != "late" {
		t.Errorf("input.query: want late got %s", req.Params.Input.Query)
	}
	if req.Method != "tools/call" {
		t.Errorf("method: want tools/call got %s", req.Method)
	}
}

// TestDecodeNestedReal_DeepArraysDiscarded — arrays at multiple levels with
// mixed types inside; all discarded without error.
func TestDecodeNestedReal_DeepArraysDiscarded(t *testing.T) {
	var req toolCall
	input := `{
		"id": 1,
		"method": "tools/call",
		"tags": [1, "two", true, null, {"x": 1}, [1, 2]],
		"params": {
			"name": "scan",
			"options": [{"key": "a"}, {"key": "b"}],
			"input": {
				"query": "deep",
				"limit": 1,
				"verbose": false,
				"ranges": [[0,10],[20,30]]
			}
		}
	}`
	if err := json.Decode(input, &req); err != nil {
		t.Fatal(err)
	}
	if req.Params.Input.Query != "deep" {
		t.Errorf("input.query: want deep got %s", req.Params.Input.Query)
	}
}

// TestDecodeNestedReal_EmptyNestedObject — nested struct is present but empty.
func TestDecodeNestedReal_EmptyNestedObject(t *testing.T) {
	var req toolCall
	input := `{"id":1,"method":"tools/call","params":{"name":"empty","input":{}}}`
	if err := json.Decode(input, &req); err != nil {
		t.Fatal(err)
	}
	if req.Params.Name != "empty" {
		t.Errorf("params.name: want empty got %s", req.Params.Name)
	}
	if req.Params.Input.Query != "" || req.Params.Input.Limit != 0 || req.Params.Input.Verbose {
		t.Error("empty nested struct should leave fields at zero values")
	}
}

// TestDecodeNestedReal_EscapedStringsInNested — escape sequences propagate
// correctly through multiple levels.
func TestDecodeNestedReal_EscapedStringsInNested(t *testing.T) {
	var req toolCall
	input := `{"id":1,"method":"tools\/call","params":{"name":"esc\"ape","input":{"query":"line1\nline2","limit":0,"verbose":false}}}`
	if err := json.Decode(input, &req); err != nil {
		t.Fatal(err)
	}
	if req.Method != "tools/call" {
		t.Errorf("method: want tools/call got %s", req.Method)
	}
	if req.Params.Name != `esc"ape` {
		t.Errorf("params.name: want esc\"ape got %s", req.Params.Name)
	}
	if req.Params.Input.Query != "line1\nline2" {
		t.Errorf("input.query: want line1\\nline2 got %q", req.Params.Input.Query)
	}
}
