package tests

import (
    stdjson "encoding/json"
    "testing"
    "github.com/tinywasm/json"
)

func BenchmarkRoundTrip_tinywasm(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        var out string
        if err := json.Encode(benchInput, &out); err != nil {
            b.Fatal(err)
        }
        u := &benchUser{}
        if err := json.Decode(out, u); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkRoundTrip_stdlib(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        data, err := stdjson.Marshal(benchInput)
        if err != nil {
            b.Fatal(err)
        }
        var u benchUser
        if err := stdjson.Unmarshal(data, &u); err != nil {
            b.Fatal(err)
        }
    }
}
