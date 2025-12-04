//go:build wasm

package tinyjson

import (
	"reflect"
	"syscall/js"

	. "github.com/cdvelop/tinystring"
)

func getJSONCodec() codec {
	return &wasmJSONCodec{
		jsJSON:   js.Global().Get("JSON"),
		jsObject: js.Global().Get("Object"),
		jsArray:  js.Global().Get("Array"),
	}
}

// wasmJSONCodec uses browser's JSON.stringify via JavaScript API
type wasmJSONCodec struct {
	jsJSON   js.Value
	jsObject js.Value
	jsArray  js.Value
}

func (j *wasmJSONCodec) Encode(data any) ([]byte, error) {
	// Convert Go value to JavaScript value
	jsValue := ConvertGoToJS(data)

	// Use browser's JSON.stringify
	jsonString := j.jsJSON.Call("stringify", jsValue).String()

	return []byte(jsonString), nil
}

func (j *wasmJSONCodec) Decode(data []byte, v any) error {
	// Use browser's JSON.parse
	jsValue := j.jsJSON.Call("parse", string(data))

	// Convert JavaScript value to Go value
	return ConvertJSToGo(jsValue, v)
}

// ConvertJSToGo converts JavaScript values to Go values
func ConvertJSToGo(jsVal js.Value, v any) error {
	// Basic implementation - extend as needed
	switch ptr := v.(type) {
	case *map[string]any:
		*ptr = make(map[string]any)
		keys := js.Global().Get("Object").Call("keys", jsVal)
		length := keys.Length()
		for i := 0; i < length; i++ {
			key := keys.Index(i).String()
			(*ptr)[key] = ConvertJSValueToGo(jsVal.Get(key))
		}
	case *[]any:
		length := jsVal.Length()
		*ptr = make([]any, length)
		for i := 0; i < length; i++ {
			(*ptr)[i] = ConvertJSValueToGo(jsVal.Index(i))
		}
	case *string:
		*ptr = jsVal.String()
	case *int:
		*ptr = jsVal.Int()
	case *float64:
		*ptr = jsVal.Float()
	case *bool:
		*ptr = jsVal.Bool()
	default:
		// Use reflection to handle pointers to slices and structs
		val := reflect.ValueOf(v)
		if val.Kind() != reflect.Ptr || val.IsNil() {
			return nil // Or error
		}
		elem := val.Elem()

		// Handle slices of any type
		if elem.Kind() == reflect.Slice {
			if !jsVal.InstanceOf(js.Global().Get("Array")) {
				return nil // Not an array
			}

			length := jsVal.Length()
			sliceType := elem.Type()
			elemType := sliceType.Elem()

			// Create new slice with correct capacity
			newSlice := reflect.MakeSlice(sliceType, length, length)

			for i := 0; i < length; i++ {
				jsItem := jsVal.Index(i)

				// Create new element of the correct type
				newElem := reflect.New(elemType)

				// Recursively decode into the new element
				if err := ConvertJSToGo(jsItem, newElem.Interface()); err != nil {
					return err
				}

				// Set the decoded value into the slice
				newSlice.Index(i).Set(newElem.Elem())
			}

			// Set the new slice to the target
			elem.Set(newSlice)
			return nil
		}

		// Handle structs
		if elem.Kind() == reflect.Struct {
			typ := elem.Type()
			for i := 0; i < elem.NumField(); i++ {
				field := typ.Field(i)
				jsonTag := field.Tag.Get("json")
				if jsonTag == "" {
					jsonTag = field.Name
				}
				if jsonTag == "-" {
					continue
				}
				if idx := Index(jsonTag, ","); idx != -1 {
					jsonTag = jsonTag[:idx]
				}

				jsField := jsVal.Get(jsonTag)
				if !jsField.IsUndefined() && !jsField.IsNull() {
					// Recursively decode
					fieldVal := elem.Field(i)
					// We need to pass a pointer to the field to ConvertJSToGo if possible,
					// or handle setting the value directly.
					// Since ConvertJSToGo takes 'any' (interface{}), we can pass a pointer to the field.
					if fieldVal.CanAddr() {
						ConvertJSToGo(jsField, fieldVal.Addr().Interface())
					}
				}
			}
		}
	}
	return nil
}

