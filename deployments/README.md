# Deployments

This directory contains deployment artifacts for AuthSentinel.

## Docker

- **deployments/docker/** — Agent and proxy Dockerfiles, `docker-compose.yaml`, and `.env.example`.
  - Copy `.env.example` to `.env` and set OIDC, Redis, and cookie secrets.
  - Compose runs Redis, authsentinel-agent, authsentinel-proxy, and a placeholder BFF service.
  - Health: use `GET /livez` or `GET /readyz` on agent (port 8080) and proxy (port 8081).

### E2E smoke

From the repo root, run `make e2e-docker` to start compose, run the E2E smoke playbook (health checks and unauthenticated proxy 401), then tear down. See `docs/ops/e2e.md` for details and API-only / gateway E2E scenarios.

Helm/Kubernetes manifests can be added later if needed.

