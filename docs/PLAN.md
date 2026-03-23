# PLAN: JSON v3 — Field v3 Migration + Post-Decode Validation (tinywasm/json)

← [README](../README.md) | Depends on: [fmt PLAN.md](../../fmt/docs/PLAN.md)

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing`.
- **Testing Runner:** Use `gotest` (install: `go install github.com/tinywasm/devflow/cmd/gotest@latest`).
- **Max 500 lines per file.** If exceeded, subdivide by domain.
- **Flat hierarchy.** No subdirectories for library code.
- **Documentation First:** Update docs before coding.
- **Publishing:** Use `gopush 'message'` after tests pass and docs are updated.

## Prerequisite

Update `go.mod` to the new `tinywasm/fmt` version (Field v3):

```bash
go get github.com/tinywasm/fmt@v0.19.0
```

## Context

The json package currently reads `Field.JSON` to determine the JSON key and omitempty behavior via `parseJSONTag()`. With Field v3:

- `Field.JSON` is removed → JSON key is always `Field.Name`.
- `Field.OmitEmpty` replaces the `,omitempty` suffix parsing.
- `json:"-"` skip behavior → Field is simply not included in Schema by ormc.
- Post-decode validation via `fmt.Validator` interface.

---

## Stage 1: Simplify `parseJSONTag` → remove entirely

**File:** `encode.go`

### 1.1 Remove `parseJSONTag` function

```go
// DELETE this function (lines 199-223):
func parseJSONTag(f fmt.Field) (key string, omitempty bool) { ... }
```

### 1.2 Update `encodeFielder` to use Field.Name and Field.OmitEmpty

```go
// BEFORE (line 49):
key, omitempty := parseJSONTag(field)
if key == "-" {
    continue
}

// AFTER:
key := field.Name
omitempty := field.OmitEmpty
```

Remove the `key == "-"` check — fields excluded from JSON are no longer in Schema.

Full updated `encodeFielder`:

```go
func encodeFielder(b *fmt.Conv, f fmt.Fielder) error {
    schema := f.Schema()
    ptrs := f.Pointers()
    if ptrs == nil && schema != nil {
        return fmt.Err("json", "encode", "failed to get pointers")
    }
    b.WriteByte('{')

    first := true
    for i, field := range schema {
        if field.OmitEmpty && isZeroPtr(ptrs[i], field.Type) {
            continue
        }

        if !first {
            b.WriteByte(',')
        }
        first = false

        b.WriteByte('"')
        fmt.JSONEscape(field.Name, b)
        b.WriteByte('"')
        b.WriteByte(':')

        encodeFromPtr(b, ptrs[i], field.Type)
    }

    b.WriteByte('}')
    return nil
}
```

---

## Stage 2: Update decoder — key matching

**File:** `decode.go` (and `parser.go` if key lookup logic lives there)

### 2.1 Update key-to-field matching

The decoder currently matches JSON keys against `Field.JSON` (or `Field.Name` as fallback). With v3, match only against `Field.Name`.

Find wherever the decoder builds a key→index map or does key comparison. Remove the `Field.JSON` lookup path.

```go
// BEFORE (conceptual):
key := field.JSON
if key == "" {
    key = field.Name
}

// AFTER:
key := field.Name
```

---

## Stage 3: Add post-decode validation

**File:** `decode.go`

### 3.1 Call Validate() after successful decode

```go
// BEFORE:
func Decode(input any, data fmt.Fielder) error {
    // ... parse raw bytes ...
    p := parser{data: raw}
    return p.parseIntoFielder(data)
}

// AFTER:
func Decode(input any, data fmt.Fielder) error {
    // ... parse raw bytes ...
    p := parser{data: raw}
    if err := p.parseIntoFielder(data); err != nil {
        return err
    }

    // Post-decode validation
    if v, ok := data.(fmt.Validator); ok {
        return v.Validate()
    }
    return nil
}
```

**Note:** Validation runs AFTER all fields are populated, so validators can cross-reference fields.

### 3.2 Add `DecodeRaw` for decode-without-validation

Some consumers may want to decode without validation (e.g., reading partial data, migration scripts).

```go
// DecodeRaw parses JSON into a Fielder without calling Validate().
func DecodeRaw(input any, data fmt.Fielder) error {
    // ... same as current Decode ...
}
```

---

## Stage 4: Update tests

### 4.1 Update test Field literals

Remove `JSON:` from all test `fmt.Field` structs. Use `OmitEmpty: true` instead.

```go
// BEFORE:
{Name: "email", Type: fmt.FieldText, JSON: "email,omitempty"}

// AFTER:
{Name: "email", Type: fmt.FieldText, OmitEmpty: true}
```

### 4.2 Update key matching tests

Tests that verified `JSON: "custom_key"` behavior → remove. JSON key is always `Field.Name`.

### 4.3 Add post-decode validation tests

Test that `Decode()` returns validation error when Fielder implements `Validator` and data is invalid.

### 4.4 Run tests

```bash
gotest
```

---

## Stage 5: Update documentation

**File:** `README.md`

- Remove references to `Field.JSON` tag semantics.
- Document that JSON key = `Field.Name` always.
- Document `OmitEmpty` flag.
- Document post-decode validation behavior.
- Document `DecodeRaw` for validation-free decoding.

---

## Stage 6: Publish

```bash
gopush 'json: Field v3 migration — use Field.Name as key, OmitEmpty flag, post-decode Validate()'
```

---

## Summary

| Stage | File(s) | Action |
|-------|---------|--------|
| 1 | `encode.go` | Remove `parseJSONTag`, use `Field.Name` + `Field.OmitEmpty` |
| 2 | `decode.go` / `parser.go` | Match keys against `Field.Name` only |
| 3 | `decode.go` | Add post-decode `Validate()` call, add `DecodeRaw` |
| 4 | `tests/` | Update Field literals, add validation tests |
| 5 | `README.md` | Update documentation |
| 6 | — | `gotest` + `gopush` |
