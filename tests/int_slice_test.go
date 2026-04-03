package tests

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/json"
	"testing"
)

// FontDef simulates fpdf's fontDefType which has a []int field (Cw = character widths).
type FontDef struct {
	Name string
	Cw   []int
}

func (f *FontDef) Schema() []fmt.Field {
	return []fmt.Field{
		{Name: "Name", Type: fmt.FieldText},
		{Name: "Cw", Type: fmt.FieldIntSlice},
	}
}

func (f *FontDef) Pointers() []any {
	return []any{&f.Name, &f.Cw}
}

func TestDecodeIntSlice(t *testing.T) {
	input := `{"Name":"Courier","Cw":[600,500,400]}`
	result := &FontDef{}
	err := json.Decode(input, result)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if result.Name != "Courier" {
		t.Errorf("Name = %q, want Courier", result.Name)
	}
	if len(result.Cw) != 3 {
		t.Fatalf("Cw length = %d, want 3", len(result.Cw))
	}
	expected := []int{600, 500, 400}
	for i, v := range expected {
		if result.Cw[i] != v {
			t.Errorf("Cw[%d] = %d, want %d", i, result.Cw[i], v)
		}
	}
}

func TestEncodeIntSlice(t *testing.T) {
	input := &FontDef{Name: "Courier", Cw: []int{600, 500, 400}}
	var result string
	err := json.Encode(input, &result)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	expected := `{"Name":"Courier","Cw":[600,500,400]}`
	if result != expected {
		t.Errorf("got %s, want %s", result, expected)
	}
}

func TestDecodeEmptyIntSlice(t *testing.T) {
	input := `{"Name":"Empty","Cw":[]}`
	result := &FontDef{}
	err := json.Decode(input, result)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if len(result.Cw) != 0 {
		t.Errorf("Cw length = %d, want 0", len(result.Cw))
	}
}

func TestDecodeNullIntSlice(t *testing.T) {
	input := `{"Name":"Null","Cw":null}`
	result := &FontDef{}
	err := json.Decode(input, result)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if result.Cw != nil {
		t.Errorf("Cw should be nil for null, got %v", result.Cw)
	}
}
