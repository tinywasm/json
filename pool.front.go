//go:build wasm

package json

// WASM is single-threaded: allocate directly instead of sync.Pool.

func getWriter() *jsonWriter           { return &jsonWriter{} }
func putWriter(_ *jsonWriter)          {}

func getArrayWriter() *jsonArrayWriter { return &jsonArrayWriter{} }
func putArrayWriter(_ *jsonArrayWriter) {}

func getReader() *jsonReader  { return &jsonReader{} }
func putReader(_ *jsonReader) {}
