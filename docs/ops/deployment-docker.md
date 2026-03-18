# Deployment with Docker

AuthSentinel runs as two services: **authsentinel-agent** (login, session, refresh, logout) and **authsentinel-proxy** (request pipeline, policy, upstream). This guide covers building and running them with Docker and docker-compose.

## Layout

| Path | Description |
|------|-------------|
| **deployments/docker/agent/Dockerfile** | Agent image. |
| **deployments/docker/proxy/Dockerfile** | Proxy image. |
| **deployments/docker/docker-compose.yaml** | Compose: Redis, agent, proxy, placeholder BFF. |
| **deployments/docker/.env.example** | Example env vars; copy to `.env` and set values. |

Build context for both Dockerfiles is the **repository root** (e.g. `docker build -f deployments/docker/agent/Dockerfile .`).

## Build

From the repo root:

```bash
# Agent
docker build -f deployments/docker/agent/Dockerfile -t authsentinel-agent:latest .

# Proxy
docker build -f deployments/docker/proxy/Dockerfile -t authsentinel-proxy:latest .
```

Compose will build these if you run `docker compose -f deployments/docker/docker-compose.yaml up --build`.

## Environment variables

Copy **deployments/docker/.env.example** to **deployments/docker/.env** and set:

- **OIDC_ISSUER**, **OIDC_REDIRECT_URI**, **OIDC_CLIENT_ID**, **OIDC_CLIENT_SECRET** — required for the agent.
- **SESSION_COOKIE_SIGNING_SECRET** — at least 32 bytes; used to sign session cookies.
- **APP_BASE_URL** — base URL of the app (e.g. `https://portal.example.com`).

Optional:

- **CONFIG_PATH** / **AGENT_CONFIG** / **PROXY_CONFIG** — config file path (see [Configuration](configuration.md)).
- **ADMIN_SECRET** — if set, enables `GET /admin` on agent and proxy (guarded by `X-Admin-Secret`).

Compose passes agent and proxy env from the compose file (and can override with `.env`). Proxy needs **AGENT_URL**, **UPSTREAM_URL**, **COOKIE_NAME** to match the agent; see the compose file and [Configuration](configuration.md).

## Run with docker-compose

From the repo root:

```bash
cd deployments/docker
cp .env.example .env
# Edit .env with your OIDC and secrets

docker compose up -d
```

Services:

- **redis** — Redis (e.g. port 6379).
- **authsentinel-agent** — Agent (e.g. port 8080).
- **authsentinel-proxy** — Proxy (e.g. port 8081).
- **bff** — Placeholder; replace with your BFF image and command.

## Health checks

Use standard health endpoints for liveness and readiness:

| Service | Port | Liveness | Readiness |
|---------|------|----------|-----------|
| Agent   | 8080 | GET /livez | GET /readyz (checks Redis if configured) |
| Proxy   | 8081 | GET /livez | GET /readyz |

Example:

```bash
curl -s http://localhost:8080/livez   # agent
curl -s http://localhost:8081/readyz  # proxy
```

In Docker or Kubernetes, set liveness/readiness probes to these URLs. See [Health, metrics, and tracing](health-metrics-tracing.md).

## E2E smoke

From the repo root:

```bash
make e2e-docker
```

This starts compose, runs the E2E playbook (health checks, unauthenticated proxy 401), and tears down. See [E2E](e2e.md) and **test/e2e/playbook.sh**.

## Reverse proxy in front

In production, put a reverse proxy (e.g. Caddy, Traefik) in front so the browser hits a single origin. Example (Caddy):

- `/auth/*` → agent (strip prefix so agent sees `/login`, `/callback`, etc.).
- `/graphql` (or your API path) → proxy.
- `/session`, `/refresh`, `/logout` → agent.
- Everything else → app or SPA.

Set **COOKIE_DOMAIN** (e.g. `.portal.example.com`) so the session cookie is sent to both agent and proxy when they are behind the same host. Both must use the same **COOKIE_NAME**. See [Configuration](configuration.md) and [Docker](docker.md) for a Caddy example.

## References

- [Configuration](configuration.md) — go-config, file vs env, configs/ and schemas.
- [Health, metrics, and tracing](health-metrics-tracing.md) — /healthz, /readyz, /metrics, /admin.
- [deployments/README.md](../../deployments/README.md) — Overview of deployment artifacts.
- [Docker](docker.md) — Additional Caddy/PeopleSuite routing example.
