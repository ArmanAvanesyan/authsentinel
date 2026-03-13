package caddy

import "github.com/ArmanAvanesyan/authsentinel/pkg/proxy"

// Middleware is a placeholder type for a Caddy middleware that delegates to the proxy Engine.
type Middleware struct {
	engine proxy.Engine
}

// NewMiddleware constructs a new Middleware instance.
func NewMiddleware(e proxy.Engine) *Middleware {
	return &Middleware{engine: e}
}
