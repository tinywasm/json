# PLAN: Reorganización de Tests + 100% Coverage + Benchmarks vs stdlib

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing` for tests.
- **Testing Runner:** Install and use `gotest`:
  ```bash
  go install github.com/tinywasm/devflow/cmd/gotest@latest
  ```
- **Max 500 lines per file.** If exceeded, subdivide by domain.
- **Flat hierarchy.** No subdirectories for library code. Exception: `tests/` for test files when >5.
- **TinyGo Compatible:** No `fmt`, `strings`, `strconv`, `errors` from stdlib. Use `tinywasm/fmt`.
- **No maps** in WASM code (binary bloat). Applies to internal code too.
- **Documentation First:** Update docs before coding.
- **Publishing:** Use `gopush 'message'` after tests pass and docs are updated.

## Context

Stages 1–5 completados. El codec es single-pass, zero-reflect, sin `map` intermedio.
Los archivos activos son: `encode.go`, `decode.go`, `parser.go`.

Tests actuales en raíz: `encode_test.go`, `decode_test.go` → deben moverse a `tests/`.
Ya existe `tests/json_test.go` (integración) — se expande con el resto.

Cobertura actual: **parcial**. Faltan ramas en `encodeValue`, `writeValue`,
`parseString`, `parser`, y paths de error en `Encode`/`Decode`.

Los benchmarks existentes en `benchmarks/clients/` son demos WASM, **no** son
`BenchmarkXxx` ejecutables con `go test -bench`. No hay medición real vs stdlib.

---

## Stage 6: Reorganizar tests en `tests/`

Mover **todos** los test files a `tests/`, con nombres que indican exactamente qué prueban.
Cada archivo debe ser pequeño (< 150 líneas), enfocado en un dominio.

### 6.1 Estructura final de `tests/`

```
tests/
├── helpers_test.go         — mockFielder compartido entre todos los archivos
├── encode_basic_test.go    — string, int, bool, nil, bytes → JSON
├── encode_tags_test.go     — JSON tag: key, omitempty, "-", fallback
├── encode_output_test.go   — output *[]byte, *string, io.Writer, error
├── encode_types_test.go    — todos los tipos numéricos de Values()
├── decode_basic_test.go    — JSON → string, int, bool, float, bytes
├── decode_types_test.go    — *int, *int32, *float32, coerciones float↔int
├── decode_input_test.go    — input []byte, string, io.Reader, error
├── decode_tags_test.go     — JSON tag: key, omitempty, "-", missing, extra
├── decode_nested_test.go   — FieldStruct recursivo, ptr no Fielder, discard
├── parser_string_test.go   — todos los escapes: \b \f \/ \uXXXX, EOF, error
├── parser_number_test.go   — int64, float64, negativo, notación científica
├── parser_bool_null_test.go — parseBool true/false/error, parseNull/error
├── parser_array_test.go    — empty, values, bad separator, error
├── parser_object_test.go   — empty, keys, bad colon/separator, error
├── parser_limits_test.go   — peek/next vacío, skipWhitespace, char desconocido
├── bench_encode_test.go    — BenchmarkEncode tinywasm vs stdlib
├── bench_decode_test.go    — BenchmarkDecode tinywasm vs stdlib
├── bench_roundtrip_test.go — BenchmarkRoundTrip tinywasm vs stdlib
└── json_test.go            — integración end-to-end (ya existe, mantener)
```

### 6.2 Eliminar de la raíz

```
encode_test.go   → DELETE (contenido migrado a tests/encode_*_test.go)
decode_test.go   → DELETE (contenido migrado a tests/decode_*_test.go)
```

### 6.3 `tests/helpers_test.go`

```go
package tests

import "github.com/tinywasm/fmt"

type mockFielder struct {
    schema   []fmt.Field
    values   []any
    pointers []any
}

func (m *mockFielder) Schema() []fmt.Field { return m.schema }
func (m *mockFielder) Values() []any       { return m.values }
func (m *mockFielder) Pointers() []any     { return m.pointers }
```

### 6.4 Migrar tests existentes

Mover cada función de test al archivo correspondiente según la tabla del 6.1.
No modificar la lógica, solo cambiar el `package` a `tests` y agregar el import
`"github.com/tinywasm/json"`.

```bash
gotest
```

---

## Stage 7: Cobertura 100% — Encoder (`tests/encode_*`)

### 7.1 `tests/encode_output_test.go` — completar con tests faltantes

```go
// TestEncodeToString — output *string (ausente en tests originales)
func TestEncodeToString(t *testing.T) { ... }

