# PLAN: Fielder v2 Migration + Performance (tinywasm/json)

← [README](../README.md) | Depends on: [fmt PLAN_FIELDER_V2](../../fmt/docs/PLAN_FIELDER_V2.md)

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing`.
- **Testing Runner:** Use `gotest` (install: `go install github.com/tinywasm/devflow/cmd/gotest@latest`).
- **Max 500 lines per file.** If exceeded, subdivide by domain.
- **Flat hierarchy.** No subdirectories for library code.
- **TinyGo Compatible:** No `fmt`, `strings`, `strconv`, `errors` from stdlib. Use `tinywasm/fmt`.
- **No maps** in WASM code (binary bloat).
- **Documentation First:** Update docs before coding.
- **Publishing:** Use `gopush 'message'` after tests pass and docs are updated.

## Prerequisite

Update `go.mod` to the new `tinywasm/fmt` version that removes `Values()` from Fielder:

```bash
go get github.com/tinywasm/fmt@latest
```

## Context

Current benchmark (struct with 4 fields: 2 string, 1 int64, 1 float64):

| Benchmark | tinywasm/json | encoding/json | Δ allocs |
|-----------|--------------|---------------|----------|
| Encode    | 1596 ns/op · 14 allocs | 548 ns/op · 1 alloc | +13 |
| Decode    | 2708 ns/op · 32 allocs | 2145 ns/op · 8 allocs | +24 |

**Target after this plan:** Encode ≤ 2 allocs, Decode ≤ 5 allocs.

Root causes of excess allocations:
1. `Values() []any` boxes strings into interface → heap escape (Encode)
2. `Schema()` allocates a new slice per call (ormc fix, not this plan)
3. `parseString()` uses `fmt.Builder` for EVERY string, even simple ASCII (Decode)
4. JSON keys parsed as strings then discarded — pure waste (Decode)
5. `Encode` creates a new Builder instead of using the pool (Encode)
6. `fmt.Convert(val).String()` for numbers boxes the value into `any` (Encode)

---

## Stage 1: Migrate Encode to use Pointers (remove Values dependency)

**File:** `encode.go`

### 1.1 Replace `encodeFielder` to read from Pointers

```go
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
```

### 1.2 Replace `encodeValue` with `encodeFromPtr`

```go
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
		case *uint:
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
			encodeFielder(b, nested)
		} else {
			b.WriteString("null")
		}
	default:
		b.WriteString("null")
	}
}
```

### 1.3 Add `isZeroPtr` (replaces `fmt.IsZero(val)`)

```go
// isZeroPtr checks if a field value is zero by reading through its pointer.
func isZeroPtr(ptr any, ft fmt.FieldType) bool {
	switch ft {
	case fmt.FieldText:
		if p, ok := ptr.(*string); ok {
			return *p == ""
		}
	case fmt.FieldInt:
		switch p := ptr.(type) {
		case *int64:
			return *p == 0
		case *int:
			return *p == 0
		case *int32:
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
```

### 1.4 Use pooled Conv in `Encode`

```go
func Encode(data fmt.Fielder, output any) error {
	b := fmt.GetConv()
	defer b.PutConv()

	if err := encodeFielder(b, data); err != nil {
		return err
	}

	switch out := output.(type) {
	case *[]byte:
		*out = b.Bytes()
	case *string:
		*out = b.String()
	case io.Writer:
		_, err := out.Write(b.OutBytes())
		return err
	default:
		return fmt.Err("json", "encode", "output must be *[]byte, *string, or io.Writer")
	}
	return nil
}
```

**Note:** Verify that `Conv` has a `Bytes()` method and an `OutBytes()` method (returns `out[:outLen]` without copy). If not, use `[]byte(b.String())` as fallback and file an issue for `fmt`.

### 1.5 Delete old `encodeValue` function

Remove the entire `encodeValue(b *fmt.Builder, v any)` function — replaced by `encodeFromPtr`.

---

## Stage 2: Optimize parseString (fast path for simple strings)

**File:** `parser.go`

### 2.1 Replace `parseString` with fast path

```go
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
			return b.String(), nil
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
```

---

## Stage 3: Zero-alloc key matching in parseIntoFielder

**File:** `parser.go`

### 3.1 Add `matchFieldIndex` method

```go
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
				k, _ := parseJSONTag(field)
				if len(k) == len(keyBytes) && matchBytesStr(k, keyBytes) {
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
				k, _ := parseJSONTag(field)
				if k == key {
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
```

### 3.2 Update `parseIntoFielder` to use `matchFieldIndex`

Replace the key parsing block (lines 197-218 approximately):

```go
// BEFORE:
//   if p.next() != '"' { return error }
//   key, err := p.parseString()
//   ... linear search by key string ...

// AFTER:
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
			if _, err := p.parseValue(); err != nil {
				return err
			}
		} else {
			// ... existing field type dispatch (unchanged) ...
		}

		// ... existing comma/brace check (unchanged) ...
	}
```

---

## Stage 4: Update Decode input handling

**File:** `decode.go`

### 4.1 Avoid `[]byte(in)` copy for string input

Currently `Decode` converts string input to `[]byte`, causing 1 alloc. Since the parser
only reads bytes and never modifies the data, we can use `unsafe` to avoid the copy:

```go
func Decode(input any, data fmt.Fielder) error {
	var raw []byte
	switch in := input.(type) {
	case []byte:
		raw = in
	case string:
		// Avoid copy: parser is read-only, never modifies data.
		raw = unsafe.Slice(unsafe.StringData(in), len(in))
	case io.Reader:
		// ... existing io.Reader handling ...
	}
	p := parser{data: raw}
	return p.parseIntoFielder(data)
}
```

**Note:** `unsafe.StringData` and `unsafe.Slice` are available since Go 1.20. Both Go and TinyGo support them. The parser MUST NOT modify `p.data` — verify this is the case (it already is: parser only reads via `p.data[p.pos]`).

If the team prefers to avoid `unsafe`, skip this optimization — it saves only 1 alloc.

---

## Stage 5: Update tests

### 5.1 Remove `Values()` from all test mocks

Search and update all files in `tests/`:

```bash
grep -rn "func.*Values().*\[\]any" tests/
```

Every mock implementing `Fielder` must remove its `Values()` method.

### 5.2 Update benchmark struct

**File:** `tests/bench_encode_test.go`

Remove `Values()` method from `benchUser`. Keep only `Schema()` and `Pointers()`.

### 5.3 Verify all tests pass

```bash
gotest
```

### 5.4 Run benchmarks and compare

```bash
go test -bench=. -benchmem -count=5 ./tests/... 2>&1 | tee /tmp/bench_v2.txt
```

Expected results:
- Encode: ≤ 2 allocs/op (down from 14)
- Decode: ≤ 5 allocs/op (down from 32)

---

## Stage 6: Update documentation

### 6.1 Update `benchmarks/README.md`

Replace the Performance Results table with new benchmark numbers.
Update "Last updated" date.

### 6.2 Update `README.md`

If it references `Values()` in API examples, update to show the new pattern.

---

## Stage 7: Publish

```bash
gopush 'json: Fielder v2 migration — encode from pointers, zero-alloc key matching, parseString fast path'
```

---

## Summary

| Stage | File(s) | Allocs eliminated |
|-------|---------|-------------------|
| 1 | `encode.go` | ~10 (Values boxing + Builder + Convert) |
| 2 | `parser.go` | ~12 (Builder per string in decode) |
| 3 | `parser.go` | ~8 (key string allocations) |
| 4 | `decode.go` | 1 (string→[]byte copy) |
| 5 | `tests/` | — (test migration) |
| 6 | docs | — |
| 7 | — | publish |

**Projected totals:** Encode 1-2 allocs, Decode 3-5 allocs (from 14 and 32).
