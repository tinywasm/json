# PLAN: tinywasm/json — Pendientes post Stage 5 (v0.2.0)

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing`.
- **Testing Runner:** Use `gotest` (install: `go install github.com/tinywasm/devflow/cmd/gotest@latest`).
- **Max 500 lines per file.** If exceeded, subdivide by domain.
- **Flat hierarchy.** No subdirectories for library code.
- **TinyGo Compatible:** No `fmt`, `strings`, `strconv`, `errors` from stdlib. Use `tinywasm/fmt`.
- **No maps** in WASM code (binary bloat).
- **Documentation First:** Update docs before coding.
- **Publishing:** Use `gopush 'message'` after tests pass and docs are updated.

## Context

Stages 6–9 se ejecutaron parcialmente. Los archivos en `tests/` existen pero:
- `encode_test.go` y `decode_test.go` en raíz aún no fueron eliminados (Stage 6.2).
- Cobertura actual: **93.5%** (objetivo: 100%). Brechas en `parseObject`, `parseArray`, `parseString`, `encodeValue`, `Encode`, etc.
- `benchmarks/clients/tinyjson/main.go` aún usa la API antigua con struct tags (Stage 9.4).
- **Bug bloqueante:** `tinywasm/fmt.Convert(s).Float64()` falla con notación científica (e.g. `"1e2"`). Resuelto en `tinywasm/fmt` antes de continuar.

**Prerequisito:** Esperar el fix de `tinywasm/fmt` (scientific notation en `parseFloatBase`). Actualizar `go.mod` con la nueva versión de `tinywasm/fmt` antes de ejecutar tests.

---

## Stage A: Eliminar test files obsoletos de raíz

### A.1 Eliminar archivos

```bash
rm encode_test.go decode_test.go
```

### A.2 Verificar

```bash
gotest
```

---

## Stage B: Alcanzar 100% de cobertura

Funciones con brechas (medidas con `go test -coverprofile -coverpkg=. ./tests/...`):

| Función | Cobertura actual | Brecha |
|---------|-----------------|--------|
| `parseObject` | 69.2% | empty object `{}`, error paths |
| `parseArray` | 89.5% | `expected [` error (wrong first char) |
| `parseString` | 96.2% | `invalid unicode escape` (short `\u00X`) |
| `encodeValue` | 91.7% | `FieldBytes` sin `[]byte` type, `FieldStruct` → not Fielder |
| `Encode` | 90.9% | `io.Writer` error path |
| `encodeFielder` | 96.2% | `omitempty` con valor cero en struct |
| `parseJSONTag` | 93.8% | tag con solo `-` (campo ignorado) |
| `writeValue` | 96.3% | `*int64` pointer type |
| `parseIntoFielder` | 95.6% | objeto vacío `{}` como input al Decoder |

### B.1 `tests/parser_object_test.go` — agregar tests faltantes

```go
// TestParseObjectEmpty — {"a":{}} → objeto vacío como campo discartado
func TestParseObjectEmpty(t *testing.T) {
    m := &mockFielder{}
    if err := json.Decode(`{"a":{}}`, m); err != nil {
        t.Fatal(err)
    }
}

// TestParseObjectNotBrace — parseObject recibe input que no empieza con '{'
// Se activa cuando parseValue encuentra '{' pero el objeto anidado no es válido.
// Esto se logra decodificando un campo FieldStruct cuyo valor no es '{':
func TestParseObjectMissingOpenBrace(t *testing.T) {
    // Usar un Fielder con un campo FieldStruct y pasarle un valor que no sea '{'
    // parseIntoFielder ya cubre '{' faltante; aquí necesitamos parseObject directamente.
    // parseObject se llama desde parseValue cuando peek() == '{'.
    // Para forzar el error de parseObject ("expected {"), necesitamos que
    // parseValue llame a parseObject pero el primer char consumido no sea '{'.
    // Esto no es posible directamente via Decode. Este error path es dead code
    // (parseValue solo llama parseObject cuando peek() == '{', y parseObject
    // consume ese '{' como primer next()). Documentar como dead code en el comentario del test.
    t.Skip("parseObject 'expected {' is unreachable: parseValue only calls it when peek()=='{', ensuring next()=='{' always")
}
```

**Nota:** Revisando el código, `parseObject` es llamado desde `parseValue` solo cuando `peek() == '{'`. Inmediatamente `parseObject` llama `p.next()` que consume ese `'{'`. Por tanto, el branch `return nil, fmt.Err("json", "decode", "expected {")` en `parseObject` es **dead code** — nunca puede ejecutarse vía la API pública. La misma situación aplica a `parseArray` con `expected [`.

**Acción:** Documentar estos branches como dead code y eliminarlos de `parseObject` y `parseArray` para que la cobertura suba a 100%, o dejar un comentario explicando por qué son inalcanzables.

### B.2 Eliminar dead code en `parser.go`

En `parseArray` (línea ~156) y `parseObject` (línea ~262), eliminar los primeros `if p.next() != '['/'{' { return error }` ya que son dead code: `parseValue` solo los llama cuando `peek()` retorna el carácter correcto.

**Alternativa (preferida):** Convertir `parseArray` y `parseObject` en funciones que asumen que el primer `[`/`{` ya fue consumido por el caller (`parseValue`). Así se elimina el dead code y la lógica es más explícita.

```go
// parseArray asume que '[' ya fue consumido por parseValue
func (p *parser) parseArray() ([]any, error) {
    var res []any
    p.skipWhitespace()
    if p.peek() == ']' {
        p.next()
        return res, nil
    }
    for { ... }
    return res, nil
}
```

Y en `parseValue`:
```go
case '[':
    p.next() // consume '['
    return p.parseArray()
case '{':
    p.next() // consume '{'
    return p.parseObject()
```

Actualizar `parseObject` igual.

### B.3 Tests restantes para brechas reales

Agregar en los archivos correspondientes de `tests/`:

**`tests/parser_string_test.go`** — ya debe tener `TestParseStringUnicodeShort`. Si falta:
```go
func TestParseStringUnicodeShort(t *testing.T) {
    m := &mockFielder{}
    if err := json.Decode(`{"a":"\u004"}`, m); err == nil {
        t.Fatal("expected error for short unicode escape")
    }
}
```

**`tests/decode_types_test.go`** — agregar `*int64`:
```go
func TestDecodeInt64Ptr(t *testing.T) {
    var v int64
    m := &mockFielder{
        schema:   []fmt.Field{{Name: "V", Type: fmt.FieldInt, JSON: "v"}},
        pointers: []any{&v},
    }
    if err := json.Decode(`{"v":42}`, m); err != nil {
        t.Fatal(err)
    }
    if v != 42 {
        t.Errorf("expected 42, got %d", v)
    }
}
```

**`tests/encode_output_test.go`** — cubrir error de `io.Writer`:
```go
type errWriter struct{}
func (e *errWriter) Write(p []byte) (n int, err error) {
    return 0, fmt.Err("test", "write", "error")  // usar errors.New del stdlib aquí
}

func TestEncodeWriterError(t *testing.T) {
    m := &mockFielder{
        schema: []fmt.Field{{Name: "V", Type: fmt.FieldText, JSON: "v"}},
        values: []any{"hello"},
    }
    if err := json.Encode(m, &errWriter{}); err == nil {
        t.Fatal("expected error from writer")
    }
}
```

**`tests/encode_basic_test.go`** — cubrir `FieldBytes` con valor no-`[]byte`:
```go
func TestEncodeFieldBytesNonBytes(t *testing.T) {
    // FieldBytes con value que no es []byte → encodeValue lo omite
    m := &mockFielder{
        schema: []fmt.Field{{Name: "V", Type: fmt.FieldBytes, JSON: "v"}},
        values: []any{"notbytes"},
    }
    var out string
    if err := json.Encode(m, &out); err != nil {
        t.Fatal(err)
    }
    // El campo se omite o produce JSON vacío
}
```

### B.4 Verificar cobertura 100%

```bash
gotest
go test -coverprofile=/tmp/cover.out -coverpkg=. ./tests/... && go tool cover -func=/tmp/cover.out
```

---

## Stage C: Actualizar cliente WASM (Stage 9.4)

Archivo: `benchmarks/clients/tinyjson/main.go`

El cliente usa la API antigua (`User` con struct tags, sin `Fielder`). Actualizar para usar la nueva API:

```go
//go:build wasm

package main

import (
    "syscall/js"
    "github.com/tinywasm/fmt"
    "github.com/tinywasm/json"
)

type User struct {
    Name  string
    Email string
    Age   int64
}

func (u *User) Schema() []fmt.Field {
    return []fmt.Field{
        {Name: "Name",  Type: fmt.FieldText, JSON: "name"},
        {Name: "Email", Type: fmt.FieldText, JSON: "email"},
        {Name: "Age",   Type: fmt.FieldInt,  JSON: "age"},
    }
}
func (u *User) Values() []any   { return []any{u.Name, u.Email, u.Age} }
func (u *User) Pointers() []any { return []any{&u.Name, &u.Email, &u.Age} }

func main() {
    console := js.Global().Get("console")
    document := js.Global().Get("document")
    body := document.Get("body")

    h1 := document.Call("createElement", "h1")
    h1.Set("innerHTML", "JSON WASM Example")
    body.Call("appendChild", h1)

    user := &User{Name: "John Doe", Email: "john@example.com", Age: 30}

    var jsonData []byte
    if err := json.Encode(user, &jsonData); err != nil {
        console.Call("error", "Encode error:", err.Error())
        return
    }

    p1 := document.Call("createElement", "p")
    p1.Set("innerHTML", "Encoded JSON: "+string(jsonData))
    body.Call("appendChild", p1)

    decoded := &User{}
    if err := json.Decode(jsonData, decoded); err != nil {
        console.Call("error", "Decode error:", err.Error())
        return
    }

    p2 := document.Call("createElement", "p")
    p2.Set("innerHTML", "Decoded User: Name="+decoded.Name)
    body.Call("appendChild", p2)

    console.Call("log", "JSON example finished successfully")
    select {}
}
```

---

## Stage D: Actualizar benchmarks/README.md

Ejecutar benchmarks y actualizar resultados:

```bash
go test -bench=. -benchmem -count=3 ./tests/... 2>&1 | tee /tmp/bench.txt
```

Reemplazar la sección **Performance Results** en `benchmarks/README.md` con los nuevos resultados.
Actualizar la fecha "Last updated".
Corregir link roto `clients/json/main.go` → `clients/tinyjson/main.go` (si aún existe).

---

## Stage E: Publicar

```bash
gopush 'json: 100% coverage, fix dead code in parser, update WASM client to Fielder API'
```

Tag objetivo: `v0.2.0`

---

## Resumen

| Stage | Archivo(s) | Acción |
|-------|-----------|--------|
| A | `encode_test.go`, `decode_test.go` (raíz) | **Eliminar** |
| B | `parser.go` | Refactorizar `parseArray`/`parseObject` para eliminar dead code |
| B | `tests/parser_string_test.go` | `TestParseStringUnicodeShort` (si falta) |
| B | `tests/decode_types_test.go` | `TestDecodeInt64Ptr` |
| B | `tests/encode_output_test.go` | `TestEncodeWriterError` |
| B | `tests/encode_basic_test.go` | `TestEncodeFieldBytesNonBytes` |
| C | `benchmarks/clients/tinyjson/main.go` | Migrar a API Fielder |
| D | `benchmarks/README.md` | Actualizar resultados + fecha |
| E | — | `gopush` con tag `v0.2.0` |
