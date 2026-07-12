package tests

import "github.com/tinywasm/model"

import (
	"github.com/tinywasm/json"
	"testing"
)

func TestEncodeNested(t *testing.T) {
	inner := &mockFielder{
		schema: []model.Field{
			{Name: "city", Type: model.Text()},
		},
		pointers: []any{ptrString("Paris")},
	}
	outer := &mockFielder{
		schema: []model.Field{
			{Name: "user", Type: model.Text()},
			{Name: "address", Type: model.Struct(nil)},
		},
		pointers: []any{ptrString("Alice"), inner},
	}
	var out string
	if err := json.Encode(outer, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"user":"Alice","address":{"city":"Paris"}}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

// TestEncodeFieldRawRoundtrip simulates a realistic MCP JSON-RPC response:
// result and error are pre-serialized JSON objects stored as FieldRaw strings.
// Verifies no double-encoding occurs — values appear inline, not as quoted strings.
func TestEncodeFieldRawRoundtrip(t *testing.T) {
	cases := []struct {
		name     string
		jsonrpc  string
		id       string
		result   string
		expected string
	}{
		{
			name:     "tools/list response",
			jsonrpc:  "2.0",
			id:       "1",
			result:   `{"tools":[{"name":"start_development"}]}`,
			expected: `{"jsonrpc":"2.0","id":"1","result":{"tools":[{"name":"start_development"}]}}`,
		},
		{
			name:     "empty result omitted",
			jsonrpc:  "2.0",
			id:       "2",
			result:   "",
			expected: `{"jsonrpc":"2.0","id":"2"}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := &mockFielder{
				schema: []model.Field{
					{Name: "jsonrpc", Type: model.Text()},
					{Name: "id", Type: model.Text()},
					{Name: "result", Type: model.Raw(), OmitEmpty: true},
				},
				pointers: []any{ptrString(c.jsonrpc), ptrString(c.id), ptrString(c.result)},
			}
			var out string
			if err := json.Encode(m, &out); err != nil {
				t.Fatal(err)
			}
			if out != c.expected {
				t.Errorf("expected %s\n     got %s", c.expected, out)
			}
		})
	}
}

func TestEncodeOmitEmpty(t *testing.T) {
	m := &mockFielder{
		schema: []model.Field{
			{Name: "name", Type: model.Text()},
			{Name: "age", Type: model.Int(), OmitEmpty: true},
		},
		pointers: []any{ptrString("Alice"), ptrInt64(0)},
	}
	var out string
	if err := json.Encode(m, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"name":"Alice"}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}
