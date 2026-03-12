package json

import (
	"github.com/tinywasm/fmt"
	"io"
)

// Encode serializes a Fielder to JSON.
// output: *[]byte | *string | io.Writer.
func Encode(data fmt.Fielder, output any) error {
	b := fmt.GetConv()
	defer b.PutConv()

	if err := encodeFielder(b, data); err != nil {
		return err
	}

	if b.GetString(fmt.BuffErr) != "" {
		return fmt.Err(b.GetString(fmt.BuffErr))
	}

	switch out := output.(type) {
	case *[]byte:
		res := b.Bytes()
		cpy := make([]byte, len(res))
		copy(cpy, res)
		*out = cpy
	case *string:
		*out = b.GetString(fmt.BuffOut)
	case io.Writer:
		_, err := out.Write(b.Bytes())
		return err
	default:
		return fmt.Err("json", "encode", "output must be *[]byte, *string, or io.Writer")
	}
	return nil
}

func encodeFielder(b *fmt.Conv, f fmt.Fielder) error {
	schema := f.Schema()
	ptrs := f.Pointers()
	if ptrs == nil && schema != nil {
		return fmt.Err("json", "encode", "failed to get pointers")
	}
	b.WriteByte('{')

	first := true
	for i, field := range schema {
		key, omitempty := parseJSONTag(field)
		if key == "-" {
			continue
		}

		ptr := ptrs[i]

		if omitempty && isZeroPtr(ptr, field.Type) {
			continue
		}

		if !first {
			b.WriteByte(',')
		}
		first = false

		b.WriteByte('"')
		fmt.JSONEscape(key, b)
		b.WriteByte('"')
		b.WriteByte(':')

		encodeFromPtr(b, ptr, field.Type)
	}

	b.WriteByte('}')
	return nil
}

// encodeFromPtr writes a JSON value by reading directly from a typed pointer.
// Avoids interface boxing — the value is never wrapped in any.
func encodeFromPtr(b *fmt.Conv, ptr any, ft fmt.FieldType) {
	switch ft {
	case fmt.FieldText:
		if p, ok := ptr.(*string); ok {
			b.WriteByte('"')
			fmt.JSONEscape(*p, b)
			b.WriteByte('"')
		} else {
			b.WriteString("null")
		}
	case fmt.FieldInt:
		switch p := ptr.(type) {
		case *int64:
			b.WriteInt(*p)
		case *int:
			b.WriteInt(int64(*p))
		case *int32:
			b.WriteInt(int64(*p))
		case *int8:
			b.WriteInt(int64(*p))
		case *int16:
			b.WriteInt(int64(*p))
		case *uint:
			b.WriteInt(int64(*p))
		case *uint8:
			b.WriteInt(int64(*p))
		case *uint16:
			b.WriteInt(int64(*p))
		case *uint32:
			b.WriteInt(int64(*p))
		case *uint64:
			b.WriteInt(int64(*p))
		default:
			b.WriteByte('0')
		}
	case fmt.FieldFloat:
		switch p := ptr.(type) {
		case *float64:
			b.WriteFloat(*p)
		case *float32:
			b.WriteFloat(float64(*p))
		default:
			b.WriteByte('0')
		}
	case fmt.FieldBool:
		if p, ok := ptr.(*bool); ok && *p {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case fmt.FieldBlob:
		if p, ok := ptr.(*[]byte); ok {
			b.WriteByte('"')
			fmt.JSONEscape(string(*p), b)
			b.WriteByte('"')
		} else {
			b.WriteString("null")
		}
	case fmt.FieldStruct:
		if nested, ok := ptr.(fmt.Fielder); ok {
			if err := encodeFielder(b, nested); err != nil {
				b.WrString(fmt.BuffErr, err.Error())
			}
		} else {
			b.WriteString("null")
		}
	default:
		b.WriteString("null")
	}
}

// isZeroPtr checks if a field value is zero by reading through its pointer.
func isZeroPtr(ptr any, ft fmt.FieldType) bool {
	switch ft {
	case fmt.FieldText:
		if p, ok := ptr.(*string); ok {
			return *p == ""
		}
	case fmt.FieldInt:
		switch p := ptr.(type) {
		case *int:
			return *p == 0
		case *int8:
			return *p == 0
		case *int16:
			return *p == 0
		case *int32:
			return *p == 0
		case *int64:
			return *p == 0
		case *uint:
			return *p == 0
		case *uint8:
			return *p == 0
		case *uint16:
			return *p == 0
		case *uint32:
			return *p == 0
		case *uint64:
			return *p == 0
		}
	case fmt.FieldFloat:
		switch p := ptr.(type) {
		case *float64:
			return *p == 0
		case *float32:
			return *p == 0
		}
	case fmt.FieldBool:
		if p, ok := ptr.(*bool); ok {
			return !*p
		}
	case fmt.FieldBlob:
		if p, ok := ptr.(*[]byte); ok {
			return len(*p) == 0
		}
	}
	return false
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
