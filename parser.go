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
		if err := p.parseNull(); err != nil {
			return nil, err
		}
		return nil, nil
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

	var b fmt.Builder
	for p.pos < len(p.data) {
		c := p.next()
		if c == '"' {
			return b.String(), nil
		}
		if c == '\\' {
			c = p.next()
			switch c {
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
				val, _ := fmt.Convert(string(p.data[p.pos : p.pos+4])).Int64(16)
				p.pos += 4
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
		key, err := p.parseString()
		if err != nil {
			return err
		}
		p.skipWhitespace()
		if p.next() != ':' {
			return fmt.Err("json", "decode", "expected :")
		}

		// Search for the field in the schema
		fieldIdx := -1
		for i, field := range schema {
			k, _ := parseJSONTag(field)
			if k == key {
				fieldIdx = i
				break
			}
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
				val, err := p.parseValue()
				if err != nil {
					return err
				}
				writeValue(ptr, field.Type, val)
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
