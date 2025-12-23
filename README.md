# JSON

A lightweight JSON wrapper for Go that optimizes WebAssembly binary size. It automatically switches between the standard encoding/json for backends and the browser's native JSON API (via syscall/js) for WASM builds.

## Usage

```go
package main

import (
    "github.com/tinywasm/json"
)

func main() {
    data := map[string]string{"message": "Hello, World!"}

    // 1. Encode to *[]byte
    var jsonBytes []byte
    if err := json.Encode(data, &jsonBytes); err != nil {
        panic(err)
    }

    // 2. Encode to io.Writer
    // var buf bytes.Buffer
    // json.Encode(data, &buf)

    // 3. Decode from []byte
    var result map[string]string
    if err := json.Decode(jsonBytes, &result); err != nil {
        panic(err)
    }

    // 4. Decode from io.Reader
    // json.Decode(bytes.NewReader(jsonBytes), &result)
}
```

## API

The API is polymorphic and avoids unnecessary allocations by accepting various input/output types.

### `Encode(input any, output any) error`

- **input**: The Go value to encode.
- **output**: The destination for the JSON output. Supported types:
    - `*[]byte`: Writes the JSON bytes to the slice.
    - `*string`: Writes the JSON string to the pointer.
    - `io.Writer`: Writes the JSON data to the writer (streaming).

### `Decode(input any, output any) error`

- **input**: The source of the JSON data. Supported types:
    - `[]byte`: Reads JSON from the byte slice.
    - `string`: Reads JSON from the string.
    - `io.Reader`: Reads JSON from the reader.
- **output**: A pointer to the Go value where the decoded data will be stored.
```

## Benchmarks

Binary size comparison using TinyGo and Gzip compression:

| Implementation | Binary Size (WASM + Gzip) |
| :--- | :--- |
| **JSON** | **27.2 KB** |
| encoding/json (stdlib) | 119 KB |

For build instructions and detailed benchmarking information, see [benchmarks/README.md](benchmarks/README.md).

### Screenshots

**JSON (27.2 KB):**
![JSON Benchmark](benchmarks/screenshots/wasm-json.png)

**Standard Library JSON (119 KB):**
![Stdlib JSON Benchmark](benchmarks/screenshots/wasm-jsonstlib.png)

---
## [Contributing](https://github.com/tinywasm/cdvelop/blob/main/CONTRIBUTING.md)

## License

See [LICENSE](LICENSE) for details.