# PLAN — Implementar `Raw()` en `jsonWriter`

> **Repo:** `github.com/tinywasm/json`
> **Archivo:** `encode.go`
> **Tipo:** implementación de nueva interfaz
> **Prerequisito:** `tinywasm/fmt` publicado con `Raw()` en `FieldWriter`/`FieldReader`

## Contexto

`tinywasm/fmt` extendió `FieldWriter` con `Raw(name, val string)` y `FieldReader`
con `Raw(name string) (string, bool)`. Este repo implementa ambas interfaces y debe
satisfacerlas.

- `jsonWriter` (encode.go) implementa `FieldWriter` → necesita `Raw()`
- `jsonReader` (parser.go) implementa `FieldReader` → `Raw()` **ya existe** en
  línea 165; solo pasa a satisfacer la interfaz extendida

## Cambios en `encode.go`

### Agregar `Raw` a `jsonWriter` (después de `Null`)

```go
func (w *jsonWriter) Raw(name, val string) {
    w.maybeComma()
    w.writeKey(name)
    if val == "" {
        w.b.WriteString("null")
        return
    }
    w.b.WriteString(val)  // emite el JSON inline, sin comillas ni escaping
}
```

### Verificar que `jsonArrayWriter` tiene `Close()` ✓

`jsonArrayWriter.Close()` ya existe en `encode.go` y escribe `]` + libera pool.
Ahora que `ArrayWriter` interface incluye `Close()`, solo necesita estar satisfecho,
no crear nada nuevo.

## Actualizar dependencia

```bash
go get github.com/tinywasm/fmt@latest
go mod tidy
```

## Verificación

```bash
go vet ./...
gotest
```

## Checklist

- [ ] `go get github.com/tinywasm/fmt@latest` actualizado
- [ ] `Raw(name, val string)` implementado en `jsonWriter` (`encode.go`)
- [ ] `jsonReader.Raw()` ya existente en `parser.go` satisface la nueva interfaz ✓
- [ ] `go vet ./...` sin errores
- [ ] `gotest` verde
