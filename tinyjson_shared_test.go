package tinyjson_test

import (
	"reflect"
	"testing"

	"github.com/cdvelop/tinyjson"
)

type TestStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func EncodeShared(t *testing.T, j *tinyjson.TinyJSON) {
	t.Run("Encode String", func(t *testing.T) {
		input := "hello"
		expected := `"hello"`
		result, err := j.Encode(input)
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
		result, err := j.Encode(input)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}
		if string(result) != expected {
			t.Errorf("Expected %s, got %s", expected, string(result))
		}
	})

	t.Run("Encode Struct", func(t *testing.T) {
		input := TestStruct{Name: "Alice", Age: 30}
		// JSON key order is not guaranteed, so we might need to check fields or use a more robust comparison if strict string match fails often.
		// For simple structs in standard json, it's usually consistent, but let's see.
		// Actually, for this simple case, let's just check if it contains the keys and values.
		result, err := j.Encode(input)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}
		resStr := string(result)
		if resStr != `{"name":"Alice","age":30}` && resStr != `{"age":30,"name":"Alice"}` {
			t.Errorf("Expected JSON representation of struct, got %s", resStr)
		}
	})

	t.Run("Encode Slice of Structs", func(t *testing.T) {
		input := []TestStruct{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
		}
		result, err := j.Encode(input)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}
		resStr := string(result)
		t.Logf("Encoded slice of structs: %s", resStr)

		// Should be a JSON array, not an empty string
		if resStr == `""` || resStr == "" {
			t.Errorf("BUG: Slice of structs encoded as empty string instead of JSON array, got: %s", resStr)
		}

		// Verify it's a valid JSON array
		if len(resStr) < 2 || resStr[0] != '[' || resStr[len(resStr)-1] != ']' {
			t.Errorf("Expected JSON array format [...], got: %s", resStr)
		}
	})
}

func DecodeShared(t *testing.T, j *tinyjson.TinyJSON) {
	t.Run("Decode String", func(t *testing.T) {
		input := `"world"`
		var result string
		err := j.Decode([]byte(input), &result)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if result != "world" {
			t.Errorf("Expected 'world', got '%s'", result)
		}
	})

	t.Run("Decode Int", func(t *testing.T) {
		input := "456"
		var result int
		err := j.Decode([]byte(input), &result)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if result != 456 {
			t.Errorf("Expected 456, got %d", result)
		}
	})

	t.Run("Decode Struct", func(t *testing.T) {
		input := `{"name":"Bob","age":25}`
		var result TestStruct
		err := j.Decode([]byte(input), &result)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		expected := TestStruct{Name: "Bob", Age: 25}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %+v, got %+v", expected, result)
		}
	})

	t.Run("Decode Slice of Structs", func(t *testing.T) {
		input := `[{"name":"Alice","age":30},{"name":"Bob","age":25}]`
		var result []TestStruct
		err := j.Decode([]byte(input), &result)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 structs, got %d", len(result))
		}

		if len(result) > 0 && result[0].Name != "Alice" {
			t.Errorf("Expected first name 'Alice', got '%s'", result[0].Name)
		}

		if len(result) > 1 && result[1].Name != "Bob" {
			t.Errorf("Expected second name 'Bob', got '%s'", result[1].Name)
		}
	})
}
