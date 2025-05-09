package protocol

import (
	"encoding/json"
	"log/slog"
	"snakelsp/internal/messages"
	"snakelsp/internal/request"
	"snakelsp/internal/workspace"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type caputredSymbol struct {
	SymbolName string
	File       *workspace.PythonFile
	Capture    tree_sitter.QueryCapture
}

func HandleWorkspaceSymbol(r *request.Request) (interface{}, error) {
	response := []messages.WorkspaceSymbol{}
	var data messages.WorkspaceSymbolParams
	err := json.Unmarshal(r.Params, &data)
	if err != nil {
		r.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	symbols, err := workspace.GetWorkspaceSymbols(data.Query)
	limit := 100
	if len(symbols) > limit {
		symbols = symbols[:limit]
	}
	if err != nil {
		return nil, err
	}

	for _, symbol := range symbols {
		response = append(response, messages.WorkspaceSymbol{
			Name: symbol.SymbolNameWithParent(),
			Kind: symbol.Kind,
			Location: messages.Location{
				URI:   symbol.File.Url,
				Range: symbol.NameRange,
			},
		})
	}

	return response, nil
}

func HandleDocumentSybmol(r *request.Request) (interface{}, error) {
	response := []messages.DocumentSymbol{}
	var data messages.DocumentSymbolParams
	err := json.Unmarshal(r.Params, &data)
	if err != nil {
		r.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	pythonFile, err := workspace.GetPythonFile(data.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	symbols, err := pythonFile.FileSymbols("")
	if err != nil {
		return nil, err
	}
	for _, symbol := range symbols {
		children := []messages.DocumentSymbol{}
		for _, child := range symbol.Children {
			children = append(children, messages.DocumentSymbol{
				Name:           child.FullName,
				Detail:         child.Parameters,
				Kind:           child.Kind,
				Range:          child.NameRange,
				SelectionRange: child.NameRange,
			})
		}
		response = append(response, messages.DocumentSymbol{
			Name:           symbol.FullName,
			Detail:         symbol.Parameters,
			Kind:           symbol.Kind,
			Range:          symbol.NameRange,
			SelectionRange: symbol.NameRange,
			Children:       children,
		})
	}
	return response, nil
}
