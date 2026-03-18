package sdk

import (
	"context"

	"github.com/ArmanAvanesyan/authsentinel/pkg/token"
)

// GraphQLContextKey is the key for storing principal in GraphQL context.
// When using SDK HTTP Middleware, the request context is populated with the principal;
// ensure your GraphQL resolver context is the same as the HTTP request context (e.g. in gqlgen
// pass r.Context() into the resolver context) so GetPrincipalFromGraphQLContext works.
const GraphQLContextKey = "authsentinel.principal"

// GetPrincipalFromGraphQLContext returns the principal from the given context.
// Use this in GraphQL resolvers when the context was derived from the HTTP request
// that passed through SDK Middleware, or when the context was set with WithPrincipal.
func GetPrincipalFromGraphQLContext(ctx context.Context) *token.Principal {
	return PrincipalFromContext(ctx)
}
