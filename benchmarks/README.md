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

You can then open `http://localhost:8080` in a browser to see the benchmark results.

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

Last updated: 2026-07-08

### Go Benchmark (`go test -bench`)

| Benchmark | tinywasm/json | encoding/json | Δ allocs |
|-----------|---------------|---------------|----------|
| Encode    | 753.2 ns/op 80 B/op 1 allocs | 758.2 ns/op 80 B/op 1 allocs | 0 |
| Decode    | 1414 ns/op 125 B/op 5 allocs | 3035 ns/op 376 B/op 8 allocs | -3 |
| RoundTrip | 2459 ns/op 221 B/op 7 allocs | 3673 ns/op 376 B/op 8 allocs | -1 |

> Run: `go test -bench=. -benchmem ./tests/...`

### WASM Binary Size

Compiled with `tinygo build -target wasm -no-debug -opt=z`.

| Implementation | Uncompressed | Gzipped |
| :--- | :--- | :--- |
| **tinywasm/json** | **51 KB** | **20 KB** |
| encoding/json (stdlib) | 270 KB | 118 KB |

> Run: `./build.sh` (tinywasm/json) and `./build.sh stlib` (stdlib) from this directory.

### Analysis

**tinywasm/json is 83% smaller gzipped** (20 KB vs 118 KB) making it ideal for web apps where bundle size matters. By eliminating the `reflect` package, it not only significantly reduces the final WASM binary size, but also makes **decoding ~2.1x faster** and **roundtripping ~1.5x faster** than the standard library.

**Use tinywasm/json when:** Bundle size and raw decoding performance are critical, or running in restricted WASM environments.
**Use Stdlib when:** You need to work dynamically with arbitrary unknown schemas (like `map[string]any`), as tinywasm/json is optimized strictly for predefined structs (`fmt.Encodable`/`fmt.Decodable`).

See the [main README](../README.md#benchmarks) for detailed benchmark results.
