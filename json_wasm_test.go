//go:build wasm

package json_test

import (
	"testing"
)

func TestWasm(t *testing.T) {
	t.Run("Encode", func(t *testing.T) { EncodeShared(t) })
	t.Run("Decode", func(t *testing.T) { DecodeShared(t) })
}
