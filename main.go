package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"snakelsp/server"
)

type discardCloser struct {
	io.Writer
}

func (d *discardCloser) Close() error {
	return nil
}

func initializeLogs() (io.WriteCloser, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	logFilePath := filepath.Join(homeDir, ".local", "state", "snakelsp", "snakelsp.log")
	err = os.MkdirAll(filepath.Dir(logFilePath), os.ModePerm)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func main() {
	stdio := server.NewStdio(nil, nil)
	f, err := initializeLogs()
	if err != nil {
		log.Printf("error initializing logs: %v", err)
		f = &discardCloser{io.Discard}
	}
	defer f.Close()
	srv := server.NewServer(f)
	srv.RunStdio(stdio)
}
