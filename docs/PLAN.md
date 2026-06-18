# PLAN — `json` al codec tipado: `Encode`/`Decode` 0-alloc (una sola forma) · BREAKING

> Este plan se despacha vía el workflow CodeJob. Ver skill: `agents-workflow`.
> **Estado:** LISTO PARA REVISIÓN DEL USUARIO.
> **Repo objetivo:** `github.com/tinywasm/json`.
> **Depende de (GATE):** `tinywasm/fmt` con el contrato del codec publicado (`fmt/docs/PLAN.md`)
> y `ormc` generando `EncodeFields`/`DecodeFields` en los modelos (`orm/docs/PLAN.md`).
> **Tipo:** breaking change (firmas de `Encode`/`Decode`: `fmt.Fielder` → `fmt.Encodable`/
> `fmt.Decodable`).
> **Objetivo:** que `json` serialice por el **mismo** contrato tipado que `jsvalue` y `binary`
> (`fmt.FieldWriter`/`FieldReader`) y sea **0-alloc** (deja de construir `Pointers() []any`).

## Reglas permanentes del repo → `AGENTS.md`

Las restricciones del ecosistema (no stdlib → `tinywasm/fmt`; **no `map`**; **no `reflect`**;
agnóstico compila wasm+backend; 0-alloc; `gotest` no `go test`) están en
`AGENTS.md`. Este plan NO las repite completas; solo inlinea lo crítico de la
tarea (ver Checklist).

## Prerequisito (PRIMERO — entorno del agente)

```bash
go install github.com/tinywasm/devflow/cmd/gotest@latest
```

Usar `gotest` (sin argumentos); **NO** `go test` directo.

## Contexto y motivación (autocontenido)

`json` HOY ya es reflect-free y map-free: serializa recorriendo `Schema()` (un `[]fmt.Field`
global, 0-alloc) y `Pointers()` (`[]any{&campo,...}`), con un `switch` por `FieldType` en
`encodeFromPtr`. **Lo único que NO es 0-alloc es `Pointers()`**: construye un `[]any` y boxea
cada puntero **en cada llamada** a `Encode`/`Decode`.

El codec tipado de `fmt` elimina eso: el modelo escribe/lee sus campos con llamadas tipadas
(`w.String("name", m.Name)`), sin `[]any`, sin slice, **0-alloc**. Migrar `json` al codec da:
(1) 0-alloc, (2) **una sola forma** de serializar en el ecosistema (la misma que `jsvalue` y `binary`).

### Contrato de `fmt` (ya publicado, referencia)

```go
type FieldWriter interface { String(name,val string); Int(name string,val int64); Uint(...); Float(...); Bool(...); Bytes(name string,val []byte); Null(name string); Object(name string,val Encodable); Array(name string,n int) ArrayWriter }
type ArrayWriter interface { String(val string); Int(val int64); Float(val float64); Bool(val bool); Bytes(val []byte); Object(val Encodable) }
type Encodable interface { EncodeFields(w FieldWriter); IsNil() bool }
type FieldReader interface { String(name string)(string,bool); Int(...); Uint(...); Float(...); Bool(...); Bytes(...)([]byte,bool); Object(name string,into Decodable) bool; Array(name string)(ArrayReader,bool) }
type ArrayReader interface { Len() int; String(i int) string; Int(i int) int64; Float(i int) float64; Bool(i int) bool; Bytes(i int) []byte; Object(i int,into Decodable) bool }
type Decodable interface { DecodeFields(r FieldReader) error; IsNil() bool }
```

## Diseño

`json` implementa los writer/reader **JSON canónicos** (es la librería con el parser; `fmt` NO
trae JSON propio para no duplicar). Migra `Encode`/`Decode` al codec y **elimina** el camino
basado en `Schema()`/`Pointers()` para serializar (queda solo en `orm` para el scan SQL).

