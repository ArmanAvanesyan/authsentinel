package config

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ArmanAvanesyan/authsentinel/pkg/session"
)

// Config holds configuration for the Agent app (from environment).
type Config struct {
	// OIDC
	OIDCIssuer       string   // OIDC_ISSUER
	OIDCRedirectURI  string   // OIDC_REDIRECT_URI
	OIDCClientID     string   // OIDC_CLIENT_ID
	OIDCClientSecret string   // OIDC_CLIENT_SECRET
	OIDCScopes       []string // OIDC_SCOPES (comma-separated)
	OIDCAudience     string   // OIDC_AUDIENCE (optional)
	OIDCClaimsSource string   // OIDC_CLAIMS_SOURCE: "id_token" or "access_token", default id_token

	// Redis
	RedisURL string // REDIS_URL

	// Session key layout and TTLs
	SessionRedisPrefix           string // SESSION_REDIS_PREFIX, default "auth"
	SessionTTLSeconds            int    // SESSION_TTL_SECONDS, default 36000
	SessionPKCETTLSeconds        int    // SESSION_PKCE_TTL_SECONDS, default 300
	SessionRefreshLockTTLSeconds int    // SESSION_REFRESH_LOCK_TTL_SECONDS, default 15
	SessionRefreshEarlySeconds   int    // SESSION_REFRESH_EARLY_SECONDS, default 60

	// Cookie
	CookieName          string        // COOKIE_NAME, e.g. __Host-ess_session
	CookieSigningSecret string        // COOKIE_SIGNING_SECRET
	CookieSecure        bool          // COOKIE_SECURE, default true
	CookieSameSite      http.SameSite // COOKIE_SAME_SITE: lax, strict, none
	CookieDomain        string        // COOKIE_DOMAIN (optional)

	// App and redirects
	AppBaseURL             string   // APP_BASE_URL, e.g. https://portal.example.com
	LoginErrorRedirectPath string   // LOGIN_ERROR_REDIRECT_PATH, e.g. /login?error=oidc_error
	AllowedRedirectOrigins []string // ALLOWED_REDIRECT_ORIGINS (comma-separated)
	AllowedRedirectPaths   []string // ALLOWED_REDIRECT_PATHS (comma-separated path prefixes)

	// HTTP
	HTTPPort string // HTTP_PORT, default 8080

	// Optional: webhook and enrichment
	PostLoginWebhookURL  string // POST_LOGIN_WEBHOOK_URL
	SessionEnrichmentAPI string // SESSION_ENRICHMENT_API (e.g. PATCH session for tenant)

	// CORS
	CORSAllowedOrigins []string // CORS_ALLOWED_ORIGINS (comma-separated)
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	c := &Config{
		OIDCIssuer:                   os.Getenv("OIDC_ISSUER"),
		OIDCRedirectURI:              os.Getenv("OIDC_REDIRECT_URI"),
		OIDCClientID:                 os.Getenv("OIDC_CLIENT_ID"),
		OIDCClientSecret:             os.Getenv("OIDC_CLIENT_SECRET"),
		OIDCScopes:                   splitTrim(os.Getenv("OIDC_SCOPES"), ","),
		OIDCAudience:                 os.Getenv("OIDC_AUDIENCE"),
		OIDCClaimsSource:             envOrDefault("OIDC_CLAIMS_SOURCE", "id_token"),
		RedisURL:                     os.Getenv("REDIS_URL"),
		SessionRedisPrefix:           envOrDefault("SESSION_REDIS_PREFIX", "auth"),
		SessionTTLSeconds:            envInt("SESSION_TTL_SECONDS", 36000),
		SessionPKCETTLSeconds:        envInt("SESSION_PKCE_TTL_SECONDS", 300),
		SessionRefreshLockTTLSeconds: envInt("SESSION_REFRESH_LOCK_TTL_SECONDS", 15),
		SessionRefreshEarlySeconds:   envInt("SESSION_REFRESH_EARLY_SECONDS", 60),
		CookieName:                   envOrDefault("COOKIE_NAME", "__Host-ess_session"),
		CookieSigningSecret:          os.Getenv("COOKIE_SIGNING_SECRET"),
		CookieSecure:                 envBool("COOKIE_SECURE", true),
		CookieSameSite:               parseSameSite(os.Getenv("COOKIE_SAME_SITE")),
		CookieDomain:                 os.Getenv("COOKIE_DOMAIN"),
		AppBaseURL:                   os.Getenv("APP_BASE_URL"),
		LoginErrorRedirectPath:       envOrDefault("LOGIN_ERROR_REDIRECT_PATH", "/login?error=oidc_error"),
		AllowedRedirectOrigins:       splitTrim(os.Getenv("ALLOWED_REDIRECT_ORIGINS"), ","),
		AllowedRedirectPaths:         splitTrim(os.Getenv("ALLOWED_REDIRECT_PATHS"), ","),
		HTTPPort:                     envOrDefault("HTTP_PORT", "8080"),
		PostLoginWebhookURL:          os.Getenv("POST_LOGIN_WEBHOOK_URL"),
		SessionEnrichmentAPI:         os.Getenv("SESSION_ENRICHMENT_API"),
		CORSAllowedOrigins:           splitTrim(os.Getenv("CORS_ALLOWED_ORIGINS"), ","),
	}
	if len(c.OIDCScopes) == 0 {
		c.OIDCScopes = []string{"openid", "profile"}
	}
	if len(c.AllowedRedirectPaths) == 0 {
		c.AllowedRedirectPaths = []string{"/"}
	}
	return c, nil
}

// Validate returns an error if required configuration is missing.
func (c *Config) Validate() error {
	if c.OIDCIssuer == "" {
		return errMissing("OIDC_ISSUER")
	}
	if c.OIDCRedirectURI == "" {
		return errMissing("OIDC_REDIRECT_URI")
	}
	if c.OIDCClientID == "" {
		return errMissing("OIDC_CLIENT_ID")
	}
	if c.RedisURL == "" {
		return errMissing("REDIS_URL")
	}
	if c.CookieSigningSecret == "" {
		return errMissing("COOKIE_SIGNING_SECRET")
	}
	if c.AppBaseURL == "" {
		return errMissing("APP_BASE_URL")
	}
	if c.OIDCClaimsSource != "id_token" && c.OIDCClaimsSource != "access_token" {
		c.OIDCClaimsSource = "id_token"
	}
	return nil
}

// KeyLayout returns the session key layout for Redis.
func (c *Config) KeyLayout() session.KeyLayout {
	p := c.SessionRedisPrefix
	if p == "" {
		p = "auth"
	}
	return session.KeyLayout{
		SessionPrefix:         p + ":session:",
		PKCEPrefix:            p + ":pkce:",
		RefreshLockPrefix:     p + ":refresh_lock:",
		SessionTTLSeconds:     c.SessionTTLSeconds,
		PKCETTLSeconds:        c.SessionPKCETTLSeconds,
		RefreshLockTTLSeconds: c.SessionRefreshLockTTLSeconds,
	}
}

func splitTrim(s, sep string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
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

func parseSameSite(s string) http.SameSite {
	switch strings.ToLower(s) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

type configError string

func (e configError) Error() string { return "config: missing " + string(e) }

func errMissing(name string) error { return configError(name) }
