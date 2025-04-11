package server

import (
	"context"
	"log"
	"snakelsp/internal/protocol"
	"snakelsp/internal/request"

	"github.com/sourcegraph/jsonrpc2"
)

func (s *Server) handle(ctx context.Context, c *jsonrpc2.Conn, r *jsonrpc2.Request) (any, error) {
	context := request.Request{
		Method:  r.Method,
		Context: ctx,
		Client: &request.Client{
			Conn:    c,
			Request: r,
		},
		Logger: s.logger,
	}
	if r.Params != nil {
		context.Params = *r.Params
	}
	switch r.Method {
	case "exit":
		protocol.Handlers[r.Method](&context)
		err := c.Close()
		return nil, err
	default:
		handler, exists := protocol.Handlers[r.Method]
		if !exists {
			err := &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: "Method not found"}
			if err := c.ReplyWithError(ctx, r.ID, err); err != nil {
				log.Println(err)
				return nil, err
			}
		}
		result, err := handler(&context)
		if err != nil {
			err := &jsonrpc2.Error{Code: jsonrpc2.CodeInternalError, Message: err.Error()}
			if err := c.ReplyWithError(ctx, r.ID, err); err != nil {
				log.Println(err)
				return nil, err
			}
			return result, nil

		}
		return result, nil
	}
}

func (s *Server) newHandler() jsonrpc2.Handler {
	return jsonrpc2.HandlerWithError(s.handle)
}
