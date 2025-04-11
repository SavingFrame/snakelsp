package protocol

import (
	"encoding/json"
	"log/slog"
	"snakelsp/internal/messages"
	"snakelsp/internal/request"
	"snakelsp/internal/workspace"
)

func HandleInitialize(r *request.Request) (interface{}, error) {
	var data messages.InitializeParams
	err := json.Unmarshal(r.Params, &data)
	if err != nil {
		r.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	userSettings := &workspace.ClientSettingsType{
		VirtualEnvPath: data.InitializationOptions.VirtualEnvPath,
	}
	workspace.ClientSettings = *userSettings

	go func() {
		workspace.ParseProject(*data.RootPath, data.InitializationOptions.VirtualEnvPath, r.Client)
		workspace.BulkParseSymbols()
	}()
	initializeResult := messages.NewInitializeResult()
	return initializeResult, nil
}

func HandleInitialized(r *request.Request) (interface{}, error) {
	return interface{}(nil), nil
}
