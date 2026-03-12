package json

import (
	"github.com/tinywasm/fmt"
	"io"
)

// Encode serializes a Fielder to JSON.
// output: *[]byte | *string | io.Writer.
func Encode(data fmt.Fielder, output any) error {
	var b fmt.Builder
	if err := encodeFielder(&b, data); err != nil {
		return err
	}
	result := b.String()

	switch out := output.(type) {
	case *[]byte:
		*out = []byte(result)
	case *string:
		*out = result
	case io.Writer:
		_, err := out.Write([]byte(result))
		return err
	default:
		return fmt.Err("json", "encode", "output must be *[]byte, *string, or io.Writer")
	}
	return nil
}

func encodeFielder(b *fmt.Builder, f fmt.Fielder) error {
	schema := f.Schema()
	values := f.Values()
	if values == nil && schema != nil {
		return fmt.Err("json", "encode", "failed to get values")
	}
	b.WriteByte('{')

	first := true
	for i, field := range schema {
		key, omitempty := parseJSONTag(field)
		if key == "-" {
			continue
		}

		val := values[i]

		if omitempty && fmt.IsZero(val) {
			continue
		}

		if !first {
			b.WriteByte(',')
		}
		first = false

		// Write key
		b.WriteByte('"')
		fmt.JSONEscape(key, b)
		b.WriteByte('"')
		b.WriteByte(':')

		// Write value
		if field.Type == fmt.FieldStruct {
			if nested, ok := val.(fmt.Fielder); ok {
				if err := encodeFielder(b, nested); err != nil {
					return err
				}
				continue
			}
		}

		encodeValue(b, val)
	}

	b.WriteByte('}')
	return nil
}

// encodeValue writes a single Go value as JSON.
// Only handles types that appear in Fielder.Values(): string, int variants,
// float variants, bool, []byte, nil.
func encodeValue(b *fmt.Builder, v any) {
	switch val := v.(type) {
	case nil:
		b.WriteString("null")
	case string:
		b.WriteByte('"')
		fmt.JSONEscape(val, b)
		b.WriteByte('"')
	case bool:
		if val {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case []byte:
		b.WriteByte('"')
		fmt.JSONEscape(string(val), b)
		b.WriteByte('"')
	default:
		// int, int8..int64, uint..uint64, float32, float64
		// All handled by fmt.Convert which already supports every numeric type.
		b.WriteString(fmt.Convert(val).String())
	}
}

// parseJSONTag extracts key and omitempty from Field.JSON.
func parseJSONTag(f fmt.Field) (key string, omitempty bool) {
	tag := f.JSON
	if tag == "" {
		return f.Name, false
	}
	if tag == "-" {
		return "-", false
	}
	comma := -1
	for i := 0; i < len(tag); i++ {
		if tag[i] == ',' {
			comma = i
			break
		}
	}
	if comma < 0 {
		return tag, false
	}
	key = tag[:comma]
	if key == "" {
		key = f.Name
	}
	return key, tag[comma+1:] == "omitempty"
}
