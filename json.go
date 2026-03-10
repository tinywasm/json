package json

// Encode converts a Go value to JSON.
// input: any supported type (fmt.Fielder, primitives, known collections).
// output: *[]byte | *string | io.Writer.
func Encode(input any, output any) error {
	return encodeWithInternal(input, output)
}

// Decode parses JSON into a Go value.
// input: []byte | string | io.Reader.
// output: fmt.Fielder | *string | *int64 | *float64 | *bool | *map[string]any | *[]any.
func Decode(input any, output any) error {
	return decodeWithInternal(input, output)
}
