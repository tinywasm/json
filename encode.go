package json

import (
	"io"

	"github.com/tinywasm/fmt"
)

type jsonWriter struct {
	b     *fmt.Conv
	first bool
}

func (w *jsonWriter) maybeComma() {
	if !w.first {
		w.b.WriteByte(',')
	}
	w.first = false
}

func (w *jsonWriter) writeKey(name string) {
	if name != "" {
		w.b.WriteByte('"')
		fmt.JSONEscape(name, w.b)
		w.b.WriteByte('"')
		w.b.WriteByte(':')
	}
}

func (w *jsonWriter) String(name, val string) {
	w.maybeComma()
	w.writeKey(name)
	w.b.WriteByte('"')
	fmt.JSONEscape(val, w.b)
	w.b.WriteByte('"')
}

func (w *jsonWriter) Int(name string, val int64) {
	w.maybeComma()
	w.writeKey(name)
	w.b.WriteInt(val)
}

func (w *jsonWriter) Uint(name string, val uint64) {
	w.maybeComma()
	w.writeKey(name)
	w.b.WriteString(fmt.Convert(val).String())
}

func (w *jsonWriter) Float(name string, val float64) {
	w.maybeComma()
	w.writeKey(name)
	w.b.WriteFloat(val)
}

func (w *jsonWriter) Bool(name string, val bool) {
	w.maybeComma()
	w.writeKey(name)
	if val {
		w.b.WriteString("true")
	} else {
		w.b.WriteString("false")
	}
}

func (w *jsonWriter) Bytes(name string, val []byte) {
	w.maybeComma()
	w.writeKey(name)
	w.b.WriteByte('"')
	fmt.JSONEscape(string(val), w.b)
	w.b.WriteByte('"')
}

func (w *jsonWriter) Null(name string) {
	w.maybeComma()
	w.writeKey(name)
	w.b.WriteString("null")
}

func (w *jsonWriter) Raw(name, val string) {
	w.maybeComma()
	w.writeKey(name)
	if val == "" {
		w.b.WriteString("null")
		return
	}
	w.b.WriteString(val)
}

func (w *jsonWriter) Object(name string, val fmt.Encodable) {
	w.maybeComma()
	w.writeKey(name)
	if val == nil || val.IsNil() {
		w.b.WriteString("null")
		return
	}
	if r, ok := val.(interface{ Raw() string }); ok {
		w.b.WriteString(r.Raw())
		return
	}

	w.b.WriteByte('{')
	iw := getWriter()
	iw.b = w.b
	iw.first = true
	val.EncodeFields(iw)
	w.b.WriteByte('}')
	putWriter(iw)
}

func (w *jsonWriter) Array(name string, n int) fmt.ArrayWriter {
	w.maybeComma()
	w.writeKey(name)
	w.b.WriteByte('[')
	aw := getArrayWriter()
	aw.b = w.b
	aw.first = true
	return aw
}

type jsonArrayWriter struct {
	b     *fmt.Conv
	first bool
}

func (w *jsonArrayWriter) maybeComma() {
	if !w.first {
		w.b.WriteByte(',')
	}
	w.first = false
}

func (w *jsonArrayWriter) String(val string) {
	w.maybeComma()
	w.b.WriteByte('"')
	fmt.JSONEscape(val, w.b)
	w.b.WriteByte('"')
}

func (w *jsonArrayWriter) Int(val int64) {
	w.maybeComma()
	w.b.WriteInt(val)
}

func (w *jsonArrayWriter) Float(val float64) {
	w.maybeComma()
	w.b.WriteFloat(val)
}

func (w *jsonArrayWriter) Bool(val bool) {
	w.maybeComma()
	if val {
		w.b.WriteString("true")
	} else {
		w.b.WriteString("false")
	}
}

func (w *jsonArrayWriter) Bytes(val []byte) {
	w.maybeComma()
	w.b.WriteByte('"')
	fmt.JSONEscape(string(val), w.b)
	w.b.WriteByte('"')
}

func (w *jsonArrayWriter) Object(val fmt.Encodable) {
	w.maybeComma()
	if val == nil || val.IsNil() {
		w.b.WriteString("null")
		return
	}
	if r, ok := val.(interface{ Raw() string }); ok {
		w.b.WriteString(r.Raw())
		return
	}
	w.b.WriteByte('{')
	iw := getWriter()
	iw.b = w.b
	iw.first = true
	val.EncodeFields(iw)
	w.b.WriteByte('}')
	putWriter(iw)
}

func (w *jsonArrayWriter) Close() {
	w.b.WriteByte(']')
	putArrayWriter(w)
}

// Encode serializes an Encodable to JSON.
// output: *[]byte | *string | io.Writer.
func Encode(data fmt.Encodable, output any) error {
	b := fmt.GetConv()
	defer b.PutConv()

	if data == nil || data.IsNil() {
		b.WriteString("null")
	} else {
		w := getWriter()
		w.b = b
		w.first = true

		var slice fmt.FielderSlice
		if s, ok := data.(interface{ FielderSlice() fmt.FielderSlice }); ok {
			slice = s.FielderSlice()
		} else if s, ok := data.(fmt.FielderSlice); ok {
			slice = s
		}

		if slice != nil {
			b.WriteByte('[')
			for i := 0; i < slice.Len(); i++ {
				if i > 0 {
					b.WriteByte(',')
				}
				if it, ok := slice.At(i).(fmt.Encodable); ok {
					iw := getWriter()
					iw.b = b
					iw.first = true
					if it.IsNil() {
						b.WriteString("null")
					} else {
						b.WriteByte('{')
						it.EncodeFields(iw)
						b.WriteByte('}')
					}
					putWriter(iw)
				} else {
					b.WriteString("null")
				}
			}
			b.WriteByte(']')
		} else {
			b.WriteByte('{')
			data.EncodeFields(w)
			b.WriteByte('}')
		}
		putWriter(w)
	}

	if b.GetString(fmt.BuffErr) != "" {
		return fmt.Err(b.GetString(fmt.BuffErr))
	}

	switch out := output.(type) {
	case *[]byte:
		res := b.Bytes()
		cpy := make([]byte, len(res))
		copy(cpy, res)
		*out = cpy
	case *string:
		*out = b.GetString(fmt.BuffOut)
	case io.Writer:
		_, err := out.Write(b.Bytes())
		return err
	default:
		return fmt.Err("json", "encode", "output must be *[]byte, *string, or io.Writer")
	}
	return nil
}
