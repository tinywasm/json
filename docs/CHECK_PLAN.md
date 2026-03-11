# PLAN: JSON Parser — Optimizaciones de Reutilización (Stage 5)

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing` for tests.
- **Testing Runner:** Install and use `gotest`:
  ```bash
  go install github.com/tinywasm/devflow/cmd/gotest@latest
  ```
- **Max 500 lines per file.** If exceeded, subdivide by domain.
- **Flat hierarchy.** No subdirectories for library code.
- **TinyGo Compatible:** No `fmt`, `strings`, `strconv`, `errors` from stdlib. Use `tinywasm/fmt`.
- **No maps** in WASM code (binary bloat). Applies to internal code too.
- **Documentation First:** Update docs before coding.
- **Publishing:** Use `gopush 'message'` after tests pass and docs are updated.

## Context

Los Stages 1–4 del plan anterior fueron ejecutados correctamente:
- `encode.go`, `decode.go`, `parser.go` creados.
- `codec_wasm.go`, `codec_stdlib.go` eliminados.
- `fmt.JSONEscape`, `fmt.IsZero`, `fmt.Convert`, `fmt.Builder`, `fmt.Err` reutilizados en encode.

Se identificaron **3 oportunidades de reutilización pendientes** en `parser.go` y `decode.go`:

| # | Problema | Archivo | Impacto |
|---|----------|---------|---------|
| 1 | `decodeHex` duplica `fmt.Convert(s).Int64(16)` | `parser.go:114` | Eliminar 15 líneas propias |
| 2 | `parseString` usa `[]byte`+`append` en lugar de `fmt.Builder` | `parser.go:69` | Consistencia + menos allocations |
| 3 | `map[string]any` en decoder → binary bloat en WASM | `parser.go:202`, `decode.go:40` | Eliminar `map` del binario |

---

## Stage 5: Eliminar Duplicación y `map` del Decoder

### 5.1 Fix `decodeHex` → usar `fmt.Convert(s).Int64(16)`

**Archivo:** `parser.go`

Eliminar la función `decodeHex` completa (líneas 114–128) y reemplazar su uso inline:

```go
// ANTES — función propia de 15 líneas:
func decodeHex(s string) int { ... }
// uso: val := decodeHex(hex)

// DESPUÉS — cero líneas extras, reutiliza fmt:
val, _ := fmt.Convert(string(p.data[p.pos : p.pos+4])).Int64(16)
```

Cambio en `parseString` donde se llama `decodeHex`:
```go
case 'u':
    if p.pos+4 > len(p.data) {
        return "", fmt.Err("json", "decode", "invalid unicode escape")
    }
    val, _ := fmt.Convert(string(p.data[p.pos : p.pos+4])).Int64(16)
    p.pos += 4
    b.WriteByte(byte(val))
```

### 5.2 Fix `parseString` → usar `fmt.Builder`

**Archivo:** `parser.go`

Reemplazar `var b []byte` + `append` por `fmt.Builder`:

```go
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
```

### 5.3 Eliminar `map[string]any` — Decode de una sola pasada

**Problema:** `parseObject` retorna `map[string]any` y `decodeFielder` la itera después. Esto requiere que `map` exista en el binario WASM.

**Solución:** Decodificar directamente mientras se parsea. Eliminar `parseObject` del flujo principal del decoder y agregar `parseIntoFielder` en `parser.go`.

**Archivo:** `parser.go` — agregar método nuevo:

```go
// parseIntoFielder parsea un objeto JSON escribiendo directamente en el Fielder.
// Elimina la necesidad de map[string]any como intermediario.
func (p *parser) parseIntoFielder(f fmt.Fielder) error {
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

        // Buscar el campo en el schema
        fieldIdx := -1
        for i, field := range schema {
            k, _ := parseJSONTag(field)
            if k == key {
                fieldIdx = i
                break
            }
        }

        if fieldIdx < 0 {
            // Campo desconocido: parsear y descartar
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
```

**Archivo:** `decode.go` — simplificar `Decode`:

```go
func Decode(input any, data fmt.Fielder) error {
    var raw []byte
    switch in := input.(type) {
    case []byte:
        raw = in
    case string:
        raw = []byte(in)
    case io.Reader:
        var buf []byte
        tmp := make([]byte, 4096)
        for {
            n, err := in.Read(tmp)
            if n > 0 {
                buf = append(buf, tmp[:n]...)
            }
            if err != nil {
                break
            }
        }
        raw = buf
    default:
        return fmt.Err("json", "decode", "input must be []byte, string, or io.Reader")
    }

    p := &parser{data: raw}
    p.skipWhitespace()
    return p.parseIntoFielder(data)
}
```

**Eliminar de `decode.go`:** la función `decodeFielder` completa (ya no se necesita).

**Nota:** `parseObject` y `parseArray` se mantienen en `parser.go` porque `parseValue` los necesita para manejar valores anidados desconocidos al descartar campos. Sin embargo, ya **no son el camino principal** del decode de Fielder.

### 5.4 Tests

Verificar que todos los tests existentes siguen pasando sin cambios:

```bash
gotest
```

Tests adicionales que deben pasar:
- `TestDecodeSimple` — decode directo sin mapa intermedio
- `TestDecodeNested` — structs anidados vía `parseIntoFielder` recursivo
- `TestDecodeExtraField` — campos desconocidos descartados silenciosamente
- `TestDecodeStringEscapes` — `\uXXXX` vía `fmt.Convert(...).Int64(16)`

### 5.5 Verificación final

```bash
# Sin map en el decoder principal:
grep -n "map\[string\]any" decode.go    # Solo debe aparecer en decodeFielder si se mantiene como fallback, o cero

# Sin decodeHex propio:
grep -n "decodeHex" parser.go           # Cero resultados

# Sin imports stdlib prohibidos:
grep -rn "\"strings\"\|\"strconv\"\|\"errors\"\|\"fmt\"" *.go  # Cero resultados
```

### 5.6 Publicar

```bash
gopush 'json parser: reuse fmt.Convert for hex, fmt.Builder in parseString, single-pass decode eliminates map'
```

---

## Resumen de Cambios

| Archivo | Cambio |
|---------|--------|
| `parser.go` | Eliminar `decodeHex`; `parseString` usa `fmt.Builder`; agregar `parseIntoFielder` |
| `decode.go` | `Decode` llama `parseIntoFielder` directamente; eliminar `decodeFielder` |

## Reutilización de `tinywasm/fmt` tras este Stage

| Qué | Función fmt | Usado en |
|-----|-------------|----------|
| Escape de strings | `fmt.JSONEscape(s, b)` | `encodeValue`, `encodeFielder` |
| Check zero | `fmt.IsZero(v)` | `encodeFielder` (omitempty) |
| Número → string | `fmt.Convert(v).String()` | `encodeValue` |
| String → int64 | `fmt.Convert(s).Int64()` | `parser.parseNumber` |
| String → int64 base 16 | `fmt.Convert(s).Int64(16)` | `parser.parseString` (`\uXXXX`) ✨ nuevo |
| String → float64 | `fmt.Convert(s).Float64()` | `parser.parseNumber` |
| String builder | `fmt.Builder` | Encoder + `parseString` ✨ nuevo |
| Creación de errores | `fmt.Err(...)` | Todos los errores |
