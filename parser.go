package json

import "github.com/tinywasm/fmt"

type parser struct {
	data []byte
	pos  int
}

type jsonReader struct {
	p     *parser
	start int // offset of the object '{'
	err   error
}

func (r *jsonReader) String(name string) (string, bool) {
	found, err := r.scanToKey(name)
	if err != nil {
		r.err = err
		return "", false
	}
	if !found {
		return "", false
	}
	r.p.skipWhitespace()
	if r.p.peek() != ':' {
		return "", false
	}
	r.p.next() // consume ':'
	r.p.skipWhitespace()
	if r.p.peek() != '"' {
		return "", false
	}
	r.p.next() // consume '"'
	val, err := r.p.parseString()
	if err != nil {
		r.err = err
		return "", false
	}
	return val, true
}

func (r *jsonReader) Int(name string) (int64, bool) {
	found, err := r.scanToKey(name)
	if err != nil {
		r.err = err
		return 0, false
	}
	if !found {
		return 0, false
	}
	r.p.skipWhitespace()
	if r.p.peek() != ':' {
		return 0, false
	}
	r.p.next()
	r.p.skipWhitespace()
	start := r.p.pos
	if err := r.p.skipNumber(); err != nil {
		r.err = err
		return 0, false
	}
	c := fmt.GetConv()
	defer c.PutConv()
	c.LoadBytes(r.p.data[start:r.p.pos])
	val, err := c.Int64()
	if err != nil {
		r.err = err
		return 0, false
	}
	return val, true
}

func (r *jsonReader) Uint(name string) (uint64, bool) {
	found, err := r.scanToKey(name)
	if err != nil {
		r.err = err
		return 0, false
	}
	if !found {
		return 0, false
	}
	r.p.skipWhitespace()
	if r.p.next() != ':' {
		return 0, false
	}
	r.p.skipWhitespace()
	start := r.p.pos
	if err := r.p.skipNumber(); err != nil {
		r.err = err
		return 0, false
	}
	c := fmt.GetConv()
	defer c.PutConv()
	c.LoadBytes(r.p.data[start:r.p.pos])
	val, err := c.Uint64()
	if err != nil {
		r.err = err
		return 0, false
	}
	return val, true
}

func (r *jsonReader) Float(name string) (float64, bool) {
	found, err := r.scanToKey(name)
	if err != nil {
		r.err = err
		return 0, false
	}
	if !found {
		return 0, false
	}
	r.p.skipWhitespace()
	if r.p.next() != ':' {
		return 0, false
	}
	r.p.skipWhitespace()
	start := r.p.pos
	if err := r.p.skipNumber(); err != nil {
		r.err = err
		return 0, false
	}
	c := fmt.GetConv()
	defer c.PutConv()
	c.LoadBytes(r.p.data[start:r.p.pos])
	val, err := c.Float64()
	if err != nil {
		r.err = err
		return 0, false
	}
	return val, true
}

func (r *jsonReader) Bool(name string) (bool, bool) {
	found, err := r.scanToKey(name)
	if err != nil {
		r.err = err
		return false, false
	}
	if !found {
		return false, false
	}
	r.p.skipWhitespace()
	if r.p.peek() != ':' {
		return false, false
	}
	r.p.next()
	r.p.skipWhitespace()
	val, err := r.p.parseBool()
	if err != nil {
		r.err = err
		return false, false
	}
	return val, true
}

func (r *jsonReader) Bytes(name string) ([]byte, bool) {
	s, ok := r.String(name)
	if !ok {
		return nil, false
	}
	return []byte(s), true
}

func (r *jsonReader) Raw(name string) (string, bool) {
	found, err := r.scanToKey(name)
	if err != nil {
		r.err = err
		return "", false
	}
	if !found {
		return "", false
	}
	r.p.skipWhitespace()
	if r.p.peek() != ':' {
		return "", false
	}
	r.p.next() // consume ':'
	r.p.skipWhitespace()
	start := r.p.pos
	if err := r.p.skipValue(); err != nil {
		r.err = err
		return "", false
	}
	return string(r.p.data[start:r.p.pos]), true
}

func (r *jsonReader) Object(name string, into fmt.Decodable) bool {
	found, err := r.scanToKey(name)
	if err != nil {
		r.err = err
		return false
	}
	if !found {
		return false
	}
	r.p.skipWhitespace()
	if r.p.peek() != ':' {
		return false
	}
	r.p.next() // consume ':'
	r.p.skipWhitespace()
	if r.p.peek() == 'n' {
		r.p.parseNull()
		return true
	}
	if r.p.peek() != '{' {
		return false
	}
	objStart := r.p.pos
	innerReader := getReader()
	innerReader.p = r.p
	innerReader.start = objStart
	innerReader.err = nil
	err = into.DecodeFields(innerReader)
	if err != nil {
		r.err = err
		putReader(innerReader)
		return false
	}
	if innerReader.err != nil {
		r.err = innerReader.err
		putReader(innerReader)
		return false
	}
	putReader(innerReader)
	r.p.pos = objStart
	r.p.skipObject()
	return true
}

