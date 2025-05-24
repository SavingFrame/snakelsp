package protocol

import (
	"encoding/json"
	"log/slog"
	"snakelsp/internal/messages"
	"snakelsp/internal/request"
	"snakelsp/internal/workspace"
)

func HandleSymbolDeclaration(r *request.Request) (any, error) {
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
	if len(symbol.SuperObjects) > 0 {
		superMethod := symbol.SuperObjects[0]
		return messages.Location{
			URI:   superMethod.File.Url,
			Range: superMethod.NameRange,
		}, nil
	}
	return nil, nil
}
