package protocol

import (
	"encoding/json"
	"log/slog"
	"strings"

	"snakelsp/internal/messages"
	"snakelsp/internal/request"
	"snakelsp/internal/workspace"
)

func HandleDidOpen(r *request.Request) (interface{}, error) {
	var data messages.DidOpenTextDocumentParams
	var external bool
	err := json.Unmarshal(r.Params, &data)
	if err != nil {
		r.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	if data.TextDocument.LanguageID != "python" {
		return nil, nil
	}
	if strings.Contains(data.TextDocument.URI, workspace.ClientSettings.WorkspaceRoot) {
		external = false
	} else {
		external = true
	}
	workspace.NewPythonFile(data.TextDocument.URI, data.TextDocument.Text, external, true)

	return interface{}(nil), nil
}

func HandleDidChange(r *request.Request) (interface{}, error) {
	var data messages.DidChangeTextDocumentParams
	err := json.Unmarshal(r.Params, &data)
	if err != nil {
		r.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	pythonFile, err := workspace.GetPythonFile(data.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	pythonFile.ApplyChange(data.ContentChanges)

	return nil, nil
}

func HandleDidClose(r *request.Request) (interface{}, error) {
	var data messages.DidCloseTextDocumentParams
	err := json.Unmarshal(r.Params, &data)
	if err != nil {
		r.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	file, err := workspace.GetPythonFile(data.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	file.CloseFile()
	return nil, nil
}
