# Dependency Refactor: Adopt fmt.FielderSlice

## 1. Objective
Refactor the `json` module's implementation of encoding and parsing nested struct slices. 
This involves removing the locally defined technical debt interface `FielderSlice` and switching exclusively to the standardized `fmt.FielderSlice` API.

**Dependency prerequisite**: The `fmt` package must already have `fmt.FielderSlice` implemented. 

## 2. Implementation Steps

### 2.1 Remove Local Interface
**File**: `encode.go`
1. Locate and completely delete the `FielderSlice` interface declaration (lines 8-14):
```go
// FielderSlice is implemented by generated code to allow
// iteration over a slice of structs without reflection.
type FielderSlice interface {
	Len() int
	At(i int) fmt.Fielder
	Append() fmt.Fielder
}
```

### 2.2 Re-wire Encoder
**File**: `encode.go`
1. Locate the type assertion inside the switch block handling `fmt.FieldStructSlice` (around line 160).
2. Change the assertion type:
```diff
- if p, ok := ptr.(FielderSlice); ok {
+ if p, ok := ptr.(fmt.FielderSlice); ok {
```

### 2.3 Re-wire Parser
**File**: `parser.go`
1. Update `parseIntoPtr`: Locate the type assertion handling `fmt.FieldStructSlice` (around line 428).
```diff
- fs, ok := ptr.(FielderSlice)
+ fs, ok := ptr.(fmt.FielderSlice)
```
2. Update `parseIntoFielder`: Locate the type assertion handling `fmt.FieldStructSlice` (around line 557).
```diff
- if nested, ok := ptr.(FielderSlice); ok {
+ if nested, ok := ptr.(fmt.FielderSlice); ok {
```

## 3. Post-Refactor Verification

### 3.1 Cleanup Check
Verify that no residual usages of the old package-local `FielderSlice` remain across `json` logic.
```bash
grep -rn "ptr.(FielderSlice)" . --include="*.go"
# Should yield nothing
```

### 3.2 Testing
Run the comprehensive test suite for the `json` module. All deserialization scenarios for `FieldStructSlice` must continue parsing natively through the new `fmt` contract without compilation issues or panics.
```bash
go test ./...
```
