# PLAN: Rewrite JSON Codec — Fielder-Only, Zero Reflect

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing` for tests.
- **Testing Runner:** Install and use `gotest`:
  ```bash
  go install github.com/tinywasm/devflow/cmd/gotest@latest
  ```
- **Max 500 lines per file.** If exceeded, subdivide by domain.
- **Flat hierarchy.** No subdirectories for library code.
- **Documentation First:** Update docs before coding.
- **Publishing:** Use `gopush 'message'` after tests pass and docs are updated.
- **TinyGo Compatible:** No `fmt`, `strings`, `strconv`, `errors` from stdlib. Use `tinywasm/fmt`.
- **No maps** in exported types (binary bloat in WASM).

## Prerequisites

- **`tinywasm/fmt`** must be published with:
  - `Field.JSON string` for JSON key + modifiers.
  - `FieldStruct` constant for nested struct detection.
  - `Fielder` interface (`Schema()`, `Values()`, `Pointers()`).
  - `JSONEscape(s string, b *Builder)` for string escaping.
  - `IsZero(v any) bool` for omitempty support.
- Update `go.mod` to require the new `fmt` version before starting.

## Context

The `tinywasm/json` package currently has two codec implementations:
- **`codec_wasm.go`** (`//go:build wasm`): Browser JSON API + ~30 reflect operations.
- **`codec_stdlib.go`** (`//go:build !wasm`): Delegates to `encoding/json` (reflect internally).

### Problems
1. **Two codecs** → potential inconsistency, double maintenance.
2. **reflect** → WASM binary bloat.
3. **Redundant code** → logic that already exists in `tinywasm/fmt` is duplicated.

### New Architecture

**Single codec, no build tags, no reflect, no `encoding/json`, no `syscall/js`.**

- **Public API accepts only `fmt.Fielder`** — no primitives, no maps, no slices in the signature.
- **Encoding:** Iterates `Schema()` + `Values()` → builds JSON string via `fmt.Builder` + `fmt.JSONEscape`.
- **Decoding:** Minimal JSON parser → populates struct via `Pointers()`.
- **All number/bool formatting and parsing** reuses `fmt.Convert()` — zero duplication.
- **Unknown types → error.** All structs must implement `Fielder` via `ormc`.

### What This Plan Does NOT Cover

- Changes to `tinywasm/fmt` or `tinywasm/orm` — those have their own independent plans.

---

## Stage 1: Create Encoder

