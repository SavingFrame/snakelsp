package protocol

import (
	"encoding/json"
	"log/slog"

	"snakelsp/internal/messages"
	"snakelsp/internal/request"
)

func HandleCancelRequest(r *request.Request) (any, error) {
	var params messages.CancelParams
	err := json.Unmarshal(r.Params, &params)
	if err != nil {
		r.Logger.Error("Failed to unmarshal cancel request params", slog.Any("error", err))
		return nil, err
	}

	// Try to cancel the request
	cancelled := request.CancelRequest(params.ID.Value)
	if cancelled {
		r.Logger.Debug("Successfully cancelled request", slog.Any("id", params.ID.Value))
	} else {
		r.Logger.Debug("Request not found or already completed", slog.Any("id", params.ID.Value))
	}

	// Cancel requests don't return a response according to LSP spec
	return nil, nil
}
