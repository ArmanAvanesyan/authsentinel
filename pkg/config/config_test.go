package config

import "testing"

func TestLoadReturnsSharedWithoutError(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatalf("Load() returned nil Shared")
	}
}
