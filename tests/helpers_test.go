package tests

import "github.com/tinywasm/fmt"

type mockFielder struct {
	schema   []fmt.Field
	values   []any
	pointers []any
	err      error
}

func (m *mockFielder) Schema() []fmt.Field { return m.schema }
func (m *mockFielder) Values() []any       {
	if m.err != nil {
		return nil
	}
	return m.values
}
func (m *mockFielder) Pointers() []any { return m.pointers }
