package protocol

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"snakelsp/internal/messages"
	"snakelsp/internal/request"
)

// BackgroundProcessor handles long-running requests with cancellation and progress reporting
type BackgroundProcessor struct {
	client *request.Client
}

// NewBackgroundProcessor creates a new background processor
func NewBackgroundProcessor(client *request.Client) *BackgroundProcessor {
	return &BackgroundProcessor{
		client: client,
	}
}

// ProcessRequest processes a request in the background with cancellation support
func (bp *BackgroundProcessor) ProcessRequest(
	requestID any,
	workDoneToken *messages.ProgressToken,
	partialResultToken *messages.ProgressToken,
	processor func(ctx context.Context, partial PartialResultSender) (any, error),
) (any, error) {
	// Create cancellable context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register for cancellation
	// request.RegisterRequest(requestID, cancel)
	defer request.UnregisterRequest(requestID)

	// Create partial result sender
	partialSender := NewPartialResultSender(bp.client, partialResultToken)

	// Process in background
	resultChan := make(chan any, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Background processor panic", slog.Any("panic", r))
				errorChan <- fmt.Errorf("internal error during processing")
			}
		}()

		result, err := processor(ctx, partialSender)
		if err != nil {
			slog.Error("Error during background processing", slog.Any("error", err))
			errorChan <- err
		} else {
			slog.Debug("Background processing completed successfully", slog.Any("id", requestID))
			resultChan <- result
		}
	}()

	slog.Debug("Background processing completed")
	// Wait for completion or cancellation
	select {
	case <-ctx.Done():
		slog.Debug("Request cancelled in the ctx done", slog.Any("id", requestID))
		return nil, ctx.Err()
	case err := <-errorChan:
		slog.Error("Error during processing", slog.Any("error", err))
		return nil, err
	case result := <-resultChan:
		slog.Debug("Request completed successfully")
		return result, nil
	}
}

// PartialResultSender handles partial result notifications
type PartialResultSender interface {
	Send(result interface{}) error
}

type partialResultSender struct {
	client *request.Client
	token  *messages.ProgressToken
}

func NewPartialResultSender(client *request.Client, token *messages.ProgressToken) PartialResultSender {
	return &partialResultSender{
		client: client,
		token:  token,
	}
}

func (prs *partialResultSender) Send(result interface{}) error {
	if prs.token == nil {
		return nil // No partial result support requested
	}

	// Send partial result notification
	params := map[string]interface{}{
		"token": prs.token.Value,
		"value": result,
	}

	prs.client.Notify("$/partialResult", params)
	slog.Debug("Sent partial result")
	return nil
}

// Helper function to check if context is cancelled and return appropriate error
func CheckCancellation(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// Helper function to add cancellation checks in loops
func ShouldContinue(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		slog.Debug("Operation cancelled", slog.Any("error", ctx.Err()))
		return false
	default:
		return true
	}
}

// Helper to create a timeout context for operations
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
