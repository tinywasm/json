package json

import (
	"github.com/tinywasm/fmt"
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
	c := p.next()
	switch c {
	case '"':
		return p.parseString()
	case '{':
		return p.parseObject()
	case '[':
		return p.parseArray()
	case 't', 'f':
		p.pos-- // back to 't' or 'f'
		return p.parseBool()
	case 'n':
		p.pos--
		if err := p.parseNull(); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		if (c >= '0' && c <= '9') || c == '-' {
			p.pos--
			return p.parseNumber()
		}
		return nil, fmt.Err("json", "decode", "unexpected character")
	}
}

// parseString parses a JSON string. The opening '"' must already be consumed.
// Fast path: if no escape sequences, returns string(data[start:end]) — 1 alloc.
// Slow path: uses Conv builder for strings with escapes — 2 allocs.
func (p *parser) parseString() (string, error) {
	start := p.pos
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if c == '"' {
			s := string(p.data[start:p.pos])
			p.pos++ // consume closing '"'
			return s, nil
		}
		if c == '\\' {
			// Escape found — fall back to slow path with builder
			return p.parseStringEscape(start)
		}
		p.pos++
	}
	return "", fmt.Err("json", "decode", "unexpected EOF")
}

// parseStringEscape handles strings with escape sequences.
// Called when parseString encounters '\' at some position.
// start is the position of the first character after the opening '"'.
func (p *parser) parseStringEscape(start int) (string, error) {
	b := fmt.GetConv()
	defer b.PutConv()
	// Write the part before the escape that was already scanned
	for i := start; i < p.pos; i++ {
		b.WriteByte(p.data[i])
	}
	// Continue parsing with escape handling
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		p.pos++
		if c == '"' {
			return b.GetString(fmt.BuffOut), nil
		}
		if c == '\\' {
			if p.pos >= len(p.data) {
				return "", fmt.Err("json", "decode", "unexpected EOF")
			}
			esc := p.data[p.pos]
			p.pos++
			switch esc {
			case '"':
				b.WriteByte('"')
			case '\\':
				b.WriteByte('\\')
			case '/':
				b.WriteByte('/')
			case 'b':
				b.WriteByte('\b')
			case 'f':
				b.WriteByte('\f')
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case 'u':
				if p.pos+4 > len(p.data) {
					return "", fmt.Err("json", "decode", "invalid unicode escape")
				}
				s := string(p.data[p.pos : p.pos+4])
				p.pos += 4
				val, _ := fmt.Convert(s).Int64(16)
				b.WriteByte(byte(val))
			default:
				return "", fmt.Err("json", "decode", "invalid escape sequence")
			}
		} else {
			b.WriteByte(c)
		}
	}
	return "", fmt.Err("json", "decode", "unexpected EOF")
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
		return fmt.Convert(s).Float64()
	}
	return fmt.Convert(s).Int64()
}

// parseNumberInto parses a JSON number directly into a typed pointer.
// Uses fmt.GetConv + LoadBytes + Int64/Float64 + PutConv — 0 allocations
// (reuses fmt's existing parseIntBase/parseFloatBase via pool).
func (p *parser) parseNumberInto(ptr any, ft fmt.FieldType) error {
	p.skipWhitespace()
	start := p.pos
	isFloat := false
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if (c >= '0' && c <= '9') || c == '-' || c == '+' {
			p.pos++
		} else if c == '.' || c == 'e' || c == 'E' {
			isFloat = true
			p.pos++
		} else {
			break
		}
	}
	if p.pos == start {
		return fmt.Err("json", "decode", "expected number")
	}

	numBytes := p.data[start:p.pos]

	c := fmt.GetConv()
	c.LoadBytes(numBytes)

	if ft == fmt.FieldFloat || isFloat {
		v, err := c.Float64()
		c.PutConv()
		if err != nil {
			return err
		}
		if ft == fmt.FieldFloat {
			switch fp := ptr.(type) {
			case *float64:
				*fp = v
			case *float32:
				*fp = float32(v)
			}
		} else {
			// FieldInt but JSON has decimal/exponent — truncate
			switch ip := ptr.(type) {
			case *int64:
				*ip = int64(v)
			case *int:
				*ip = int(v)
			case *int32:
				*ip = int32(v)
			}
		}
	} else {
		v, err := c.Int64()
		c.PutConv()
		if err != nil {
			return err
		}
		switch ip := ptr.(type) {
		case *int64:
			*ip = v
		case *int:
			*ip = int(v)
		case *int32:
			*ip = int32(v)
		}
	}
	return nil
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

func (p *parser) parseNull() error {
	if p.pos+4 <= len(p.data) && string(p.data[p.pos:p.pos+4]) == "null" {
		p.pos += 4
		return nil
	}
	return fmt.Err("json", "decode", "expected null")
}