// TestEncodeInvalidOutput — output desconocido → error
func TestEncodeInvalidOutput(t *testing.T) { ... }
```

### 7.2 `tests/encode_types_test.go` — todos los tipos numéricos

```go
// TestEncodeNumericTypes — int, int32, int64, uint, uint64, float32, float64
func TestEncodeNumericTypes(t *testing.T) {
    cases := []struct {
        name     string
        val      any
        expected string
    }{
        {"int", int(5), `{"v":5}`},
        {"int32", int32(5), `{"v":5}`},
        {"int64", int64(5), `{"v":5}`},
        {"float32", float32(1.5), `{"v":1.5}`},
        {"float64", float64(1.5), `{"v":1.5}`},
        {"uint", uint(5), `{"v":5}`},
        {"uint64", uint64(5), `{"v":5}`},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            m := &mockFielder{
                schema: []fmt.Field{{Name: "V", Type: fmt.FieldInt, JSON: "v"}},
                values: []any{c.val},
            }
            var out string
            if err := json.Encode(m, &out); err != nil {
                t.Fatal(err)
            }
            if out != c.expected {
                t.Errorf("expected %s, got %s", c.expected, out)
            }
        })
    }
}
```

### 7.3 `tests/encode_basic_test.go` — ramas faltantes

```go
// TestEncodeStructNotFielder — FieldStruct cuyo value no implementa Fielder → omitido
func TestEncodeStructNotFielder(t *testing.T) { ... }

// TestEncodeControlChars — chars < 0x20 → \u00XX
func TestEncodeControlChars(t *testing.T) { ... }
```

```bash
gotest
```

---

## Stage 8: Cobertura 100% — Decoder y Parser (`tests/decode_*`, `tests/parser_*`)

### 8.1 `tests/decode_input_test.go`

```go
// TestDecodeInvalidInput — tipo desconocido → error
func TestDecodeInvalidInput(t *testing.T) { ... }

// TestDecodeNotObject — JSON no es objeto → error
func TestDecodeNotObject(t *testing.T) { ... }

// TestDecodeFromBytes — input []byte
func TestDecodeFromBytes(t *testing.T) { ... }
```

### 8.2 `tests/decode_types_test.go`

```go
// TestDecodeInt        — writeValue con *int
// TestDecodeInt32      — writeValue con *int32
// TestDecodeFloat32    — writeValue con *float32
// TestDecodeInt32FromFloat  — parser retorna float64 → *int32
// TestDecodeIntFromFloat    — parser retorna float64 → *int
// TestDecodeFloat32FromInt  — parser retorna int64  → *float32
```

### 8.3 `tests/decode_nested_test.go`

```go
// TestDecodeStructNotFielder  — ptr no implementa Fielder → campo descartado
// TestDecodeExtraNestedObject — campo desconocido objeto  → descartado
// TestDecodeExtraArray        — campo desconocido array   → descartado
```

### 8.4 `tests/parser_string_test.go`

```go
// TestParseStringEscapeBF      — \b \f \/
// TestParseStringUnicode       — \u0041 → 'A'
// TestParseStringUnicodeShort  — \u004 (3 chars) → error
// TestParseStringInvalidEscape — \q → error
// TestParseStringUnexpectedEOF — sin cierre → error
// TestParseStringNotQuote      — no empieza con " → error
```

### 8.5 `tests/parser_number_test.go`

```go
// TestParseNumberNegative   — -42 → int64(-42)
// TestParseNumberScientific — 1e2 → float64(100)
```

### 8.6 `tests/parser_bool_null_test.go`

```go
// TestParseBoolFalse        — false → bool(false)
// TestParseBoolInvalid      — tru   → error
// TestParseBoolFalseInvalid — fals  → error
// TestParseNullInvalid      — nul   → error
```

### 8.7 `tests/parser_array_test.go`

```go
// TestParseArrayEmpty          — [] → []any{}
// TestParseArrayMissingBracket — {} → error
// TestParseArrayBadSeparator   — [1;2] → error
// TestParseValueArray          — [1,2,3] via parseValue
// TestParseValueUnknownChar    — @invalid → error
```

### 8.8 `tests/parser_object_test.go`

```go
// TestParseObjectBadSeparator  — {"a":1;} → error
// TestParseObjectMissingColon  — {"a" 1}  → error
```

### 8.9 `tests/parser_limits_test.go`

```go
// TestParseIntoFielderNotObject — input no es '{' → error
// TestSkipWhitespace            — " \t\r\n{" → peek '{'
// TestPeekNextEmpty             — data vacía → 0
```

```bash
gotest
```

---

## Stage 9: Benchmarks vs `encoding/json`

Cada archivo de bench es pequeño y tiene un nombre claro.

### 9.1 `tests/bench_encode_test.go`

```go
package tests

