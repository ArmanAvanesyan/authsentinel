package caddy

import (
	"context"
	"net/http"

	"github.com/ArmanAvanesyan/authsentinel/pkg/pluginapi"
	"github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
)

// Middleware is a Caddy adapter that delegates to the proxy Engine.
// It implements pluginapi.IntegrationPlugin and provides the shared AuthSentinel
// runtime integration: directive/module registration and upstream handoff.
type Middleware struct {
	engine       proxy.Engine
	desc         pluginapi.PluginDescriptor
	upstreamURL  string
}

// NewMiddleware constructs a new Middleware instance. UpstreamURL is used when
// the adapter performs upstream handoff (Handler with non-empty upstream).
func NewMiddleware(e proxy.Engine, desc pluginapi.PluginDescriptor, upstreamURL string) *Middleware {
	return &Middleware{engine: e, desc: desc, upstreamURL: upstreamURL}
}

// Handler returns an http.Handler that runs the proxy engine and hands off to next
// or to upstream. This is the entrypoint for Caddy (or any HTTP server) to use
// the same proxy/policy/token runtime.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return Handler(m.engine, next, m.upstreamURL)
}

// Descriptor implements pluginapi.Plugin.
func (m *Middleware) Descriptor() pluginapi.PluginDescriptor { return m.desc }

// Health implements pluginapi.Plugin.
func (m *Middleware) Health(ctx context.Context) pluginapi.PluginHealth {
	return pluginapi.PluginHealth{State: pluginapi.PluginStateHealthy}
}

// Serve implements pluginapi.IntegrationPlugin.
// hostCtx is gateway-specific (e.g. *caddy.Controller). When used from Caddy module,
// the directive wires this handler into the route. When used from authsentinel-proxy,
// hostCtx may be nil and the adapter is used as a registered integration only.
func (m *Middleware) Serve(ctx context.Context, hostCtx any) error {
	_ = ctx
	_ = hostCtx
	return nil
}
