# PLAN: Decode Zero-Alloc — Corrección de Performance (tinywasm/json)

← [README](../README.md) | Depends on: [fmt PLAN.md](../../fmt/docs/PLAN.md)

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing`.
- **Testing Runner:** Use `gotest` (install: `go install github.com/tinywasm/devflow/cmd/gotest@latest`).
- **Max 500 lines per file.** If exceeded, subdivide by domain.
- **Flat hierarchy.** No subdirectories for library code.
- **TinyGo Compatible:** No `fmt`, `strings`, `strconv`, `errors` from stdlib. Use `tinywasm/fmt`.
- **No maps** in WASM code (binary bloat).
- **Documentation First:** Update docs before coding.
- **Publishing:** Use `gopush 'message'` after tests pass and docs are updated.

## Prerequisite

Update `go.mod` to the new `tinywasm/fmt` version with `LoadBytes` + `GetStringZeroCopy` fix:

```bash
go get github.com/tinywasm/fmt@latest
```

## Context

Encode está optimizado (2 allocs, +1 vs stdlib). Decode tiene **20 allocs** vs 8 de stdlib.

### Diagnóstico con profiler (`go tool pprof -top -cum`)

| Fuente | % memoria | Causa raíz |
|--------|-----------|------------|
| `fmt.GetConv` (pool New) | **71.4%** | `Convert(s).Int64()` — variadic boxing + Conv leak del pool |
| `parseValue` boxing | 4.1% | Valores retornados como `any` y desboxeados en `writeValue` |
| `parseNumber` string copy | 5.6% | `string(p.data[start:p.pos])` innecesario |

### Target

| Operación | Actual | Target | stdlib |
|-----------|--------|--------|--------|
| Decode | 20 allocs | **3 allocs** | 8 allocs |

---

## Stage 1: Replace `parseNumber` calls with `fmt.LoadBytes` path

**File:** `parser.go`

Add a method that parses a number from the current position using `fmt.LoadBytes` — reuses all existing `parseIntBase`/`parseFloatBase` logic in fmt with **0 allocations**.

```go
// parseNumberInto parses a JSON number directly into a typed pointer.
// Uses fmt.GetConv + LoadBytes + Int64/Float64 + PutConv — 0 allocations
// (reuses fmt's existing parseIntBase/parseFloatBase via pool).
func (p *parser) parseNumberInto(ptr any, ft fmt.FieldType) error {
	p.skipWhitespace()
	start := p.pos
	isFloat := false
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if (c >= '0' && c <= '9') || c == '-' || c == '+' {
			p.pos++
		} else if c == '.' || c == 'e' || c == 'E' {
			isFloat = true
			p.pos++
		} else {
			break
		}
	}
	if p.pos == start {
		return fmt.Err("json", "decode", "expected number")
	}

	numBytes := p.data[start:p.pos]

	c := fmt.GetConv()
	c.LoadBytes(numBytes)

	if ft == fmt.FieldFloat || isFloat {
		v, err := c.Float64()
		c.PutConv()
		if err != nil {
			return err
		}
		if ft == fmt.FieldFloat {
			switch fp := ptr.(type) {
			case *float64:
				*fp = v
			case *float32:
				*fp = float32(v)
			}
		} else {
			// FieldInt but JSON has decimal/exponent — truncate
			switch ip := ptr.(type) {
			case *int64:
				*ip = int64(v)
			case *int:
				*ip = int(v)
			case *int32:
				*ip = int32(v)
			}
		}
	} else {
		v, err := c.Int64()
		c.PutConv()
		if err != nil {
			return err
		}
		switch ip := ptr.(type) {
		case *int64:
			*ip = v
		case *int:
			*ip = int(v)
		case *int32:
			*ip = int32(v)
		}
	}
	return nil
}
```

**Alloc analysis:**
- `fmt.GetConv()` → from pool, 0 allocs (steady state)
- `c.LoadBytes(numBytes)` → copies to existing 64-byte buffer, 0 allocs
- `c.Int64()`/`c.Float64()` → uses `GetStringZeroCopy` internally (fmt fix), 0 allocs
- `c.PutConv()` → returns to pool, 0 allocs
- **Total: 0 allocs per number field**

---

## Stage 2: Add `parseIntoPtr` — bypass `parseValue`/`writeValue` boxing

**File:** `parser.go`

```go
// parseIntoPtr parses a JSON value directly into a typed pointer.
// Bypasses parseValue()/writeValue() to avoid boxing values into any.
func (p *parser) parseIntoPtr(ptr any, ft fmt.FieldType) error {
	p.skipWhitespace()

	// Handle JSON null for any type
	if p.peek() == 'n' {
		return p.parseNull()
	}

	switch ft {
	case fmt.FieldText:
		if p.next() != '"' {
			return fmt.Err("json", "decode", "expected string")
		}
		s, err := p.parseString()
		if err != nil {
			return err
		}
		if sp, ok := ptr.(*string); ok {
			*sp = s
		}
		return nil

	case fmt.FieldInt, fmt.FieldFloat:
		return p.parseNumberInto(ptr, ft)

	case fmt.FieldBool:
		b, err := p.parseBool()
		if err != nil {
			return err
		}
		if bp, ok := ptr.(*bool); ok {
			*bp = b
		}
		return nil

	case fmt.FieldBlob:
		if p.next() != '"' {
			return fmt.Err("json", "decode", "expected string for blob")
		}
		s, err := p.parseString()
		if err != nil {
			return err
		}
		if bp, ok := ptr.(*[]byte); ok {
			*bp = []byte(s)
		}
		return nil
	}

	// Fallback for unknown field types
	_, err := p.parseValue()
	return err
}
```

---

## Stage 3: Wire `parseIntoPtr` into `parseIntoFielder`

**File:** `parser.go`

Replace the generic `parseValue()` + `writeValue()` path in `parseIntoFielder`:

```go
// BEFORE (current code, lines ~315-321):
		} else {
			val, err := p.parseValue()
			if err != nil {
				return err
			}
			writeValue(ptr, field.Type, val)
		}

