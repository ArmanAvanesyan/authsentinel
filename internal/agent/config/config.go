package config

// Config holds configuration for the Agent app.
type Config struct {
	// TODO: add OIDC issuer, client ID/secret, redirect URIs, etc.
}

// Load loads configuration from environment variables or files.
func Load() (*Config, error) {
	// TODO: implement configuration loading.
	return &Config{}, nil
}
