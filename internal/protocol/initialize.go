package protocol

import (
	"snakelsp/internal/messages"
)

func HandleInitialize(context *Context) (interface{}, error) {
	initializeResult := messages.NewInitializeResult()
	return initializeResult, nil
	// context.Connection.Reply(ctx context.Context, id jsonrpc2.ID, result interface{})
}

func HandleInitialized(context *Context) (interface{}, error) {
	return interface{}(nil), nil
}
