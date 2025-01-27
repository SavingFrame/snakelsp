package protocol

type RequestHandler func(context *Context) (interface{}, error)

var Handlers = map[string]RequestHandler{
	"initialize":              HandleInitialize,
	"initialized":             HandleInitialized,
	"textDocument/didOpen":    HandleDidOpen,
	"textDocument/didChange":  HandleDidChange,
	"textDocument/didClose":   HandleDidClose,
	"shutdown":                HandleShutdown,
	"textDocument/definition": HandleGotoDefinition,
	"workspace/symbol":        HandleWorkspaceSymbol,
}
