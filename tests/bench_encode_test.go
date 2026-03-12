package tests

import (
    stdjson "encoding/json"
    "testing"
    "github.com/tinywasm/fmt"
    "github.com/tinywasm/json"
)

type benchUser struct {
    Name  string
    Email string
    Age   int64
    Score float64
}

var benchSchema = []fmt.Field{
	{Name: "Name", Type: fmt.FieldText, JSON: "name"},
	{Name: "Email", Type: fmt.FieldText, JSON: "email"},
	{Name: "Age", Type: fmt.FieldInt, JSON: "age"},
	{Name: "Score", Type: fmt.FieldFloat, JSON: "score"},
}

func (u *benchUser) Schema() []fmt.Field { return benchSchema }
func (u *benchUser) Pointers() []any { return []any{&u.Name, &u.Email, &u.Age, &u.Score} }

var benchInput = &benchUser{Name: "Alice", Email: "alice@example.com", Age: 30, Score: 9.5}

func BenchmarkEncode_tinywasm(b *testing.B) {
    var out string
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        if err := json.Encode(benchInput, &out); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkEncode_stdlib(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        if _, err := stdjson.Marshal(benchInput); err != nil {
            b.Fatal(err)
        }
    }
}
