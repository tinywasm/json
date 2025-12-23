//go:build wasm

package main

import (
	"syscall/js"

	"github.com/tinywasm/json"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func main() {
	console := js.Global().Get("console")

	// Create h1 element
	document := js.Global().Get("document")
	body := document.Get("body")

	h1 := document.Call("createElement", "h1")
	h1.Set("innerHTML", "JSON WASM Example")
	body.Call("appendChild", h1)

	// 1. Encode Example
	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	var jsonData []byte
	err := json.Encode(user, &jsonData)
	if err != nil {
		console.Call("error", "Encode error:", err.Error())
		return
	}

	// Display JSON
	p1 := document.Call("createElement", "p")
	p1.Set("innerHTML", "Encoded JSON: "+string(jsonData))
	body.Call("appendChild", p1)

	// 2. Decode Example
	var decodedUser User
	err = json.Decode(jsonData, &decodedUser)
	if err != nil {
		console.Call("error", "Decode error:", err.Error())
		return
	}

	// Display Decoded Data
	p2 := document.Call("createElement", "p")
	// Manual string concatenation to avoid fmt.Sprintf
	info := "Decoded User: Name=" + decodedUser.Name + ", Age=" + js.ValueOf(decodedUser.Age).String()
	p2.Set("innerHTML", info)
	body.Call("appendChild", p2)

	console.Call("log", "JSON example finished successfully")

	select {}
}
