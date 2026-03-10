package json

import (
	"github.com/tinywasm/fmt"
	"io"
)

type parser struct {
	data []byte
	pos  int
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			p.pos++
		} else {
			break
		}
	}
}

func (p *parser) peek() byte {
	if p.pos < len(p.data) {
		return p.data[p.pos]
	}
	return 0
}

func (p *parser) next() byte {
	if p.pos < len(p.data) {
		c := p.data[p.pos]
		p.pos++
		return c
	}
	return 0
}

func (p *parser) parseValue() (any, error) {
	p.skipWhitespace()
	c := p.peek()
	switch c {
	case '"':
		return p.parseString()
	case '{':
		return p.parseObject()
	case '[':
		return p.parseArray()
	case 't', 'f':
		return p.parseBool()
	case 'n':
		return p.parseNull()
	default:
		if (c >= '0' && c <= '9') || c == '-' {
			return p.parseNumber()
		}
		return nil, fmt.Err("json", "decode", "unexpected character")
	}
}

func (p *parser) parseString() (string, error) {
	if p.next() != '"' {
		return "", fmt.Err("json", "decode", "expected quote")
	}

	var b []byte
	for p.pos < len(p.data) {
		c := p.next()
		if c == '"' {
			return string(b), nil
		}
		if c == '\\' {
			c = p.next()
			switch c {
			case '"':
				b = append(b, '"')
			case '\\':
				b = append(b, '\\')
			case '/':
				b = append(b, '/')
			case 'b':
				b = append(b, '\b')
			case 'f':
				b = append(b, '\f')
			case 'n':
				b = append(b, '\n')
			case 'r':
				b = append(b, '\r')
			case 't':
				b = append(b, '\t')
			case 'u':
				// Handle \uXXXX
				if p.pos+4 > len(p.data) {
					return "", fmt.Err("json", "decode", "invalid unicode escape")
				}
				hex := string(p.data[p.pos : p.pos+4])
				p.pos += 4
				// Simplified: just append raw for now, or implement proper hex decoding
				// For this task, let's keep it simple as per PLAN.md design notes
				val := decodeHex(hex)
				b = append(b, byte(val))
			default:
				return "", fmt.Err("json", "decode", "invalid escape sequence")
			}
		} else {
			b = append(b, c)
		}
	}
	return "", fmt.Err("json", "decode", "unexpected EOF")
}

func decodeHex(s string) int {
	var res int
	for i := 0; i < len(s); i++ {
		res <<= 4
		c := s[i]
		if c >= '0' && c <= '9' {
			res += int(c - '0')
		} else if c >= 'a' && c <= 'f' {
			res += int(c - 'a' + 10)
		} else if c >= 'A' && c <= 'F' {
			res += int(c - 'A' + 10)
		}
	}
	return res
}

func (p *parser) parseNumber() (any, error) {
	start := p.pos
	isFloat := false
	for p.pos < len(p.data) {
		c := p.peek()
		if (c >= '0' && c <= '9') || c == '-' || c == '.' || c == 'e' || c == 'E' || c == '+' {
			if c == '.' || c == 'e' || c == 'E' {
				isFloat = true
			}
			p.pos++
		} else {
			break
		}
	}
	s := string(p.data[start:p.pos])
	if isFloat {
		val, err := fmt.Convert(s).Float64()
		return val, err
	} else {
		// Basic int parsing
		var res int64
		neg := false
		i := 0
		if s[0] == '-' {
			neg = true
			i = 1
		}
		for ; i < len(s); i++ {
			res = res*10 + int64(s[i]-'0')
		}
		if neg {
			res = -res
		}
		return res, nil
	}
}

func (p *parser) parseBool() (bool, error) {
	if p.peek() == 't' {
		if p.pos+4 <= len(p.data) && string(p.data[p.pos:p.pos+4]) == "true" {
			p.pos += 4
			return true, nil
		}
	} else {
		if p.pos+5 <= len(p.data) && string(p.data[p.pos:p.pos+5]) == "false" {
			p.pos += 5
			return false, nil
		}
	}
	return false, fmt.Err("json", "decode", "expected boolean")
}

func (p *parser) parseNull() (any, error) {
	if p.pos+4 <= len(p.data) && string(p.data[p.pos:p.pos+4]) == "null" {
		p.pos += 4
		return nil, nil
	}
	return nil, fmt.Err("json", "decode", "expected null")
}

