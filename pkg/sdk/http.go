package sdk

import (
	"context"
	"net/http"

	"github.com/ArmanAvanesyan/authsentinel/pkg/token"
)

// PrincipalExtractor extracts a principal from an HTTP request (e.g. via session cookie or JWT).
// Implementations may call an auth agent, validate a Bearer token, or read headers.
type PrincipalExtractor interface {
	ExtractPrincipal(ctx context.Context, r *http.Request) (*token.Principal, error)
}

// Middleware returns an HTTP middleware that extracts the principal using the given extractor,
// stores it in the request context, and calls the next handler.
// If requireAuth is true and extraction yields no principal, the middleware responds with 401 Unauthorized
// and does not call next.
func Middleware(extractor PrincipalExtractor, requireAuth bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, err := extractor.ExtractPrincipal(r.Context(), r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if principal != nil {
				r = r.WithContext(WithPrincipal(r.Context(), principal))
			}
			if requireAuth && principal == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// IdentityFromHTTPRequest returns the principal from the request context.
// It only returns a non-nil value when the request has passed through SDK Middleware
// (or the context was otherwise set with WithPrincipal). For direct extraction from
// the request (e.g. without middleware), use an PrincipalExtractor and call
// ExtractPrincipal(ctx, r) on it.
func IdentityFromHTTPRequest(r *http.Request) (*token.Principal, error) {
	return PrincipalFromContext(r.Context()), nil
}
