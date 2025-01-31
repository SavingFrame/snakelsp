package protocol

import (
	"encoding/json"
	"log/slog"
	"snakelsp/internal/messages"
	"snakelsp/internal/workspace"
)

func HandleInitialize(c *Context) (interface{}, error) {
	var data messages.InitializeParams
	err := json.Unmarshal(c.Params, &data)
	if err != nil {
		c.Logger.Error("Unmarshalling error: %v", slog.Any("error", err))
		return nil, err
	}
	userSettings := &workspace.ClientSettingsType{
		VirtualEnvPath: data.InitializationOptions.VirtualEnvPath,
	}
	workspace.ClientSettings = *userSettings

	go func() {
		workspace.ParseProject(*data.RootPath, data.InitializationOptions.VirtualEnvPath)
		workspace.BulkParseSymbols()
	}()
	initializeResult := messages.NewInitializeResult()
	return initializeResult, nil
}

func HandleInitialized(context *Context) (interface{}, error) {
	return interface{}(nil), nil
}