// parseIntoPtr parses a JSON value directly into a typed pointer.
// Bypasses parseValue()/writeValue() to avoid boxing values into any.
func (p *parser) parseIntoPtr(ptr any, ft fmt.FieldType) error {
	p.skipWhitespace()

	// Handle JSON null for any type
	if p.peek() == 'n' {
		return p.parseNull()
	}

	switch ft {
	case fmt.FieldText:
		if p.next() != '"' {
			return fmt.Err("json", "decode", "expected string")
		}
		s, err := p.parseString()
		if err != nil {
			return err
		}
		if sp, ok := ptr.(*string); ok {
			*sp = s
		}
		return nil

	case fmt.FieldInt, fmt.FieldFloat:
		return p.parseNumberInto(ptr, ft)

	case fmt.FieldBool:
		b, err := p.parseBool()
		if err != nil {
			return err
		}
		if bp, ok := ptr.(*bool); ok {
			*bp = b
		}
		return nil

	case fmt.FieldBlob:
		if p.next() != '"' {
			return fmt.Err("json", "decode", "expected string for blob")
		}
		s, err := p.parseString()
		if err != nil {
			return err
		}
		if bp, ok := ptr.(*[]byte); ok {
			*bp = []byte(s)
		}
		return nil
	}

	// Fallback for unknown field types
	_, err := p.parseValue()
	return err
}

func (p *parser) parseArray() ([]any, error) {
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

// matchFieldIndex matches the current JSON key (bytes between pos and closing '"')
// against schema field names WITHOUT allocating a string.
// Returns field index or -1 if not found. Advances pos past the closing '"'.
// Falls back to parseString if escape sequences are found.
func (p *parser) matchFieldIndex(schema []fmt.Field) (int, error) {
	start := p.pos
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if c == '"' {
			keyBytes := p.data[start:p.pos]
			p.pos++ // consume closing '"'
			for i, field := range schema {
				if len(field.Name) == len(keyBytes) && matchBytesStr(field.Name, keyBytes) {
					return i, nil
				}
			}
			return -1, nil // unknown field
		}
		if c == '\\' {
			// Rare: escaped key — fall back to allocating path
			key, err := p.parseStringEscape(start)
			if err != nil {
				return -1, err
			}
			for i, field := range schema {
				if field.Name == key {
					return i, nil
				}
			}
			return -1, nil
		}
		p.pos++
	}
	return -1, fmt.Err("json", "decode", "unexpected EOF in key")
}

// matchBytesStr compares a string against a byte slice without allocation.
func matchBytesStr(s string, b []byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != b[i] {
			return false
		}
	}
	return true
}

// parseIntoFielder parses a JSON object directly into the Fielder.
// Eliminates the need for map[string]any as an intermediary.
func (p *parser) parseIntoFielder(f fmt.Fielder) error {
	p.skipWhitespace()
	if p.next() != '{' {
		return fmt.Err("json", "decode", "expected {")
	}

	schema := f.Schema()
	pointers := f.Pointers()

	p.skipWhitespace()
	if p.peek() == '}' {
		p.next()
		return nil
	}

	for {
		p.skipWhitespace()
		if p.next() != '"' {
			return fmt.Err("json", "decode", "expected quote")
		}

		fieldIdx, err := p.matchFieldIndex(schema)
		if err != nil {
			return err
		}

		p.skipWhitespace()
		if p.next() != ':' {
			return fmt.Err("json", "decode", "expected :")
		}

		if fieldIdx < 0 {
			// Unknown field: parse and discard
			if _, err := p.parseValue(); err != nil {
				return err
			}
		} else {
			field := schema[fieldIdx]
			ptr := pointers[fieldIdx]

			if field.Type == fmt.FieldStruct {
				if nested, ok := ptr.(fmt.Fielder); ok {
					if err := p.parseIntoFielder(nested); err != nil {
						return err
					}
				} else {
					if _, err := p.parseValue(); err != nil {
						return err
					}
				}
			} else {
				if err := p.parseIntoPtr(ptr, field.Type); err != nil {
					return err
				}
			}
		}

		p.skipWhitespace()
		c := p.next()
		if c == '}' {
			break
		}
		if c != ',' {
			return fmt.Err("json", "decode", "expected , or }")
		}
	}
	return nil
}

func (p *parser) parseObject() (map[string]any, error) {
	res := make(map[string]any)
	p.skipWhitespace()
	if p.peek() == '}' {
		p.next()
		return res, nil
	}
	for {
		p.skipWhitespace()
		if p.next() != '"' {
			return nil, fmt.Err("json", "decode", "expected quote")
		}
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
