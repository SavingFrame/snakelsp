package protocol

import (
	"encoding/json"
	"log/slog"
	"slices"

	"snakelsp/internal/messages"
	"snakelsp/internal/request"
	"snakelsp/internal/workspace"
)

func HandleSymbolImplementation(r *request.Request) (any, error) {
	var data messages.ImplementationParams
	err := json.Unmarshal(r.Params, &data)
	if err != nil {
		r.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	workspaceFile, err := workspace.GetPythonFile(data.TextDocument.URI)
	if err != nil {
		r.Logger.Error("Error getting Python file: %v", slog.Any("error", err))
		return nil, err
	}
	symbol, err := workspace.FindSymbolByPosition(workspaceFile, data.Position.Line, data.Position.Character)
	if err != nil {
		r.Logger.Error("Error finding symbol by position: %v", slog.Any("error", err))
		return nil, nil
	}
	var response []messages.Location
	for _, s := range workspace.FlatSymbols.AllFromFront() {
		if slices.Contains(s.SuperObjects, symbol) {
			response = append(response, messages.Location{
				URI:   s.File.Url,
				Range: s.NameRange,
			})
		}
	}
	if len(response) == 1 {
		return response[0], nil
	}
	return response, nil
}