← None | Next → [Stage 2](#stage-2-create-decoder)

### 1.1 Create `encode.go`

Single file, no build tags. The only public entry point for encoding.

**Public API (breaking change):**

```go
// Encode serializes a Fielder to JSON.
// output: *[]byte | *string | io.Writer.
func Encode(data fmt.Fielder, output any) error
```

### 1.2 Internal implementation

```go
func encodeFielder(b *fmt.Builder, f fmt.Fielder) error {
    schema := f.Schema()
    values := f.Values()
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
```

**Key design decisions:**
- `encodeValue` only handles types that `Values()` can return. No maps, no slices, no generic `any` — those don't exist inside a Fielder.
- `fmt.JSONEscape` handles all string escaping — no duplication.
- `fmt.Convert(val).String()` handles all numeric formatting — no duplication.
- `fmt.IsZero(val)` handles omitempty — no duplication.

### 1.3 Helper: `parseJSONTag`

```go
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
```

**Note:** No `fmt.Convert().Split()` needed — a simple byte scan for one comma is faster and avoids allocation.

### 1.4 Output handling

```go
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
```

### 1.5 Tests

- `TestEncodeSimple`: Fielder with string/int/bool fields → correct JSON.
- `TestEncodeNested`: Fielder with `FieldStruct` → recursive JSON object.
- `TestEncodeOmitEmpty`: Zero-value fields with `omitempty` → skipped.
- `TestEncodeJSONExclude`: `JSON: "-"` → field absent.
- `TestEncodeJSONKey`: `JSON: "custom"` → uses custom key.
- `TestEncodeJSONKeyFallback`: `JSON: ""` → uses `Field.Name`.
- `TestEncodeStringEscaping`: Special chars correctly escaped.
- `TestEncodeNilField`: `nil` value → `"null"`.
- `TestEncodeBytes`: `[]byte` → JSON string.
- `TestEncodeToBytes`: Output `*[]byte`.
- `TestEncodeToString`: Output `*string`.
- `TestEncodeToWriter`: Output `io.Writer`.

```bash
gotest
```

---

## Stage 2: Create Decoder

← [Stage 1](#stage-1-create-encoder) | Next → [Stage 3](#stage-3-remove-old-codecs)

### 2.1 Create `decode.go`

**Public API (breaking change):**

```go
// Decode parses JSON into a Fielder.
// input: []byte | string | io.Reader.
func Decode(input any, data fmt.Fielder) error
```

### 2.2 Create `parser.go`

Minimal JSON parser (~200 lines). Parses JSON into Go primitives:

```go
type parser struct {
    data []byte
    pos  int
}

func (p *parser) skipWhitespace()
func (p *parser) peek() byte
func (p *parser) next() byte
func (p *parser) parseValue() (any, error)
func (p *parser) parseString() (string, error)
func (p *parser) parseNumber() (any, error)         // int64 if no decimal, float64 otherwise
func (p *parser) parseBool() (bool, error)
func (p *parser) parseNull() error
func (p *parser) parseArray() ([]any, error)
func (p *parser) parseObject() (map[string]any, error)
```

**Number parsing:** Uses `fmt.Convert(numberString).Int64()` / `.Float64()` instead of reimplementing number parsing. The parser extracts the raw number string, `fmt` does the conversion.

**String unescaping:** Handles JSON escape sequences (`\"`, `\\`, `\/`, `\b`, `\f`, `\n`, `\r`, `\t`, `\uXXXX`). This is the inverse of `fmt.JSONEscape` and is specific to the JSON parser — it stays in `json`, not `fmt`.

### 2.3 Implement `decodeFielder`

```go
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
```

### 2.4 Decode entry point

```go
func Decode(input any, data fmt.Fielder) error {
    var raw []byte
    switch in := input.(type) {
    case []byte:
        raw = in
    case string:
        raw = []byte(in)
    case io.Reader:
        var buf fmt.Builder
        tmp := make([]byte, 4096)
        for {
            n, err := in.Read(tmp)
            if n > 0 {
                buf.Write(tmp[:n])
            }
            if err != nil {
                break
            }
        }
        raw = []byte(buf.String())
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
```

### 2.5 Tests

- `TestDecodeSimple`: JSON object → Fielder fields populated.
- `TestDecodeNested`: Nested JSON object → nested Fielder.
- `TestDecodeJSONKey`: `JSON: "custom"` → reads from `"custom"` key.
- `TestDecodeJSONExclude`: `JSON: "-"` → field skipped.
- `TestDecodeMissingField`: JSON missing a key → field unchanged.
- `TestDecodeExtraField`: JSON has extra key → silently ignored.
- `TestDecodeIntFromFloat`: JSON `1.0` → `int64(1)` when `FieldInt`.
- `TestDecodeFloatFromInt`: JSON `1` → `float64(1.0)` when `FieldFloat`.
- `TestDecodeBytes`: JSON string → `[]byte`.
- `TestDecodeFromBytes`: Input `[]byte`.
- `TestDecodeFromString`: Input `string`.
- `TestDecodeFromReader`: Input `io.Reader`.
- `TestDecodeStringEscapes`: `\"`, `\\`, `\n`, `\uXXXX` correctly unescaped.
- `TestDecodeNumbers`: No decimal → `int64`, with decimal → `float64`.
- `TestDecodeNull`: `null` value → field unchanged.
- `TestDecodeInvalidJSON`: Malformed JSON → error.

```bash
gotest
```

---

## Stage 3: Remove Old Codecs

← [Stage 2](#stage-2-create-decoder) | Next → [Stage 4](#stage-4-documentation-and-publish)

### 3.1 Delete platform-specific files

- **Delete** `codec_wasm.go`
- **Delete** `codec_stdlib.go`

### 3.2 Simplify `json.go`

Remove `codec` interface, `getJSONCodec()`, `instance` var, build tags. The file becomes minimal — just package declaration and imports shared between `encode.go` and `decode.go`. Or remove it entirely if `encode.go` and `decode.go` are self-contained.

### 3.3 Reorganize tests

- **Delete** `json_stlib_test.go` and `json_wasm_test.go`.
- **Update** `json_shared_test.go`: remove platform runner pattern, test `Encode`/`Decode` directly with mock `Fielder` types.
- If >5 test files, move all to `tests/` directory.

### 3.4 Verify clean codebase

```bash
grep -r "\"reflect\"" *.go         # Zero in non-test files
grep -r "//go:build" *.go          # Zero (no platform split)
grep -r "syscall/js" *.go          # Zero
grep -r "encoding/json" *.go       # Zero
```

### 3.5 Tests

```bash
gotest
```

---

## Stage 4: Documentation and Publish

← [Stage 3](#stage-3-remove-old-codecs) | None →

### 4.1 Update `README.md`

- New architecture: single codec, zero reflect, platform-agnostic, Fielder-only.
- Updated API signatures: `Encode(fmt.Fielder, any) error`, `Decode(any, fmt.Fielder) error`.
- Breaking changes: structs must implement `Fielder` via `ormc`, no primitive encode/decode.
- Migration guide.

### 4.2 Run full test suite

```bash
gotest
```

### 4.3 Publish

```bash
gopush 'rewrite json codec: Fielder-only, zero reflect, single platform-agnostic implementation'
```

---

## Summary of Changes

| File | Action |
|------|--------|
| `encode.go` (new) | Fielder → JSON via `fmt.Builder` + `fmt.JSONEscape` + `fmt.Convert` |
| `decode.go` (new) | JSON → Fielder via `Pointers()` + `fmt.Convert` |
| `parser.go` (new) | Minimal JSON tokenizer (~200 lines) |
| `json.go` | Simplified or removed |
| `codec_wasm.go` | **Deleted** |
| `codec_stdlib.go` | **Deleted** |
| `json_stlib_test.go` | **Deleted** |
| `json_wasm_test.go` | **Deleted** |
| `json_shared_test.go` | Updated for Fielder-only API |
| `README.md` | Rewritten |

## Breaking Changes

1. **`Encode` signature**: `Encode(any, any)` → `Encode(fmt.Fielder, any) error`.
2. **`Decode` signature**: `Decode(any, any)` → `Decode(any, fmt.Fielder)`.
3. **Structs MUST implement `fmt.Fielder`** via `ormc`. Raw structs → error.
4. **No primitive encode/decode** — only Fielder types. Use `fmt.Convert` for standalone conversions.
5. **No `syscall/js`** — browser JSON API no longer used.
6. **No `encoding/json`** — manual codec.

## Code Reuse from `tinywasm/fmt`

| What | fmt function | Used in |
|------|-------------|---------|
| String escaping | `fmt.JSONEscape(s, b)` | `encodeValue`, `encodeFielder` (keys) |
| Zero-value check | `fmt.IsZero(v)` | `encodeFielder` (omitempty) |
| Number → string | `fmt.Convert(v).String()` | `encodeValue` (int, float) |
| String → int64 | `fmt.Convert(s).Int64()` | `parser.parseNumber` |
| String → float64 | `fmt.Convert(s).Float64()` | `parser.parseNumber` |
| String builder | `fmt.Builder` | Entire encoder |
| Error creation | `fmt.Err(...)` | All error paths |
