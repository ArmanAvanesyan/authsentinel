package pluginapi

import (
	"context"

	"github.com/ArmanAvanesyan/authsentinel/pkg/policy"
	"github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
	"github.com/ArmanAvanesyan/authsentinel/pkg/token"
)

// VersionInfo describes the compatibility between a plugin and the host runtime.
type VersionInfo struct {
	// APIVersion is the semantic version of the plugin API this plugin was built against.
	APIVersion string
	// MinRuntimeVersion and MaxRuntimeVersion describe the supported AuthSentinel runtime versions.
	MinRuntimeVersion string
	MaxRuntimeVersion string
}

// PluginKind identifies the high-level class of a plugin.
type PluginKind string

const (
	PluginKindPipeline    PluginKind = "pipeline"
	PluginKindProvider    PluginKind = "provider"
	PluginKindIntegration PluginKind = "integration"
)

// Capability is a symbolic name that describes a specific behavior a plugin provides.
// Examples:
//   - "pipeline:ratelimit"
//   - "pipeline:headers"
//   - "provider:oidc"
//   - "integration:krakend"
type Capability string

// PluginID is the stable identifier for a plugin instance.
type PluginID string

// PluginState represents the lifecycle state of a plugin instance.
type PluginState string

const (
	PluginStateDiscovered  PluginState = "discovered"
	PluginStateVerified    PluginState = "verified"
	PluginStateRegistered  PluginState = "registered"
	PluginStateConfigured  PluginState = "configured"
	PluginStateInitialized PluginState = "initialized"
	PluginStateStarted     PluginState = "started"
	PluginStateHealthy     PluginState = "healthy"
	PluginStateDegraded    PluginState = "degraded"
	PluginStateStopped     PluginState = "stopped"
)

// PluginHealth describes the health of a plugin as reported by the plugin itself.
type PluginHealth struct {
	State   PluginState
	Message string
	// Details may contain arbitrary plugin-specific health metadata (e.g. counters, recent errors).
	Details map[string]any
}

// PluginDescriptor is the static description of a plugin implementation.
type PluginDescriptor struct {
	ID           PluginID
	Kind         PluginKind
	Name         string
	Description  string
	Version      string
	Capabilities []Capability

	// DependsOn lists required plugins or capabilities that must be available before this plugin can start.
	DependsOn []Capability

	// ConfigSchemaRef is an optional reference to a JSON Schema under schemas/plugins/**.
	ConfigSchemaRef string

	VersionInfo VersionInfo
}

// Plugin is the base interface implemented by all plugins.
type Plugin interface {
	Descriptor() PluginDescriptor
	Health(ctx context.Context) PluginHealth
}

// ConfigurablePlugin is implemented by plugins that accept configuration.
// The host passes a typed configuration struct (loaded by go-config) specific to the plugin.
type ConfigurablePlugin interface {
	Configure(ctx context.Context, cfg any) error
}

// StartablePlugin is implemented by plugins that need explicit start/stop hooks.
type StartablePlugin interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// PipelinePlugin is a plugin that participates in the proxy request pipeline.
// Implementations should be side-effect free aside from interacting with the request/response.
type PipelinePlugin interface {
	Plugin

	// Handle is invoked with the normalized proxy.Request and an optional already-resolved principal.
	// It returns an updated request/response decision or an error.
	Handle(ctx context.Context, req *proxy.Request, principal *token.Principal) (*policy.Decision, error)
}

// ProviderPlugin encapsulates interactions with an identity provider (e.g. OIDC, Keycloak, Authentik).
// It is used by the agent runtime to drive browser-based login flows and session establishment.
type ProviderPlugin interface {
	Plugin

	// AuthorizationURL builds the authorization redirect URL for a login start request.
	AuthorizationURL(ctx context.Context, state string, nonce string, extraParams map[string]string) (string, error)

	// ExchangeCode exchanges an authorization code for tokens and returns a normalized principal.
	ExchangeCode(ctx context.Context, code string, redirectURI string) (*token.Principal, error)

	// Refresh refreshes the principal using a refresh token or session information, if supported.
	Refresh(ctx context.Context, principal *token.Principal) (*token.Principal, error)
}

// IntegrationPlugin represents a gateway integration (e.g. Caddy, Traefik, KrakenD).
// Implementations are responsible for wiring a configured proxy engine into the host gateway.
type IntegrationPlugin interface {
	Plugin

	// Serve attaches the integration to the underlying gateway. The concrete type of hostCtx
	// is gateway-specific (e.g. *caddy.Controller, traefik middleware context, etc.).
	Serve(ctx context.Context, hostCtx any) error
}

