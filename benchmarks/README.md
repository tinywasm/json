# JSON Benchmarks

This directory contains benchmarking tools to compare JSON against the standard library `encoding/json` when compiled to WebAssembly using TinyGo.

## Build Script

The `build.sh` script compiles the WASM binaries with different JSON implementations.

### Usage

```bash
# Build with JSON (default)
./build.sh

# Build with encoding/json (stdlib)
./build.sh stlib
```

# Run the benchmark server

```bash
# Start the local server to serve the compiled WASM
go run ./web/server.go
```

You can then open `http://localhost:6060` in a browser to see the benchmark results.

### Output

The compiled WASM binary is output to `web/public/main.wasm`.

The script will display:
- Uncompressed binary size
- Gzipped size (for realistic deployment comparison)

## Source Files

- [`clients/json/main.go`](clients/json/main.go) - Implementation using JSON
- [`clients/stdlib/main.go`](clients/stdlib/main.go) - Implementation using encoding/json (stdlib)

Both files implement the same functionality to ensure fair comparison.

## Performance Results

Last updated: 2025-12-11 12:59:16

```
BenchmarkTinyJSON_Encode       	      10	     69990 ns/op	     930 B/op	      24 allocs/op
BenchmarkTinyJSON_Decode       	      10	     69990 ns/op	     704 B/op	      32 allocs/op
BenchmarkTinyJSON_EncodeDecode 	      10	    119987 ns/op	    1634 B/op	      56 allocs/op
BenchmarkStdlib_Encode         	      10	      9984 ns/op	     557 B/op	       9 allocs/op
BenchmarkStdlib_Decode         	      10	     49997 ns/op	     968 B/op	      27 allocs/op
BenchmarkStdlib_EncodeDecode   	      10	     60006 ns/op	    1525 B/op	      36 allocs/op
PASS
```

### Analysis

**JSON is 77% smaller** (27.2 KB vs 119 KB) making it ideal for web apps where bundle size matters. While stdlib is faster at encoding, JSON decode performance is competitive and the size advantage outweighs the microsecond differences for most browser applications.

**Use JSON when:** Bundle size is critical, slow connections, or decode-heavy workloads.  
**Use Stdlib when:** Encode-intensive operations or bundle size isn't a concern.

## Binary Size Results

| Implementation | Binary Size (WASM + Gzip) |
| :--- | :--- |
| **JSON** | **27.2 KB** |
| encoding/json (stdlib) | 119 KB |

See the [main README](../README.md#benchmarks) for detailed benchmark results and screenshots.
