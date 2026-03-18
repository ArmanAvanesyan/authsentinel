# Agent flows (login, callback, refresh, logout)

This document describes the OAuth Agent flows and how they map to the **pkg/agent** contract. The implementation lives in **internal/agent** (service and OIDC client).

## Flow overview

- **Browser → Agent**: login start, callback, session, refresh, logout.
- **Cookie**: signed session ID; session body lives in Redis (see [Cookie model](../architecture/cookie-model.md)).
- **Proxy**: resolves session via `GET /internal/resolve` and enforces policy before forwarding to upstream.

## Login start

- **Contract**: `Service.LoginStart(ctx, LoginStartRequest) (*LoginStartResponse, error)`.
- **HTTP**: `GET /login` (optional query: `redirect_to`).
- **Behavior**: Agent builds the IdP authorization URL with PKCE (code_verifier, code_challenge), state (CSRF), and nonce. It stores PKCE state and nonce in Redis (PKCE store) keyed by state. Response is **302** redirect to the IdP authorization URL.
- **CSRF / state**: The `state` parameter is generated and stored; the callback must present the same state so the agent can look up PKCE and nonce and prevent cross-site request forgery.

## Callback

- **Contract**: `Service.LoginEnd(ctx, LoginEndRequest) (*LoginEndResponse, error)`.
- **HTTP**: `GET /callback` (or POST for error binding). Query: `code`, `state` (required); or `error`, `error_description` from IdP.
- **Behavior**: Agent validates `state` (lookup PKCE state), exchanges `code` for tokens at the IdP token endpoint (using code_verifier), validates the ID token (issuer, audience, nonce). It creates a session (tokens, claims), stores it in Redis, sets the session cookie, and redirects to the app (or `redirect_to`). On error, redirects to the configured error path.
- **Nonce**: Validated against the value stored at login start to prevent token replay.

## Session bootstrap

- **Contract**: `Service.Session(ctx, SessionRequest) (*SessionResponse, error)`.
- **HTTP**: `GET /session`. Cookie: session cookie (signed session ID).
- **Behavior**: Agent reads the session ID from the cookie, loads the session from Redis, and returns `is_authenticated` and user payload. Used by the app to bootstrap UI state.

## Refresh

- **Contract**: `Service.Refresh(ctx, RefreshRequest) (*RefreshResponse, error)`.
- **HTTP**: `GET /refresh` (or POST). Cookie: session cookie.
- **Behavior**: Agent obtains a refresh lock (per session) to avoid concurrent refresh, exchanges the refresh token for new tokens at the IdP, updates the session in Redis, and may set a new session cookie. On failure (e.g. refresh token expired), returns **401**.

## Logout

- **Contract**: `Service.Logout(ctx, LogoutRequest) (*LogoutResponse, error)`.
- **HTTP**: `GET /logout` or `POST /logout`. Optional query: `redirect_to`.
- **Behavior**: Agent clears the session cookie and optionally redirects to the IdP end_session endpoint, then back to the app (or `redirect_to`). Session can be deleted from Redis; for “logout everywhere” or token revocation, a revocation list (e.g. JTI/session ID in Redis) can be used in addition (see internal/store/redis SetRevoked/IsRevoked).

## Types (pkg/agent)

Request/response types for these flows are defined in **pkg/agent**:

- `LoginStartRequest`, `LoginStartResponse`
- `LoginEndRequest`, `LoginEndResponse`
- `SessionRequest`, `SessionResponse`
- `RefreshRequest`, `RefreshResponse`
- `LogoutRequest`, `LogoutResponse`

Implementation details (HTTP handlers, OIDC client, Redis keys, cookie signing) are in **internal/agent** and **internal/store/redis**.
