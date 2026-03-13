# Monorepo structure

This repository uses a standard Go layout at the root plus non-Go packages:

- **cmd/** — entrypoints: `cmd/agent`, `cmd/proxy`.
- **internal/** — private app code: `internal/agent`, `internal/proxy` (each with httpserver and config).
- **pkg/** — public Go packages (cookie, token, policy, proxy, agent, graphql, grpc, sdk, config, testing, plugins).
- **packages/** — non-Go: `packages/js`, `packages/flutter`, `packages/policies`.
- **proto/**, **docs/**, **deployments/** — APIs, documentation, and deployment assets.

See the root `README.md` for build commands and a high-level map.
