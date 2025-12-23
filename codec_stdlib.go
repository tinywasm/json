//go:build !wasm

package json

import (
	"encoding/json"
	"errors"
	"io"
)

func getJSONCodec() codec { return &stdlibJSONCodec{} }

// stdlibJSONCodec encodes Go values to JSON []byte
type stdlibJSONCodec struct{}

func (e *stdlibJSONCodec) Encode(input any, output any) error {
	switch out := output.(type) {
	case io.Writer:
		return json.NewEncoder(out).Encode(input)
	case *[]byte:
		b, err := json.Marshal(input)
		if err != nil {
			return err
		}
		*out = b
		return nil
	case *string:
		b, err := json.Marshal(input)
		if err != nil {
			return err
		}
		*out = string(b)
		return nil
	default:
		return errors.New("json: unsupported output type")
	}
}

func (e *stdlibJSONCodec) Decode(input any, output any) error {
	switch in := input.(type) {
	case io.Reader:
		return json.NewDecoder(in).Decode(output)
	case []byte:
		return json.Unmarshal(in, output)
	case string:
		return json.Unmarshal([]byte(in), output)
	default:
		return errors.New("json: unsupported input type")
	}
}
