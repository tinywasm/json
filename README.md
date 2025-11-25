# tinyjson

A lightweight JSON wrapper for Go that optimizes WebAssembly binary size. It automatically switches between the standard encoding/json for backends and the browser's native JSON API (via syscall/js) for WASM builds.
