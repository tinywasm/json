# PLAN: Soporte FieldRaw en encode y decode

## Contexto

`tinywasm/fmt` v0.23.3 añadió `FieldRaw` — campo string con JSON pre-serializado que debe emitirse inline sin quotes. Este plan implementa el soporte en `tinywasm/json`.

## Restricciones del ecosistema

- Solo se usa `github.com/tinywasm/fmt` — ningún import de la stdlib (`encoding/json`, `strings`, etc.)
- Tamaño de binario mínimo: sin allocaciones innecesarias, reutilizar lógica existente del parser

## Reutilización de código existente

### Encode — `fmt.JSONEscape` NO se usa para FieldRaw

`FieldText` llama a `fmt.JSONEscape` + quotes. `FieldRaw` escribe el string directamente con `b.WriteString(*p)` — el string ya es JSON válido, no necesita escape ni quotes. Cero lógica nueva, cero allocaciones.

### Decode — `skipValue` como base de `captureValue`

El parser ya tiene `skipValue()` que recorre cualquier valor JSON (objeto, array, string, número, bool, null) avanzando `p.pos`. `captureValue` es la misma lógica pero guarda `p.data[start:p.pos]` antes de retornar. No duplica código — extrae el rango de bytes que `skipValue` ya recorrería de todas formas.

## Cambios en `encode.go`

### `encodeFromPtr` — añadir case `fmt.FieldRaw`

```go
case fmt.FieldRaw:
    if p, ok := ptr.(*string); ok && *p != "" {
        b.WriteString(*p)
    } else {
        b.WriteString("null")
    }
```

### `isZeroPtr` — añadir case `fmt.FieldRaw`

```go
case fmt.FieldRaw:
    if p, ok := ptr.(*string); ok {
        return *p == ""
    }
    return true
```

## Cambios en `parser.go`

### Añadir `captureValue` — captura bytes de un valor JSON sin allocar

```go
// captureValue captura los bytes del próximo valor JSON sin interpretarlos.
// Reutiliza la misma lógica de avance que skipValue.
func (p *parser) captureValue() ([]byte, error) {
    p.skipWhitespace()
    start := p.pos
    if err := p.skipValue(); err != nil {
        return nil, err
    }
    return p.data[start:p.pos], nil
}
```

No alloca — retorna un slice de `p.data` (read-only, mismo backing array).

### `parseIntoPtr` — añadir case `fmt.FieldRaw`

```go
case fmt.FieldRaw:
    raw, err := p.captureValue()
    if err != nil {
        return err
    }
    if sp, ok := ptr.(*string); ok {
        *sp = string(raw)
    }
    return nil
```

La única allocación es `string(raw)` — necesaria para guardar el valor fuera del slice de entrada.

## Tests a añadir en `tests/`

Siguiendo la convención de archivos existentes (`encode_types_test.go`, `decode_types_test.go`):

### `encode_types_test.go` — casos `FieldRaw`

```go
// FieldRaw: objeto JSON inline (sin double-encoding)
{"raw object", ptrString(`{"a":1}`), fmt.FieldRaw, `{"v":{"a":1}}`},
// FieldRaw: array JSON inline
{"raw array",  ptrString(`[1,2,3]`), fmt.FieldRaw, `{"v":[1,2,3]}`},
// FieldRaw: string vacía → null
{"raw empty",  ptrString(""),        fmt.FieldRaw, `{"v":null}`},
// FieldRaw con omitempty: string vacía → campo omitido
{"raw omitempty", ptrString(""),     fmt.FieldRaw, `{}`},  // campo con OmitEmpty:true
```

### `decode_types_test.go` — casos `FieldRaw`

```go
// FieldRaw: captura objeto JSON tal cual
{`{"v":{"a":1}}`, fmt.FieldRaw, `{"a":1}`},
// FieldRaw: captura array
{`{"v":[1,2,3]}`, fmt.FieldRaw, `[1,2,3]`},
// FieldRaw: captura null
{`{"v":null}`,    fmt.FieldRaw, `null`},
```

### `encode_tags_test.go` / `decode_tags_test.go` — roundtrip MCP realista

Struct que simula una respuesta MCP con `Result string` de tipo `FieldRaw`:
- Encode produce `{"jsonrpc":"2.0","result":{"tools":[...]}}` (sin quotes en result)
- Decode recupera el JSON del result como string sin modificar

## Documentación

Actualizar `docs/` con nueva entrada en el diagrama/tabla de tipos:

| FieldType | Go type | JSON encode | JSON decode |
|---|---|---|---|
| `FieldRaw` | `*string` | valor inline sin quotes | captura bytes como string |
