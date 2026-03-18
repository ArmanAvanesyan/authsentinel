package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	jsonformat "github.com/ArmanAvanesyan/go-config/format/json"
	"github.com/ArmanAvanesyan/go-config/source/env"
	"github.com/ArmanAvanesyan/go-config/source/file"
)

func TestValidate_RequiredFields(t *testing.T) {
	cfg := &Config{}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for empty config")
	}
	cfg.OIDCIssuer = "https://idp.example.com"
	cfg.OIDCRedirectURI = "https://app/callback"
	cfg.OIDCClientID = "client"
	cfg.RedisURL = "redis://localhost"
	cfg.CookieSigningSecret = "secret"
	cfg.AppBaseURL = "https://app.example.com"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error: %v", err)
	}
}

func TestKeyLayout(t *testing.T) {
	cfg := &Config{SessionRedisPrefix: "auth", SessionTTLSeconds: 3600, SessionPKCETTLSeconds: 300, SessionRefreshLockTTLSeconds: 15}
	layout := cfg.KeyLayout()
	if layout.SessionPrefix != "auth:session:" {
		t.Errorf("SessionPrefix: got %q", layout.SessionPrefix)
	}
	if layout.PKCEPrefix != "auth:pkce:" {
		t.Errorf("PKCEPrefix: got %q", layout.PKCEPrefix)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("OIDC_ISSUER", "https://idp.example.com")
	os.Setenv("OIDC_REDIRECT_URI", "https://app/callback")
	os.Setenv("OIDC_CLIENT_ID", "client")
	os.Setenv("OIDC_CLIENT_SECRET", "secret")
	os.Setenv("REDIS_URL", "redis://localhost:6379")
	os.Setenv("COOKIE_SIGNING_SECRET", "cookie-secret")
	os.Setenv("APP_BASE_URL", "https://app.example.com")
	defer func() {
		os.Unsetenv("OIDC_ISSUER")
		os.Unsetenv("OIDC_REDIRECT_URI")
		os.Unsetenv("OIDC_CLIENT_ID")
		os.Unsetenv("OIDC_CLIENT_SECRET")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("COOKIE_SIGNING_SECRET")
		os.Unsetenv("APP_BASE_URL")
	}()

	var cfg Config
	err := config.New().AddSource(env.New("")).Load(context.Background(), &cfg)
	if err != nil {
		t.Fatalf("Load from env: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if cfg.OIDCIssuer != "https://idp.example.com" {
		t.Errorf("OIDCIssuer: got %q", cfg.OIDCIssuer)
	}
}

// TestLoadFromFile_AgentExampleJSON runs the same load path as cmd/validateconfig
// (file source + go-config + ApplyDefaults + Validate) for configs/agent.example.json.
func TestLoadFromFile_AgentExampleJSON(t *testing.T) {
	// From internal/agent/config we need ../../../configs to reach repo configs/.
	path := filepath.Join("..", "..", "..", "configs", "agent.example.json")
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	if _, err := os.Stat(path); err != nil {
		t.Skipf("config file not found (run from repo root or ensure configs/ exists): %v", err)
	}
	ctx := context.Background()
	loader := config.New().AddSource(file.New(path), jsonformat.New())
	var cfg Config
	if err := loader.Load(ctx, &cfg); err != nil {
		t.Fatalf("Load config: %v", err)
	}
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if cfg.OIDCIssuer == "" || cfg.AppBaseURL == "" {
		t.Errorf("expected non-empty OIDCIssuer and AppBaseURL from example config")
	}
}
