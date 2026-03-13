package krakend

import "github.com/ArmanAvanesyan/authsentinel/pkg/proxy"

// Plugin is a placeholder type for a KrakenD plugin that delegates to the proxy Engine.
type Plugin struct {
	engine proxy.Engine
}

// NewPlugin constructs a new Plugin instance.
func NewPlugin(e proxy.Engine) *Plugin {
	return &Plugin{engine: e}
}
