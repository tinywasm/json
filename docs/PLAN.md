# PLAN: json — Eliminar validación de Decode

← [README](../README.md)

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing`.
- **Testing Runner:** Use `gotest` (install: `go install github.com/tinywasm/devflow/cmd/gotest@latest`).
- **Max 500 lines per file.** If exceeded, subdivide by domain.
- **Flat hierarchy.** No subdirectories for library code.
- **Documentation First:** Update docs before coding.
- **Publishing:** Use `gopush 'message'` after tests pass and docs are updated.

## Contexto

`tinywasm/json` es una librería de codificación/decodificación únicamente.
No debe validar datos — esa responsabilidad es del caller (form, orm, crudp).

El post-decode `Validate()` que se agregó en v3 fue un error de diseño.

---

## Stage 1: Simplificar `decode.go`

**File:** `decode.go`

### 1.1 Eliminar post-decode Validate de `Decode()`

`Decode` pasa a ser idéntico al actual `DecodeRaw`:

```go
// ANTES:
func Decode(input any, data fmt.Fielder) error {
    if err := DecodeRaw(input, data); err != nil {
        return err
    }
    if v, ok := data.(fmt.Validator); ok {
        return v.Validate()
    }
    return nil
}

// DESPUÉS:
func Decode(input any, data fmt.Fielder) error {
    var raw []byte
    switch in := input.(type) {
    case []byte:
        raw = in
    case string:
        raw = unsafe.Slice(unsafe.StringData(in), len(in))
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
    p := parser{data: raw}
    return p.parseIntoFielder(data)
}
```

### 1.2 Eliminar `DecodeRaw`

Con `Decode` sin validación, `DecodeRaw` es redundante. Eliminar la función.

### 1.3 Eliminar import `fmt.Validator`

Si `fmt.Validator` solo se usaba para el post-decode validate, el import puede simplificarse.

---

## Stage 2: Actualizar tests

### 2.1 Eliminar `tests/validation_test.go`

El archivo `validation_test.go` testea el comportamiento de validación post-decode que ya no existe. Eliminar el archivo.

### 2.2 Verificar que otros tests no dependan de `DecodeRaw`

Actualizar cualquier test que llame `DecodeRaw` → reemplazar con `Decode`.

```bash
gotest
```

---

## Stage 3: Actualizar documentación

**File:** `README.md`

- Actualizar sección `Decode` (línea 72-78): eliminar referencia a que llama `Validate()`

```
// ANTES:
Parses JSON into a Fielder and calls Validate() if the fielder implements fmt.Validator.

// DESPUÉS:
Parses JSON into a Fielder.
```

- Eliminar sección `DecodeRaw` (líneas 79-84) — ya no existe

---

```bash
gopush 'json: eliminar post-decode Validate y DecodeRaw — json solo codifica/decodifica'
```

---

## Stage 5: Alineación con Validación Contextual

**Motivación:** Con el cambio en `fmt` a `ValidateFields(action byte, data Fielder)`, 
la validación post-decodificación es ahora responsabilidad del llamador (caller). 
`json.Decode` se mantiene como una herramienta agnóstica de transporte.

### 5.1 Documentación del flujo recomendado

Actualizar `README.md` para mostrar explícitamente el patrón de uso:

```go
// 1. Decodificar (solo llena los campos del struct)
if err := json.Decode(input, &user); err != nil {
    return err
}

// 2. Validar con contexto (acción 'c'reate, 'u'pdate, etc.)
// Nota: user implementa fmt.Fielder (o fmt.Model)
if err := fmt.ValidateFields('c', &user); err != nil {
    return err
}
```

### 5.2 Independencia de Interfases de Validación

Confirmar que `json` no importa `fmt.Validator` ni depende de métodos `Validate()`. 
Su única dependencia fuerte en `fmt` es la interfaz `Fielder` para acceder 
a `Schema()` y `Pointers()`.
