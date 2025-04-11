package request

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

type Client struct {
	Conn    *jsonrpc2.Conn
	Request *jsonrpc2.Request
	Context *context.Context
}

type Request struct {
	Method  string
	Params  json.RawMessage
	Context context.Context
	Client  *Client
	Logger  *slog.Logger
}

func (c *Client) Notify(method string, params any) {
	if err := c.Conn.Notify(*c.Context, method, params); err != nil {
		slog.Error(err.Error())
	}
}

func (c *Client) Call(method string, params any) (any, error) {
	result := interface{}(nil)
	if err := c.Conn.Call(*c.Context, method, params, &result); err != nil {
		slog.Error(err.Error())
		return result, err
	}
	return result, nil
}
