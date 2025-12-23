package json

type codec interface {
	Encode(input any, output any) error
	Decode(input any, output any) error
}

var instance codec = getJSONCodec()

func Encode(input any, output any) error {
	return instance.Encode(input, output)
}

func Decode(input any, output any) error {
	return instance.Decode(input, output)
}
