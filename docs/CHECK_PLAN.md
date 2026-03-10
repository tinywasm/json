# PLAN: Rewrite JSON Codec — Platform-Agnostic, Zero Reflect

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing` and `reflect` for test helpers only.
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
- **`tinywasm/orm`** must be published with `ormc` generating `Field.JSON` from `json:` tags.
- Update `go.mod` to require the new `fmt` version before starting.

## Context

The `tinywasm/json` package currently has two codec implementations:
- **`codec_wasm.go`** (`//go:build wasm`): Uses browser's `JSON.parse`/`JSON.stringify` + ~30 reflect operations to bridge Go ↔ JS values.
- **`codec_stdlib.go`** (`//go:build !wasm`): Delegates to `encoding/json` (which uses reflect internally).

This creates two problems:
1. **WASM binary bloat** — reflect is expensive in TinyGo builds.
2. **Potential inconsistency** — two different codecs could produce different JSON for edge cases.

### New Architecture

A **single, platform-agnostic codec** with zero reflect:
- **Structs** that implement `fmt.Fielder` → manual JSON encoding/decoding using `Schema()`, `Values()`, `Pointers()`.
- **Primitives** (`string`, `int`, `float64`, `bool`, `[]byte`) → type switches (no reflect needed).
- **Known collections** (`[]string`, `[]int`, `map[string]any`, etc.) → type switches.
- **Unknown types** (struct without `Fielder`, unknown slices) → **error**. All structs in the tinywasm ecosystem must go through `ormc`.

### What This Plan Does NOT Cover

- Changes to `tinywasm/fmt` or `tinywasm/orm` — those have their own independent plans.
- Support for arbitrary `any` types without `Fielder` — explicitly out of scope.

---

## Stage 1: Create Unified Encoder

