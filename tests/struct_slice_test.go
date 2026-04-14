package tests

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/json"
	"testing"
)

type mockFielderSlice struct {
	items []fmt.Fielder
}

func (s *mockFielderSlice) Len() int           { return len(s.items) }
func (s *mockFielderSlice) At(i int) fmt.Fielder { return s.items[i] }
func (s *mockFielderSlice) Append() fmt.Fielder {
	it := &item{}
	s.items = append(s.items, it)
	return it
}

func (s *mockFielderSlice) Schema() []fmt.Field { return nil }
func (s *mockFielderSlice) Pointers() []any     { return nil }

type item struct {
	ID   int
	Name string
}

func (i *item) Schema() []fmt.Field {
	return []fmt.Field{
		{Name: "id", Type: fmt.FieldInt},
		{Name: "name", Type: fmt.FieldText},
	}
}

func (i *item) Pointers() []any {
	return []any{&i.ID, &i.Name}
}

func TestEncodeFieldStructSlice(t *testing.T) {
	slice := &mockFielderSlice{
		items: []fmt.Fielder{
			&item{ID: 1, Name: "Alice"},
			&item{ID: 2, Name: "Bob"},
		},
	}

	parent := &mockFielder{
		schema: []fmt.Field{
			{Name: "staff_name", Type: fmt.FieldText},
			{Name: "schedule", Type: fmt.FieldStructSlice},
		},
		pointers: []any{ptrString("Dra. Ana"), slice},
	}

	var out string
	if err := json.Encode(parent, &out); err != nil {
		t.Fatalf("failed to encode: %v", err)
	}

	expected := `{"staff_name":"Dra. Ana","schedule":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeFieldStructSlice_Empty(t *testing.T) {
	slice := &mockFielderSlice{
		items: []fmt.Fielder{},
	}

	parent := &mockFielder{
		schema: []fmt.Field{
			{Name: "schedule", Type: fmt.FieldStructSlice},
		},
		pointers: []any{slice},
	}

	var out string
	if err := json.Encode(parent, &out); err != nil {
		t.Fatalf("failed to encode: %v", err)
	}

	expected := `{"schedule":[]}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeFieldStructSlice_Nil(t *testing.T) {
	parent := &mockFielder{
		schema: []fmt.Field{
			{Name: "schedule", Type: fmt.FieldStructSlice},
		},
		pointers: []any{nil},
	}

	var out string
	if err := json.Encode(parent, &out); err != nil {
		t.Fatalf("failed to encode: %v", err)
	}

	expected := `{"schedule":null}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestDecodeFieldStructSlice(t *testing.T) {
	input := `{"staff_name":"Dra. Ana","schedule":[{"id":101,"name":"Alice"},{"id":102,"name":"Bob"}]}`
	
	slice := &mockFielderSlice{}
	
	// Wait! mockFielder in helpers_test.go doesn't have a way to store data outside pointers.
	var staffName string
	parent := &mockFielder{
		schema: []fmt.Field{
			{Name: "staff_name", Type: fmt.FieldText},
			{Name: "schedule", Type: fmt.FieldStructSlice},
		},
		pointers: []any{&staffName, slice},
	}

	if err := json.Decode(input, parent); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if staffName != "Dra. Ana" {
		t.Errorf("expected staff_name 'Dra. Ana', got %q", staffName)
	}

	if len(slice.items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(slice.items))
	}

	it1 := slice.items[0].(*item)
	if it1.ID != 101 || it1.Name != "Alice" {
		t.Errorf("item 1 mismatch: %+v", it1)
	}

	it2 := slice.items[1].(*item)
	if it2.ID != 102 || it2.Name != "Bob" {
		t.Errorf("item 2 mismatch: %+v", it2)
	}
}

func TestEncode_RootSlice(t *testing.T) {
	slice := &mockFielderSlice{items: []fmt.Fielder{
		&item{ID: 1, Name: "Alice"},
		&item{ID: 2, Name: "Bob"},
	}}
	var out string
	if err := json.Encode(slice, &out); err != nil {
		t.Fatal(err)
	}
	want := `[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`
	if out != want {
		t.Errorf("got %s", out)
	}
}

func TestDecode_RootSlice(t *testing.T) {
	slice := &mockFielderSlice{}
	if err := json.Decode(`[{"id":1,"name":"Alice"}]`, slice); err != nil {
		t.Fatal(err)
	}
	if slice.Len() != 1 || slice.items[0].(*item).Name != "Alice" {
		t.Error("decode mismatch")
	}
}
