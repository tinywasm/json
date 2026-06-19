# PLAN — `json` al codec tipado: `Encode`/`Decode` 0-alloc (una sola forma) · BREAKING

> Este plan se despacha vía el workflow CodeJob. Ver skill: `agents-workflow`.
> **Estado:** ✅ DONE (2026-06-19).
> **Repo objetivo:** `github.com/tinywasm/json`.
> **Depende de (GATE):** `tinywasm/fmt` con el contrato del codec publicado (`fmt/docs/PLAN.md`)
> y `ormc` generando `EncodeFields`/`DecodeFields` en los modelos (`orm/docs/PLAN.md`).
> **Tipo:** breaking change (firmas de `Encode`/`Decode`: `fmt.Fielder` → `fmt.Encodable`/
> `fmt.Decodable`).
> **Objetivo:** que `json` serialice por el **mismo** contrato tipado que `jsvalue`
> (`fmt.FieldWriter`/`FieldReader`) y sea **0-alloc** (deja de construir `Pointers() []any`).

## Reglas permanentes del repo → `AGENTS.md`

Las restricciones del ecosistema (no stdlib → `tinywasm/fmt`; **no `map`**; **no `reflect`**;
agnóstico compila wasm+backend; 0-alloc; `gotest` no `go test`) están en
[`AGENTS.md`](../AGENTS.md). Este plan NO las repite completas; solo inlinea lo crítico de la
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
(1) 0-alloc, (2) **una sola forma** de serializar en el ecosistema (la misma que `jsvalue`).

### Contrato de `fmt` (ya publicado, referencia)

```go
type FieldWriter interface { String(name,val string); Int(name string,val int64); Uint(...); Float(...); Bool(...); Bytes(name string,val []byte); Null(name string); Object(name string,val Encodable); Array(name string,n int,each func(i int,a ArrayWriter)) }
type Encodable interface { EncodeFields(w FieldWriter) }
type FieldReader interface { String(name string)(string,bool); Int(...); Uint(...); Float(...); Bool(...); Bytes(...)([]byte,bool); Object(name string,into Decodable) bool; Array(name string)(ArrayReader,bool) }
type Decodable interface { DecodeFields(r FieldReader) error }
```

## Diseño (resuelto — para revisión del usuario)

`json` implementa los writer/reader **JSON canónicos** (es la librería con el parser; `fmt` NO
trae JSON propio para no duplicar). Migra `Encode`/`Decode` al codec y **elimina** el camino
basado en `Schema()`/`Pointers()` para serializar (queda solo en `orm` para el scan SQL).

- **`jsonWriter` (`fmt.FieldWriter`)**: escribe JSON a un `*fmt.Conv` reusado. Portar la lógica
  ya existente de `encodeFromPtr` a los métodos tipados:
  - `String` → `"` + `fmt.JSONEscape` + `"`; `Int`/`Uint` → `WriteInt`; `Float` → `WriteFloat`;
    `Bool` → `true`/`false`; `Bytes` → `"` + `JSONEscape(string(b))` + `"`; `Null` → `null`;
    `Object` → `{...}` recursivo; `Array` → `[...]`. Manejo de comas/`OmitEmpty` en el writer.
  - 0-alloc: reusa el buffer del `Conv`; nunca crea `[]any`.
- **`jsonReader` (`fmt.FieldReader`)**: lee un campo **por nombre** desde el objeto JSON ya
  posicionado por el parser. Ver la decisión de control de flujo abajo.

### Decisión: decode push (actual) → pull (codec)

El parser actual (`parseIntoFielder`) es **push** (el parser recorre el JSON y escribe en
`pointers[idx]`). El codec es **pull** (`DecodeFields` pide `r.String("name")`). Resolución:

