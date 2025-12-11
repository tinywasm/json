package json

type codec interface {
	Encode(data any) ([]byte, error)
	Decode(data []byte, v any) error
}

type JSON struct {
	codec codec
}

func New() *JSON {

	t := &JSON{
		codec: getJSONCodec(),
	}

	return t
}

func (t *JSON) Encode(data any) ([]byte, error) {
	return t.codec.Encode(data)
}

func (t *JSON) Decode(data []byte, v any) error {
	return t.codec.Decode(data, v)
}
