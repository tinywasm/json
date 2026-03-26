package json

import (
	"github.com/tinywasm/fmt"
	"io"
	"unsafe"
)

// Decode parses JSON into a Fielder.
// input: []byte | string | io.Reader.
func Decode(input any, data fmt.Fielder) error {
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
	return p.parseIntoFielder(data)
}
