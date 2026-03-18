package krakend

import (
	"context"
	"net/http"

	"github.com/ArmanAvanesyan/authsentinel/pkg/pluginapi"
	"github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
)

// Plugin is a KrakenD adapter that delegates to the proxy Engine.
// It implements pluginapi.IntegrationPlugin and provides the endpoint/auth
// middleware bridge with shared principal decision output (headers).
type Plugin struct {
	engine      proxy.Engine
	desc        pluginapi.PluginDescriptor
	upstreamURL string
}

// NewPlugin constructs a new Plugin instance. upstreamURL is used when the
// adapter performs upstream handoff for the protected endpoint.
func NewPlugin(e proxy.Engine, desc pluginapi.PluginDescriptor, upstreamURL string) *Plugin {
	return &Plugin{engine: e, desc: desc, upstreamURL: upstreamURL}
}

// Handler returns an http.Handler that runs the proxy engine and sets principal
// decision output (X-User-Id, X-Roles, etc.) on the response for KrakenD backends.
func (p *Plugin) Handler(next http.Handler) http.Handler {
	return Handler(p.engine, next, p.upstreamURL)
}

// Descriptor implements pluginapi.Plugin.
func (p *Plugin) Descriptor() pluginapi.PluginDescriptor { return p.desc }

// Health implements pluginapi.Plugin.
func (p *Plugin) Health(ctx context.Context) pluginapi.PluginHealth {
	return pluginapi.PluginHealth{State: pluginapi.PluginStateHealthy}
}

// Serve implements pluginapi.IntegrationPlugin.
// hostCtx is KrakenD-specific; use Handler() to wire the auth bridge into the gateway.
func (p *Plugin) Serve(ctx context.Context, hostCtx any) error {
	_ = ctx
	_ = hostCtx
	return nil
}
