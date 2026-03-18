# Go SDK

The Go SDK (**pkg/sdk**) provides server-side helpers to integrate AuthSentinel with your Go services: HTTP middleware for principal extraction, GraphQL and gRPC context, session-aware agent client, and a proxy-backed principal extractor.

## Entrypoints

- **HTTP**: Middleware and principal extraction for BFFs or APIs that sit behind the proxy or call the agent.
- **GraphQL**: Inject principal into GraphQL request context.
- **gRPC**: Unary interceptor and JWT validation for gRPC services.
- **Agent client**: Build login/logout URLs and call `/session` and `/refresh` with the user’s session cookie.
- **Proxy adapter**: Use the proxy’s principal resolution as a `PrincipalExtractor` in SDK middleware.

## HTTP middleware

Use **sdk.Middleware(extractor, requireAuth)** to protect HTTP handlers:

- **PrincipalExtractor**: Extracts a **token.Principal** from the request (e.g. via session cookie by calling the agent, or via Bearer token).
- **requireAuth**: If true, a missing principal yields **401 Unauthorized** and the next handler is not called.
- The principal is stored in the request context; read it with **sdk.PrincipalFromContext(ctx)** or **sdk.IdentityFromHTTPRequest(r)**.

Example (agent-backed BFF): implement **PrincipalExtractor** by reading the session cookie from the request, calling **AgentClient.GetSession(ctx, cookie)**, and mapping **SessionResponse.User** to **token.Principal**; then pass that extractor to **sdk.Middleware(extractor, true)**.

If your app is behind the AuthSentinel **proxy** and the proxy resolves the principal (e.g. via its resolve endpoint), use **sdk.NewProxyPrincipalExtractor(resolver)** where **resolver** is a **proxy.PrincipalResolver** (e.g. **internal/proxy.AgentPrincipalResolver**). The adapter builds a **proxy.Request** from the HTTP request and calls **Resolve**; use with **sdk.Middleware**.

## Context and identity

- **sdk.WithPrincipal(ctx, principal)** — Store principal in context (used by middleware/interceptor).
- **sdk.PrincipalFromContext(ctx)** — Retrieve principal from context; returns nil if not set.
- **sdk.IdentityFromHTTPRequest(r)** — Returns principal from `r.Context()`; use after middleware.

## Agent client

**sdk.AgentClient** talks to the agent’s public endpoints:

- **NewAgentClient(baseURL, cookieName)** — baseURL is the agent base (e.g. `https://auth.example.com`).
- **GetLoginURL(returnURL)** — URL to redirect the user for login.
- **GetLogoutURL(redirectTo)** — URL for logout.
- **GetSession(ctx, sessionCookie)** — GET /session with the given cookie; returns **SessionResponse** (IsAuthenticated, User, optional SetCookie from response header).
- **Refresh(ctx, sessionCookie)** — POST /refresh; returns updated session and optional new Set-Cookie.

Use the agent client in a BFF that forwards the browser’s session cookie to the agent and forwards back any Set-Cookie for session refresh.

## GraphQL

- **sdk.GetPrincipalFromGraphQLContext(ctx)** — Returns the principal from the GraphQL request context. Wire the SDK so that the principal is set on the GraphQL context (e.g. from HTTP middleware that runs before GraphQL).

See **pkg/graphql** for parsing and principal injection into GraphQL requests.

## gRPC

- **sdk.UnaryServerInterceptor(jwtValidator, requireAuth)** — gRPC unary interceptor that validates JWT (or extracts principal) and sets it on context. If requireAuth is true and no principal, returns Unauthenticated.
- **sdk.JWTValidator** — Interface for validating Bearer tokens and returning **token.Principal**.

See **pkg/grpc** for metadata extraction and adapter types.

## Proxy adapter

When your Go service receives requests that have already passed through the AuthSentinel proxy (e.g. proxy sets `X-User-Id`, or you have a **proxy.PrincipalResolver** that calls the proxy’s resolve endpoint):

- **sdk.NewProxyPrincipalExtractor(resolver)** — Wraps **proxy.PrincipalResolver** as **PrincipalExtractor**. Builds **proxy.Request** from the HTTP request (without consuming body) and calls **Resolve**. Use with **sdk.Middleware**.

## References

- [pkg/sdk](https://pkg.go.dev/github.com/ArmanAvanesyan/authsentinel/pkg/sdk) — Go doc.
- [Agent flows](../runtime/agent-flows.md) — Login, callback, session, refresh, logout.
- [Configuration](../ops/configuration.md) — Agent and proxy config (cookie name, agent URL).
