# Cookie model

The cookie model is defined in `pkg/cookie` and used by the Agent (and optionally the Proxy) to issue and validate session cookies.

## Contract

- **Cookie value**: Opaque signed string containing only the **session ID**. No sensitive data (tokens) is stored in the cookie. The agent signs the session ID with `COOKIE_SIGNING_SECRET` (HMAC-SHA256) so that only the agent can create or validate it.
- **Cookie attributes**: `HttpOnly`, `Secure` (when TLS), `SameSite=Lax` (or configurable), `Path=/`, `Max-Age` aligned with session TTL. `Domain` is configurable (e.g. empty for current host, or `.portal.example.com` for subdomains).
- **Proxy**: The proxy does not decode the cookie; it forwards the cookie to the agent on `GET /internal/resolve`. The agent is the only writer of the cookie.

## Session record (Redis)

The agent stores the full session in Redis keyed by session ID. Record shape (see `pkg/session`):

- `id`: Session ID (opaque).
- `access_token`, `refresh_token`, `id_token`: OAuth tokens.
- `expires_at`: Unix seconds (access token expiry).
- `claims`: Map (e.g. `sub`, `email`, `preferred_username`, `name`, `realm_access.roles`, `groups`) used to build user and upstream headers.
- `tenant_context` (optional): `{ "tenant_id", "tenant_slug", "status", "locale", "timezone" }` for multi-tenancy (Option A: set via PATCH `/internal/session`).

## Key prefixes and TTLs

- Session: `{SESSION_REDIS_PREFIX}:session:{id}`, TTL `SESSION_TTL_SECONDS`.
- PKCE state: `{SESSION_REDIS_PREFIX}:pkce:{state}`, TTL `SESSION_PKCE_TTL_SECONDS`.
- Refresh lock: `{SESSION_REDIS_PREFIX}:refresh_lock:{session_id}`, TTL `SESSION_REFRESH_LOCK_TTL_SECONDS`.

## Tenant options

- **Option A**: Agent supports `POST_LOGIN_WEBHOOK_URL` (POST with session_id, subject, email, claims, host) and `PATCH /internal/session` to attach `tenant_context` to a session. An enrichment service can call the agent after login and then PATCH the session with tenant data.
- **Option C**: Tenant is resolved at the proxy or BFF from the agent’s session response plus request `Host`. The proxy sets `X-Tenant-Id` from `session.tenant_context.tenant_id` when present, or the BFF resolves tenant separately.
