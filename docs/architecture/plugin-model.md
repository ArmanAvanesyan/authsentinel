# Plugin model

AuthSentinel’s plugin system extends the runtime with **pipeline** steps, **provider** backends (e.g. OIDC), and **integration** adapters for gateways (Caddy, Traefik, KrakenD). All plugins use the same API and lifecycle; config is schema-driven under `schemas/plugins/`.

## Plugin kinds

| Kind | Purpose | Implemented by |
|------|---------|----------------|
| **Pipeline** | Participate in the proxy request pipeline (e.g. rate limit, headers, audit). | `pluginapi.PipelinePlugin` |
| **Provider** | Identity provider: authorization URL, code exchange, refresh. Used by the agent. | `pluginapi.ProviderPlugin` |
| **Integration** | Gateway adapter: wire proxy engine into Caddy, Traefik, KrakenD. | `pluginapi.IntegrationPlugin` |

Each plugin has a **PluginDescriptor** (ID, kind, name, capabilities, optional config schema ref, version info).

## Core interfaces

- **Plugin**: `Descriptor() PluginDescriptor`, `Health(ctx) PluginHealth`.
- **ConfigurablePlugin** (optional): `Configure(ctx, cfg) error` — host passes typed config from go-config.
- **StartablePlugin** (optional): `Start(ctx)`, `Stop(ctx)` for lifecycle.
- **PipelinePlugin**: `Handle(ctx, req *proxy.Request, principal *token.Principal) (*policy.Decision, error)`.
- **ProviderPlugin**: `AuthorizationURL`, `ExchangeCode`, `Refresh` — used by the agent for login/callback/refresh.
- **IntegrationPlugin**: `Serve(ctx, hostCtx any) error` — attaches to the gateway (hostCtx is gateway-specific).

Interfaces live in **pkg/pluginapi**. Gateway adapters in **pkg/plugins/{caddy,traefik,krakend}** implement IntegrationPlugin (and expose Handler/Middleware for HTTP wiring).

## Lifecycle states

Plugins move through a state machine used by the registry and admin API:

| State | Meaning |
|-------|--------|
| discovered | Found by discovery (e.g. manifest). |
| verified | Checksum/signature verified (if enabled). |
| registered | Registered in the registry. |
| configured | Configure() called with config. |
| initialized | Ready to start. |
| started | Start() called. |
| healthy / degraded | Running; health reported by plugin. |
| stopped | Stop() called or failure. |

Health is reported via **PluginHealth** (State, Message, Details). The proxy’s `/admin` endpoint exposes plugin list and per-plugin health when `AdminSecret` is set.

## Registry and discovery

- **pkg/pluginregistry**: Register plugins, enable/disable, resolve by capability, build dependency graph (`DependsOn` in descriptor). Used by proxy (and optionally agent) at startup.
- **pkg/plugindiscovery**: Discover plugins from a directory (manifest-based). `DiscoverFromDir(ctx, registry, dir, verifier)` reads manifests and registers plugins; no Go `.so` loading in v1.

Proxy config can set **PluginsManifestDir**; `cmd/proxy` calls discovery then `BuildDependencyGraph()`. Built-in plugins (e.g. Caddy, Traefik, KrakenD) are registered by the application or by discovery manifests that reference built-in implementations.

## Config and schema

- Plugin config is loaded by the host (go-config) and passed to `Configure(ctx, cfg)`.
- **ConfigSchemaRef** in the descriptor points to a JSON Schema under `schemas/plugins/**` (e.g. `schemas/plugins/provider/oidc.schema.json`, `schemas/plugins/integration/krakend.schema.json`).
- **pkg/pluginconfig**: Envelope and helpers for plugin config; aligned with go-config where applicable.

Contract tests in **test/contract** validate plugin configs against these schemas.

## Host services

**pkg/pluginhost** provides safe host services for plugins: logger, metrics, cache access, secret resolution, Redis abstraction. Plugins do not receive raw config secrets; they receive resolved values or callbacks from the host.

## Packaging (v1)

- **Built-in plugins**: Compiled into the binary (e.g. gateway adapters).
- **Manifest-based plugins**: Discovered from a directory; manifest describes the plugin and config schema. No dynamic Go `.so` loading in v1.
- Optional: WASM for untrusted or custom policy/filter plugins (future).

## References

- [pluginapi](https://pkg.go.dev/github.com/ArmanAvanesyan/authsentinel/pkg/pluginapi) — interfaces and types.
- [pluginregistry](https://pkg.go.dev/github.com/ArmanAvanesyan/authsentinel/pkg/pluginregistry), [plugindiscovery](https://pkg.go.dev/github.com/ArmanAvanesyan/authsentinel/pkg/plugindiscovery) — registration and discovery.
- [Compatibility matrix](../integration/compatibility-matrix.md) — gateway adapter versions and modes.
- [Implementation plan](implementation-plan.md) — Phase 2 plugin platform.
