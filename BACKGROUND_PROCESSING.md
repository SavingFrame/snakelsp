# Background Processing with Cancellation Support

This document explains how to implement background processing for LSP requests with cancellation and partial response support.

## Architecture Overview

The background processing system consists of:

1. **BackgroundProcessor**: Main orchestrator for async request processing
2. **ProgressReporter**: Sends work done progress notifications to client
3. **PartialResultSender**: Streams partial results as they become available
4. **Cancellation Integration**: Checks context cancellation during processing

## Key Features

- **Asynchronous Processing**: Requests run in background goroutines
- **Cancellation Support**: Requests can be cancelled via `$/cancelRequest`
- **Progress Reporting**: Work done progress notifications with percentages
- **Partial Results**: Stream results as they become available
- **Error Recovery**: Panic recovery and proper error handling

## Usage Examples

### 1. Full Background Processing (workspace/symbol)

```go
func HandleWorkspaceSymbol(r *request.Request) (any, error) {
    var params struct {
        messages.WorkspaceSymbolParams
        messages.WorkDoneProgressParams
        messages.PartialResultParams
    }

    err := json.Unmarshal(r.Params, &params)
    if err != nil {
        return nil, err
    }

    processor := NewBackgroundProcessor(r.Client, r.Logger)

    return processor.ProcessRequest(
        r.Context,
        r.ID,
        params.WorkDoneToken,
        params.PartialResultToken,
        func(ctx context.Context, progress ProgressReporter, partial PartialResultSender) (any, error) {
            return processWorkspaceSymbolsBackground(ctx, params.Query, progress, partial, r.Logger)
        },
    )
}
```

### 2. Using the Background Wrapper

```go
// Define your background handler
var handleMyRequest = WithBackground(func(
    ctx context.Context,
    params json.RawMessage,
    progress ProgressReporter,
    partial PartialResultSender,
) (any, error) {
    // Parse your specific params
    var myParams MyRequestParams
    if err := json.Unmarshal(params, &myParams); err != nil {
        return nil, err
    }

    progress.Begin("Processing request", true, "Starting...", 0)

    // Your processing logic with cancellation checks
    for i, item := range items {
        if !ShouldContinue(ctx) {
            return nil, ctx.Err()
        }

        // Process item
        result := processItem(item)

        // Send partial result
        partial.Send(result)

        // Update progress
        progress.Report(fmt.Sprintf("Processed %d/%d", i+1, len(items)),
                       uint16((i+1)*100/len(items)))
    }

    progress.End("Completed")
    return finalResult, nil
})
```

### 3. Simple Handler with Cancellation Only

```go
var handleSimpleRequest = SimpleHandler(func(r *request.Request) (any, error) {
    // Check for cancellation
    if err := CheckCancellation(r.Context); err != nil {
        return nil, err
    }

    // Your synchronous processing logic
    return processRequest(r.Params)
})
```

## Progress Reporting

Progress reporting follows the LSP work done progress specification:

```go
// Start progress
progress.Begin("Operation Title", true, "Initial message", 0)

// Report progress updates
progress.Report("Processing file 1/10", 10)
progress.Report("Processing file 5/10", 50)

// End progress
progress.End("Operation completed")
```

## Partial Results

Send partial results as they become available:

```go
// Send individual results
partial.Send(singleResult)

// Send batch of results
partial.Send([]Result{result1, result2, result3})
```

## Cancellation Handling

### In Processing Loops

```go
for _, item := range items {
    // Check if request was cancelled
    if !ShouldContinue(ctx) {
        logger.Debug("Processing cancelled")
        return nil, ctx.Err()
    }

    // Process item
    processItem(item)
}
```

### Before Expensive Operations

```go
// Check before starting expensive operation
if err := CheckCancellation(ctx); err != nil {
    return nil, err
}

expensiveOperation()
```

### With Timeouts

```go
// Add timeout to context
timeoutCtx, cancel := WithTimeout(ctx, 30*time.Second)
defer cancel()

result, err := longRunningOperation(timeoutCtx)
```

## Client Integration

Clients can:

1. **Request work done progress**: Include `workDoneToken` in request params
2. **Request partial results**: Include `partialResultToken` in request params
3. **Cancel requests**: Send `$/cancelRequest` with the request ID

Example client request:

```json
{
  "id": 1,
  "method": "workspace/symbol",
  "params": {
    "query": "MyClass",
    "workDoneToken": "progress-token-1",
    "partialResultToken": "partial-token-1"
  }
}
```

## Error Handling

The system handles:

- **Panic recovery**: Panics are caught and converted to errors
- **Context cancellation**: Proper cleanup when requests are cancelled
- **Timeout handling**: Requests can have timeouts applied
- **Partial result errors**: Failures in partial result sending are logged but don't stop processing

## Performance Considerations

- **Batch partial results**: Don't send individual items, batch them for efficiency
- **Progress granularity**: Don't update progress too frequently (every 1-5% is sufficient)
- **Memory management**: Be careful with large result sets in memory
- **Goroutine cleanup**: The system properly cleans up goroutines and contexts

## Migration Guide

To migrate existing handlers:

1. **Add background processing**: Wrap handler with `WithBackground()` or use `BackgroundProcessor` directly
2. **Add progress reporting**: Use `ProgressReporter` interface for long operations
3. **Add partial results**: Use `PartialResultSender` for streaming results
4. **Add cancellation checks**: Use `ShouldContinue()` and `CheckCancellation()` in loops
5. **Update params parsing**: Include `WorkDoneProgressParams` and `PartialResultParams`

The system is backward compatible - handlers work without these features if clients don't request them.
