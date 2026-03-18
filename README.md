AuthSentinel Monorepo
======================

AuthSentinel is a security runtime that provides an OAuth Agent, OAuth Proxy, embedded policy engine, and SDKs for multiple platforms.

This repository is a single monorepo that contains:

- **Go code** in standard layout: `cmd/` (agent, proxy binaries), `internal/` (app-specific code), `pkg/` (cookie, token, policy, proxy, agent, graphql, grpc, sdk, testing, plugins). Configuration is loaded via **go-config**; example configs are in `configs/`.
- JS/TS SDKs under `packages/js/*`
- Flutter SDK under `packages/flutter/*`
- Gateway plugins under `pkg/plugins/*` (caddy, traefik, krakend)
- Policy bundles under `packages/policies/*`
- Protobuf APIs under `proto/*`
- Documentation under `docs/*`
- Docker builds under `deployments/docker/agent` and `deployments/docker/proxy`

Build from repo root:

- `go build -o authsentinel-agent ./cmd/agent`
- `go build -o authsentinel-proxy ./cmd/proxy`

For a detailed overview of the architecture and package boundaries, see `docs/architecture/overview.md` and `docs/architecture/package-boundaries.md`.
