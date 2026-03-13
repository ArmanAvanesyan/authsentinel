# PeopleSuite integration

This document describes how to run AuthSentinel as two separate services (OAuth Agent + OAuth Proxy) via Docker to replace the current PeopleSuite platform-edge auth gateway.

## Architecture

- **authsentinel-agent**: Handles `/auth/*` and `/me`. OIDC PKCE login, session in Redis, session cookie. Endpoints: `/login`, `/callback`, `/session`, `/me`, `/logout`, `/refresh`, and internal `/internal/resolve`, `/internal/session`.
- **authsentinel-proxy**: Proxies a path prefix (e.g. `/graphql`) to the BFF. Resolves session via the agent (`GET /internal/resolve`), adds `Authorization` and `X-User-Id`, `X-Tenant-Id`, etc., forwards to the BFF.

Reverse proxy (Caddy) in front:

- `/auth/*` → agent
- `/graphql`, `/graphql/*` → proxy
- `/me` → agent
- `/*` → SPA

Cookie domain (e.g. `.portal.example.com`) must be shared so agent and proxy receive the same session cookie.

## Deployment

1. Use `deployments/docker/docker-compose.yaml` or equivalent. Ensure `REDIS_URL`, `AGENT_URL`, `UPSTREAM_URL`, and OIDC/COOKIE env vars are set.
2. Configure the reverse proxy as in [Docker / Caddy example](../ops/docker.md).
3. Replace the platform-edge gateway with the agent and proxy; keep the existing BFF. The BFF continues to read `Authorization`, `X-User-Id`, `X-Tenant-Id`, `X-Roles`, etc., from the request headers set by the proxy.

## Tenant and identity (Options A and C)

- **Option A**: Use `POST_LOGIN_WEBHOOK_URL` to notify an enrichment service after login. The enrichment service can call the agent `PATCH /internal/session` to attach `tenant_context` to the session. The proxy then sends `X-Tenant-Id` from the session.
- **Option C**: Resolve tenant at the proxy or BFF from the agent session response and the request `Host` (e.g. tenant slug from hostname). The proxy sets `X-Tenant-Id` when `tenant_context.tenant_id` is present; otherwise the BFF can resolve tenant itself.

## Acceptance

- PeopleSuite can remove the custom platform-edge gateway and use only the `authsentinel-agent` and `authsentinel-proxy` images.
- All behaviour is configurable via environment variables.
- Session and cookie format are documented in [Cookie model](../architecture/cookie-model.md) and [OAuth Agent](../runtime/oauth-agent.md) / [OAuth Proxy](../runtime/oauth-proxy.md).
