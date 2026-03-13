package config

// Shared holds configuration shared across Agent and Proxy.
// TODO: add common fields (logging, metrics, etc.) and loader.
type Shared struct{}

// Load loads shared configuration.
func Load() (*Shared, error) {
	return &Shared{}, nil
}
