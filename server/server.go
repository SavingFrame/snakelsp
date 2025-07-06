// Package server provides the core LSP server implementation.
// It handles server initialization, connection management, and request routing
// for the SnakeLSP language server.
package server

import (
	"io"
	"log/slog"
	"time"
)

const (
	defaultTimeout = time.Minute
)

type Server struct {
	debug  bool
	logger *slog.Logger

	timeout      time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewServer(log_writter io.Writer) *Server {
	debug := true
	logger := createLogger(log_writter, debug)
	return &Server{
		debug:        true,
		logger:       logger,
		timeout:      defaultTimeout,
		readTimeout:  defaultTimeout,
		writeTimeout: defaultTimeout,
	}
}
