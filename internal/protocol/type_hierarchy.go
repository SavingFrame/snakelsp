package protocol

import (
	"encoding/json"
	"log/slog"
	"snakelsp/internal/messages"
	"snakelsp/internal/request"
	"snakelsp/internal/workspace"

	"github.com/google/uuid"
)

func HandlePrepareTypeHierarchy(r *request.Request) (any, error) {
	var data messages.CallHierarchyPrepareParams
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
	slog.Debug("Response: %v", slog.Any("response", symbol))
	return []messages.CallHierarchyItem{
		{
			Name:           symbol.Name,
			Kind:           symbol.Kind,
			URI:            symbol.File.Url,
			Range:          &symbol.Range,
			SelectionRange: &symbol.NameRange,
			Data:           symbol.UUID.String(),
		},
	}, nil
}

func HandleTypeHierarchySuperTypes(r *request.Request) (any, error) {
	var data messages.TypeHierarchySupertypesParams
	err := json.Unmarshal(r.Params, &data)
	if err != nil {
		r.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	symbolId, err := uuid.Parse(data.Item.Data.(string))
	if err != nil {
		r.Logger.Error("Error parsing UUID: %v", slog.Any("error", err))
		return nil, err
	}
	symbol, err := workspace.SearchSymbolByUUID(symbolId)
	if err != nil {
		r.Logger.Error("Error searching symbol by UUID: %v", slog.Any("error", err))
		return nil, err
	}
	superClasses := []messages.TypeHierarchyItem{}
	for _, superClass := range symbol.SuperClasses {
		superClasses = append(superClasses, messages.TypeHierarchyItem{
			Name:           superClass.Name,
			Kind:           superClass.Kind,
			Detail:         superClass.FullName,
			URI:            superClass.File.Url,
			Range:          &superClass.Range,
			SelectionRange: &superClass.NameRange,
			Data:           superClass.UUID.String(),
		})
	}
	return superClasses, nil
}
