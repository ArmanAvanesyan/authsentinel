# Schemas

This directory is the **source of truth** for configuration JSON Schemas and related definitions used across AuthSentinel.

## Layout

| Path | Purpose |
|------|--------|
| **runtime.schema.json** | Base runtime schema: shared concepts (key layout prefixes, cookie/session naming). Used for documentation; validation is per-binary. |
| **agent.schema.json** | Per-binary: AuthSentinel Agent config. Generated from `internal/agent/config` via `make schema`. |
| **proxy.schema.json** | Per-binary: AuthSentinel Proxy config. Generated from `internal/proxy/config` via `make schema`. |
| **plugins/provider/** | Per-plugin: OIDC and other identity provider plugin configs (e.g. `oidc.schema.json`). |
| **plugins/pipeline/** | Per-plugin: Pipeline plugin configs (e.g. `ratelimit.schema.json`). |
| **plugins/integration/** | Per-plugin: Gateway integration configs (Caddy, Traefik, KrakenD). |

## Regenerating binary schemas

From the repo root:

```bash
make schema
```

This runs `cmd/schema`, which uses go-config’s schema generator to produce `agent.schema.json` and `proxy.schema.json` from the Go config structs. Plugin schemas under `schemas/plugins/` are maintained by hand.

## Tooling

- **validate-config**: Validate a config file against the runtime (load + `Validate()`). Use `make validate-config CONFIG_PATH=configs/agent.example.json BINARY=agent`.
- **print-schema**: Print a schema to stdout. Use `make print-schema BINARY=agent` or `make print-schema SCHEMA=schemas/plugins/integration/caddy.schema.json`.
- **render-config-example**: Render example config from defaults (go-config struct + `ApplyDefaults()`). Use `make render-config-example BINARY=agent FORMAT=json`.

See the root **Makefile** for target details.
