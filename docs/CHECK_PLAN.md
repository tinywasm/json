# PLAN: Support root-level JSON array encoding/decoding

## Problem

`json.Encode` and `json.Decode` only handle `fmt.Fielder` (single struct → `{}`).
Encoding a list as a root JSON array `[...]` is not supported.

## Solution — internal type switch, no API change

After `tinywasm/fmt` makes `FielderSlice` embed `Fielder`, any list type satisfies
`fmt.Fielder`. `Encode` and `Decode` remain **unchanged in signature**. A type
assertion inside each function routes to array or object encoding transparently.

```go
// Signatures stay exactly the same
func Encode(data fmt.Fielder, output any) error
func Decode(input any, data fmt.Fielder) error
```

---

## Change to `encode.go`

Add private `encodeSlice`, then check for `FielderSlice` at the top of `Encode`
before falling through to `encodeFielder`:

```go
func Encode(data fmt.Fielder, output any) error {
	b := fmt.GetConv()
	defer b.PutConv()

	if slice, ok := data.(fmt.FielderSlice); ok {
		encodeSlice(b, slice)
	} else {
		if err := encodeFielder(b, data); err != nil {
			return err
		}
	}

	if b.GetString(fmt.BuffErr) != "" {
		return fmt.Err(b.GetString(fmt.BuffErr))
	}
	// ... existing output switch unchanged ...
}

func encodeSlice(b *fmt.Conv, s fmt.FielderSlice) {
	b.WriteByte('[')
	for i := 0; i < s.Len(); i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		if err := encodeFielder(b, s.At(i)); err != nil {
			b.WrString(fmt.BuffErr, err.Error())
		}
	}
	b.WriteByte(']')
}
```

---

## Change to `decode.go`

Check for `FielderSlice` at the top of `Decode` before calling `parseIntoFielder`:

```go
func Decode(input any, data fmt.Fielder) error {
	// ... existing raw bytes extraction unchanged ...

	p := parser{data: raw}
	if slice, ok := data.(fmt.FielderSlice); ok {
		return p.parseArray(slice)
	}
	return p.parseIntoFielder(data)
}
```

Add `parseArray` to `parser.go`. Based on the `FieldStructSlice` branch in
`parseValue` (parser.go:422-451) but with two differences: called at the root
(no field-type dispatch), and adds `p.skipWhitespace()` before the `[` check
to tolerate leading whitespace in root-level input:

```go
func (p *parser) parseArray(fs fmt.FielderSlice) error {
	p.skipWhitespace()
	if p.peek() != '[' {
		return fmt.Err("json", "decode", "expected array")
	}
	p.next()
	p.skipWhitespace()
	if p.peek() == ']' {
		p.next()
		return nil
	}
	for {
		nested := fs.Append()
		if err := p.parseIntoFielder(nested); err != nil {
			return err
		}
		p.skipWhitespace()
		c := p.next()
		if c == ']' {
			break
		}
		if c != ',' {
			return fmt.Err("json", "decode", "expected , or ]")
		}
		p.skipWhitespace()
	}
	return nil
}
```

---

## Tests

> **`mockFielderSlice` will not compile until the stubs are added.** `fmt.FielderSlice`
> already embeds `fmt.Fielder` — add these methods first, before running any test.

Extend `tests/struct_slice_test.go` — add `Schema()`/`Pointers()` stubs to
`mockFielderSlice`:

```go
func (s *mockFielderSlice) Schema() []fmt.Field { return nil }
func (s *mockFielderSlice) Pointers() []any     { return nil }
```

Add two new test cases in the same file:

```go
func TestEncode_RootSlice(t *testing.T) {
	slice := &mockFielderSlice{items: []fmt.Fielder{
		&item{ID: 1, Name: "Alice"},
		&item{ID: 2, Name: "Bob"},
	}}
	var out string
	if err := json.Encode(slice, &out); err != nil {
		t.Fatal(err)
	}
	want := `[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`
	if out != want {
		t.Errorf("got %s", out)
	}
}

func TestDecode_RootSlice(t *testing.T) {
	slice := &mockFielderSlice{}
	if err := json.Decode(`[{"id":1,"name":"Alice"}]`, slice); err != nil {
		t.Fatal(err)
	}
	if slice.Len() != 1 || slice.items[0].(*item).Name != "Alice" {
		t.Error("decode mismatch")
	}
}
```

---

## Dependency

Requires `tinywasm/fmt` with `FielderSlice` embedding `Fielder` (see fmt/docs/PLAN.md).
Update `go.mod` to the new fmt version before implementing.

---

## Checklist

- [ ] Add `Schema()`/`Pointers()` stubs to `mockFielderSlice` in tests (**do this first — breaks compilation**)
- [ ] Bump `tinywasm/fmt` in `go.mod` to version with `FielderSlice` embedding `Fielder`
- [ ] Add `encodeSlice` (with `BuffErr` error handling) and type-switch in `Encode` (`encode.go`)
- [ ] Add `parseArray` to `parser.go` and type-switch in `Decode` (`decode.go`)
- [ ] Add stubs to `mockFielderSlice` in tests; add two new test cases
- [ ] `go test ./...` passes
- [ ] Bump minor version
