package tests

import "github.com/tinywasm/model"

import (
	stdjson "encoding/json"
	"github.com/tinywasm/json"
	"testing"
)

type benchUser struct {
	Name  string
	Email string
	Age   int64
	Score float64
}

func (u *benchUser) IsNil() bool { return u == nil }
func (u *benchUser) EncodeFields(w model.FieldWriter) {
	w.String("name", u.Name)
	w.String("email", u.Email)
	w.Int("age", u.Age)
	w.Float("score", u.Score)
}
func (u *benchUser) DecodeFields(r model.FieldReader) {
	u.Name, _ = r.String("name")
	u.Email, _ = r.String("email")
	u.Age, _ = r.Int("age")
	u.Score, _ = r.Float("score")
}

func (u *benchUser) Schema() []model.Field { return nil }
func (u *benchUser) Pointers() []any       { return nil }

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
