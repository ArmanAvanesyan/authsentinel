# Configuration

AuthSentinel loads configuration using the external [go-config](https://github.com/ArmanAvanesyan/go-config) library. No in-tree config loader exists; the agent and proxy binaries use go-config to merge file and environment sources.

## Precedence

Sources are applied in order; later sources override earlier ones:

1. **Config file** (optional)
2. **Environment variables**

So environment variables override values from the config file. Keys in the merged tree use **lowercase with underscores** (e.g. `oidc_issuer`, `redis_url`), matching the standard env style (e.g. `OIDC_ISSUER`, `REDIS_URL`).

## Config file

- **Agent**: set `CONFIG_PATH` or `AGENT_CONFIG` to the path of a JSON config file (e.g. `configs/agent.example.json`).
- **Proxy**: set `CONFIG_PATH` or `PROXY_CONFIG` to the path of a JSON config file (e.g. `configs/proxy.example.json`).

If no file path is set, only environment variables are used. This keeps env-only deployments (e.g. Docker with `environment:` or `env_file`) working.

## Example configs

| File | Use |
|------|-----|
| `configs/agent.example.json` | Agent: OIDC, Redis, session, cookie, HTTP, CORS. |
| `configs/proxy.example.json` | Proxy: upstream URL, agent URL, header claim mapping. |
| `configs/agent.example.yaml` | Same as agent JSON (for YAML-capable loaders). |
| `configs/proxy.example.yaml` | Same as proxy JSON (for YAML-capable loaders). Includes policy engine keys. |

Copy an example and adjust values. Required fields are validated at startup; see the example files and `internal/agent/config` / `internal/proxy/config` for the full set of keys.

## Running with a config file

**Agent (file + env):**

```bash
export CONFIG_PATH=configs/agent.example.json
# Override specific values with env
export OIDC_ISSUER=https://your-idp.example.com
export REDIS_URL=redis://localhost:6379
./authsentinel-agent
```

**Proxy (file + env):**

```bash
export PROXY_CONFIG=configs/proxy.example.json
export AGENT_URL=http://localhost:8080
export UPSTREAM_URL=http://localhost:3000
./authsentinel-proxy
```

## Env-only (no file)

Omit `CONFIG_PATH` / `AGENT_CONFIG` / `PROXY_CONFIG` and set all required variables in the environment. Same variable names as before (e.g. `OIDC_ISSUER`, `REDIS_URL`, `COOKIE_SIGNING_SECRET`, `APP_BASE_URL` for the agent; `UPSTREAM_URL`, `AGENT_URL` for the proxy).

## Schema and tooling

- **Schemas**: `schemas/` holds JSON Schema for agent and proxy configs (and plugin configs under `schemas/plugins/`). Regenerate binary schemas with `make schema`.
- **Validate config**: `make validate-config CONFIG_PATH=configs/agent.example.json BINARY=agent` (or `BINARY=proxy`). Supports JSON and YAML; uses the same load and `Validate()` as the runtime.
- **Print schema**: `make print-schema BINARY=agent` or `make print-schema SCHEMA=schemas/plugins/integration/caddy.schema.json`.
- **Render example**: `make render-config-example BINARY=agent FORMAT=json` outputs an example config from defaults (go-config struct + `ApplyDefaults()`).

See [schemas/README.md](../../schemas/README.md) and [configs/README.md](../../configs/README.md) for layout and usage.

## References

- [go-config](https://github.com/ArmanAvanesyan/go-config) — config loading, merge, and decode.
- [configs/README.md](../../configs/README.md) — location of example configs in this repo.
