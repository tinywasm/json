package json

type codec interface {
	Encode(data any) ([]byte, error)
	Decode(data []byte, v any) error
}

type TinyJSON struct {
	codec codec
}

func New() *TinyJSON {

	t := &TinyJSON{
		codec: getJSONCodec(),
	}

	return t
}

func (t *TinyJSON) Encode(data any) ([]byte, error) {
	return t.codec.Encode(data)
}

func (t *TinyJSON) Decode(data []byte, v any) error {
	return t.codec.Decode(data, v)
}
