//go:build wasm

package json

import (
	"io"
	"reflect"
	"syscall/js"

	. "github.com/tinywasm/fmt"
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

func (j *wasmJSONCodec) Encode(input any, output any) error {
	// Convert Go value to JavaScript value
	jsValue := j.convertGoToJS(input)

	// Use browser's JSON.stringify
	jsonString := j.jsJSON.Call("stringify", jsValue).String()

	switch out := output.(type) {
	case *[]byte:
		*out = []byte(jsonString)
	case *string:
		*out = jsonString
	case io.Writer:
		_, err := out.Write([]byte(jsonString))
		return err
	default:
		return Err("json: unsupported output type")
	}

	return nil
}

func (j *wasmJSONCodec) Decode(input any, output any) error {
	var jsonStr string

	switch in := input.(type) {
	case string:
		jsonStr = in
	case []byte:
		jsonStr = string(in)
	case io.Reader:
		// WASM currently reads all into memory for JSON.parse
		b, err := io.ReadAll(in)
		if err != nil {
			return err
		}
		jsonStr = string(b)
	default:
		return Err("json: unsupported input type")
	}

	// Use browser's JSON.parse
	jsValue := j.jsJSON.Call("parse", jsonStr)

	// Convert JavaScript value to Go value
	return j.convertJSToGo(jsValue, output)
}

// convertJSToGo converts JavaScript values to Go values
func (j *wasmJSONCodec) convertJSToGo(jsVal js.Value, v any) error {
	// Basic implementation - extend as needed
	switch ptr := v.(type) {
	case *map[string]any:
		*ptr = make(map[string]any)
		keys := j.jsObject.Call("keys", jsVal)
		length := keys.Length()
		for i := 0; i < length; i++ {
			key := keys.Index(i).String()
			(*ptr)[key] = j.convertJSValueToGo(jsVal.Get(key))
		}
	case *[]any:
		length := jsVal.Length()
		*ptr = make([]any, length)
		for i := 0; i < length; i++ {
			(*ptr)[i] = j.convertJSValueToGo(jsVal.Index(i))
		}
	case *string:
		*ptr = jsVal.String()
	case *int:
		*ptr = jsVal.Int()
	case *float64:
		*ptr = jsVal.Float()
	case *bool:
		*ptr = jsVal.Bool()
	case *byte: // byte is alias for uint8
		*ptr = byte(jsVal.Int())
	case *[]byte:
		// []byte is encoded as string, so decode from string
		if jsVal.Type() == js.TypeString {
			*ptr = []byte(jsVal.String())
		}
	default:
		// Use reflection to handle pointers to slices and structs
		val := reflect.ValueOf(v)
		if val.Kind() != reflect.Ptr || val.IsNil() {
			return nil // Or error
		}
		elem := val.Elem()

		// Handle slices of any type
		if elem.Kind() == reflect.Slice {
			sliceType := elem.Type()
			elemType := sliceType.Elem()

			// Special case: []byte is encoded as string, not array
			// Check if this is a []byte (slice of uint8)
			if elemType.Kind() == reflect.Uint8 {
				if jsVal.Type() == js.TypeString {
					elem.SetBytes([]byte(jsVal.String()))
					return nil
				}
			}

			if !jsVal.InstanceOf(j.jsArray) {
				return nil // Not an array
			}

			length := jsVal.Length()

			// Create new slice with correct capacity
			newSlice := reflect.MakeSlice(sliceType, length, length)

			for i := 0; i < length; i++ {
				jsItem := jsVal.Index(i)

				// Special case for [][]byte: inner []byte is encoded as string
				if elemType.Kind() == reflect.Slice && elemType.Elem().Kind() == reflect.Uint8 {
					if jsItem.Type() == js.TypeString {
						newSlice.Index(i).SetBytes([]byte(jsItem.String()))
						continue
					}
				}

				// Create new element of the correct type
				newElem := reflect.New(elemType)

				// Recursively decode into the new element
				if err := j.convertJSToGo(jsItem, newElem.Interface()); err != nil {
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
			// Check if jsVal is actually an object (not string, number, etc.)
			if jsVal.Type() != js.TypeObject {
				return nil // Cannot decode non-object into struct
			}

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
					// We need to pass a pointer to the field to convertJSToGo if possible,
					// or handle setting the value directly.
					// Since convertJSToGo takes 'any' (interface{}), we can pass a pointer to the field.
					if fieldVal.CanAddr() {
						j.convertJSToGo(jsField, fieldVal.Addr().Interface())
					}
				}
			}
		}
	}
	return nil
}

// convertJSValueToGo converts a js.Value to a Go value recursively
// eg: js.Value representing { "key": [1, 2, 3], "flag": true } becomes map[string]any{ "key": []any{1, 2, 3}, "flag": true }
func (j *wasmJSONCodec) convertJSValueToGo(jsVal js.Value) any {
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
		if jsVal.InstanceOf(j.jsArray) {
			length := jsVal.Length()
			arr := make([]any, length)
			for i := 0; i < length; i++ {
				arr[i] = j.convertJSValueToGo(jsVal.Index(i))
			}
			return arr
		}
		obj := make(map[string]any)
		keys := j.jsObject.Call("keys", jsVal)
		length := keys.Length()
		for i := 0; i < length; i++ {
			key := keys.Index(i).String()
			obj[key] = j.convertJSValueToGo(jsVal.Get(key))
		}
		return obj
	default:
		return jsVal.String()
	}
}

// convertGoToJS converts Go values to JavaScript values recursively
func (j *wasmJSONCodec) convertGoToJS(data any) js.Value {
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
		arr := j.jsArray.New(len(v))
		for i, item := range v {
			arr.SetIndex(i, j.convertGoToJS(item))
		}
		return arr
	case map[string]any:
		obj := j.jsObject.New()
		for key, val := range v {
			obj.Set(key, j.convertGoToJS(val))
		}
		return obj
	case map[string]string:
		obj := j.jsObject.New()
		for key, val := range v {
			obj.Set(key, js.ValueOf(val))
		}
		return obj
	case map[string]int:
		obj := j.jsObject.New()
		for key, val := range v {
			obj.Set(key, js.ValueOf(val))
		}
		return obj
	case []string:
		arr := j.jsArray.New(len(v))
		for i, item := range v {
			arr.SetIndex(i, js.ValueOf(item))
		}
		return arr
	case []int:
		arr := j.jsArray.New(len(v))
		for i, item := range v {
			arr.SetIndex(i, js.ValueOf(item))
		}
		return arr
	default:
		// Use reflection to handle structs and slices
		val := reflect.ValueOf(data)

		// Handle pointers - dereference and recurse
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return js.Null()
			}
			return j.convertGoToJS(val.Elem().Interface())
		}

		// Handle slices of any type using reflection
		if val.Kind() == reflect.Slice {
			arr := j.jsArray.New(val.Len())
			for i := 0; i < val.Len(); i++ {
				arr.SetIndex(i, j.convertGoToJS(val.Index(i).Interface()))
			}
			return arr
		}

		// Handle structs
		if val.Kind() == reflect.Struct {
			obj := j.jsObject.New()
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

				obj.Set(jsonTag, j.convertGoToJS(val.Field(i).Interface()))
			}
			return obj
		}

		// For other types, try to convert to string using tinystring
		return js.ValueOf(Convert(v).String())
	}
}