// AFTER:
		} else {
			if err := p.parseIntoPtr(ptr, field.Type); err != nil {
				return err
			}
		}
```

**Note:** `parseValue()`, `writeValue()`, and `parseNumber()` remain unchanged — they're still used by:
- `parseObject()` (generic `map[string]any` parsing)
- `parseArray()` (generic `[]any` parsing)
- Unknown field discard in `parseIntoFielder` (line 298)

---

## Stage 4: Stack-allocate parser

**File:** `decode.go`

```go
// BEFORE (line 36):
	p := &parser{data: raw}
	return p.parseIntoFielder(data)

// AFTER:
	p := parser{data: raw}
	return p.parseIntoFielder(data)
```

Parser methods use `*parser` receiver but never store the pointer externally.
Escape analysis should keep `p` on the stack (saves 1 alloc).

Verify:
```bash
go build -gcflags='-m' 2>&1 | grep 'parser.*escapes'
```

---

## Stage 5: Run tests and benchmarks

```bash
gotest
```

Then:

```bash
go test -bench=. -benchmem -count=5 ./tests/... 2>&1 | tee /tmp/bench_fix.txt
```

Expected results:

| Benchmark | Before | After |
|-----------|--------|-------|
| Encode | 2 allocs | 2 allocs (unchanged) |
| Decode | 20 allocs | **3 allocs** |

Verify with profiler:
```bash
go test -bench=BenchmarkDecode_tinywasm -memprofile=/tmp/fix_mem.prof -run=^$ ./tests/...
go tool pprof -top /tmp/fix_mem.prof
```

`fmt.GetConv` / `fmt.Convert` / `fmt.init.func1` should **NOT** appear in the profile.

---

## Stage 6: Update `benchmarks/README.md`

Replace the Performance Results table with the new numbers.
Update "Last updated" date.

---

## Stage 7: Publish

```bash
gopush 'json: zero-alloc decode — use fmt.LoadBytes for numbers, direct pointer writes'
```

---

## Alloc Breakdown (target)

For `{"name":"Alice","email":"alice@example.com","age":30,"score":9.5}`:

| Step | Allocs | Detail |
|------|--------|--------|
| `parser{}` on stack | 0 | escape analysis |
| `f.Schema()` | 0 | package-level var |
| `f.Pointers()` | **1** | `[]any` slice (inevitable) |
| key matching × 4 | 0 | `matchFieldIndex` byte comparison |
| "Alice" string | **1** | `string(data[start:end])` copy (inevitable) |
| "alice@example.com" | **1** | `string(data[start:end])` copy (inevitable) |
| age=30 | 0 | `fmt.GetConv + LoadBytes + Int64 + PutConv` (pool) |
| score=9.5 | 0 | `fmt.GetConv + LoadBytes + Float64 + PutConv` (pool) |
| **Total** | **3** | **vs 20 before, vs 8 stdlib** |

## Summary

| Stage | File(s) | Change |
|-------|---------|--------|
| 1 | `parser.go` | Add `parseNumberInto` — uses `fmt.LoadBytes` (0 allocs) |
| 2 | `parser.go` | Add `parseIntoPtr` — direct pointer writes |
| 3 | `parser.go` | Wire into `parseIntoFielder` |
| 4 | `decode.go` | Stack-allocate parser |
| 5 | tests | Verify 3 allocs target |
| 6 | docs | Update benchmarks README |
| 7 | — | `gopush` |

**Código nuevo en json:** ~80 líneas (`parseNumberInto` + `parseIntoPtr`).
**Código duplicado de fmt:** 0. Toda la lógica de parsing numérico se reutiliza via `LoadBytes + Int64/Float64`.
