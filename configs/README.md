# Configurations

Example and template configuration files for AuthSentinel. Configuration is loaded via [go-config](https://github.com/ArmanAvanesyan/go-config) (file + env).

## Binary configs

- **agent.example.json** / **agent.example.yaml** — Agent (OIDC, session, cookie, Redis).
- **proxy.example.json** / **proxy.example.yaml** — Proxy (upstream, agent URL, header mapping).

Optional **dev/prod variants** (YAML):

- **agent.dev.yaml**, **agent.prod.yaml** — Agent with dev- or prod-oriented defaults (e.g. `cookie_secure`, redirects).
- **proxy.dev.yaml**, **proxy.prod.yaml** — Proxy with local vs production URLs.

## Plugin examples

- **plugins/caddy.Caddyfile**, **plugins/caddy.example.json** — Caddy forward-auth and optional directive.
- **plugins/traefik.example.yaml** — Traefik ForwardAuth dynamic config.
- **plugins/krakend.example.json** — KrakenD endpoint/auth plugin config.

## Usage

Use `CONFIG_PATH` or `AGENT_CONFIG` / `PROXY_CONFIG` to point at a config file. Environment variables override file values (keys lowercase with underscores, e.g. `OIDC_ISSUER`).

## Tooling

From the repo root (see root `Makefile`):

- **validate-config** — Validate a config file (same load + `Validate()` as runtime):
  ```bash
  make validate-config CONFIG_PATH=configs/agent.example.json BINARY=agent
  make validate-config CONFIG_PATH=configs/proxy.dev.yaml BINARY=proxy
  ```
- **print-schema** — Print JSON Schema for a binary or plugin:
  ```bash
  make print-schema BINARY=agent
  make print-schema SCHEMA=schemas/plugins/integration/caddy.schema.json
  ```
- **render-config-example** — Render example config from defaults (go-config struct + `ApplyDefaults()`):
  ```bash
  make render-config-example BINARY=agent FORMAT=json
  make render-config-example BINARY=proxy FORMAT=yaml
  ```

