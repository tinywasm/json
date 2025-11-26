//go:build wasm

package tinyjson_test

import (
	"encoding/json"
	"testing"

	"github.com/cdvelop/tinyjson"
)

type BenchStruct struct {
	ID       int               `json:"id"`
	Name     string            `json:"name"`
	Email    string            `json:"email"`
	Active   bool              `json:"active"`
	Score    float64           `json:"score"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
}

var benchData = BenchStruct{
	ID:     12345,
	Name:   "John Doe",
	Email:  "john.doe@example.com",
	Active: true,
	Score:  98.5,
	Tags:   []string{"golang", "wasm", "json", "benchmark"},
	Metadata: map[string]string{
		"department": "engineering",
		"level":      "senior",
		"location":   "remote",
	},
}

var benchJSON = []byte(`{"id":12345,"name":"John Doe","email":"john.doe@example.com","active":true,"score":98.5,"tags":["golang","wasm","json","benchmark"],"metadata":{"department":"engineering","level":"senior","location":"remote"}}`)

// TinyJSON Benchmarks
func BenchmarkTinyJSON_Encode(b *testing.B) {
	tj := tinyjson.New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tj.Encode(benchData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTinyJSON_Decode(b *testing.B) {
	tj := tinyjson.New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result BenchStruct
		err := tj.Decode(benchJSON, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTinyJSON_EncodeDecode(b *testing.B) {
	tj := tinyjson.New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoded, err := tj.Encode(benchData)
		if err != nil {
			b.Fatal(err)
		}
		var result BenchStruct
		err = tj.Decode(encoded, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Standard Library JSON Benchmarks
func BenchmarkStdlib_Encode(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(benchData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdlib_Decode(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result BenchStruct
		err := json.Unmarshal(benchJSON, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdlib_EncodeDecode(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoded, err := json.Marshal(benchData)
		if err != nil {
			b.Fatal(err)
		}
		var result BenchStruct
		err = json.Unmarshal(encoded, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}
