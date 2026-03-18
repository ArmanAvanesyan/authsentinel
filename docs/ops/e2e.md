# E2E testing

End-to-end tests run the full stack (agent, proxy, optional upstream and IdP) and assert key flows.

## Smoke E2E (make e2e-docker)

From the repo root, with Docker and docker-compose available:

1. Copy `deployments/docker/.env.example` to `deployments/docker/.env` and set required variables (OIDC_ISSUER, OIDC_REDIRECT_URI, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, SESSION_COOKIE_SIGNING_SECRET, APP_BASE_URL).
2. Run: `make e2e-docker`

This will:

- Start `docker-compose` in `deployments/docker` (Redis, authsentinel-agent, authsentinel-proxy, BFF placeholder).
- Wait for services to be up, then run `test/e2e/playbook.sh`.
- The playbook checks: agent and proxy health endpoints return 200; proxy returns 401 for unauthenticated requests to the proxy path; agent `/login` returns 302 or 200.
- Tear down with `docker-compose down`.

Requirements: `docker`, `docker-compose`, `curl`, and a valid `.env` (a real OIDC IdP for full login, or use placeholder values for smoke-only health checks).

## API-only E2E scenario

For API-only clients (no browser): the client sends a Bearer JWT to the proxy. The proxy validates the token and forwards the request to the upstream with principal-derived headers.

To run this scenario:

1. Start the stack (e.g. `make e2e-docker` without the playbook, or run compose manually).
2. Obtain a JWT from your IdP (or use a test token if the proxy is configured to accept it).
3. `curl -H "Authorization: Bearer <JWT>" http://localhost:8081/graphql` (or your proxy path). Expect 200 and upstream to receive headers such as `X-User-Id`.

Document this flow in your runbook; the same compose and proxy config support both browser (cookie) and API-only (Bearer) flows.

## Caddy / Traefik / KrakenD E2E

Optional: run the same stack with a gateway in front of the proxy. Use the example configs in `configs/plugins/` (Caddy, Traefik, KrakenD). Start the gateway pointing at the proxy URL; run the smoke playbook against the gateway URL (e.g. `PROXY_URL=http://localhost:443` if the gateway listens on 443). See `docs/integration/` for each gateway’s setup.

## CI

- **Unit and contract tests**: Run `make test` (and `make proto-lint`, `make proto-breaking`). No extra services required.
- **Integration tests (Redis)**: Tests in `internal/store/redis` require Redis. Set **REDIS_URL** (e.g. `redis://localhost:6379/1`) in CI to run them; otherwise they are skipped. Optionally use testcontainers to start Redis in CI.
- **E2E (make e2e-docker)**: Optional CI job that runs the smoke playbook against docker-compose. Requires Docker, docker-compose, and a valid `.env` (or skip this job if no OIDC IdP is configured for CI).