import (
    stdjson "encoding/json"
    "testing"
    "github.com/tinywasm/fmt"
    "github.com/tinywasm/json"
)

type benchUser struct {
    Name  string
    Email string
    Age   int64
    Score float64
}

func (u *benchUser) Schema() []fmt.Field {
    return []fmt.Field{
        {Name: "Name", Type: fmt.FieldText, JSON: "name"},
        {Name: "Email", Type: fmt.FieldText, JSON: "email"},
        {Name: "Age", Type: fmt.FieldInt, JSON: "age"},
        {Name: "Score", Type: fmt.FieldFloat, JSON: "score"},
    }
}
func (u *benchUser) Values() []any   { return []any{u.Name, u.Email, u.Age, u.Score} }
func (u *benchUser) Pointers() []any { return []any{&u.Name, &u.Email, &u.Age, &u.Score} }

var benchInput = &benchUser{Name: "Alice", Email: "alice@example.com", Age: 30, Score: 9.5}

func BenchmarkEncode_tinywasm(b *testing.B) {
    var out string
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        if err := json.Encode(benchInput, &out); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkEncode_stdlib(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        if _, err := stdjson.Marshal(benchInput); err != nil {
            b.Fatal(err)
        }
    }
}
```

**Nota:** `benchUser` y `benchInput` se declaran en este archivo. Los otros bench files
los usan porque comparten `package tests`.

### 9.2 `tests/bench_decode_test.go`

```go
package tests

import (
    stdjson "encoding/json"
    "testing"
    "github.com/tinywasm/json"
)

var benchJSONStr = `{"name":"Alice","email":"alice@example.com","age":30,"score":9.5}`

func BenchmarkDecode_tinywasm(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        u := &benchUser{}
        if err := json.Decode(benchJSONStr, u); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkDecode_stdlib(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        var u benchUser
        if err := stdjson.Unmarshal([]byte(benchJSONStr), &u); err != nil {
            b.Fatal(err)
        }
    }
}
```

### 9.3 `tests/bench_roundtrip_test.go`

```go
package tests

import (
    stdjson "encoding/json"
    "testing"
    "github.com/tinywasm/json"
)

func BenchmarkRoundTrip_tinywasm(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        var out string
        if err := json.Encode(benchInput, &out); err != nil {
            b.Fatal(err)
        }
        u := &benchUser{}
        if err := json.Decode(out, u); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkRoundTrip_stdlib(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        data, err := stdjson.Marshal(benchInput)
        if err != nil {
            b.Fatal(err)
        }
        var u benchUser
        if err := stdjson.Unmarshal(data, &u); err != nil {
            b.Fatal(err)
        }
    }
}
```

### 9.4 Actualizar `benchmarks/clients/tinyjson/main.go`

El cliente WASM usa la API antigua (struct sin `Fielder`). Actualizar:
- Hacer que `User` implemente `fmt.Fielder` (Schema/Values/Pointers).
- Eliminar `json:"..."` struct tags.

### 9.5 Verificación

```bash
gotest
go test -bench=. -benchmem -count=3 ./tests/...
```

---

## Stage 10: Documentación y Publish

### 10.1 Actualizar `benchmarks/README.md`

Reemplazar la sección **Performance Results** (resultados obsoletos de la arquitectura
anterior con reflect) con los nuevos resultados de `go test -bench`:

```bash
go test -bench=. -benchmem -count=3 ./tests/... 2>&1 | tee /tmp/bench.txt
```

Estructura de la sección actualizada:

```markdown
## Performance Results

Last updated: <fecha>

### Go Benchmark (`go test -bench`)

| Benchmark | tinywasm/json | encoding/json | Δ allocs |
|-----------|--------------|---------------|----------|
| Encode    | X ns/op Y B/op Z allocs | ... | ... |
| Decode    | X ns/op Y B/op Z allocs | ... | ... |
| RoundTrip | X ns/op Y B/op Z allocs | ... | ... |

> Run: `go test -bench=. -benchmem ./tests/...`

### WASM Binary Size

| Implementation | Binary Size (WASM + Gzip) |
| :--- | :--- |
| **tinywasm/json** | **~27 KB** |
| encoding/json (stdlib) | ~119 KB |
```

También corregir el link roto: `clients/json/main.go` → `clients/tinyjson/main.go`.

### 10.2 Actualizar `README.md` — sección Benchmarks

El `benchmarks/README.md` ya referencia `../README.md#benchmarks` pero esa ancla
**no existe**. Agregar la sección al README principal:

