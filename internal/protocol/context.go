package protocol

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/sourcegraph/jsonrpc2"
)

type (
	NotifyFunc func(method string, params any)
	CallFunc   func(method string, params any) (any, error)
)

type Context struct {
	Method     string
	Params     json.RawMessage
	Context    context.Context
	Connection *jsonrpc2.Conn
	Request    *jsonrpc2.Request
	Logger     *slog.Logger
}

func (c *Context) Notify(method string, params any) {
	if err := c.Connection.Notify(c.Context, method, params); err != nil {
		slog.Error(err.Error())
	}
}

func (c *Context) Call(method string, params any) (any, error) {
	result := interface{}(nil)
	if err := c.Connection.Call(c.Context, method, params, &result); err != nil {
		slog.Error(err.Error())
		return result, err
	}
	return result, nil
}
