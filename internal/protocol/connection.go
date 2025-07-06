// Package protocol implements the Language Server Protocol handlers and
// request processing logic. It provides the core functionality for handling
// LSP requests like shutdown, initialization, and other protocol operations.
package protocol

import "snakelsp/internal/request"

func HandleShutdown(r *request.Request) (interface{}, error) {
	return interface{}(nil), nil
}
