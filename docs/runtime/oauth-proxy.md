# OAuth Proxy runtime

Describes the behavior of the AuthSentinel OAuth Proxy (`cmd/proxy` binary and `internal/proxy`).

## Role

- Receives requests for a configured path prefix (e.g. `/graphql`, `/graphql/*`).
- Resolves session by calling the agent `GET /internal/resolve` with the session cookie.
- Adds to the request to upstream: `Authorization: Bearer <access_token>` and custom headers from session claims.
- Forwards the request to the configured upstream (BFF). Returns upstream response to the client.

## Upstream headers (PeopleSuite BFF)

The proxy sets (configurable via env):

| Header | Source |
|--------|--------|
| `Authorization` | `Bearer ` + session `access_token` |
| `X-User-Id` | Claim (default `sub`) |
| `X-User-Email` | Claim (default `email`) |
| `X-User-Full-Name` | Claim (default `name`) |
| `X-User-Preferred-Username` | Claim (default `preferred_username`) |
| `X-Roles` | Comma-separated from claim path (default `realm_access.roles`) |
| `X-Groups` | Comma-separated from `groups` |
| `X-Is-Admin` | `true` or `false` from role list |
| `X-Tenant-Id` | `tenant_context.tenant_id` or claim path |

## Environment variables

| Variable | Description | Example |
|----------|-------------|---------|
| `UPSTREAM_URL` | Base URL of upstream (BFF) | `http://bff:8002` |
| `PROXY_PATH_PREFIX` | Path prefix to proxy | `/graphql` |
| `REQUIRE_AUTH` | Require valid session for proxied requests | `true` |
| `AGENT_URL` | Base URL of authsentinel-agent | `http://authsentinel-agent:8080` |
| `COOKIE_NAME` | Session cookie name (same as agent) | `__Host-ess_session` |
| `HTTP_PORT` | Listen port | `8081` |
| `HEADERS_USER_ID_CLAIM` | Claim key for user id | `sub` |
| `HEADERS_EMAIL_CLAIM` | Claim key for email | `email` |
| `HEADERS_NAME_CLAIM` | Claim key for full name | `name` |
| `HEADERS_PREFERRED_USERNAME_CLAIM` | Claim key | `preferred_username` |
| `HEADERS_ROLES_CLAIM` | Claim path for roles (e.g. `realm_access.roles` or `roles`) | |
| `HEADERS_GROUPS_CLAIM` | Claim path for groups | `groups` |
| `HEADERS_TENANT_ID_CLAIM` | Claim or session field for tenant id | |
| `HEADERS_ADMIN_ROLE` | Role name that sets X-Is-Admin=true | `admin` |

## Auth requirement

- If `REQUIRE_AUTH=true` and no valid session (agent resolve returns 401), the proxy responds **401** with body `{"errors":[{"message":"unauthorized"}]}` (GraphQL-style).
- If auth is not required, requests without a session are forwarded without identity headers; the upstream may still enforce auth.

## Session resolution (Option 2)

The proxy calls the agent `GET {AGENT_URL}/internal/resolve` with the session cookie. The agent returns JSON with `access_token`, `claims`, and `tenant_context`. The proxy does not read Redis directly; the agent owns all session and refresh logic.
