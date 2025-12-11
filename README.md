# TinyJSON

A lightweight JSON wrapper for Go that optimizes WebAssembly binary size. It automatically switches between the standard encoding/json for backends and the browser's native JSON API (via syscall/js) for WASM builds.

## Usage

```go
package main

import (
    "github.com/tinywasm/json"
)

func main() {
    tj := tinyjson.New()

    // Encode data to JSON
    data := map[string]string{"message": "Hello, World!"}
    jsonBytes, err := tj.Encode(data)
    if err != nil {
        panic(err)
    }
    println(string(jsonBytes)) // {"message":"Hello, World!"}

    // Decode JSON back to data
    var result map[string]string
    err = tj.Decode(jsonBytes, &result)
    if err != nil {
        panic(err)
    }
    println(result) // map[message:Hello, World!]
}
```

## Benchmarks

Binary size comparison using TinyGo and Gzip compression:

| Implementation | Binary Size (WASM + Gzip) |
| :--- | :--- |
| **TinyJSON** | **27.2 KB** |
| encoding/json (stdlib) | 119 KB |

For build instructions and detailed benchmarking information, see [benchmarks/README.md](benchmarks/README.md).

### Screenshots

**TinyJSON (27.2 KB):**
![TinyJSON Benchmark](benchmarks/screenshots/wasm-tinyjson.png)

**Standard Library JSON (119 KB):**
![Stdlib JSON Benchmark](benchmarks/screenshots/wasm-jsonstlib.png)

---
## [Contributing](https://github.com/tinywasm/cdvelop/blob/main/CONTRIBUTING.md)

## License

See [LICENSE](LICENSE) for details.