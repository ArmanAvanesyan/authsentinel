package sdk

import (
	"context"
	"strings"

	"github.com/ArmanAvanesyan/authsentinel/pkg/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JWTValidator validates a JWT and returns the principal (e.g. pkg/token.JWTValidator).
type JWTValidator interface {
	ValidateJWT(ctx context.Context, raw string) (*token.Principal, error)
}

// UnaryServerInterceptor returns a gRPC unary interceptor that reads the Authorization header
// (Bearer token), validates it with the given validator, and stores the principal in the context.
// Handlers can retrieve it via PrincipalFromContext or IdentityFromGRPCContext.
// If no Authorization header is present, the request continues with no principal in context.
// If the token is invalid, the interceptor returns Unauthenticated.
func UnaryServerInterceptor(validator JWTValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}
		vals := md.Get("authorization")
		if len(vals) == 0 {
			return handler(ctx, req)
		}
		raw := strings.TrimSpace(vals[0])
		if strings.HasPrefix(strings.ToLower(raw), "bearer ") {
			raw = strings.TrimSpace(raw[7:])
		}
		if raw == "" {
			return handler(ctx, req)
		}
		principal, err := validator.ValidateJWT(ctx, raw)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}
		if principal != nil {
			ctx = WithPrincipal(ctx, principal)
		}
		return handler(ctx, req)
	}
}

// IdentityFromGRPCContext returns the principal stored in the context by UnaryServerInterceptor.
// Returns (nil, nil) if no principal is present.
func IdentityFromGRPCContext(ctx context.Context) (*token.Principal, error) {
	return PrincipalFromContext(ctx), nil
}
