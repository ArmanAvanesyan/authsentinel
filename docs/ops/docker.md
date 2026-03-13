# Docker

Dockerfiles for the Agent and Proxy live under `deployments/docker/agent` and `deployments/docker/proxy`. Build from the repository root as build context.

## docker-compose

A minimal compose file is at `deployments/docker/docker-compose.yaml` with:

- `redis`
- `authsentinel-agent` (builds from `deployments/docker/agent/Dockerfile`)
- `authsentinel-proxy` (builds from `deployments/docker/proxy/Dockerfile`)
- `bff` (placeholder; replace with your BFF image)

Set environment variables (e.g. in `.env` or export): `OIDC_ISSUER`, `OIDC_REDIRECT_URI`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `SESSION_COOKIE_SIGNING_SECRET`, `APP_BASE_URL`.

## Reverse proxy (Caddy) example for PeopleSuite

Place Caddy (or another reverse proxy) in front so that the browser hits a single origin (e.g. `https://portal.example.com`). Route by path:

- `https://portal.example.com/auth/*` → `http://authsentinel-agent:8080/*` (strip `/auth` so the agent sees `/login`, `/callback`, etc.)
- `https://portal.example.com/graphql` and `https://portal.example.com/graphql/*` → `http://authsentinel-proxy:8081`
- `https://portal.example.com/me` → `http://authsentinel-agent:8080/me`
- `https://portal.example.com/*` → SPA (or other app)

Example Caddyfile (conceptual):

```caddyfile
portal.example.com {
    handle /auth/* {
        uri strip_prefix /auth
        reverse_proxy authsentinel-agent:8080
    }
    handle /graphql* {
        reverse_proxy authsentinel-proxy:8081
    }
    handle /me {
        reverse_proxy authsentinel-agent:8080
    }
    handle {
        reverse_proxy spa:80
    }
}
```

**Cookie domain**: Set `COOKIE_DOMAIN` to `.portal.example.com` (or leave empty for current host) so that the session cookie is sent to both the agent and the proxy when they are reached via the same host (e.g. through Caddy). Both services must see the same cookie name (`COOKIE_NAME`, e.g. `__Host-ess_session`).

See [PeopleSuite integration](../integration/peoplesuite.md) for the full flow and replacement of the platform-edge gateway.
