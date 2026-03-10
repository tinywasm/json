package json

import (
	"github.com/tinywasm/fmt"
	"io"
)

// Decode parses JSON into a Fielder.
// input: []byte | string | io.Reader.
func Decode(input any, data fmt.Fielder) error {
	var raw []byte
	switch in := input.(type) {
	case []byte:
		raw = in
	case string:
		raw = []byte(in)
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

	p := &parser{data: raw}
	parsed, err := p.parseValue()
	if err != nil {
		return err
	}

	obj, ok := parsed.(map[string]any)
	if !ok {
		return fmt.Err("json", "decode", "expected JSON object for Fielder")
	}
	return decodeFielder(obj, data)
}

func decodeFielder(obj map[string]any, f fmt.Fielder) error {
	schema := f.Schema()
	pointers := f.Pointers()

	for i, field := range schema {
		key, _ := parseJSONTag(field)
		if key == "-" {
			continue
		}

		val, exists := obj[key]
		if !exists {
			continue
		}

		ptr := pointers[i]

		// Nested struct: recurse
		if field.Type == fmt.FieldStruct {
			if nested, ok := ptr.(fmt.Fielder); ok {
				if innerObj, ok := val.(map[string]any); ok {
					if err := decodeFielder(innerObj, nested); err != nil {
						return err
					}
					continue
				}
			}
		}

		writeValue(ptr, field.Type, val)
	}
	return nil
}

// writeValue writes a parsed JSON value into a Go pointer.
// Uses fmt.Convert for type coercion where needed.
func writeValue(ptr any, ft fmt.FieldType, val any) {
	switch ft {
	case fmt.FieldText:
		if p, ok := ptr.(*string); ok {
			if s, ok := val.(string); ok {
				*p = s
			}
		}
	case fmt.FieldInt:
		// Parser returns int64 for integers, float64 for decimals.
		// Support *int, *int32, *int64 via type switches on ptr.
		switch p := ptr.(type) {
		case *int64:
			switch v := val.(type) {
			case int64:
				*p = v
			case float64:
				*p = int64(v)
			}
		case *int:
			switch v := val.(type) {
			case int64:
				*p = int(v)
			case float64:
				*p = int(v)
			}
		case *int32:
			switch v := val.(type) {
			case int64:
				*p = int32(v)
			case float64:
				*p = int32(v)
			}
		}
	case fmt.FieldFloat:
		switch p := ptr.(type) {
		case *float64:
			switch v := val.(type) {
			case float64:
				*p = v
			case int64:
				*p = float64(v)
			}
		case *float32:
			switch v := val.(type) {
			case float64:
				*p = float32(v)
			case int64:
				*p = float32(v)
			}
		}
	case fmt.FieldBool:
		if p, ok := ptr.(*bool); ok {
			if b, ok := val.(bool); ok {
				*p = b
			}
		}
	case fmt.FieldBlob:
		if p, ok := ptr.(*[]byte); ok {
			if s, ok := val.(string); ok {
				*p = []byte(s)
			}
		}
	}
}
