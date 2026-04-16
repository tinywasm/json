package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

func TestEncodeNested(t *testing.T) {
	inner := &mockFielder{
		schema: []fmt.Field{
			{Name: "city", Type: fmt.FieldText},
		},
		pointers: []any{ptrString("Paris")},
	}
	outer := &mockFielder{
		schema: []fmt.Field{
			{Name: "user", Type: fmt.FieldText},
			{Name: "address", Type: fmt.FieldStruct},
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
				schema: []fmt.Field{
					{Name: "jsonrpc", Type: fmt.FieldText},
					{Name: "id", Type: fmt.FieldText},
					{Name: "result", Type: fmt.FieldRaw, OmitEmpty: true},
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
		schema: []fmt.Field{
			{Name: "name", Type: fmt.FieldText},
			{Name: "age", Type: fmt.FieldInt, OmitEmpty: true},
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