← None | Next → [Stage 2](#stage-2-create-unified-decoder)

### 1.1 Create `encode.go`

New file replacing both `convertGoToJS` (wasm) and `json.Marshal` (stdlib). Single implementation, no build tags.

**Public API (unchanged):**
```go
func Encode(input any, output any) error
```

**Internal encoding strategy:**

```go
func encode(b *fmt.Builder, data any) error {
    switch v := data.(type) {
    // 1. Fielder — zero reflect, uses Schema + Values
    case fmt.Fielder:
        return encodeFielder(b, v)

    // 2. Primitives
    case string:
        encodeString(b, v)
    case bool:
        encodeBool(b, v)
    case int:
        b.WriteString(fmt.Convert(v).String())
    case int64:
        b.WriteString(fmt.Convert(v).String())
    case float64:
        b.WriteString(fmt.Convert(v).String())
    // ... int8, int16, int32, uint, uint8, uint16, uint32, uint64, float32

    // 3. []byte → JSON string (base64 or raw, matching current behavior)
    case []byte:
        encodeString(b, string(v))

    // 4. Known slice types
    case []string:
        encodeSlice(b, v, func(b *fmt.Builder, s string) { encodeString(b, s) })
    case []int:
        encodeSlice(b, v, func(b *fmt.Builder, n int) { b.WriteString(fmt.Convert(n).String()) })
    case []any:
        return encodeAnySlice(b, v)

    // 5. Known map types
    case map[string]any:
        return encodeMap(b, v)
    case map[string]string:
        return encodeStringMap(b, v)

    // 6. nil
    case nil:
        b.WriteString("null")

    // 7. Unknown → error
    default:
        return fmt.Err("json", "encode", "unsupported type, implement fmt.Fielder via ormc")
    }
    return nil
}
```

### 1.2 Implement `encodeFielder()`

The core function that replaces all struct reflection:

```go
func encodeFielder(b *fmt.Builder, f fmt.Fielder) error {
    schema := f.Schema()
    values := f.Values()
    b.WriteByte('{')

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
            b.WriteByte(',')
        }
        first = false

        // Write key
        encodeString(b, key)
        b.WriteByte(':')

        // Write value — recurse if nested Fielder
        if field.Type == fmt.FieldStruct {
            if nested, ok := val.(fmt.Fielder); ok {
                if err := encodeFielder(b, nested); err != nil {
                    return err
                }
                continue
            }
        }

        if err := encode(b, val); err != nil {
            return err
        }
    }

    b.WriteByte('}')
    return nil
}
```

### 1.3 Implement helper functions

```go
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
func encodeString(b *fmt.Builder, s string) {
    b.WriteByte('"')
    // Escape: \", \\, \n, \r, \t, control chars
    for i := 0; i < len(s); i++ {
        c := s[i]
        switch c {
        case '"':
            b.WriteString(`\"`)
        case '\\':
            b.WriteString(`\\`)
        case '\n':
            b.WriteString(`\n`)
        case '\r':
            b.WriteString(`\r`)
        case '\t':
            b.WriteString(`\t`)
        default:
            if c < 0x20 {
                b.WriteString(`\u00`)
                b.WriteByte("0123456789abcdef"[c>>4])
                b.WriteByte("0123456789abcdef"[c&0xf])
            } else {
                b.WriteByte(c)
            }
        }
    }
    b.WriteByte('"')
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
    case int64:
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
```

### 1.4 Output handling

The `Encode` function writes to the requested output type:

```go
func Encode(input any, output any) error {
    var b fmt.Builder
    if err := encode(&b, input); err != nil {
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

- `TestEncodeFielderSimple`: Encode a mock `Fielder` with string/int/bool fields.
- `TestEncodeFielderNested`: Encode a `Fielder` with `FieldStruct` field containing nested `Fielder`.
- `TestEncodeFielderOmitEmpty`: Verify zero-value fields skipped when `JSON: "name,omitempty"`.
- `TestEncodeFielderJSONExclude`: Field with `JSON: "-"` is not in output.
- `TestEncodeFielderJSONKey`: Field with `JSON: "custom_key"` uses that key.
- `TestEncodePrimitives`: All primitive types produce correct JSON.
- `TestEncodeSlices`: `[]string`, `[]int`, `[]any` produce correct JSON arrays.
- `TestEncodeMaps`: `map[string]any` produces correct JSON object.
- `TestEncodeNil`: `nil` produces `"null"`.
- `TestEncodeUnsupportedType`: Struct without `Fielder` returns error.
- `TestEncodeStringEscaping`: Special chars (`"`, `\`, `\n`, control chars) correctly escaped.
- `TestEncodeToBytes`: Output to `*[]byte`.
- `TestEncodeToString`: Output to `*string`.
- `TestEncodeToWriter`: Output to `io.Writer`.

```bash
gotest
```

---

## Stage 2: Create Unified Decoder

← [Stage 1](#stage-1-create-unified-encoder) | Next → [Stage 3](#stage-3-remove-old-codecs)

### 2.1 Create `decode.go`

New file replacing both `convertJSToGo` (wasm) and `json.Unmarshal` (stdlib).

**Decoding strategy — two-phase:**
1. **Parse** JSON into a generic token stream or intermediate representation.
2. **Populate** target using `Pointers()` from `Fielder`, or direct assignment for primitives.

### 2.2 Implement JSON parser

A minimal, zero-allocation JSON tokenizer that works character-by-character:

```go
type parser struct {
    data []byte
    pos  int
}

func (p *parser) skipWhitespace()
func (p *parser) peek() byte
func (p *parser) next() byte
func (p *parser) parseValue() (any, error)        // Dispatches by first char
func (p *parser) parseString() (string, error)     // "..." with escape handling
func (p *parser) parseNumber() (any, error)        // int64 or float64
func (p *parser) parseBool() (bool, error)         // true/false
func (p *parser) parseNull() error                 // null
func (p *parser) parseArray() ([]any, error)       // [...]
func (p *parser) parseObject() (map[string]any, error) // {...}
```

**Design notes:**
- Numbers are parsed as `int64` if no decimal point, `float64` otherwise (matching current behavior).
- Strings handle all JSON escape sequences (`\"`, `\\`, `\/`, `\b`, `\f`, `\n`, `\r`, `\t`, `\uXXXX`).
- The parser is ~200 lines — well within the 500-line limit for a single file.

### 2.3 Implement `decodeFielder()`

```go
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

// writeJSONValue writes a parsed JSON value into a Go pointer.
func writeJSONValue(ptr any, ft fmt.FieldType, val any) {
    switch ft {
    case fmt.FieldText:
        if p, ok := ptr.(*string); ok {
            if s, ok := val.(string); ok {
                *p = s
            }
        }
    case fmt.FieldInt:
        if p, ok := ptr.(*int64); ok {
            switch v := val.(type) {
            case int64:
                *p = v
            case float64:
                *p = int64(v)
            }
        }
        // Also handle *int, *int32, etc. via type switches
    case fmt.FieldFloat:
        if p, ok := ptr.(*float64); ok {
            switch v := val.(type) {
            case float64:
                *p = v
            case int64:
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
```

### 2.4 Decode entry point

```go
func Decode(input any, output any) error {
    // 1. Get JSON bytes from input
    var data []byte
    switch in := input.(type) {
    case []byte:
        data = in
    case string:
        data = []byte(in)
    case io.Reader:
        // Read all bytes
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
        data = []byte(buf.String())
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
    case *int64:
        if n, ok := parsed.(int64); ok {
            *out = n
        }
    case *float64:
        if n, ok := parsed.(float64); ok {
            *out = n
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
```

### 2.5 Tests

- `TestDecodeFielderSimple`: Decode JSON object into mock `Fielder`.
- `TestDecodeFielderNested`: Decode JSON with nested object into nested `Fielder`.
- `TestDecodeFielderJSONKey`: Field with `JSON: "custom"` reads from `"custom"` key.
- `TestDecodeFielderJSONExclude`: Field with `JSON: "-"` is skipped.
- `TestDecodeFielderMissingField`: JSON missing a field → field unchanged.
- `TestDecodeFielderExtraField`: JSON has extra field → silently ignored.
- `TestDecodePrimitives`: Decode string, int, float, bool, null.
- `TestDecodeArrays`: Decode JSON arrays into `*[]any`.
- `TestDecodeObjects`: Decode JSON objects into `*map[string]any`.
- `TestDecodeFromBytes`: Input as `[]byte`.
- `TestDecodeFromString`: Input as `string`.
- `TestDecodeFromReader`: Input as `io.Reader`.
- `TestDecodeUnsupportedOutput`: Struct without `Fielder` returns error.
- `TestDecodeStringEscapes`: JSON escape sequences (`\"`, `\\`, `\n`, `\uXXXX`) correctly parsed.
- `TestDecodeNumbers`: Integer vs float detection (no decimal → int64, with decimal → float64).

```bash
gotest
```

---

## Stage 3: Remove Old Codecs

← [Stage 2](#stage-2-create-unified-decoder) | Next → [Stage 4](#stage-4-documentation-and-publish)

### 3.1 Delete platform-specific files

- **Delete** `codec_wasm.go` (all reflect + js.Value logic).
- **Delete** `codec_stdlib.go` (encoding/json delegation).

### 3.2 Simplify `json.go`

The codec interface pattern is no longer needed. `json.go` becomes:

```go
package json

import "io"

// Encode converts a Go value to JSON.
// input: any supported type (fmt.Fielder, primitives, known collections).
// output: *[]byte | *string | io.Writer.
func Encode(input any, output any) error { ... }

// Decode parses JSON into a Go value.
// input: []byte | string | io.Reader.
// output: fmt.Fielder | *string | *int64 | *float64 | *bool | *map[string]any | *[]any.
func Decode(input any, output any) error { ... }
```

No `codec` interface, no `getJSONCodec()`, no build tags, no `instance` var.

### 3.3 Reorganize test files

- **Delete** `json_stlib_test.go` and `json_wasm_test.go` (test runners for platform split).
- **Move** all remaining tests (e.g., `json_shared_test.go`) to a new `tests/` directory to keep the library more organized (e.g. rename to `tests/json_test.go` or split into `tests/encode_test.go` and `tests/decode_test.go`).
- Update test files to call `Encode`/`Decode` directly instead of going through shared runner pattern.

### 3.4 Verify no reflect and no build tags

```bash
grep -r "\"reflect\"" *.go        # Must be zero in non-test files
grep -r "//go:build" *.go          # Must be zero (no platform split)
grep -r "syscall/js" *.go          # Must be zero
grep -r "encoding/json" *.go       # Must be zero
```

### 3.5 Tests

All existing tests from `json_shared_test.go` must pass. Run:

```bash
gotest
```

---

## Stage 4: Documentation and Publish

← [Stage 3](#stage-3-remove-old-codecs) | None →

### 4.1 Update `README.md`

- New architecture description: single codec, zero reflect, platform-agnostic.
- Updated API: supported input/output types.
- `fmt.Fielder` integration: how structs are encoded/decoded via `Schema()`/`Values()`/`Pointers()`.
- Breaking change notice: unsupported types now return error instead of using reflect fallback.
- Migration guide: ensure all structs implement `Fielder` via `ormc`.

### 4.2 Run full test suite

```bash
gotest
```

### 4.3 Publish

```bash
gopush 'rewrite json codec: platform-agnostic, zero reflect, Fielder-based'
```

---

## Summary of Changes

| File | Action |
|------|--------|
| `encode.go` (new) | Unified encoder: `Fielder` → JSON, primitives, collections |
| `decode.go` (new) | Unified decoder: JSON parser + `Fielder` population via `Pointers()` |
| `json.go` | Simplified: direct `Encode`/`Decode` functions, no codec interface |
| `codec_wasm.go` | **Deleted** |
| `codec_stdlib.go` | **Deleted** |
| `json_stlib_test.go` | **Deleted** |
| `json_wasm_test.go` | **Deleted** |
| `json_shared_test.go` | Moved to `tests/` directory and updated |
| `README.md` | Updated architecture and API docs |

## Breaking Changes

1. **Struct types MUST implement `fmt.Fielder`** (via `ormc`). Passing a raw struct without `Fielder` to `Encode`/`Decode` returns an error.
2. **No more `syscall/js` dependency** in WASM builds — browser JSON API is no longer used.
3. **`encoding/json` no longer used** in stdlib builds — manual JSON building.
4. **Removed `codec` interface** — internal implementation detail, not part of public API.

## Downstream Impact

- **All packages using `json.Encode`/`json.Decode` with structs** must ensure those structs implement `fmt.Fielder`. Run `ormc` to generate the implementation.
- **WASM binary size** should decrease significantly (no reflect, no js bridge for JSON).
