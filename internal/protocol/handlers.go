package protocol

import "snakelsp/internal/request"

type RequestHandler func(r *request.Request) (interface{}, error)

var Handlers = map[string]RequestHandler{
	"initialize":                        HandleInitialize,
	"initialized":                       HandleInitialized,
	"textDocument/didOpen":              HandleDidOpen,
	"textDocument/didChange":            HandleDidChange,
	"textDocument/didClose":             HandleDidClose,
	"shutdown":                          HandleShutdown,
	"textDocument/definition":           HandleGotoDefinition,
	"workspace/symbol":                  HandleWorkspaceSymbol,
	"textDocument/documentSymbol":       HandleDocumentSybmol,
	"$/cancelRequest":                   HandleCancelRequest,
	"textDocument/prepareTypeHierarchy": HandlePrepareTypeHierarchy,
	"typeHierarchy/supertypes":          HandleTypeHierarchySuperTypes,
	"textDocument/implementation":       HandleSymbolImplementation,
}