> **`jsonReader` por re-escaneo, 0-alloc y map-free.** Guarda el offset de inicio del objeto;
> cada `r.Tipo("name")` re-escanea las claves del objeto desde ese offset hasta encontrar
> `"name"` y parsea su valor. **Sin `map`, sin slices intermedios → 0-alloc.** Es O(campos²) por
> objeto, despreciable para structs normales (el caso común). `DecodeFields` generado por `ormc`
> pide los campos en orden de schema, que coincide con el orden que produce este `Encode`, así
> que el re-escaneo casi siempre acierta al primer/siguiente intento.

(Alternativa descartada: bufferizar el objeto en slices paralelas `names[]/values[]` → O(campos)
pero **asigna**; viola 0-alloc. Y `map` está prohibido.)

## Pasos de ejecución

### Stage 1 — `jsonWriter` y migrar `Encode`
1. Crear `jsonWriter` (`fmt.FieldWriter`) portando `encodeFromPtr`/`encodeFielder` a métodos
   tipados sobre `*fmt.Conv`. Manejar comas y `OmitEmpty` dentro del writer.
2. Cambiar la firma: `func Encode(data fmt.Encodable, output any) error`. `Encode` crea el
   `jsonWriter`, llama `data.EncodeFields(w)`, y vuelca a `*[]byte`/`*string`/`io.Writer` (igual
   que hoy). Para slices, soportar `fmt.FielderSlice` → o un `Encodable` de slice / iteración con
   `[` `]` (mantener el comportamiento de `encodeSlice`).
3. **Eliminar** `encodeFromPtr`, `encodeFielder`, `isZeroPtr` (basados en `Pointers()`): su
   lógica vive ahora en `jsonWriter` (no duplicar).

### Stage 2 — `jsonReader` y migrar `Decode`
4. Crear `jsonReader` (`fmt.FieldReader`) por re-escaneo (ver decisión). Reusar el lexer/parser
   existente para parsear el valor de una clave.
5. Cambiar la firma de decode a `fmt.Decodable` y llamar `data.DecodeFields(r)`. **Eliminar**
   `parseIntoFielder` basado en `Schema()`/`Pointers()`.

### Stage 3 — tests
6. Adaptar tests: los tipos de test implementan `fmt.Encodable`/`fmt.Decodable` (a mano o vía
   `ormc`). Cubrir round-trip: primitivos, anidado, slices, `[]byte`, `OmitEmpty`, claves
   desconocidas (ignorar), y campos ausentes.
7. **0-alloc**: `testing.AllocsPerRun` sobre `Encode` (con writer/buffer reusado) → **0
   asignaciones del heap Go** (excluyendo la copia final a `*[]byte`, que es inherente a la API).
8. `gotest` verde.

### Stage 4 — actualizar el benchmark existente (antes/después) — OBLIGATORIO
Esta migración es un cambio de **rendimiento**: hay que medirlo y dejarlo registrado.
`json` es donde más se mueve información del ecosistema. **YA EXISTE** la infra de benchmark en
`benchmarks/` — **NO crear** un doc nuevo; **actualizar** lo que hay:
- `benchmarks/clients/tinyjson/main.go` (usa tinywasm/json) y `benchmarks/clients/stdlib/main.go`
  (usa `encoding/json`).
- `benchmarks/build.sh` → compila wasm e imprime tamaño sin comprimir + gzip.
- `benchmarks/README.md` → sección **"Performance Results"** con tabla `go test -bench`
  (tinywasm/json vs encoding/json, Δ allocs) + tamaños wasm + `screenshots/`.

Pasos:
8. **Baseline ANTES de migrar:** correr el benchmark Go actual (`go test -bench=. -benchmem`) y
   `benchmarks/build.sh` / `build.sh stlib` sobre el código basado en `Pointers()`; anotar
   `ns/op`/`B/op`/`allocs/op` y tamaño wasm (sin comprimir + gzip).
9. **Asegurar que `benchmarks/clients/tinyjson/main.go` ejercita el camino del codec** tras la
   migración (el modelo de ejemplo debe implementar `fmt.Encodable`/`fmt.Decodable`); si el
   benchmark Go vive en otro `_test.go`, ajustarlo igual.