### `jsonWriter` (`fmt.FieldWriter` & `fmt.ArrayWriter`)

Escribe JSON a un `*fmt.Conv` reusado:
- Métodos estándar: `String` → `"` + `fmt.JSONEscape` + `"`; `Int`/`Uint` → `WriteInt`; `Float` → `WriteFloat`;
  `Bool` → `true`/`false`; `Bytes` → `"` + `JSONEscape(string(b))` + `"`; `Null` → `null`;
  `Object` → si es nil (`fmt.IsNil(val)`) escribe `null`, si no `{...}` recursivo y llama `EncodeFields`.
- `Array(name, n)` → escribe `[` + retorna `jsonArrayWriter` (pre-alocado o estructurado) para escribir elementos separados por comas.
- 0-alloc: reusa el buffer del `Conv`; nunca crea `[]any`.

### `jsonReader` (`fmt.FieldReader` & `fmt.ArrayReader`)

Lee un campo **por nombre** desde el objeto JSON ya posicionado por el parser:
- **`jsonReader` por re-escaneo, 0-alloc y map-free.** Guarda el offset de inicio del objeto;
  cada `r.Tipo("name")` re-escanea las claves del objeto desde ese offset hasta encontrar
  `"name"` y parsea su valor. **Sin `map`, sin slices intermedios → 0-alloc.**
- `Array(name)` → lee `[` + retorna `jsonArrayReader` que permite leer por índice sin alocar en Go.

## Pasos de ejecución

### Stage 1 — `jsonWriter` y migrar `Encode`
1. Crear `jsonWriter` (`fmt.FieldWriter` y `fmt.ArrayWriter`) portando `encodeFromPtr` a métodos
   tipados sobre `*fmt.Conv`. Manejar comas y `OmitEmpty` en el writer.
2. Cambiar la firma: `func Encode(data fmt.Encodable, output any) error`. `Encode` crea el
   `jsonWriter`, valida `fmt.IsNil(data)` (escribiendo `"null"`), llama `data.EncodeFields(w)`, y vuelca a `*[]byte`/`*string`/`io.Writer`.
3. **Eliminar** `encodeFromPtr`, `encodeFielder`, `isZeroPtr` (basados en `Pointers()`): su
   lógica vive ahora en `jsonWriter` (no duplicar).

### Stage 2 — `jsonReader` y migrar `Decode`
4. Crear `jsonReader` (`fmt.FieldReader` y `fmt.ArrayReader`) por re-escaneo. Reusar el lexer/parser
   existente.
5. Cambiar la firma de decode: `func Decode(input any, data fmt.Decodable) error`. Validar `fmt.IsNil(data)` (retornando error), y llamar `data.DecodeFields(r)`. **Eliminar**
   `parseIntoFielder` basado en `Schema()`/`Pointers()`.

### Stage 3 — tests
6. Adaptar tests: los tipos de test implementan `fmt.Encodable`/`fmt.Decodable` e `IsNil() bool { return m == nil }`. Cubrir round-trip: primitivos, anidado, slices, `[]byte`, `OmitEmpty`, typed nil pointer, claves desconocidas (ignorar), y campos ausentes.
7. **0-alloc**: `testing.AllocsPerRun` sobre `Encode` (con writer/buffer reusado) → **0 asignaciones del heap Go**.

### Stage 4 — actualizar el benchmark existente (antes/después) — OBLIGATORIO
8. Correr benchmarks (`benchmarks/` y `benchmarks/README.md`) anotando `ns/op`/`B/op`/`allocs/op` antes y después de la migración.
9. Actualizar la tabla de "Performance Results" reflejando la mejora a **0 allocs** y el tamaño WASM reducido.

### Stage 5 — documentación (OBLIGATORIO)
10. Actualizar `README.md` y `docs/*` documentando las nuevas firmas basadas en `Encodable`/`Decodable` e `IsNil()`. Enlazar a `benchmarks/README.md`.
