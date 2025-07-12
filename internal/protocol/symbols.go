package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"

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

func HandleWorkspaceSymbol(r *request.Request) (any, error) {
	var params messages.WorkspaceSymbolParams

	err := json.Unmarshal(r.Params, &params)
	if err != nil {
		slog.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}

	// Create background processor
	processor := NewBackgroundProcessor(r.Client)

	// Process in background with cancellation and progress support
	return processor.ProcessRequest(
		r.ID,
		params.WorkDoneToken,
		params.PartialResultToken,
		func(ctx context.Context, partial PartialResultSender) (any, error) {
			return processWorkspaceSymbolsBackground(ctx, params.Query, partial)
		},
	)
}

func processWorkspaceSymbolsBackground(
	ctx context.Context,
	query string,
	partial PartialResultSender,
) (any, error) {
	// Start progress reporting

	if err := CheckCancellation(ctx); err != nil {
		slog.Debug("Request cancelled before processing", slog.Any("error", err))
		return nil, err
	}

	response := []messages.WorkspaceSymbol{}

	for symbols := range slices.Chunk(slices.Collect(workspace.FlatSymbols.Values()), 100) {

		if !ShouldContinue(ctx) {
			slog.Debug("Request dhouldn't continue", slog.Any("context", ctx))
			return nil, ctx.Err()
		}
		var err error
		if query != "" {
			symbols, err = workspace.FilterSymbols(symbols, query)
			if err != nil {
				slog.Error("Error filtering symbols", slog.Any("error", err))
				return nil, fmt.Errorf("error filtering symbols: %w", err)
			}
		}

		for _, symbol := range symbols {

			workspaceSymbol := messages.WorkspaceSymbol{
				Name: symbol.SymbolNameWithParent(),
				Kind: symbol.Kind,
				Location: messages.Location{
					URI:   symbol.File.Url,
					Range: symbol.NameRange,
				},
			}

			response = append(response, workspaceSymbol)

			if err := partial.Send(response); err != nil {
				slog.Error("Failed to send partial result", slog.Any("error", err))
			}
		}
	}
	return response, nil
}

func HandleDocumentSybmol(r *request.Request) (any, error) {
	response := []messages.DocumentSymbol{}
	var data messages.DocumentSymbolParams
	err := json.Unmarshal(r.Params, &data)
	if err != nil {
		slog.Error("Unmarshalling error: %v", slog.Any("error", err))
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
