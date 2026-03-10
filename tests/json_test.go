package tests

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/json"
	"testing"
)

type mockFielder struct {
	schema   []fmt.Field
	values   []any
	pointers []any
}

func (m *mockFielder) Schema() []fmt.Field { return m.schema }
func (m *mockFielder) Values() []any       { return m.values }
func (m *mockFielder) Pointers() []any     { return m.pointers }

type TestStruct struct {
	Name string
	Age  int64
}

func (s *TestStruct) Schema() []fmt.Field {
	return []fmt.Field{
		{Name: "Name", Type: fmt.FieldText, JSON: "name"},
		{Name: "Age", Type: fmt.FieldInt, JSON: "age"},
	}
}

func (s *TestStruct) Values() []any {
	return []any{s.Name, s.Age}
}

func (s *TestStruct) Pointers() []any {
	return []any{&s.Name, &s.Age}
}

func TestEncode(t *testing.T) {
	t.Run("Encode String", func(t *testing.T) {
		input := "hello"
		expected := `"hello"`
		var result []byte
		err := json.Encode(input, &result)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}
		if string(result) != expected {
			t.Errorf("Expected %s, got %s", expected, string(result))
		}
	})

	t.Run("Encode Int", func(t *testing.T) {
		input := 123
		expected := "123"
		var result []byte
		err := json.Encode(input, &result)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}
		if string(result) != expected {
			t.Errorf("Expected %s, got %s", expected, string(result))
		}
	})

	t.Run("Encode Fielder", func(t *testing.T) {
		input := &TestStruct{Name: "Alice", Age: 30}
		var result []byte
		err := json.Encode(input, &result)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}
		resStr := string(result)
		expected := `{"name":"Alice","age":30}`
		if resStr != expected {
			t.Errorf("Expected %s, got %s", expected, resStr)
		}
	})
}

func TestDecode(t *testing.T) {
	t.Run("Decode String", func(t *testing.T) {
		input := `"world"`
		var result string
		err := json.Decode([]byte(input), &result)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if result != "world" {
			t.Errorf("Expected 'world', got '%s'", result)
		}
	})

	t.Run("Decode Int", func(t *testing.T) {
		input := "456"
		var result int64
		err := json.Decode([]byte(input), &result)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if result != 456 {
			t.Errorf("Expected 456, got %d", result)
		}
	})

	t.Run("Decode Fielder", func(t *testing.T) {
		input := `{"name":"Bob","age":25}`
		result := &TestStruct{}
		err := json.Decode([]byte(input), result)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if result.Name != "Bob" || result.Age != 25 {
			t.Errorf("Expected {Bob 25}, got %+v", result)
		}
	})
}