```markdown
## Benchmarks

tinywasm/json es **77% más pequeño** que `encoding/json` en WASM (~27 KB vs ~119 KB)
y **zero-reflect**, eliminando overhead de introspección.

| Benchmark | tinywasm/json | encoding/json |
|-----------|--------------|---------------|
| Encode    | (resultado go test) | (resultado go test) |
| Decode    | (resultado go test) | (resultado go test) |

Ver resultados completos y análisis en [benchmarks/README.md](benchmarks/README.md).
```

También agregar el link a `benchmarks/README.md` en el índice del README principal
junto a los demás documentos en `docs/`.

### 10.3 Publicar

```bash
gopush 'json: reorganize tests by domain, 100% coverage, benchmarks vs encoding/json'
```

---

## Resumen

| Stage | Archivos | Acción |
|-------|----------|--------|
| 6 | `tests/helpers_test.go` + 10 archivos migrate | Mover todos los tests existentes a `tests/`, divididos por dominio |
| 6 | `encode_test.go`, `decode_test.go` (raíz) | **Eliminar** |
| 7 | `tests/encode_output_test.go` | `TestEncodeToString`, `TestEncodeInvalidOutput` |
| 7 | `tests/encode_types_test.go` | `TestEncodeNumericTypes` (int/int32/uint/float32/float64) |
| 7 | `tests/encode_basic_test.go` | `TestEncodeStructNotFielder`, `TestEncodeControlChars` |
| 8 | `tests/decode_input_test.go` | `TestDecodeInvalidInput`, `TestDecodeNotObject`, `TestDecodeFromBytes` |
| 8 | `tests/decode_types_test.go` | `TestDecodeInt`, `TestDecodeInt32`, `TestDecodeFloat32`, coerciones |
| 8 | `tests/decode_nested_test.go` | `TestDecodeStructNotFielder`, `TestDecodeExtraNestedObject`, `TestDecodeExtraArray` |
| 8 | `tests/parser_string_test.go` | 6 tests: todos los escapes |
| 8 | `tests/parser_number_test.go` | negativo, científica |
| 8 | `tests/parser_bool_null_test.go` | bool/null y sus errores |
| 8 | `tests/parser_array_test.go` | empty, bad sep, unknown char |
| 8 | `tests/parser_object_test.go` | bad sep, missing colon |
| 8 | `tests/parser_limits_test.go` | peek/next vacío, skipWhitespace, no `{` |
| 9 | `tests/bench_encode_test.go` | BenchmarkEncode tinywasm vs stdlib |
| 9 | `tests/bench_decode_test.go` | BenchmarkDecode tinywasm vs stdlib |
| 9 | `tests/bench_roundtrip_test.go` | BenchmarkRoundTrip tinywasm vs stdlib |
| 9 | `benchmarks/clients/tinyjson/main.go` | Actualizar User → Fielder |
| 10 | `benchmarks/README.md` | Actualizar resultados obsoletos + corregir link roto `clients/json` → `clients/tinyjson` |
| 10 | `README.md` | Agregar sección `#benchmarks` (ancla requerida por `benchmarks/README.md`) + link a `benchmarks/README.md` |

## Estructura final de `tests/`

```
tests/
├── helpers_test.go           — mockFielder
├── encode_basic_test.go      — string, int, bool, nil, bytes, control chars, FieldStruct not Fielder
├── encode_tags_test.go       — key, omitempty, "-", fallback, nested
├── encode_output_test.go     — *[]byte, *string, io.Writer, error output
├── encode_types_test.go      — todos los tipos numéricos
├── decode_basic_test.go      — JSON → campos básicos
├── decode_types_test.go      — *int, *int32, *float32, coerciones
├── decode_input_test.go      — []byte, string, io.Reader, error
├── decode_tags_test.go       — key, "-", missing, extra field
├── decode_nested_test.go     — recursivo, ptr no Fielder, discard object/array
├── parser_string_test.go     — todos los escapes, EOF, error
├── parser_number_test.go     — int64, float64, negativo, científica
├── parser_bool_null_test.go  — true, false, null, errores
├── parser_array_test.go      — empty, values, errores
├── parser_object_test.go     — keys, errores
├── parser_limits_test.go     — peek/next vacío, whitespace, char desconocido, no `{`
├── bench_encode_test.go      — BenchmarkEncode
├── bench_decode_test.go      — BenchmarkDecode
├── bench_roundtrip_test.go   — BenchmarkRoundTrip
└── json_test.go              — integración end-to-end (existente)
```
