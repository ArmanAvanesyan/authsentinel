package config

import "strings"

// Config holds configuration for the Proxy app (loaded via go-config from file + env).
type Config struct {
	UpstreamURL     string `json:"upstream_url"`
	ProxyPathPrefix string `json:"proxy_path_prefix"`
	RequireAuth     bool   `json:"require_auth"`
	AgentURL        string `json:"agent_url"`
	CookieName      string `json:"cookie_name"`
	HTTPPort        string `json:"http_port"`

	// PipelinePlugins lists pipeline plugin configs (id, type, raw config). Used by proxy startup to configure and enable pipeline plugins from the registry.
	PipelinePlugins []PipelinePluginEntry `json:"pipeline_plugins"`
	// PluginsManifestDir optional directory to discover plugin manifests (JSON). Empty disables filesystem discovery.
	PluginsManifestDir string `json:"plugins_manifest_dir"`

	// AdminSecret if set guards /admin; requests must include header X-Admin-Secret: <value>. Empty disables admin endpoint.
	AdminSecret string `json:"admin_secret"`

	// Header claim mapping
	HeaderUserIDClaim            string `json:"headers_user_id_claim"`
	HeaderEmailClaim             string `json:"headers_email_claim"`
	HeaderNameClaim              string `json:"headers_name_claim"`
	HeaderPreferredUsernameClaim string `json:"headers_preferred_username_claim"`
	HeaderRolesClaim             string `json:"headers_roles_claim"`
	HeaderGroupsClaim            string `json:"headers_groups_claim"`
	HeaderTenantIDClaim          string `json:"headers_tenant_id_claim"`
	HeaderAdminRole              string `json:"headers_admin_role"`
}

// Validate returns an error if required configuration is missing.
func (c *Config) Validate() error {
	if c.UpstreamURL == "" {
		return errMissing("UPSTREAM_URL")
	}
	if c.AgentURL == "" {
		return errMissing("AGENT_URL")
	}
	if !strings.HasPrefix(c.ProxyPathPrefix, "/") {
		c.ProxyPathPrefix = "/" + c.ProxyPathPrefix
	}
	return nil
}

// ApplyDefaults sets default values for optional fields when not set.
func (c *Config) ApplyDefaults() {
	if c.ProxyPathPrefix == "" {
		c.ProxyPathPrefix = "/graphql"
	}
	if c.CookieName == "" {
		c.CookieName = "__Host-ess_session"
	}
	if c.HTTPPort == "" {
		c.HTTPPort = "8081"
	}
	if c.HeaderUserIDClaim == "" {
		c.HeaderUserIDClaim = "sub"
	}
	if c.HeaderEmailClaim == "" {
		c.HeaderEmailClaim = "email"
	}
	if c.HeaderNameClaim == "" {
		c.HeaderNameClaim = "name"
	}
	if c.HeaderPreferredUsernameClaim == "" {
		c.HeaderPreferredUsernameClaim = "preferred_username"
	}
	if c.HeaderRolesClaim == "" {
		c.HeaderRolesClaim = "realm_access.roles"
	}
	if c.HeaderGroupsClaim == "" {
		c.HeaderGroupsClaim = "groups"
	}
	if c.HeaderAdminRole == "" {
		c.HeaderAdminRole = "admin"
	}
}

// PipelinePluginEntry is a single pipeline plugin config entry (id, type, raw map for go-config / pluginconfig).
type PipelinePluginEntry struct {
	ID   string         `json:"id"`
	Type string         `json:"type"`
	Raw  map[string]any `json:"raw"`
}

type configError string

func (e configError) Error() string { return "config: missing " + string(e) }

func errMissing(name string) error { return configError(name) }
