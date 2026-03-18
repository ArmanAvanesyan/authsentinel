package sdk

import (
	"context"

	"github.com/ArmanAvanesyan/authsentinel/pkg/token"
)

type contextKey int

const (
	principalContextKey contextKey = iota
)

// PrincipalFromContext returns the principal stored in the context by SDK middleware or gRPC interceptor.
// Returns nil if not set.
func PrincipalFromContext(ctx context.Context) *token.Principal {
	p, _ := ctx.Value(principalContextKey).(*token.Principal)
	return p
}

// WithPrincipal returns a context that contains the given principal.
// Used by SDK middleware and gRPC interceptor; typically not called by application code.
func WithPrincipal(ctx context.Context, p *token.Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, p)
}
