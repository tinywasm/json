package tests

import (
	"github.com/tinywasm/json"
	"github.com/tinywasm/fmt"
	"testing"
)

type rawFielder struct {
	raw string
}

func (r *rawFielder) EncodeFields(w fmt.FieldWriter) {
	w.Raw("raw", r.raw)
	w.Raw("null_raw", "")
}

func (r *rawFielder) DecodeFields(rdr fmt.FieldReader) {}

func (r *rawFielder) IsNil() bool {
	return r == nil
}

func TestEncodeRaw(t *testing.T) {
	r := &rawFielder{raw: `{"foo":"bar"}`}
	var out string
	if err := json.Encode(r, &out); err != nil {
		t.Fatal(err)
	}
	expected := `{"raw":{"foo":"bar"},"null_raw":null}`
	if out != expected {
		t.Errorf("expected %s, got %s", expected, out)
	}
}

type decodeRawStruct struct {
	rawVal  string
	textVal string
}

func (d *decodeRawStruct) EncodeFields(w fmt.FieldWriter) {}
func (d *decodeRawStruct) DecodeFields(r fmt.FieldReader) {
	d.rawVal, _ = r.Raw("raw")
	d.textVal, _ = r.String("text")
}
func (d *decodeRawStruct) IsNil() bool { return d == nil }

func TestDecodeRaw(t *testing.T) {
	input := `{"raw":{"foo":"bar"},"text":"hello"}`
	d := &decodeRawStruct{}

	if err := json.Decode([]byte(input), d); err != nil {
		t.Fatal(err)
	}

	if d.rawVal != `{"foo":"bar"}` {
		t.Errorf("expected %s, got %s", `{"foo":"bar"}`, d.rawVal)
	}
	if d.textVal != "hello" {
		t.Errorf("expected %s, got %s", "hello", d.textVal)
	}
}
