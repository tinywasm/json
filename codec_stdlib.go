//go:build !wasm

package json

import "encoding/json"

func getJSONCodec() codec { return &stdlibJSONCodec{} }

// stdlibJSONCodec encodes Go values to JSON []byte
type stdlibJSONCodec struct{}

func (e *stdlibJSONCodec) Encode(data any) ([]byte, error) {
	return json.Marshal(data)
}

func (e *stdlibJSONCodec) Decode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
