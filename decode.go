package json

import "github.com/tinywasm/model"

import (
	"github.com/tinywasm/fmt"
	"unsafe"
)

// Reader is io.Reader redeclared here, for the same reason as Writer.
type Reader interface {
	Read(p []byte) (n int, err error)
}

// Decode parses JSON into a Decodable.
// input: []byte | string | Reader.
func Decode[T model.Decodable](input any, data T) error {
	if any(data) == nil || data.IsNil() {
		return fmt.Err("json", "decode", "destination is nil")
	}

	var raw []byte
	switch in := input.(type) {
	case []byte:
		raw = in
	case string:
		// Avoid copy: parser is read-only, never modifies data.
		raw = unsafe.Slice(unsafe.StringData(in), len(in))
	case Reader:
		var buf []byte
		tmp := make([]byte, 4096)
		for {
			n, err := in.Read(tmp)
			if n > 0 {
				buf = append(buf, tmp[:n]...)
			}
			if err != nil {
				break
			}
		}
		raw = buf
	default:
		return fmt.Err("json", "decode", "input must be []byte, string, or json.Reader")
	}

	p := parser{data: raw}
	p.skipWhitespace()

	r := getReader()
	r.p = &p
	r.err = nil
	defer putReader(r)

	var slice model.FielderSlice
	if s, ok := any(data).(interface{ FielderSlice() model.FielderSlice }); ok {
		slice = s.FielderSlice()
	} else if s, ok := any(data).(model.FielderSlice); ok {
		slice = s
	}

	if slice != nil {
		// Special case for root level arrays of Fielders
		if p.peek() != '[' {
			return fmt.Err("json", "decode", "expected array, got "+string(p.peek()))
		}
		arrayStart := p.pos
		ar := jsonArrayReader{p: &p, start: arrayStart}
		for i := 0; i < ar.Len(); i++ {
			it := slice.Append()
			if dec, ok := it.(model.Decodable); ok {
				ar.Object(i, dec)
			}
		}
		return nil
	}

	if p.peek() != '{' {
		return fmt.Err("json", "decode", "expected object, got "+string(p.peek()))
	}
	start := p.pos
	r.start = start
	data.DecodeFields(r)
	return r.err
}