// ConvertJSValueToGo converts a js.Value to a Go value recursively
// eg: js.Value representing { "key": [1, 2, 3], "flag": true } becomes map[string]any{ "key": []any{1, 2, 3}, "flag": true }
func ConvertJSValueToGo(jsVal js.Value) any {
	switch jsVal.Type() {
	case js.TypeNull, js.TypeUndefined:
		return nil
	case js.TypeBoolean:
		return jsVal.Bool()
	case js.TypeNumber:
		return jsVal.Float()
	case js.TypeString:
		return jsVal.String()
	case js.TypeObject:
		if jsVal.InstanceOf(js.Global().Get("Array")) {
			length := jsVal.Length()
			arr := make([]any, length)
			for i := 0; i < length; i++ {
				arr[i] = ConvertJSValueToGo(jsVal.Index(i))
			}
			return arr
		}
		obj := make(map[string]any)
		keys := js.Global().Get("Object").Call("keys", jsVal)
		length := keys.Length()
		for i := 0; i < length; i++ {
			key := keys.Index(i).String()
			obj[key] = ConvertJSValueToGo(jsVal.Get(key))
		}
		return obj
	default:
		return jsVal.String()
	}
}

// ConvertGoToJS converts Go values to JavaScript values recursively
func ConvertGoToJS(data any) js.Value {
	if data == nil {
		return js.Null()
	}

	switch v := data.(type) {
	case string:
		return js.ValueOf(v)
	case bool:
		return js.ValueOf(v)
	case int:
		return js.ValueOf(v)
	case int8:
		return js.ValueOf(int(v))
	case int16:
		return js.ValueOf(int(v))
	case int32:
		return js.ValueOf(int(v))
	case int64:
		return js.ValueOf(int(v))
	case uint:
		return js.ValueOf(int(v))
	case uint8:
		return js.ValueOf(int(v))
	case uint16:
		return js.ValueOf(int(v))
	case uint32:
		return js.ValueOf(int(v))
	case uint64:
		return js.ValueOf(int(v))
	case float32:
		return js.ValueOf(float64(v))
	case float64:
		return js.ValueOf(v)
	case []byte:
		return js.ValueOf(string(v))
	case []any:
		arr := js.Global().Get("Array").New(len(v))
		for i, item := range v {
			arr.SetIndex(i, ConvertGoToJS(item))
		}
		return arr
	case map[string]any:
		obj := js.Global().Get("Object").New()
		for key, val := range v {
			obj.Set(key, ConvertGoToJS(val))
		}
		return obj
	case map[string]string:
		obj := js.Global().Get("Object").New()
		for key, val := range v {
			obj.Set(key, js.ValueOf(val))
		}
		return obj
	case map[string]int:
		obj := js.Global().Get("Object").New()
		for key, val := range v {
			obj.Set(key, js.ValueOf(val))
		}
		return obj
	case []string:
		arr := js.Global().Get("Array").New(len(v))
		for i, item := range v {
			arr.SetIndex(i, js.ValueOf(item))
		}
		return arr
	case []int:
		arr := js.Global().Get("Array").New(len(v))
		for i, item := range v {
			arr.SetIndex(i, js.ValueOf(item))
		}
		return arr
	default:
		// Use reflection to handle structs and slices
		val := reflect.ValueOf(data)

		// Handle slices of any type using reflection
		if val.Kind() == reflect.Slice {
			arr := js.Global().Get("Array").New(val.Len())
			for i := 0; i < val.Len(); i++ {
				arr.SetIndex(i, ConvertGoToJS(val.Index(i).Interface()))
			}
			return arr
		}

		// Handle structs
		if val.Kind() == reflect.Struct {
			obj := js.Global().Get("Object").New()
			typ := val.Type()
			for i := 0; i < val.NumField(); i++ {
				field := typ.Field(i)
				jsonTag := field.Tag.Get("json")
				if jsonTag == "" {
					jsonTag = field.Name
				}
				// Handle "omitempty" and other tag options if needed, but for now basic name support
				if jsonTag == "-" {
					continue
				}

				// Simple tag parsing to get the name
				if idx := Index(jsonTag, ","); idx != -1 {
					jsonTag = jsonTag[:idx]
				}

				obj.Set(jsonTag, ConvertGoToJS(val.Field(i).Interface()))
			}
			return obj
		}

		// For other types, try to convert to string using tinystring
		return js.ValueOf(Convert(v).String())
	}
}
