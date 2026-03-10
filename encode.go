package json

import (
	"github.com/tinywasm/fmt"
	"io"
)

// encodeWithInternal is the internal implementation of Encode.
func encodeWithInternal(input any, output any) error {
	c := fmt.Convert()
	if err := encode(c, input); err != nil {
		c.Reset()
		return err
	}
	result := c.String()

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

func encode(c *fmt.Conv, data any) error {
	switch v := data.(type) {
	// 1. Fielder — zero reflect, uses Schema + Values
	case fmt.Fielder:
		return encodeFielder(c, v)

	// 2. Primitives
	case string:
		encodeString(c, v)
	case bool:
		if v {
			c.Write("true")
		} else {
			c.Write("false")
		}
	case int:
		c.Write(v)
	case int8:
		c.Write(v)
	case int16:
		c.Write(v)
	case int32:
		c.Write(v)
	case int64:
		c.Write(v)
	case uint:
		c.Write(v)
	case uint8:
		c.Write(v)
	case uint16:
		c.Write(v)
	case uint32:
		c.Write(v)
	case uint64:
		c.Write(v)
	case float32:
		c.Write(v)
	case float64:
		c.Write(v)

	// 3. []byte → JSON string
	case []byte:
		encodeString(c, string(v))

	// 4. Known slice types
	case []string:
		return encodeSlice(c, len(v), func(i int) error {
			encodeString(c, v[i])
			return nil
		})
	case []int:
		return encodeSlice(c, len(v), func(i int) error {
			c.Write(v[i])
			return nil
		})
	case []any:
		return encodeSlice(c, len(v), func(i int) error {
			return encode(c, v[i])
		})

	// 5. Known map types
	case map[string]any:
		return encodeMap(c, v)
	case map[string]string:
		return encodeStringMap(c, v)

	// 6. nil
	case nil:
		c.Write("null")

	// 7. Unknown → error
	default:
		return fmt.Err("json", "encode", "unsupported type, implement fmt.Fielder via ormc")
	}
	return nil
}

func encodeSlice(c *fmt.Conv, length int, encodeItem func(int) error) error {
	c.Write("[")
	for i := 0; i < length; i++ {
		if i > 0 {
			c.Write(",")
		}
		if err := encodeItem(i); err != nil {
			return err
		}
	}
	c.Write("]")
	return nil
}

func encodeMap(c *fmt.Conv, m map[string]any) error {
	c.Write("{")
	first := true
	for k, v := range m {
		if !first {
			c.Write(",")
		}
		first = false
		encodeString(c, k)
		c.Write(":")
		if err := encode(c, v); err != nil {
			return err
		}
	}
	c.Write("}")
	return nil
}

func encodeStringMap(c *fmt.Conv, m map[string]string) error {
	c.Write("{")
	first := true
	for k, v := range m {
		if !first {
			c.Write(",")
		}
		first = false
		encodeString(c, k)
		c.Write(":")
		encodeString(c, v)
	}
	c.Write("}")
	return nil
}

func encodeFielder(c *fmt.Conv, f fmt.Fielder) error {
	schema := f.Schema()
	values := f.Values()
	c.Write("{")

	first := true
	for i, field := range schema {
		// Parse JSON key from Field.JSON (or fall back to Field.Name)
		key, omitempty := parseJSONTag(field)
		if key == "-" {
			continue
		}

		val := values[i]

		// Handle omitempty: skip zero values
		if omitempty && isZero(val) {
			continue
		}

		if !first {
			c.Write(",")
		}
		first = false

		// Write key
		encodeString(c, key)
		c.Write(":")

		// Write value — recurse if nested Fielder
		if field.Type == fmt.FieldStruct {
			if nested, ok := val.(fmt.Fielder); ok {
				if err := encodeFielder(c, nested); err != nil {
					return err
				}
				continue
			}
		}

		if err := encode(c, val); err != nil {
			return err
		}
	}

	c.Write("}")
	return nil
}

// parseJSONTag extracts key and omitempty from Field.JSON.
// If Field.JSON is empty, returns Field.Name as key.
func parseJSONTag(f fmt.Field) (key string, omitempty bool) {
	tag := f.JSON
	if tag == "" {
		return f.Name, false
	}
	if tag == "-" {
		return "-", false
	}
	// Split on comma: "email,omitempty" → key="email", omitempty=true
	parts := fmt.Convert(tag).Split(",")
	key = parts[0]
	if key == "" {
		key = f.Name
	}
	for i := 1; i < len(parts); i++ {
		if parts[i] == "omitempty" {
			omitempty = true
		}
	}
	return key, omitempty
}

// encodeString writes a JSON-escaped string with quotes.
func encodeString(c *fmt.Conv, s string) {
	c.Write("\"")
	// Escape: \", \\, \n, \r, \t, control chars
	for i := 0; i < len(s); i++ {
		b := s[i]
		switch b {
		case '"':
			c.Write(`\"`)
		case '\\':
			c.Write(`\\`)
		case '\n':
			c.Write(`\n`)
		case '\r':
			c.Write(`\r`)
		case '\t':
			c.Write(`\t`)
		default:
			if b < 0x20 {
				c.Write(`\u00`)
				c.Write(string("0123456789abcdef"[b>>4]))
				c.Write(string("0123456789abcdef"[b&0xf]))
			} else {
				c.Write(string(b))
			}
		}
	}
	c.Write("\"")
}

// isZero returns true if a value is its zero value (for omitempty).
func isZero(v any) bool {
	switch val := v.(type) {
	case string:
		return val == ""
	case bool:
		return !val
	case int:
		return val == 0
	case int8:
		return val == 0
	case int16:
		return val == 0
	case int32:
		return val == 0
	case int64:
		return val == 0
	case uint:
		return val == 0
	case uint8:
		return val == 0
	case uint16:
		return val == 0
	case uint32:
		return val == 0
	case uint64:
		return val == 0
	case float32:
		return val == 0
	case float64:
		return val == 0
	case []byte:
		return len(val) == 0
	case nil:
		return true
	}
	return false
}
