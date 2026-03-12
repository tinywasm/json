# Bug Report: tinywasm/fmt - Scientific Notation Conversion

## Description
The `tinywasm/fmt` library's `fmt.Convert(s).Float64()` method fails to correctly parse strings in scientific notation (e.g., "1e2"). It returns "character invalid" error.

## Steps to Reproduce
1. In `github.com/tinywasm/json`, run the following test:
   ```go
   func TestParseNumberScientific(t *testing.T) {
       var f float64
       m := &mockFielder{
           schema: []fmt.Field{{Name: "F", Type: fmt.FieldFloat, JSON: "f"}},
           pointers: []any{&f},
       }
       input := `{"f":1e2}`
       if err := json.Decode(input, m); err != nil {
           t.Fatal(err)
       }
       if f != 100.0 {
           t.Errorf("expected 100.0, got %f", f)
       }
   }
   ```
2. The `json.Decode` call fails because `parser.go` identifies it as a float and calls `fmt.Convert(s).Float64()`, which returns an error.

## Observations
- `parser.go` correctly extracts the string "1e2" because its `parseNumber` method includes 'e', 'E', and '+' in its character set.
- The error "character invalid" is returned by the underlying conversion logic in `tinywasm/fmt`.

## Impact
JSON numbers using scientific notation cannot be decoded into float fields.
