package tests

import (
    stdjson "encoding/json"
    "testing"
    "github.com/tinywasm/json"
)

var benchJSONStr = `{"name":"Alice","email":"alice@example.com","age":30,"score":9.5}`

func BenchmarkDecode_tinywasm(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        u := &benchUser{}
        if err := json.Decode(benchJSONStr, u); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkDecode_stdlib(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        var u benchUser
        if err := stdjson.Unmarshal([]byte(benchJSONStr), &u); err != nil {
            b.Fatal(err)
        }
    }
}
