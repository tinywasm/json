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
	{Name: "name", Type: fmt.FieldText},
	{Name: "email", Type: fmt.FieldText},
	{Name: "age", Type: fmt.FieldInt},
	{Name: "score", Type: fmt.FieldFloat},
}

func (u *benchUser) Schema() []fmt.Field { return benchSchema }
func (u *benchUser) Pointers() []any { return []any{&u.Name, &u.Email, &u.Age, &u.Score} }

var benchInput = &benchUser{Name: "alice", Email: "alice@example.com", Age: 30, Score: 9.5}

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
