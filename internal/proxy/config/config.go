package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds configuration for the Proxy app (from environment).
type Config struct {
	UpstreamURL     string // UPSTREAM_URL (BFF base URL)
	ProxyPathPrefix string // PROXY_PATH_PREFIX, e.g. /graphql
	RequireAuth     bool   // REQUIRE_AUTH
	AgentURL        string // AGENT_URL (authsentinel-agent base for session/refresh)
	CookieName      string // COOKIE_NAME (same as agent)
	HTTPPort        string // HTTP_PORT, default 8081

	// Header claim mapping (claim path or key for upstream headers)
	HeaderUserIDClaim            string // HEADERS_USER_ID_CLAIM, default "sub"
	HeaderEmailClaim             string // HEADERS_EMAIL_CLAIM, default "email"
	HeaderNameClaim              string // HEADERS_NAME_CLAIM, default "name"
	HeaderPreferredUsernameClaim string // HEADERS_PREFERRED_USERNAME_CLAIM
	HeaderRolesClaim             string // HEADERS_ROLES_CLAIM, e.g. "realm_access.roles" or "roles"
	HeaderGroupsClaim            string // HEADERS_GROUPS_CLAIM, default "groups"
	HeaderTenantIDClaim          string // HEADERS_TENANT_ID_CLAIM or from session tenant_context
	HeaderAdminRole              string // role name that sets X-Is-Admin=true
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	c := &Config{
		UpstreamURL:                  os.Getenv("UPSTREAM_URL"),
		ProxyPathPrefix:              envOrDefault("PROXY_PATH_PREFIX", "/graphql"),
		RequireAuth:                  envBool("REQUIRE_AUTH", true),
		AgentURL:                     os.Getenv("AGENT_URL"),
		CookieName:                   envOrDefault("COOKIE_NAME", "__Host-ess_session"),
		HTTPPort:                     envOrDefault("HTTP_PORT", "8081"),
		HeaderUserIDClaim:            envOrDefault("HEADERS_USER_ID_CLAIM", "sub"),
		HeaderEmailClaim:             envOrDefault("HEADERS_EMAIL_CLAIM", "email"),
		HeaderNameClaim:              envOrDefault("HEADERS_NAME_CLAIM", "name"),
		HeaderPreferredUsernameClaim: envOrDefault("HEADERS_PREFERRED_USERNAME_CLAIM", "preferred_username"),
		HeaderRolesClaim:             envOrDefault("HEADERS_ROLES_CLAIM", "realm_access.roles"),
		HeaderGroupsClaim:            envOrDefault("HEADERS_GROUPS_CLAIM", "groups"),
		HeaderTenantIDClaim:          os.Getenv("HEADERS_TENANT_ID_CLAIM"),
		HeaderAdminRole:              envOrDefault("HEADERS_ADMIN_ROLE", "admin"),
	}
	return c, nil
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

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

type configError string

func (e configError) Error() string { return "config: missing " + string(e) }

func errMissing(name string) error { return configError(name) }
