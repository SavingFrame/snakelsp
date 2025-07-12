package protocol

import (
	"context"
	"encoding/json"

	"snakelsp/internal/messages"
	"snakelsp/internal/request"
)

// BackgroundHandler wraps any handler to support background processing
type BackgroundHandler func(
	ctx context.Context,
	params json.RawMessage,
	partial PartialResultSender,
) (any, error)

// WithBackground wraps a handler to support background processing with cancellation
func WithBackground(handler BackgroundHandler) func(r *request.Request) (any, error) {
	return func(r *request.Request) (any, error) {
		// Parse work done and partial result tokens from params
		var baseParams struct {
			messages.WorkDoneProgressParams
			messages.PartialResultParams
		}

		// Try to unmarshal tokens (ignore errors as they're optional)
		json.Unmarshal(r.Params, &baseParams)

		// Create background processor
		processor := NewBackgroundProcessor(r.Client)

		// Process in background
		return processor.ProcessRequest(
			r.ID,
			baseParams.WorkDoneToken,
			baseParams.PartialResultToken,
			func(ctx context.Context, partial PartialResultSender) (any, error) {
				return handler(ctx, r.Params, partial)
			},
		)
	}
}

// Example usage for a simple handler that doesn't need background processing
func SimpleHandler(handler func(r *request.Request) (any, error)) func(r *request.Request) (any, error) {
	return func(r *request.Request) (any, error) {
		// Register for cancellation but process synchronously
		ctx, cancel := context.WithCancel(r.Context)
		defer cancel()

		request.RegisterRequest(r.ID, cancel)
		defer request.UnregisterRequest(r.ID)

		// Check for cancellation before processing
		if err := CheckCancellation(ctx); err != nil {
			return nil, err
		}

		// Update request context
		r.Context = ctx
		return handler(r)
	}
}
