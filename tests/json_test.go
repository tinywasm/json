package tests

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/json"
	"testing"
)

type TestStruct struct {
	Name string
	Age  int64
}

func (s *TestStruct) Schema() []fmt.Field {
	return []fmt.Field{
		{Name: "name", Type: fmt.FieldText},
		{Name: "age", Type: fmt.FieldInt},
	}
}

func (s *TestStruct) Pointers() []any {
	return []any{&s.Name, &s.Age}
}

func TestIntegration(t *testing.T) {
	t.Run("Encode Fielder", func(t *testing.T) {
		input := &TestStruct{Name: "alice", Age: 30}
		var result string
		err := json.Encode(input, &result)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}
		expected := `{"name":"alice","age":30}`
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("Decode Fielder", func(t *testing.T) {
		input := `{"name":"Bob","age":25}`
		result := &TestStruct{}
		err := json.Decode(input, result)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if result.Name != "Bob" || result.Age != 25 {
			t.Errorf("Expected {Bob 25}, got %+v", result)
		}
	})
}
