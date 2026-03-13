package token

import "time"

// Principal is the normalized identity model extracted from tokens.
type Principal struct {
	Subject   string
	Scopes    []string
	Roles     []string
	Claims    map[string]any
	ExpiresAt time.Time
}
