//go:build !wasm

package json_test

import (
	"testing"
)

func TestStdlib(t *testing.T) {
	t.Run("Encode", func(t *testing.T) { EncodeShared(t) })
	t.Run("Decode", func(t *testing.T) { DecodeShared(t) })
}
