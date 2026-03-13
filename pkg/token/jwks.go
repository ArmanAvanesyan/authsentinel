package token

import "context"

// JWKSSource provides raw JWKS for JWT verification.
// TODO: implement fetch and cache.
type JWKSSource interface {
	GetJWKS(ctx context.Context, issuer string) ([]byte, error)
}
