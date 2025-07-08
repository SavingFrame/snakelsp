// Package request provides client and request handling utilities for LSP communication.
// It wraps JSON-RPC connections and provides convenient methods for making
// requests and notifications to LSP clients.
package request

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/sourcegraph/jsonrpc2"
)

type (
	NotifyFunc func(method string, params any)
	CallFunc   func(method string, params any) (any, error)
)

type Client struct {
	Conn    *jsonrpc2.Conn
	Request *jsonrpc2.Request
	Context context.Context
}

type Request struct {
	Method  string
	Params  json.RawMessage
	Context context.Context
	Client  *Client
	Logger  *slog.Logger
	ID      any // Request ID for cancellation tracking
}

var (
	activeRequests = make(map[any]context.CancelFunc)
	requestsMutex  sync.RWMutex
)

func RegisterRequest(id any, cancel context.CancelFunc) {
	requestsMutex.Lock()
	defer requestsMutex.Unlock()
	activeRequests[id] = cancel
}

func CancelRequest(id any) bool {
	requestsMutex.Lock()
	defer requestsMutex.Unlock()
	if cancel, exists := activeRequests[id]; exists {
		cancel()
		delete(activeRequests, id)
		return true
	}
	return false
}

func UnregisterRequest(id any) {
	requestsMutex.Lock()
	defer requestsMutex.Unlock()
	delete(activeRequests, id)
}

func (c *Client) Notify(method string, params any) {
	if err := c.Conn.Notify(c.Context, method, params); err != nil {
		slog.Error(err.Error())
	}
}

func (c *Client) Call(method string, params any) (any, error) {
	result := interface{}(nil)
	if err := c.Conn.Call(c.Context, method, params, &result); err != nil {
		slog.Error(err.Error())
		return result, err
	}
	return result, nil
}
