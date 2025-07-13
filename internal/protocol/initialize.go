package protocol

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"snakelsp/internal/messages"
	"snakelsp/internal/progress"
	"snakelsp/internal/request"
	"snakelsp/internal/workspace"
)

func HandleInitialize(r *request.Request) (any, error) {
	var data messages.InitializeParams
	err := json.Unmarshal(r.Params, &data)
	if err != nil {
		r.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	if data.RootPath == "" {
		return nil, fmt.Errorf("rootPath is required")
	}
	workspace.SetClientSettings(data.InitializationOptions.VirtualEnvPath, data.RootPath)

	go func() {
		filesProgress := progress.NewWorkDone(r.Client)
		workspace.ParseProjectFiles(data.RootPath, data.InitializationOptions.VirtualEnvPath, filesProgress)
		importsProgress := progress.NewWorkDone(r.Client)
		workspace.BulkParseImports(importsProgress)
		symbolsProgress := progress.NewWorkDone(r.Client)
		workspace.BulkParseSymbols(symbolsProgress)
	}()
	initializeResult := messages.NewInitializeResult(&data)
	return initializeResult, nil
}

func HandleInitialized(r *request.Request) (any, error) {
	return any(nil), nil
}
