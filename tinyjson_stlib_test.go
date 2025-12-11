//go:build !wasm

package json

import (
	"testing"

	"github.com/tinywasm/json"
)

func TestStdlib(t *testing.T) {
	j := json.New()

	t.Run("Encode", func(t *testing.T) { EncodeShared(t, j) })
	t.Run("Decode", func(t *testing.T) { DecodeShared(t, j) })
}
