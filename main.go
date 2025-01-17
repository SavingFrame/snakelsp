package main

import (
	"log"
	"os"
	"snakelsp/server"
)

func main() {
	stdio := server.NewStdio(nil, nil)
	f, err := os.OpenFile("/Users/user/projects/snakelsp/log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	srv := server.NewServer(f)
	srv.RunStdio(stdio)
}
