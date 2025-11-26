# TinyJSON Benchmarks

This directory contains benchmarking tools to compare TinyJSON against the standard library `encoding/json` when compiled to WebAssembly using TinyGo.

## Build Script

The `build.sh` script compiles the WASM binaries with different JSON implementations.

### Usage

```bash
# Build with TinyJSON (default)
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

- [`clients/tinyjson/main.go`](clients/tinyjson/main.go) - Implementation using TinyJSON
- [`clients/stdlib/main.go`](clients/stdlib/main.go) - Implementation using encoding/json (stdlib)

Both files implement the same functionality to ensure fair comparison.

## Performance Results

Last updated: 2025-11-26 09:11:23

```
BenchmarkTinyJSON_Encode       	      10	     99994 ns/op	     952 B/op	      27 allocs/op
BenchmarkTinyJSON_Decode       	      10	     30003 ns/op	     456 B/op	      17 allocs/op
BenchmarkTinyJSON_EncodeDecode 	      10	    119987 ns/op	    1408 B/op	      44 allocs/op
BenchmarkStdlib_Encode         	      10	     19994 ns/op	     557 B/op	       9 allocs/op
BenchmarkStdlib_Decode         	      10	     50022 ns/op	     968 B/op	      27 allocs/op
BenchmarkStdlib_EncodeDecode   	      10	     69990 ns/op	    1525 B/op	      36 allocs/op
PASS
```

### Analysis

**TinyJSON is 77% smaller** (27.2 KB vs 119 KB) making it ideal for web apps where bundle size matters. While stdlib is faster at encoding, TinyJSON decode performance is competitive and the size advantage outweighs the microsecond differences for most browser applications.

**Use TinyJSON when:** Bundle size is critical, slow connections, or decode-heavy workloads.  
**Use Stdlib when:** Encode-intensive operations or bundle size isn't a concern.

## Binary Size Results

| Implementation | Binary Size (WASM + Gzip) |
| :--- | :--- |
| **TinyJSON** | **27.2 KB** |
| encoding/json (stdlib) | 119 KB |

See the [main README](../README.md#benchmarks) for detailed benchmark results and screenshots.
