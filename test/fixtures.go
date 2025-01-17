package test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"snakelsp/server"
	"testing"
	"time"
)

func ServerFixture(t *testing.T) (stdinWriter io.WriteCloser, stdoutReader *bufio.Reader, done chan bool, ready chan struct{}) {
	t.Helper() // Mark this as a test helper

	// Create pipes for communication
	mockInputReader, mockInputWriter := io.Pipe()   // Mock stdin
	mockOutputReader, mockOutputWriter := io.Pipe() // Mock stdout
	stdio := server.NewStdio(mockInputReader, mockOutputWriter)

	log_writer := os.Stdout
	srv := server.NewServer(log_writer)
	done = make(chan bool)
	ready = make(chan struct{})
	// Start the LSP server in a separate goroutine
	go func() {
		defer close(done) // Notify completion
		close(ready)
		srv.RunStdio(stdio) // Runs the server in stdio communication mode
	}()

	// Return the writer for stdin, reader for stdout, and a done channel for cleanup
	return mockInputWriter, bufio.NewReader(mockOutputReader), done, ready
}

func SendLSPRequest(t *testing.T, stdin io.WriteCloser, request map[string]interface{}) error {
	t.Helper() // Mark this as a test helper
	rawReq, err := json.Marshal(request)
	if err != nil {
		return err
	}
	println("send lsp request")

	// Write the request as JSON followed by a newline (LSP framing)
	_, err = stdin.Write(append(rawReq, '\n'))
	// err = stdin.Close()
	// if err != nil {
	// 	log.Println(err)
	// }
	return err
}

// ReadLSPResponse reads a JSON-RPC response from the stdout reader.
// It handles timeout and ensures valid unmarshalling into a Go map.
func ReadLSPResponse(t *testing.T, stdout *bufio.Reader) map[string]interface{} {
	t.Helper() // Mark this as a helper function

	responseBuffer := bytes.Buffer{}
	done := make(chan struct{})    // Signal for completion
	errChan := make(chan error, 1) // Signal for errors

	// Start a goroutine to read the response
	go func() {
		println("test")
		defer close(done) // Notify completion
		for {
			println("test2")
			line, err := stdout.ReadBytes('\n') // Read output line by line
			println(string(line))
			if err != nil {
				if err == io.EOF { // Graceful completion on EOF
					return
				}
				errChan <- err // Signal errors
				return
			}
			responseBuffer.Write(line) // Accumulate response lines
		}
	}()

	// Wait for reading to finish, detect errors, or timeout
	select {
	case <-done: // Successfully finished reading
	case err := <-errChan: // Error occurred while reading
		t.Fatalf("Error reading from mock stdout: %v", err)
	case <-time.After(2 * time.Second): // Timeout
		t.Fatal("Timeout waiting for response")
	}

	// Parse JSON-RPC response into a Go map
	var response map[string]interface{}
	err := json.Unmarshal(responseBuffer.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON response: %v", err)
	}
	return response
}
