package tests

import "github.com/tinywasm/fmt"

type mockFielder struct {
    schema   []fmt.Field
    values   []any
    pointers []any
}

func (m *mockFielder) Schema() []fmt.Field { return m.schema }
func (m *mockFielder) Values() []any       { return m.values }
func (m *mockFielder) Pointers() []any     { return m.pointers }
