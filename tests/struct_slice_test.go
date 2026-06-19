package tests

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/json"
	"testing"
)

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

func (i *item) IsNil() bool { return i == nil }
func (i *item) EncodeFields(w fmt.FieldWriter) {
	w.Int("id", int64(i.ID))
	w.String("name", i.Name)
}
func (i *item) DecodeFields(r fmt.FieldReader) error {
	id, _ := r.Int("id")
	i.ID = int(id)
	i.Name, _ = r.String("name")
	return nil
}

type itemSlice struct {
	items []*item
}

func (s *itemSlice) Len() int           { return len(s.items) }
func (s *itemSlice) At(i int) fmt.Fielder { return s.items[i] }
func (s *itemSlice) Append() fmt.Fielder {
	it := &item{}
	s.items = append(s.items, it)
	return it
}
func (s *itemSlice) IsNil() bool                       { return s == nil }
func (s *itemSlice) EncodeFields(w fmt.FieldWriter)    {}
func (s *itemSlice) DecodeFields(r fmt.FieldReader) error { return nil }
func (s *itemSlice) Schema() []fmt.Field { return nil }
func (s *itemSlice) Pointers() []any     { return nil }
func (s *itemSlice) FielderSlice() fmt.FielderSlice    { return s }

type rootModel struct {
	Name     string
	Schedule *itemSlice
}

func (m *rootModel) IsNil() bool { return m == nil }
func (m *rootModel) EncodeFields(w fmt.FieldWriter) {
	w.String("staff_name", m.Name)
	if m.Schedule != nil {
		n := m.Schedule.Len()
		aw := w.Array("schedule", n)
		for i := 0; i < n; i++ {
			if it, ok := m.Schedule.At(i).(fmt.Encodable); ok {
				aw.Object(it)
			}
		}
		if closer, ok := aw.(interface{ Close() }); ok {
			closer.Close()
		}
	} else {
		w.Null("schedule")
	}
}
func (m *rootModel) DecodeFields(r fmt.FieldReader) error {
	m.Name, _ = r.String("staff_name")
	if ar, ok := r.Array("schedule"); ok {
		n := ar.Len()
		m.Schedule = &itemSlice{}
		for i := 0; i < n; i++ {
			it := &item{}
			m.Schedule.items = append(m.Schedule.items, it)
			ar.Object(i, it)
		}
	}
	return nil
}

func TestEncodeFieldStructSlice(t *testing.T) {
	slice := &itemSlice{
		items: []*item{
			{ID: 1, Name: "Alice"},
			{ID: 2, Name: "Bob"},
		},
	}

	parent := &rootModel{
		Name:     "Dra. Ana",
		Schedule: slice,
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
	slice := &itemSlice{
		items: []*item{},
	}

	parent := &rootModel{
		Schedule: slice,
	}

	var out string
	if err := json.Encode(parent, &out); err != nil {
		t.Fatalf("failed to encode: %v", err)
	}

	expected := `{"staff_name":"","schedule":[]}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestEncodeFieldStructSlice_Nil(t *testing.T) {
	parent := &rootModel{
		Schedule: nil,
	}

	var out string
	if err := json.Encode(parent, &out); err != nil {
		t.Fatalf("failed to encode: %v", err)
	}

	expected := `{"staff_name":"","schedule":null}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

func TestDecodeFieldStructSlice(t *testing.T) {
	input := `{"staff_name":"Dra. Ana","schedule":[{"id":101,"name":"Alice"},{"id":102,"name":"Bob"}]}`
	
	parent := &rootModel{}

	if err := json.Decode(input, parent); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if parent.Name != "Dra. Ana" {
		t.Errorf("expected staff_name 'Dra. Ana', got %q", parent.Name)
	}

	if parent.Schedule == nil || len(parent.Schedule.items) != 2 {
		t.Fatalf("expected 2 items, got %v", parent.Schedule)
	}

	it1 := parent.Schedule.items[0]
	if it1.ID != 101 || it1.Name != "Alice" {
		t.Errorf("item 1 mismatch: %+v", it1)
	}

	it2 := parent.Schedule.items[1]
	if it2.ID != 102 || it2.Name != "Bob" {
		t.Errorf("item 2 mismatch: %+v", it2)
	}
}

func TestEncode_RootSlice(t *testing.T) {
	slice := &itemSlice{items: []*item{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
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

func TestEncode_RootSlice_Empty(t *testing.T) {
	slice := &itemSlice{}
	var out string
	if err := json.Encode(slice, &out); err != nil {
		t.Fatal(err)
	}
	if out != "[]" {
		t.Errorf("got %s, want []", out)
	}
}

func TestDecode_RootSlice(t *testing.T) {
	slice := &itemSlice{}
	if err := json.Decode(`[{"id":1,"name":"Alice"}]`, slice); err != nil {
		t.Fatal(err)
	}
	if slice.Len() != 1 || slice.items[0].Name != "Alice" {
		t.Error("decode mismatch")
	}
}

func TestDecode_RootSlice_Empty(t *testing.T) {
	slice := &itemSlice{}
	if err := json.Decode(`[]`, slice); err != nil {
		t.Fatal(err)
	}
	if slice.Len() != 0 {
		t.Errorf("expected empty slice, got %d items", slice.Len())
	}
}

func TestDecode_RootSlice_InvalidInput(t *testing.T) {
	slice := &itemSlice{}
	if err := json.Decode(`{"id":1}`, slice); err == nil {
		t.Fatal("expected error for object input, got nil")
	}
}
