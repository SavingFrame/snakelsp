package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"time"
)

type JsonRpcLogger struct{}

func (l *JsonRpcLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func tempLog(s string) {
	file, err := os.OpenFile("/Users/user/projects/snakelsp/output.txt", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("File does not exists or cannot be created")
		os.Exit(1)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintln(w, s)

	w.Flush()
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
	// Format the log message manually
	timestamp := time.Now().Format(time.RFC3339) // Customize timestamp if needed
	level := record.Level.String()               // Get the log level as string
	message := record.Message                    // Default log message (no escaping)

	// Print basic log message (no escaping)
	fmt.Fprintf(h.writer, "time=%s level=%s msg=%s\n", timestamp, level, message)
	// fmt.Fprintln(h.writer, message)
	// fmt.Fprintln(h.writer, message)

	// Print any additional attributes
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
	// slog.SetDefault(logger)
	return logger
}
