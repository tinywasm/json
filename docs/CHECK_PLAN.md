# tinywasm/json: Eliminar map[string]any y []any del parser

## Contexto

`tinywasm/json` debe compilar con TinyGo para correr en el navegador. TinyGo tiene soporte limitado de reflection y problemas conocidos con `map[string]any` y `[]any`:

- `map[string]any` requiere reflection para acceder a valores arbitrarios
- `[]any` aloca en heap y requiere boxing
- TinyGo puede fallar en compilación o producir binarios inválidos cuando estos tipos están en el árbol de llamadas alcanzable, aunque el resultado sea descartado

### Estado actual del problema

`parser.go` tiene tres funciones que usan estos tipos:

| Función | Retorna | Callers | Resultado usado |
|---|---|---|---|
| `parseValue()` | `(any, error)` | líneas 421, 434, 311 | **Nunca** — siempre `_, err := p.parseValue()` |
| `parseObject()` | `(map[string]any, error)` | `parseValue()` únicamente | **Nunca** |
| `parseArray()` | `([]any, error)` | `parseValue()` únicamente | **Nunca** |
| `parseNumber()` | `(any, error)` | `parseValue()` únicamente | **Nunca** |

Los tres call sites que usan `parseValue()`:
- `parseIntoFielder` línea 421: campo JSON desconocido (no está en `Schema()`) → descartar
- `parseIntoFielder` línea 434: campo es `FieldStruct` pero el pointer no implementa `Fielder` → descartar
- `parseIntoPtr` línea 311: tipo de campo desconocido (fallback) → descartar

**Conclusión**: `parseValue()`, `parseObject()`, `parseArray()` y `parseNumber()` pueden ser reemplazadas por funciones `skip*` equivalentes que consumen tokens sin alocar.

---

## Stage 1 — Reemplazar parseValue() con skipValue()

**Archivo**: `parser.go`

### 1.1 — Agregar skipValue()

```go
// skipValue consumes a JSON value without allocating or returning it.
// Used to discard unknown fields and unresolvable struct pointers.
func (p *parser) skipValue() error {
	p.skipWhitespace()
	c := p.next()
	switch c {
	case '"':
		_, err := p.parseString()
		return err
	case '{':
		return p.skipObject()
	case '[':
		return p.skipArray()
	case 't', 'f':
		p.pos--
		_, err := p.parseBool()
		return err
	case 'n':
		p.pos--
		return p.parseNull()
	default:
		if (c >= '0' && c <= '9') || c == '-' {
			p.pos--
			return p.skipNumber()
		}
		return fmt.Err("json", "decode", "unexpected character")
	}
}
```

### 1.2 — Agregar skipObject()

```go
// skipObject consumes a JSON object without allocating map or keys.
func (p *parser) skipObject() error {
	p.skipWhitespace()
	if p.peek() == '}' {
		p.next()
		return nil
	}
	for {
		p.skipWhitespace()
		if p.next() != '"' {
			return fmt.Err("json", "decode", "expected quote")
		}
		if _, err := p.parseString(); err != nil {
			return err
		}
		p.skipWhitespace()
		if p.next() != ':' {
			return fmt.Err("json", "decode", "expected :")
		}
		if err := p.skipValue(); err != nil {
			return err
		}
		p.skipWhitespace()
		c := p.next()
		if c == '}' {
			return nil
		}
		if c != ',' {
			return fmt.Err("json", "decode", "expected , or }")
		}
	}
}
```

### 1.3 — Agregar skipArray()

```go
// skipArray consumes a JSON array without allocating []any.
func (p *parser) skipArray() error {
	p.skipWhitespace()
	if p.peek() == ']' {
		p.next()
		return nil
	}
	for {
		if err := p.skipValue(); err != nil {
			return err
		}
		p.skipWhitespace()
		c := p.next()
		if c == ']' {
			return nil
		}
		if c != ',' {
			return fmt.Err("json", "decode", "expected , or ]")
		}
	}
}
```

### 1.4 — Agregar skipNumber()

```go
// skipNumber consumes a JSON number without returning any value.
func (p *parser) skipNumber() error {
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if (c >= '0' && c <= '9') || c == '-' || c == '+' || c == '.' || c == 'e' || c == 'E' {
			p.pos++
		} else {
			break
		}
	}
	return nil
}
```

---

## Stage 2 — Actualizar call sites en parseIntoFielder y parseIntoPtr

Reemplazar todas las llamadas a `parseValue()` por `skipValue()`:

### 2.1 — parseIntoFielder línea 421 (campo desconocido)

```go
// Antes
if _, err := p.parseValue(); err != nil {
    return err
}

// Después
if err := p.skipValue(); err != nil {
    return err
}
```

### 2.2 — parseIntoFielder línea 434 (FieldStruct sin Fielder)

```go
// Antes
if _, err := p.parseValue(); err != nil {
    return err
}

// Después
if err := p.skipValue(); err != nil {
    return err
}
```

### 2.3 — parseIntoPtr línea 311 (tipo desconocido, fallback)

```go
// Antes
_, err := p.parseValue()
return err

// Después
return p.skipValue()
```

---

## Stage 3 — Eliminar funciones obsoletas

Una vez que ningún código llama a `parseValue()`, `parseObject()`, `parseArray()` ni `parseNumber()`, **eliminarlas**:

- Eliminar `parseValue()` (líneas 39–65)
- Eliminar `parseObject()` (líneas 457–492)
- Eliminar `parseArray()` (líneas 315–338)
- Eliminar `parseNumber()` (líneas 146–165)

---

## Stage 4 — Actualizar tests

Los tests en `tests/parser_object_test.go` y `tests/parser_array_test.go` prueban las funciones eliminadas. Deben eliminarse o reemplazarse:

- Eliminar `tests/parser_object_test.go` — probaba `parseObject()` directamente
- Eliminar `tests/parser_array_test.go` — probaba `parseArray()` directamente
- `tests/parser_unreachable_test.go` — revisar si prueba alguna función eliminada

Los tests de integración en `tests/decode_*.go` y `tests/encode_*.go` deben seguir pasando sin cambios — son los que realmente validan el comportamiento observable.

---

## Stage 5 — Verificar

```bash
gotest
```

Verificar además que no hay ninguna referencia a `map[string]any` ni `[]any` en el paquete:

```bash
grep -n "map\[string\]any\|\\[\\]any" parser.go
```

No debe producir ningún resultado.

---

## Acciones prohibidas

- NO agregar soporte público para `map[string]any` — si se necesita en el futuro, será un paquete separado explícitamente non-WASM
- NO usar `encoding/json` en ningún archivo del paquete
- NO usar `reflect` ni `unsafe`
- NO romper la API pública (`Encode`, `Decode`) — solo cambia la implementación interna
