# OAuth Agent runtime

Describes the behavior and endpoints of the AuthSentinel OAuth Agent (`pkg/agent` and `cmd/agent` binary).

## HTTP API (path contract)

All paths are relative to the agent base URL. When behind a reverse proxy (e.g. Caddy), route `/auth/*` to the agent and strip the `/auth` prefix so the agent receives `/login`, `/callback`, etc.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/login` | Start login (OIDC PKCE). Query: `redirect_to` (optional). Responds **302** to IdP authorization URL. |
| GET | `/callback` | OIDC callback. Query: `code`, `state` (required); or `error`, `error_description`. Responds **302** to app or error path. |
| POST | `/callback` | Error callback (IdP POST binding). Same redirect behavior. |
| GET | `/session` | Current session. Cookie: session cookie. **200** JSON: `{ "is_authenticated": bool, "user": { ... } \| null }`. |
| GET | `/me` | Current user. **200** JSON user object or **401**. |
| GET / POST | `/logout` | Logout. Query: `redirect_to` (optional). Clears session, **302** to IdP end_session then back to app. |
| GET | `/refresh` | Token refresh. **200** (optional Set-Cookie) or **401**. |

Internal (for proxy / enrichment):

| Method | Path | Description |
|--------|------|-------------|
| GET | `/internal/resolve` | Resolve session for proxy. Cookie required. **200** JSON: `{ "access_token", "claims", "tenant_context" }` or **401**. |
| PATCH / POST | `/internal/session` | Attach tenant_context to session (Option A). Body: `{ "session_id", "tenant_context": { ... } }`. |

## Environment variables

| Variable | Description | Example |
|----------|-------------|---------|
| `OIDC_ISSUER` | IdP issuer URL | `https://auth.example.com/realms/peopleplatform` |
| `OIDC_REDIRECT_URI` | Callback URL | `https://portal.example.com/auth/callback` |
| `OIDC_CLIENT_ID` | OAuth client id | `peoplespace.agent.security` |
| `OIDC_CLIENT_SECRET` | Client secret (optional) | (secret) |
| `OIDC_SCOPES` | Scopes (comma-separated) | `openid,profile` |
| `OIDC_AUDIENCE` | Expected audience (optional) | |
| `OIDC_CLAIMS_SOURCE` | `id_token` or `access_token` | `id_token` |
| `REDIS_URL` | Redis connection | `redis://localhost:6379/0` |
| `SESSION_REDIS_PREFIX` | Key prefix for session keys | `auth` |
| `SESSION_TTL_SECONDS` | Session lifetime | `36000` |
| `SESSION_PKCE_TTL_SECONDS` | PKCE state TTL | `300` |
| `SESSION_REFRESH_LOCK_TTL_SECONDS` | Refresh lock TTL | `15` |
| `SESSION_REFRESH_EARLY_SECONDS` | Refresh when access token expires in less than this | `60` |
| `COOKIE_NAME` | Session cookie name | `__Host-ess_session` |
| `COOKIE_SIGNING_SECRET` | Secret for signing cookie | (long random string) |
| `COOKIE_SECURE` | Set Secure flag | `true` |
| `COOKIE_SAME_SITE` | `lax`, `strict`, `none` | `lax` |
| `COOKIE_DOMAIN` | Optional cookie domain | `` or `.portal.example.com` |
| `APP_BASE_URL` | Base URL of the app | `https://portal.example.com` |
| `LOGIN_ERROR_REDIRECT_PATH` | Path on login error | `/login?error=oidc_error` |
| `ALLOWED_REDIRECT_ORIGINS` | Comma-separated origins for `redirect_to` | `https://portal.example.com` |
| `ALLOWED_REDIRECT_PATHS` | Comma-separated path prefixes | `/` |
| `HTTP_PORT` | Listen port | `8080` |
| `POST_LOGIN_WEBHOOK_URL` | Optional webhook after login | |
| `SESSION_ENRICHMENT_API` | Optional (documentation) | |
| `CORS_ALLOWED_ORIGINS` | Comma-separated CORS origins | |

## Session and Redis key layout

- Session keys: `{SESSION_REDIS_PREFIX}:session:{session_id}`, TTL `SESSION_TTL_SECONDS`.
- PKCE keys: `{SESSION_REDIS_PREFIX}:pkce:{state}`, TTL `SESSION_PKCE_TTL_SECONDS`.
- Refresh lock keys: `{SESSION_REDIS_PREFIX}:refresh_lock:{session_id}`, TTL `SESSION_REFRESH_LOCK_TTL_SECONDS`.

See [Cookie model](../architecture/cookie-model.md) for session record shape and cookie signing.
