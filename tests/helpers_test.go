package tests

import "github.com/tinywasm/fmt"

type mockFielder struct {
	schema   []fmt.Field
	pointers []any
	err      error
}

func (m *mockFielder) Schema() []fmt.Field { return m.schema }
func (m *mockFielder) Pointers() []any {
	if m.err != nil {
		return nil
	}
	return m.pointers
}

func (m *mockFielder) IsNil() bool { return m == nil }

func (m *mockFielder) DecodeFields(r fmt.FieldReader) error {
	for i, f := range m.schema {
		ptr := m.pointers[i]
		switch f.Type {
		case fmt.FieldText:
			if p, ok := ptr.(*string); ok {
				if val, ok := r.String(f.Name); ok {
					*p = val
				}
			}
		case fmt.FieldInt:
			if p, ok := ptr.(*int); ok {
				if val, ok := r.Int(f.Name); ok {
					*p = int(val)
				}
			} else if p, ok := ptr.(*int64); ok {
				if val, ok := r.Int(f.Name); ok {
					*p = val
				}
			} else if p, ok := ptr.(*int32); ok {
				if val, ok := r.Int(f.Name); ok {
					*p = int32(val)
				}
			}
		case fmt.FieldFloat:
			if p, ok := ptr.(*float64); ok {
				if val, ok := r.Float(f.Name); ok {
					*p = val
				}
			} else if p, ok := ptr.(*float32); ok {
				if val, ok := r.Float(f.Name); ok {
					*p = float32(val)
				}
			}
		case fmt.FieldBool:
			if p, ok := ptr.(*bool); ok {
				if val, ok := r.Bool(f.Name); ok {
					*p = val
				}
			}
		case fmt.FieldBlob:
			if p, ok := ptr.(*[]byte); ok {
				if val, ok := r.Bytes(f.Name); ok {
					*p = val
				}
			}
		case fmt.FieldStruct:
			if p, ok := ptr.(fmt.Decodable); ok {
				r.Object(f.Name, p)
			}
		case fmt.FieldStructSlice:
			if p, ok := ptr.(fmt.FielderSlice); ok {
				if ar, ok := r.Array(f.Name); ok {
					n := ar.Len()
					for j := 0; j < n; j++ {
						it := p.Append()
						if dec, ok := it.(fmt.Decodable); ok {
							ar.Object(j, dec)
						}
					}
				}
			}
		case fmt.FieldRaw:
			if p, ok := ptr.(*string); ok {
				if r2, ok := r.(interface{ Raw(string) (string, bool) }); ok {
					if val, ok := r2.Raw(f.Name); ok {
						*p = val
					}
				}
			}
		}
	}
	return nil
}

func (m *mockFielder) EncodeFields(w fmt.FieldWriter) {
	for i, f := range m.schema {
		ptr := m.pointers[i]
		if f.OmitEmpty && isZero(ptr) {
			continue
		}
		switch f.Type {
		case fmt.FieldText:
			if p, ok := ptr.(*string); ok {
				if p == nil {
					w.Null(f.Name)
				} else {
					w.String(f.Name, *p)
				}
			} else {
				w.Null(f.Name)
			}
		case fmt.FieldInt:
			if p, ok := ptr.(*int); ok {
				w.Int(f.Name, int64(*p))
			} else if p, ok := ptr.(*int64); ok {
				w.Int(f.Name, *p)
			} else if p, ok := ptr.(*int32); ok {
				w.Int(f.Name, int64(*p))
			} else if p, ok := ptr.(*uint); ok {
				w.Uint(f.Name, uint64(*p))
			} else if p, ok := ptr.(*uint64); ok {
				w.Uint(f.Name, *p)
			}
		case fmt.FieldFloat:
			if p, ok := ptr.(*float64); ok {
				w.Float(f.Name, *p)
			} else if p, ok := ptr.(*float32); ok {
				w.Float(f.Name, float64(*p))
			}
		case fmt.FieldBool:
			if p, ok := ptr.(*bool); ok {
				w.Bool(f.Name, *p)
			}
		case fmt.FieldBlob:
			if p, ok := ptr.(*[]byte); ok {
				if p == nil {
					w.Null(f.Name)
				} else {
					w.Bytes(f.Name, *p)
				}
			} else {
				w.Null(f.Name)
			}
		case fmt.FieldStruct:
			if p, ok := ptr.(fmt.Encodable); ok {
				w.Object(f.Name, p)
			} else {
				w.Null(f.Name)
			}
		case fmt.FieldStructSlice:
			if ptr == nil {
				w.Null(f.Name)
			} else if p, ok := ptr.(fmt.FielderSlice); ok {
				n := p.Len()
				aw := w.Array(f.Name, n)
				for j := 0; j < n; j++ {
					if it, ok := p.At(j).(fmt.Encodable); ok {
						aw.Object(it)
					}
				}
				if closer, ok := aw.(interface{ Close() }); ok {
					closer.Close()
				}
			} else {
				w.Null(f.Name)
			}
		case fmt.FieldRaw:
			if p, ok := ptr.(*string); ok {
				if *p != "" {
					w.Object(f.Name, &rawEncodable{val: *p})
				} else {
					w.Null(f.Name)
				}
			}
		}
	}
}


func ptrString(s string) *string { return &s }
func ptrInt(i int) *int { return &i }
func ptrInt8(i int8) *int8 { return &i }
func ptrInt16(i int16) *int16 { return &i }
func ptrInt32(i int32) *int32 { return &i }
func ptrInt64(i int64) *int64 { return &i }
func ptrUint(i uint) *uint { return &i }
func ptrUint8(i uint8) *uint8 { return &i }
func ptrUint16(i uint16) *uint16 { return &i }
func ptrUint32(i uint32) *uint32 { return &i }
func ptrUint64(i uint64) *uint64 { return &i }
func ptrFloat32(f float32) *float32 { return &f }
func ptrFloat64(f float64) *float64 { return &f }
func ptrBool(b bool) *bool { return &b }
func ptrBytes(b []byte) *[]byte { return &b }

func isZero(ptr any) bool {
	if ptr == nil {
		return true
	}
	if s, ok := ptr.(fmt.FielderSlice); ok {
		return s.Len() == 0
	}
	switch p := ptr.(type) {
	case *string:
		return *p == ""
	case *int:
		return *p == 0
	case *int64:
		return *p == 0
	case *int32:
		return *p == 0
	case *uint:
		return *p == 0
	case *uint64:
		return *p == 0
	case *float64:
		return *p == 0
	case *float32:
		return *p == 0
	case *bool:
		return !*p
	case *[]byte:
		return len(*p) == 0
	}
	return false
}

type rawEncodable struct {
	val string
}

func (r *rawEncodable) EncodeFields(w fmt.FieldWriter) {}
func (r *rawEncodable) IsNil() bool { return r == nil }
func (r *rawEncodable) Raw() string { return r.val }

type simpleModel struct {
	Name   string
	Age    int64
	Active bool
}

func (m *simpleModel) EncodeFields(w fmt.FieldWriter) {
	w.String("name", m.Name)
	w.Int("age", m.Age)
	w.Bool("active", m.Active)
}

func (m *simpleModel) DecodeFields(r fmt.FieldReader) error {
	m.Name, _ = r.String("name")
	m.Age, _ = r.Int("age")
	m.Active, _ = r.Bool("active")
	return nil
}

func (m *simpleModel) IsNil() bool { return m == nil }
