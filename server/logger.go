package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/fatih/color"
)

type JsonRpcLogger struct{}

func (l *JsonRpcLogger) Printf(format string, v ...any) {
	log.Printf(format, v...)
}

type CustomTextHandler struct {
	writer io.Writer
	level  *slog.LevelVar
}

func NewCustomTextHandler(w io.Writer, level *slog.LevelVar) *CustomTextHandler {
	return &CustomTextHandler{
		writer: w,
		level:  level,
	}
}

func (h *CustomTextHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.level.Level() <= level // Use the level to enable/disable logging
}

func (h *CustomTextHandler) Handle(_ context.Context, record slog.Record) error {
	timestamp := time.Now().Format(time.DateTime) // Customize timestamp if needed
	level := record.Level.String()                // Get the log level as string
	message := record.Message                     // Default log message (no escaping)

	// Force colors to be enabled even when writing to files
	color.NoColor = false

	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgHiCyan).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	if !strings.HasPrefix(message, "jsonrpc2") {
		message = blue(message)
	}
	fmt.Fprintf(h.writer, "%s | %s | %s\n", green(timestamp), cyan(level), message)

	record.Attrs(func(attr slog.Attr) bool {
		// Special handling to ensure attributes with "\n" remain unescaped
		if str, ok := attr.Value.Any().(string); ok && len(str) > 0 {
			fmt.Fprintf(h.writer, "%s: %s\n", attr.Key, str)
		} else {
			fmt.Fprintf(h.writer, "%s: %v\n", attr.Key, attr.Value.Any())
		}
		return true
	})

	// Add a separator between multiline log entries
	return nil
}

func (h *CustomTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Add attributes to the handler (e.g., for contextual logging)
	return h
}

func (h *CustomTextHandler) WithGroup(name string) slog.Handler {
	// Add grouping, if necessary
	return h
}

func createLogger(writer io.Writer, debug bool) *slog.Logger {
	programLevel := new(slog.LevelVar)
	handler := NewCustomTextHandler(writer, programLevel)
	if debug {
		programLevel.Set(slog.LevelDebug)
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
