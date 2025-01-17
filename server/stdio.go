package server

import (
	"io"
	"os"
)

type Stdio struct {
	input  io.ReadCloser
	output io.WriteCloser
}

func NewStdio(input io.ReadCloser, output io.WriteCloser) *Stdio {
	if input == nil {
		input = os.Stdin
	}
	if output == nil {
		output = os.Stdout
	}
	return &Stdio{
		input:  input,
		output: output,
	}
}

func (s *Server) RunStdio(stdio *Stdio) {
	<-s.newStreamConnection(stdio).DisconnectNotify()
}

func (s *Stdio) Read(p []byte) (n int, err error) {
	n, err = s.input.Read(p)
	return n, err
}

func (s *Stdio) Write(p []byte) (n int, err error) {
	return s.output.Write(p)
}

// Close closes both the input and output streams, if supported.
func (s *Stdio) Close() error {
	if err := s.input.Close(); err != nil {
		return err
	}
	return s.output.Close()
}
