package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/source/env"
	"gopkg.in/yaml.v3"
)

func TestValidate_RequiredFields(t *testing.T) {
	cfg := &Config{}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for empty config")
	}
	cfg.UpstreamURL = "http://localhost:3000"
	cfg.AgentURL = "http://localhost:8080"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error: %v", err)
	}
}

func TestValidate_ProxyPathPrefix(t *testing.T) {
	cfg := &Config{UpstreamURL: "http://u", AgentURL: "http://a", ProxyPathPrefix: "graphql"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if cfg.ProxyPathPrefix != "/graphql" {
		t.Errorf("expected ProxyPathPrefix to get leading slash: %q", cfg.ProxyPathPrefix)
	}
}

func TestLoadFromEnv(t *testing.T) {
	if err := os.Setenv("UPSTREAM_URL", "http://localhost:3000"); err != nil {
		t.Fatalf("Setenv UPSTREAM_URL: %v", err)
	}
	if err := os.Setenv("AGENT_URL", "http://localhost:8080"); err != nil {
		t.Fatalf("Setenv AGENT_URL: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("UPSTREAM_URL"); err != nil {
			t.Fatalf("Unsetenv UPSTREAM_URL: %v", err)
		}
		if err := os.Unsetenv("AGENT_URL"); err != nil {
			t.Fatalf("Unsetenv AGENT_URL: %v", err)
		}
	}()

	var cfg Config
	err := config.New().AddSource(env.New("")).Load(context.Background(), &cfg)
	if err != nil {
		t.Fatalf("Load from env: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if cfg.UpstreamURL != "http://localhost:3000" {
		t.Errorf("UpstreamURL: got %q", cfg.UpstreamURL)
	}
}

// TestLoadFromFile_ProxyExampleYAML runs the same load path as cmd/validateconfig
// for YAML (read file + YAML unmarshal + ApplyDefaults + Validate) for configs/proxy.example.yaml.
func TestLoadFromFile_ProxyExampleYAML(t *testing.T) {
	// From internal/proxy/config we need ../../../configs to reach repo configs/.
	path := filepath.Join("..", "..", "..", "configs", "proxy.example.yaml")
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("config file not found (run from repo root): %v", err)
	}
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse YAML: %v", err)
	}
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("convert to JSON: %v", err)
	}
	var cfg Config
	if err := json.Unmarshal(jsonBytes, &cfg); err != nil {
		t.Fatalf("decode config: %v", err)
	}
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if cfg.UpstreamURL == "" || cfg.AgentURL == "" {
		t.Errorf("expected non-empty UpstreamURL and AgentURL from example config")
	}
}
