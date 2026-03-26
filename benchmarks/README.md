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

- [`clients/tinyjson/main.go`](clients/tinyjson/main.go) - Implementation using JSON
- [`clients/stdlib/main.go`](clients/stdlib/main.go) - Implementation using encoding/json (stdlib)

Both files implement the same functionality to ensure fair comparison.

## Performance Results

Last updated: 2026-03-26

### Go Benchmark (`go test -bench`)

| Benchmark | tinywasm/json | encoding/json | Δ allocs |
|-----------|---------------|---------------|----------|
| Encode    | 268 ns/op 144 B/op 2 allocs | 253 ns/op 80 B/op 1 allocs | +1 |
| Decode    | 308 ns/op 157 B/op 5 allocs | 995 ns/op 376 B/op 8 allocs | -3 |
| RoundTrip | 632 ns/op 317 B/op 8 allocs | 1215 ns/op 376 B/op 8 allocs | 0 |

> Run: `go test -bench=. -benchmem ./tests/...`

### WASM Binary Size

| Implementation | Binary Size (WASM + Gzip) |
| :--- | :--- |
| **tinywasm/json** | **~27 KB** |
| encoding/json (stdlib) | ~119 KB |

### Analysis

**tinywasm/json is 77% smaller** (~27 KB vs ~119 KB) making it ideal for web apps where bundle size matters. By eliminating the `reflect` package, it not only significantly reduces the final WASM binary size, but also makes **decoding ~3.2x faster** and **roundtripping ~1.9x faster** than the standard library.

**Use tinywasm/json when:** Bundle size and raw decoding performance are critical, or running in restricted WASM environments.
**Use Stdlib when:** You need to work dynamically with arbitrary unknown schemas (like `map[string]any`), as tinywasm/json is optimized strictly for predefined structs (`fmt.Fielder`).

See the [main README](../README.md#benchmarks) for detailed benchmark results.
