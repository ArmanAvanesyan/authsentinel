# AuthSentinel Documentation

High-level and deep-dive documentation lives under this directory.

- **Big picture**: `architecture/overview.md`
- **Implementation plan**: `architecture/implementation-plan.md` — consolidated roadmap, phases, milestones, and technology choices (config, WASM, plugins, SDKs).
- **Architecture**: `architecture/runtime-map.md` (package responsibilities and allowed imports), `architecture/plugin-model.md` (plugin kinds, lifecycle, registry).
- **Component behavior**: `runtime/*` — agent flows, proxy pipeline, policy engine, gRPC, GraphQL.
- **Integration**: `integration/*` — Caddy, Traefik, KrakenD, compatibility matrix.
- **SDK**: `sdk/go.md`, `sdk/js.md`, `sdk/react.md`, `sdk/flutter.md`.
- **Ops**: `ops/configuration.md`, `ops/deployment-docker.md`, `ops/health-metrics-tracing.md`, `ops/e2e.md`.

