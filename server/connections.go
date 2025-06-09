package server

import (
	"context"
	"io"

	"github.com/sourcegraph/jsonrpc2"
)

func (s *Server) newStreamConnection(stream io.ReadWriteCloser) *jsonrpc2.Conn {
	handler := s.newHandler()
	connectionOptions := s.newConnectionOptions()
	context, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return jsonrpc2.NewConn(context, jsonrpc2.NewBufferedStream(stream, jsonrpc2.VSCodeObjectCodec{}), handler, connectionOptions...)
}

func (s *Server) newConnectionOptions() []jsonrpc2.ConnOpt {
	// if s.debug {
	// 	return []jsonrpc2.ConnOpt{
	// 		jsonrpc2.LogMessages(&JsonRpcLogger{}),
	// 	}
	// 	// return nil
	// } else {
	// 	return nil
	// }
	return nil
}
