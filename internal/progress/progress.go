// Package progress provides work-done progress reporting functionality for LSP operations.
// It implements the LSP work done progress protocol to show progress indicators
// to clients during long-running operations.
package progress

import (
	"snakelsp/internal/messages"
	"snakelsp/internal/request"

	"github.com/google/uuid"
)

type WorkDone struct {
	token     uuid.UUID
	client    *request.Client
	isStarted bool
}

func NewWorkDone(client *request.Client) *WorkDone {
	token := uuid.New()
	return &WorkDone{
		token:  token,
		client: client,
	}
}

func (w *WorkDone) Start(message string) {
	w.client.Notify("window/workDoneProgress/create", map[string]any{"token": w.token.String()})
	w.client.Notify("$/progress", messages.ServerWorkDoneProgress{
		Token: w.token.String(),
		WorkDoneProgress: messages.WorkDoneProgress{
			Kind:        "begin",
			Title:       message,
			Cancellable: false,
			Message:     message,
		},
	})
	w.isStarted = true
}

func (w *WorkDone) Report(message string, percentage uint16) {
	if !w.isStarted {
		w.Start(message)
	}
	w.client.Notify("$/progress", messages.ServerWorkDoneProgress{
		Token: w.token.String(),
		WorkDoneProgress: messages.WorkDoneProgress{
			Kind:       "report",
			Message:    message,
			Percentage: percentage,
		},
	})
}

func (w *WorkDone) End(message string) {
	w.client.Notify("$/progress", messages.ServerWorkDoneProgress{
		Token: w.token.String(),
		WorkDoneProgress: messages.WorkDoneProgress{
			Kind: "end",
		},
	})
}
