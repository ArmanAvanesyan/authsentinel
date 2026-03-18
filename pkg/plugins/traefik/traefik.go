package traefik

import (
	"context"
	"net/http"

	"github.com/ArmanAvanesyan/authsentinel/pkg/pluginapi"
	"github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
)

// Middleware is a Traefik adapter that delegates to the proxy Engine.
// It implements pluginapi.IntegrationPlugin and provides the middleware plugin
// interface: request context handoff, header and deny response mapping.
type Middleware struct {
	engine                 proxy.Engine
	desc                   pluginapi.PluginDescriptor
	upstreamURL             string
	authResponseHeaders     []string
}

// NewMiddleware constructs a new Middleware instance. upstreamURL is used when
// the adapter performs upstream handoff; authResponseHeaders filters which headers
// are copied to the forwarded request (empty = forward X-* and Authorization).
func NewMiddleware(e proxy.Engine, desc pluginapi.PluginDescriptor, upstreamURL string, authResponseHeaders []string) *Middleware {
	return &Middleware{engine: e, desc: desc, upstreamURL: upstreamURL, authResponseHeaders: authResponseHeaders}
}

// Handler returns an http.Handler that runs the proxy engine and maps response
// for Traefik (headers and deny mapping). Use as Traefik ForwardAuth target or
// wrap in Traefik plugin middleware.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return Handler(m.engine, next, m.upstreamURL, m.authResponseHeaders)
}

// Descriptor implements pluginapi.Plugin.
func (m *Middleware) Descriptor() pluginapi.PluginDescriptor { return m.desc }

// Health implements pluginapi.Plugin.
func (m *Middleware) Health(ctx context.Context) pluginapi.PluginHealth {
	return pluginapi.PluginHealth{State: pluginapi.PluginStateHealthy}
}

// Serve implements pluginapi.IntegrationPlugin.
// hostCtx is gateway-specific. When used as ForwardAuth, Traefik calls the handler
// at AuthURL; when used as plugin, hostCtx wires the middleware into the route.
func (m *Middleware) Serve(ctx context.Context, hostCtx any) error {
	_ = ctx
	_ = hostCtx
	return nil
}