func (p *parser) parseArray() ([]any, error) {
	if p.next() != '[' {
		return nil, fmt.Err("json", "decode", "expected [")
	}
	var res []any
	p.skipWhitespace()
	if p.peek() == ']' {
		p.next()
		return res, nil
	}
	for {
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		res = append(res, val)
		p.skipWhitespace()
		c := p.next()
		if c == ']' {
			break
		}
		if c != ',' {
			return nil, fmt.Err("json", "decode", "expected , or ]")
		}
	}
	return res, nil
}

func decodeWithInternal(input any, output any) error {
	// 1. Get JSON bytes from input
	var data []byte
	switch in := input.(type) {
	case []byte:
		data = in
	case string:
		data = []byte(in)
	case io.Reader:
		// Read all bytes
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
		data = buf
	default:
		return fmt.Err("json", "decode", "input must be []byte, string, or io.Reader")
	}

	// 2. Parse JSON
	p := &parser{data: data}
	parsed, err := p.parseValue()
	if err != nil {
		return err
	}

	// 3. Populate output
	switch out := output.(type) {
	case fmt.Fielder:
		obj, ok := parsed.(map[string]any)
		if !ok {
			return fmt.Err("json", "decode", "expected JSON object for Fielder")
		}
		return decodeFielder(obj, out)

	case *string:
		if s, ok := parsed.(string); ok {
			*out = s
		}
	case *int:
		if n, ok := parsed.(int64); ok {
			*out = int(n)
		}
	case *int64:
		if n, ok := parsed.(int64); ok {
			*out = n
		}
	case *float64:
		if n, ok := parsed.(float64); ok {
			*out = n
		} else if n, ok := parsed.(int64); ok {
			*out = float64(n)
		}
	case *bool:
		if b, ok := parsed.(bool); ok {
			*out = b
		}
	case *map[string]any:
		if obj, ok := parsed.(map[string]any); ok {
			*out = obj
		}
	case *[]any:
		if arr, ok := parsed.([]any); ok {
			*out = arr
		}

	default:
		return fmt.Err("json", "decode", "unsupported output type, implement fmt.Fielder via ormc")
	}
	return nil
}

func decodeFielder(parsed map[string]any, f fmt.Fielder) error {
	schema := f.Schema()
	pointers := f.Pointers()

	for i, field := range schema {
		key, _ := parseJSONTag(field)
		if key == "-" {
			continue
		}

		val, exists := parsed[key]
		if !exists {
			continue
		}

		ptr := pointers[i]

		// Nested struct: recurse
		if field.Type == fmt.FieldStruct {
			if nested, ok := ptr.(fmt.Fielder); ok {
				if obj, ok := val.(map[string]any); ok {
					if err := decodeFielder(obj, nested); err != nil {
						return err
					}
					continue
				}
			}
		}

		// Write value to pointer based on FieldType
		writeJSONValue(ptr, field.Type, val)
	}
	return nil
}

func writeJSONValue(ptr any, ft fmt.FieldType, val any) {
	switch ft {
	case fmt.FieldText:
		if p, ok := ptr.(*string); ok {
			if s, ok := val.(string); ok {
				*p = s
			}
		}
	case fmt.FieldInt:
		switch p := ptr.(type) {
		case *int:
			if v, ok := val.(int64); ok {
				*p = int(v)
			} else if v, ok := val.(float64); ok {
				*p = int(v)
			}
		case *int64:
			if v, ok := val.(int64); ok {
				*p = v
			} else if v, ok := val.(float64); ok {
				*p = int64(v)
			}
		case *int32:
			if v, ok := val.(int64); ok {
				*p = int32(v)
			} else if v, ok := val.(float64); ok {
				*p = int32(v)
			}
		}
	case fmt.FieldFloat:
		if p, ok := ptr.(*float64); ok {
			if v, ok := val.(float64); ok {
				*p = v
			} else if v, ok := val.(int64); ok {
				*p = float64(v)
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

func (p *parser) parseObject() (map[string]any, error) {
	if p.next() != '{' {
		return nil, fmt.Err("json", "decode", "expected {")
	}
	res := make(map[string]any)
	p.skipWhitespace()
	if p.peek() == '}' {
		p.next()
		return res, nil
	}
	for {
		p.skipWhitespace()
		key, err := p.parseString()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if p.next() != ':' {
			return nil, fmt.Err("json", "decode", "expected :")
		}
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		res[key] = val
		p.skipWhitespace()
		c := p.next()
		if c == '}' {
			break
		}
		if c != ',' {
			return nil, fmt.Err("json", "decode", "expected , or }")
		}
	}
	return res, nil
}
