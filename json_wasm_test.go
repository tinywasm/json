//go:build wasm

package json_test

import (
	"testing"

	"github.com/tinywasm/json"
)

func TestWasm(t *testing.T) {
	j := json.New()

	t.Run("Encode", func(t *testing.T) { EncodeShared(t, j) })
	t.Run("Decode", func(t *testing.T) { DecodeShared(t, j) })
}