10. **Medir DESPUÉS** (codec): re-correr bench + `build.sh`. Esperado: **0 `allocs/op`** en
    `Encode` (excluyendo la copia final a `*[]byte`); `encoding/json` sigue con allocs > 0 y
    binario mucho mayor (arrastra `reflect`).
11. **Actualizar `benchmarks/README.md`**: la tabla "Performance Results" con los números nuevos
    y un renglón/columna **Antes (Pointers) | Después (codec)** mostrando la mejora de allocs.
    Refrescar tamaños wasm y, si cambian visiblemente, los `screenshots/`. Actualizar
    "Last updated".

### Stage 5 — documentación (OBLIGATORIO)
12. **`README.md`** (raíz) y `docs/*`: `Encode`/`Decode` ahora reciben `fmt.Encodable`/
    `fmt.Decodable` (no `fmt.Fielder`); serialización por el codec tipado 0-alloc; misma forma
    que `jsvalue`. Quitar referencias a que serializa vía `Pointers()`. Enlazar
    `benchmarks/README.md` para los números.

## Verificación (repo-local, ejecutable por el agente)

```bash
# 1. json ya no serializa vía Pointers()/Schema() (eso queda en orm para SQL):
grep -nE '\.Pointers\(\)|encodeFromPtr|parseIntoFielder' *.go | grep -v _test && echo "FALLA: queda el camino viejo" || echo "OK"

# 2. sin map ni reflect:
grep -nE 'map\[|"reflect"' *.go | grep -v _test && echo "FALLA" || echo "OK"

# 3. tests + 0-alloc:
gotest
```

## Checklist de calidad (obligatorio)

- **0-alloc** en `Encode` (medido), reusando el buffer del `Conv`; nunca `[]any`.
- **Sin `map`, sin `reflect`, sin `any`** en el camino de serialización (el `output any` de
  `Encode` es solo el destino `*[]byte`/`*string`/`io.Writer`, no el dato).
- **Sin duplicación:** la lógica por tipo vive SOLO en `jsonWriter` (borrar `encodeFromPtr` &co.).
- **Una sola forma:** el mismo contrato `fmt.FieldWriter`/`FieldReader` que `jsvalue`.
- Reglas genéricas del ecosistema (no stdlib/map/reflect/any): ver [`AGENTS.md`](../AGENTS.md).

## Tabla de stages

| Stage | Objetivo | Entregable | Criterio de salida |
|---|---|---|---|
| 1 | Encode al codec | `jsonWriter` + `Encode(fmt.Encodable,...)` | borra `encodeFromPtr`/`isZeroPtr` |
| 2 | Decode al codec | `jsonReader` + `Decode(...fmt.Decodable)` | borra `parseIntoFielder` |
| 3 | Tests + 0-alloc | tipos test `Encodable`/`Decodable`; `AllocsPerRun==0` | `gotest` verde |
| 4 | Comparativa antes/después | **actualizar `benchmarks/`** (README "Performance Results" + `build.sh` + clients): Antes\|Después\|stdlib | 0 allocs en `Encode`; tabla actualizada |
| 5 | Documentación | `README.md`/`docs/*` + enlace a `benchmarks/README.md` | sin mención a `Pointers()` en serialización |

## Nota (coordinación)

GATEs: `fmt` (contrato) y `ormc` (genera `EncodeFields`/`DecodeFields` en los modelos reales).
`json` puede testear con tipos propios que implementen el contrato a mano, pero sus consumidores
pasan modelos generados → mergear DESPUÉS de `ormc`. `fmt.Fielder`/`Pointers()` siguen existiendo
para el scan SQL en `orm` (operación distinta, no serialización). Ver
`~/Dev/Project/tinywasm/docs/SIZE_OPTIMIZATION_MASTER_PLAN.md`.
