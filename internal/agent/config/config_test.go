package config

import "testing"

func TestLoadReturnsConfig(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg == nil {
		t.Fatalf("expected non-nil Config")
	}
}
