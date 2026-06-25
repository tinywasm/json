package json

import (
	"github.com/tinywasm/fmt"
	"io"
	"unsafe"
)

// Decode parses JSON into a Decodable.
// input: []byte | string | io.Reader.
func Decode(input any, data fmt.Decodable) error {
	if data == nil || data.IsNil() {
		return fmt.Err("json", "decode", "destination is nil")
	}

	var raw []byte
	switch in := input.(type) {
	case []byte:
		raw = in
	case string:
		// Avoid copy: parser is read-only, never modifies data.
		raw = unsafe.Slice(unsafe.StringData(in), len(in))
	case io.Reader:
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
		return fmt.Err("json", "decode", "input must be []byte, string, or io.Reader")
	}

	p := parser{data: raw}
	p.skipWhitespace()

	r := getReader()
	r.p = &p
	r.err = nil
	defer putReader(r)

	var slice fmt.FielderSlice
	if s, ok := data.(interface{ FielderSlice() fmt.FielderSlice }); ok {
		slice = s.FielderSlice()
	} else if s, ok := data.(fmt.FielderSlice); ok {
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
			if dec, ok := it.(fmt.Decodable); ok {
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
	if err := data.DecodeFields(r); err != nil {
		return err
	}
	return r.err
}