func (r *jsonReader) Array(name string) (fmt.ArrayReader, bool) {
	found, err := r.scanToKey(name)
	if err != nil {
		r.err = err
		return nil, false
	}
	if !found {
		return nil, false
	}
	r.p.skipWhitespace()
	if r.p.peek() != ':' {
		return nil, false
	}
	r.p.next() // consume ':'
	r.p.skipWhitespace()
	if r.p.peek() != '[' {
		return nil, false
	}
	arrayStart := r.p.pos
	return &jsonArrayReader{p: r.p, start: arrayStart}, true
}

func (r *jsonReader) scanToKey(name string) (bool, error) {
	r.p.pos = r.start
	r.p.skipWhitespace()
	if r.p.peek() != '{' {
		return false, fmt.Err("json", "decode", "expected { at "+string(r.p.peek()))
	}
	r.p.next()
	for {
		r.p.skipWhitespace()
		if r.p.peek() == '}' {
			return false, nil
		}
		if r.p.next() != '"' {
			return false, fmt.Err("json", "decode", "expected quote")
		}
		keyStart := r.p.pos
		if err := r.p.skipString(); err != nil {
			return false, err
		}
		keyBytes := r.p.data[keyStart : r.p.pos-1]
		r.p.skipWhitespace()
		if r.p.peek() != ':' {
			return false, fmt.Err("json", "decode", "expected :")
		}
		if equalStringBytes(name, keyBytes) {
			return true, nil
		}
		r.p.next() // consume ':'
		if err := r.p.skipValue(); err != nil {
			return false, err
		}
		r.p.skipWhitespace()
		c := r.p.peek()
		if c == '}' {
			continue // will be caught by loop start
		}
		if c == ',' {
			r.p.next()
			continue
		}
		return false, fmt.Err("json", "decode", "expected , or }")
	}
}

type jsonArrayReader struct {
	p     *parser
	start int // offset of '['
}

func (r *jsonArrayReader) Len() int {
	r.p.pos = r.start
	r.p.skipWhitespace()
	if r.p.next() != '[' {
		return 0
	}
	r.p.skipWhitespace()
	if r.p.peek() == ']' {
		return 0
	}
	count := 0
	for {
		if err := r.p.skipValue(); err != nil {
			return count
		}
		count++
		r.p.skipWhitespace()
		c := r.p.next()
		if c == ']' {
			return count
		}
		if c != ',' {
			return count
		}
	}
}

func (r jsonArrayReader) seekToIndex(i int) bool {
	r.p.pos = r.start
	r.p.skipWhitespace()
	if r.p.peek() != '[' {
		return false
	}
	r.p.next() // consume '['
	for j := 0; j <= i; j++ {
		r.p.skipWhitespace()
		if r.p.peek() == ']' {
			return false
		}
		if j == i {
			return true
		}
		if err := r.p.skipValue(); err != nil {
			return false
		}
		r.p.skipWhitespace()
		if r.p.next() != ',' {
			return false
		}
	}
	return false
}

func (r jsonArrayReader) String(i int) string {
	if !r.seekToIndex(i) {
		return ""
	}
	r.p.skipWhitespace()
	if r.p.next() != '"' {
		return ""
	}
	val, _ := r.p.parseString()
	return val
}

func (r jsonArrayReader) Int(i int) int64 {
	if !r.seekToIndex(i) {
		return 0
	}
	r.p.skipWhitespace()
	start := r.p.pos
	if err := r.p.skipNumber(); err != nil {
		return 0
	}
	c := fmt.GetConv()
	defer c.PutConv()
	c.LoadBytes(r.p.data[start:r.p.pos])
	val, _ := c.Int64()
	return val
}

func (r jsonArrayReader) Float(i int) float64 {
	if !r.seekToIndex(i) {
		return 0
	}
	r.p.skipWhitespace()
	start := r.p.pos
	if err := r.p.skipNumber(); err != nil {
		return 0
	}
	c := fmt.GetConv()
	defer c.PutConv()
	c.LoadBytes(r.p.data[start:r.p.pos])
	val, _ := c.Float64()
	return val
}

func (r jsonArrayReader) Bool(i int) bool {
	if !r.seekToIndex(i) {
		return false
	}
	val, _ := r.p.parseBool()
	return val
}

func (r jsonArrayReader) Bytes(i int) []byte {
	return []byte(r.String(i))
}

