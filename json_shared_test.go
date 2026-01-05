package json_test

import (
	"reflect"
	"testing"

	"github.com/tinywasm/json"
)

type TestStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func EncodeShared(t *testing.T) {
	t.Run("Encode String", func(t *testing.T) {
		input := "hello"
		expected := `"hello"`
		var result []byte
		err := json.Encode(input, &result)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}
		if string(result) != expected {
			t.Errorf("Expected %s, got %s", expected, string(result))
		}
	})

	t.Run("Encode Int", func(t *testing.T) {
		input := 123
		expected := "123"
		var result []byte
		err := json.Encode(input, &result)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}
		if string(result) != expected {
			t.Errorf("Expected %s, got %s", expected, string(result))
		}
	})

	t.Run("Encode Struct", func(t *testing.T) {
		input := TestStruct{Name: "Alice", Age: 30}
		var result []byte
		err := json.Encode(input, &result)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}
		resStr := string(result)
		if resStr != `{"name":"Alice","age":30}` && resStr != `{"age":30,"name":"Alice"}` {
			t.Errorf("Expected JSON representation of struct, got %s", resStr)
		}
	})

	t.Run("Encode Slice of Structs", func(t *testing.T) {
		input := []TestStruct{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
		}
		var result []byte
		err := json.Encode(input, &result)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}
		resStr := string(result)
		t.Logf("Encoded slice of structs: %s", resStr)

		// Should be a JSON array, not an empty string
		if resStr == `""` || resStr == "" {
			t.Errorf("BUG: Slice of structs encoded as empty string instead of JSON array, got: %s", resStr)
		}

		// Verify it's a valid JSON array
		if len(resStr) < 2 || resStr[0] != '[' || resStr[len(resStr)-1] != ']' {
			t.Errorf("Expected JSON array format [...], got: %s", resStr)
		}
	})

	t.Run("Encode Skip Private and Tagged Fields", func(t *testing.T) {
		type SkipStruct struct {
			Public  string `json:"public"`
			private string
			Skipped string `json:"-"`
		}
		input := SkipStruct{
			Public:  "visible",
			private: "hidden",
			Skipped: "should-skip",
		}
		var result []byte
		err := json.Encode(input, &result)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}
		resStr := string(result)
		// Should only contain "public":"visible"
		expected := `{"public":"visible"}`
		if resStr != expected {
			t.Errorf("Expected %s, got %s", expected, resStr)
		}
	})
}

