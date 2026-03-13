package config

// Config holds configuration for the Proxy app.
type Config struct {
	// TODO: add upstream config, policy bundle locations, etc.
}

// Load loads configuration from environment variables or files.
func Load() (*Config, error) {
	// TODO: implement configuration loading.
	return &Config{}, nil
}