func (r jsonArrayReader) Object(i int, into fmt.Decodable) bool {
	if !r.seekToIndex(i) {
		return false
	}
	r.p.skipWhitespace()
	if r.p.peek() != '{' {
		return false
	}
	objStart := r.p.pos
	innerReader := getReader()
	innerReader.p = r.p
	innerReader.start = objStart
	innerReader.err = nil
	err := into.DecodeFields(innerReader)
	r.p.pos = objStart
	r.p.skipObject()
	putReader(innerReader)
	return err == nil
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

func (p *parser) skipString() error {
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if c == '"' {
			p.pos++
			return nil
		}
		if c == '\\' {
			p.pos++
			if p.pos >= len(p.data) {
				return fmt.Err("json", "decode", "unexpected EOF")
			}
			esc := p.data[p.pos]
			p.pos++
			switch esc {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			case 'u':
				if p.pos+4 > len(p.data) {
					return fmt.Err("json", "decode", "invalid unicode escape")
				}
				p.pos += 4
			default:
				return fmt.Err("json", "decode", "invalid escape sequence")
			}
			continue
		}
		p.pos++
	}
	return fmt.Err("json", "decode", "unexpected EOF")
}

func (p *parser) captureValue() ([]byte, error) {
	start := p.pos
	if err := p.skipValue(); err != nil {
		return nil, err
	}
	return p.data[start:p.pos], nil
}

func (p *parser) skipValue() error {
	p.skipWhitespace()
	c := p.peek()
	switch c {
	case '"':
		p.next()
		return p.skipString()
	case '{':
		p.next()
		return p.skipObject()
	case '[':
		p.next()
		return p.skipArray()
	case 't', 'f':
		_, err := p.parseBool()
		return err
	case 'n':
		return p.parseNull()
	default:
		if (c >= '0' && c <= '9') || c == '-' {
			return p.skipNumber()
		}
		return fmt.Err("json", "decode", "unexpected character")
	}
}

func (p *parser) skipObject() error {
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
		if err := p.skipString(); err != nil {
			return err
		}
		p.skipWhitespace()
		if p.next() != ':' {
			return fmt.Err("json", "decode", "expected :")
		}
		if err := p.skipValue(); err != nil {
			return err
		}
		p.skipWhitespace()
		c := p.next()
		if c == '}' {
			return nil
		}
		if c != ',' {
			return fmt.Err("json", "decode", "expected , or }")
		}
	}
}

func (p *parser) skipArray() error {
	p.skipWhitespace()
	if p.peek() == ']' {
		p.next()
		return nil
	}
	for {
		if err := p.skipValue(); err != nil {
			return err
		}
		p.skipWhitespace()
		c := p.next()
		if c == ']' {
			return nil
		}
		if c != ',' {
			return fmt.Err("json", "decode", "expected , or ]")
		}
	}
}

func (p *parser) skipNumber() error {
	p.skipWhitespace()
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if (c >= '0' && c <= '9') || c == '-' || c == '+' || c == '.' || c == 'e' || c == 'E' {
			p.pos++
		} else {
			break
		}
	}
	return nil
}

func (p *parser) parseString() (string, error) {
	start := p.pos
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if c == '"' {
			s := string(p.data[start:p.pos])
			p.pos++
			return s, nil
		}
		if c == '\\' {
			return p.parseStringEscape(start)
		}
		p.pos++
	}
	return "", fmt.Err("json", "decode", "unexpected EOF")
}

func (p *parser) parseStringEscape(start int) (string, error) {
	b := fmt.GetConv()
	defer b.PutConv()
	for i := start; i < p.pos; i++ {
		b.WriteByte(p.data[i])
	}
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

func (p *parser) parseBool() (bool, error) {
	p.skipWhitespace()
	if p.peek() == 't' {
		if p.pos+4 <= len(p.data) && string(p.data[p.pos:p.pos+4]) == "true" {
			p.pos += 4
			return true, nil
		}
	} else if p.peek() == 'f' {
		if p.pos+5 <= len(p.data) && string(p.data[p.pos:p.pos+5]) == "false" {
			p.pos += 5
			return false, nil
		}
	}
	return false, fmt.Err("json", "decode", "expected boolean")
}

func (p *parser) parseNull() error {
	p.skipWhitespace()
	if p.pos+4 <= len(p.data) && equalStringBytes("null", p.data[p.pos:p.pos+4]) {
		p.pos += 4
		return nil
	}
	return fmt.Err("json", "decode", "expected null")
}

func equalStringBytes(s string, b []byte) bool {
	if len(s) != len(b) {
		return false
	}
	for i := range b {
		if b[i] != s[i] {
			return false
		}
	}
	return true
}