func DecodeShared(t *testing.T) {
	t.Run("Decode String", func(t *testing.T) {
		input := `"world"`
		var result string
		err := json.Decode([]byte(input), &result)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if result != "world" {
			t.Errorf("Expected 'world', got '%s'", result)
		}
	})

	t.Run("Decode Int", func(t *testing.T) {
		input := "456"
		var result int
		err := json.Decode([]byte(input), &result)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		if result != 456 {
			t.Errorf("Expected 456, got %d", result)
		}
	})

	t.Run("Decode Struct", func(t *testing.T) {
		input := `{"name":"Bob","age":25}`
		var result TestStruct
		err := json.Decode([]byte(input), &result)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		expected := TestStruct{Name: "Bob", Age: 25}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %+v, got %+v", expected, result)
		}
	})

	t.Run("Decode Slice of Structs", func(t *testing.T) {
		input := `[{"name":"Alice","age":30},{"name":"Bob","age":25}]`
		var result []TestStruct
		err := json.Decode([]byte(input), &result)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 structs, got %d", len(result))
		}

		if len(result) > 0 && result[0].Name != "Alice" {
			t.Errorf("Expected first name 'Alice', got '%s'", result[0].Name)
		}

		if len(result) > 1 && result[1].Name != "Bob" {
			t.Errorf("Expected second name 'Bob', got '%s'", result[1].Name)
		}
	})

	// Test case that replicates crudp Packet structure with [][]byte
	t.Run("Decode Struct with byte and [][]byte fields", func(t *testing.T) {
		// This replicates crudp.Packet structure
		type Packet struct {
			Action    byte     `json:"action"`
			HandlerID uint8    `json:"handler_id"`
			ReqID     string   `json:"req_id"`
			Data      [][]byte `json:"data"`
		}

		// First encode a packet
		innerData := []byte(`{"name":"John"}`)
		packet := Packet{
			Action:    'c',
			HandlerID: 0,
			ReqID:     "test-1",
			Data:      [][]byte{innerData},
		}

		var encoded []byte
		err := json.Encode(packet, &encoded)
		if err != nil {
			t.Fatalf("Failed to encode packet: %v", err)
		}
		t.Logf("Encoded packet: %s", string(encoded))

		// Now decode it back
		var decoded Packet
		err = json.Decode(encoded, &decoded)
		if err != nil {
			t.Fatalf("Failed to decode packet: %v", err)
		}

		if decoded.Action != 'c' {
			t.Errorf("Expected Action 'c' (%d), got %d", 'c', decoded.Action)
		}
		if decoded.HandlerID != 0 {
			t.Errorf("Expected HandlerID 0, got %d", decoded.HandlerID)
		}
		if decoded.ReqID != "test-1" {
			t.Errorf("Expected ReqID 'test-1', got '%s'", decoded.ReqID)
		}
		if len(decoded.Data) != 1 {
			t.Fatalf("Expected 1 data item, got %d", len(decoded.Data))
		}
		if string(decoded.Data[0]) != string(innerData) {
			t.Errorf("Expected data '%s', got '%s'", string(innerData), string(decoded.Data[0]))
		}
	})

	// Test BatchRequest structure (nested structs with [][]byte)
	t.Run("Decode BatchRequest with nested Packets", func(t *testing.T) {
		type Packet struct {
			Action    byte     `json:"action"`
			HandlerID uint8    `json:"handler_id"`
			ReqID     string   `json:"req_id"`
			Data      [][]byte `json:"data"`
		}
		type BatchRequest struct {
			Packets []Packet `json:"packets"`
		}

		innerData := []byte(`{"name":"John","email":"john@example.com"}`)
		batch := BatchRequest{
			Packets: []Packet{
				{
					Action:    'c',
					HandlerID: 0,
					ReqID:     "test-create",
					Data:      [][]byte{innerData},
				},
			},
		}

		var encoded []byte
		err := json.Encode(batch, &encoded)
		if err != nil {
			t.Fatalf("Failed to encode batch: %v", err)
		}
		t.Logf("Encoded batch: %s", string(encoded))

		var decoded BatchRequest
		err = json.Decode(encoded, &decoded)
		if err != nil {
			t.Fatalf("Failed to decode batch: %v", err)
		}

		if len(decoded.Packets) != 1 {
			t.Fatalf("Expected 1 packet, got %d", len(decoded.Packets))
		}

		pkt := decoded.Packets[0]
		if pkt.Action != 'c' {
			t.Errorf("Expected Action 'c' (%d), got %d", 'c', pkt.Action)
		}
		if pkt.ReqID != "test-create" {
			t.Errorf("Expected ReqID 'test-create', got '%s'", pkt.ReqID)
		}
		if len(pkt.Data) != 1 {
			t.Fatalf("Expected 1 data item, got %d", len(pkt.Data))
		}
		if string(pkt.Data[0]) != string(innerData) {
			t.Errorf("Expected data '%s', got '%s'", string(innerData), string(pkt.Data[0]))
		}
	})

	// Test embedded struct (like crudp.PacketResult embeds Packet)
	t.Run("Decode struct with embedded struct", func(t *testing.T) {
		type Packet struct {
			Action    byte     `json:"action"`
			HandlerID uint8    `json:"handler_id"`
			ReqID     string   `json:"req_id"`
			Data      [][]byte `json:"data"`
		}
		type PacketResult struct {
			Packet             // Embedded struct
			MessageType uint8  `json:"message_type"`
			Message     string `json:"message"`
		}

		innerData := []byte(`{"id":123,"name":"John"}`)
		result := PacketResult{
			Packet: Packet{
				Action:    'c',
				HandlerID: 0,
				ReqID:     "test-1",
				Data:      [][]byte{innerData},
			},
			MessageType: 4, // Success
			Message:     "OK",
		}

		var encoded []byte
		err := json.Encode(result, &encoded)
		if err != nil {
			t.Fatalf("Failed to encode PacketResult: %v", err)
		}
		t.Logf("Encoded PacketResult: %s", string(encoded))

		var decoded PacketResult
		err = json.Decode(encoded, &decoded)
		if err != nil {
			t.Fatalf("Failed to decode PacketResult: %v", err)
		}

		if decoded.Action != 'c' {
			t.Errorf("Expected Action 'c' (%d), got %d", 'c', decoded.Action)
		}
		if decoded.ReqID != "test-1" {
			t.Errorf("Expected ReqID 'test-1', got '%s'", decoded.ReqID)
		}
		if decoded.MessageType != 4 {
			t.Errorf("Expected MessageType 4, got %d", decoded.MessageType)
		}
		if decoded.Message != "OK" {
			t.Errorf("Expected Message 'OK', got '%s'", decoded.Message)
		}
		if len(decoded.Data) != 1 {
			t.Fatalf("Expected 1 data item, got %d", len(decoded.Data))
		}
		if string(decoded.Data[0]) != string(innerData) {
			t.Errorf("Expected data '%s', got '%s'", string(innerData), string(decoded.Data[0]))
		}
	})

	// Test BatchResponse with embedded structs
	t.Run("Decode BatchResponse with PacketResult", func(t *testing.T) {
		type Packet struct {
			Action    byte     `json:"action"`
			HandlerID uint8    `json:"handler_id"`
			ReqID     string   `json:"req_id"`
			Data      [][]byte `json:"data"`
		}
		type PacketResult struct {
			Packet
			MessageType uint8  `json:"message_type"`
			Message     string `json:"message"`
		}
		type BatchResponse struct {
			Results []PacketResult `json:"results"`
		}

		innerData := []byte(`{"id":123,"name":"John"}`)
		batch := BatchResponse{
			Results: []PacketResult{
				{
					Packet: Packet{
						Action:    'c',
						HandlerID: 0,
						ReqID:     "test-1",
						Data:      [][]byte{innerData},
					},
					MessageType: 4,
					Message:     "OK",
				},
			},
		}

		var encoded []byte
		err := json.Encode(batch, &encoded)
		if err != nil {
			t.Fatalf("Failed to encode BatchResponse: %v", err)
		}
		t.Logf("Encoded BatchResponse: %s", string(encoded))

		var decoded BatchResponse
		err = json.Decode(encoded, &decoded)
		if err != nil {
			t.Fatalf("Failed to decode BatchResponse: %v", err)
		}

		if len(decoded.Results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(decoded.Results))
		}

		r := decoded.Results[0]
		if r.Action != 'c' {
			t.Errorf("Expected Action 'c' (%d), got %d", 'c', r.Action)
		}
		if r.MessageType != 4 {
			t.Errorf("Expected MessageType 4, got %d", r.MessageType)
		}
		if r.Message != "OK" {
			t.Errorf("Expected Message 'OK', got '%s'", r.Message)
		}
		if len(r.Data) != 1 {
			t.Fatalf("Expected 1 data item, got %d", len(r.Data))
		}
	})

	// Test decode into existing pointer (like crudp.decodeWithKnownType does)
	t.Run("Decode into existing struct pointer", func(t *testing.T) {
		type User struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		// This replicates how crudp.decodeWithKnownType works
		// It has a handler instance and decodes directly into it
		handler := &User{} // Existing pointer like in handler registry

		userData := `{"id":123,"name":"John","email":"john@example.com"}`

		err := json.Decode([]byte(userData), handler)
		if err != nil {
			t.Fatalf("Failed to decode into handler: %v", err)
		}

		if handler.ID != 123 {
			t.Errorf("Expected ID 123, got %d", handler.ID)
		}
		if handler.Name != "John" {
			t.Errorf("Expected Name 'John', got '%s'", handler.Name)
		}
		if handler.Email != "john@example.com" {
			t.Errorf("Expected Email 'john@example.com', got '%s'", handler.Email)
		}
	})

	// Test the full crudp flow: encode batch -> decode batch -> decode inner data
	t.Run("Full crudp flow simulation", func(t *testing.T) {
		type User struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		type Packet struct {
			Action    byte     `json:"action"`
			HandlerID uint8    `json:"handler_id"`
			ReqID     string   `json:"req_id"`
			Data      [][]byte `json:"data"`
		}
		type BatchRequest struct {
			Packets []Packet `json:"packets"`
		}

		// Step 1: Create user and encode
		user := User{Name: "John", Email: "john@example.com"}
		var userData []byte
		err := json.Encode(&user, &userData)
		if err != nil {
			t.Fatalf("Failed to encode user: %v", err)
		}
		t.Logf("Encoded user: %s", string(userData))

		// Step 2: Create packet with user data
		packet := Packet{
			Action:    'c',
			HandlerID: 0,
			ReqID:     "test-create",
			Data:      [][]byte{userData},
		}

		// Step 3: Create batch request
		batch := BatchRequest{Packets: []Packet{packet}}
		var batchBytes []byte
		err = json.Encode(batch, &batchBytes)
		if err != nil {
			t.Fatalf("Failed to encode batch: %v", err)
		}
		t.Logf("Encoded batch: %s", string(batchBytes))

		// Step 4: Decode batch request
		var decodedBatch BatchRequest
		err = json.Decode(batchBytes, &decodedBatch)
		if err != nil {
			t.Fatalf("Failed to decode batch: %v", err)
		}

		if len(decodedBatch.Packets) != 1 {
			t.Fatalf("Expected 1 packet, got %d", len(decodedBatch.Packets))
		}

		pkt := decodedBatch.Packets[0]
		t.Logf("Decoded packet: Action=%d, ReqID=%s, Data len=%d", pkt.Action, pkt.ReqID, len(pkt.Data))

		if len(pkt.Data) != 1 {
			t.Fatalf("Expected 1 data item, got %d", len(pkt.Data))
		}

		t.Logf("Packet data[0]: %s", string(pkt.Data[0]))

		// Step 5: Decode user from packet data (like decodeWithKnownType does)
		handler := &User{}
		err = json.Decode(pkt.Data[0], handler)
		if err != nil {
			t.Fatalf("Failed to decode user from packet data: %v", err)
		}

		if handler.Email != "john@example.com" {
			t.Errorf("Expected Email 'john@example.com', got '%s'", handler.Email)
		}
	})

	t.Run("Decode Skip Private and Tagged Fields", func(t *testing.T) {
		type SkipStruct struct {
			Public  string `json:"public"`
			private string
			Skipped string `json:"-"`
		}
		input := `{"public":"visible","private":"should-be-ignored","Skipped":"should-be-ignored"}`
		var result SkipStruct
		// Initialize with values to ensure they are cleared/not overwritten incorrectly
		result.private = "initial"
		result.Skipped = "initial"

		err := json.Decode([]byte(input), &result)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if result.Public != "visible" {
			t.Errorf("Expected Public='visible', got %q", result.Public)
		}
		if result.private != "initial" {
			t.Errorf("private field should not have been modified, got %q", result.private)
		}
		if result.Skipped != "initial" {
			t.Errorf("Skipped field should not have been modified, got %q", result.Skipped)
		}
	})
}
