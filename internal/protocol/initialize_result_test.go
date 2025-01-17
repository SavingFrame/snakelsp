package protocol

import (
	// "log"
	// "snakelsp/test"
	"testing"
)

func TestInitializeResult(t *testing.T) {
	// mockStdinWriter, _, done, ready := test.ServerFixture(t)
	// defer close(done)
	// <-ready
	// log.Println("ready")
	// initializeRequest := map[string]interface{}{
	// 	"jsonrpc": "2.0",
	// 	"id":      1,
	// 	"method":  "initialize",
	// 	"params": map[string]interface{}{
	// 		"rootUri": "file:///path/to/project",
	// 		"capabilities": map[string]interface{}{
	// 			"textDocument": map[string]interface{}{
	// 				"completion": true,
	// 			},
	// 		},
	// 	},
	// }
	// if err := test.SendLSPRequest(t, mockStdinWriter, initializeRequest); err != nil {
	// 	t.Fatalf("Failed to send initialize request: %v", err)
	// }

	// Read and validate the response
	// response := test.ReadLSPResponse(t, mockStdoutReader)
	// log.Printf("response: %v", response)
	// if response["id"] != float64(1) { // JSON numbers are unmarshalled as float64
	// 	t.Errorf("Expected response id 1, got %v", response["id"])
	// }
	// if response["jsonrpc"] != "2.0" {
	// 	t.Errorf("Expected jsonrpc '2.0', got %v", response["jsonrpc"])
	// }
}

// func TestSayHello(t *testing.T) {
// 	// Setup the mocked server
// 	mockStdinWriter, mockStdoutReader, done := test.ServerFixture(t)
// 	defer close(done)
//
// 	// Request for "sayHello"
// 	request := map[string]interface{}{
// 		"jsonrpc": "2.0",
// 		"id":      1,
// 		"method":  "sayHello",
// 		"params":  nil,
// 	}
//
// 	// Send the request to the server
// 	if err := test.SendLSPRequest(mockStdinWriter, request); err != nil {
// 		t.Fatalf("Failed to send request: %v", err)
// 	}
//
// 	// Read and validate the response
// 	response := test.ReadLSPResponse(t, mockStdoutReader)
//
// 	// Check the basics of the response
// 	if response["id"] != float64(1) { // JSON numbers unmarshalled as float64
// 		t.Errorf("Expected response id 1, got %v", response["id"])
// 	}
// 	if response["jsonrpc"] != "2.0" {
// 		t.Errorf("Expected jsonrpc '2.0', got %v", response["jsonrpc"])
// 	}
//
// 	// Check response result
// 	result, ok := response["result"]
// 	if !ok || result != "hello world" {
// 		t.Errorf("Expected result 'hello world', got %v", result)
// 	}
// }
