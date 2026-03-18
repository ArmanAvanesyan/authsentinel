# AuthSentinel: Consolidated Implementation Plan

This document is the **single source of truth** for turning the AuthSentinel monorepo from an architectural declaration into a production-ready security runtime. It consolidates the full implementation plan, Rust/WASM guidance, and the decision to use the external **go-config** library for configuration.

---

## Part 1: Deep dive summary (current state)

### What is in good shape

- **Binaries**: `cmd/agent` and `cmd/proxy` build and run; release config (e.g. GoReleaser) targets both.
- **Core packages**: `pkg/agent`, `pkg/cookie`, `pkg/token`, `pkg/session`, `pkg/proxy`, `pkg/graphql`, `pkg/grpc`, `pkg/sdk`, `pkg/testing` have real logic, tests, and clear boundaries.
- **Internal orchestration**: `internal/agent` (OIDC, HTTP, redirect, service), `internal/proxy` (HTTP, agent client, headers), `internal/store/redis` (session, PKCE, refresh lock) are implemented.
- **Proto**: `proto/authsentinel/{agent,policy,proxy,sdk}/v1` with full contracts (login/callback/refresh/logout/introspect, request decision/principal/route/deny reason, evaluation/obligations/trace, principal/session/auth context); buf lint, breaking checks, and Go + TS codegen; Makefile targets `proto-lint`, `proto-breaking`, `proto-generate`; policy in `proto/README.md` for Go/TS/Flutter generated clients.
- **SDKs**: **Done** — Go `pkg/sdk` (HTTP middleware, principal extraction, GraphQL context, gRPC interceptor, session-aware AgentClient, proxy adapter); JS/TS under `packages/sdk/js` (core browser client, node middleware, React AuthProvider/useAuth, typed models, typescript re-exports); Flutter under `packages/sdk/flutter` (AuthClient, SessionClient, principal/session models, AuthScope with refresh-on-resume, platform adapter for mobile/web).
- **Docs**: `docs/` has substantive architecture, runtime, SDK, and ops content (overview, package-boundaries, docker, release, etc.).
- **Deployments**: `deployments/docker/` has agent and proxy Dockerfiles and a working docker-compose.
- **Workspace**: Root package.json is a pnpm workspace with build/lint/test/changeset/release; JS/TS and Flutter SDK packages exist under `packages/`.
- **Configuration**: Loaded via **go-config** in `cmd/agent` and `cmd/proxy`; config types in `internal/agent/config` and `internal/proxy/config`; no `pkg/config`. See `docs/ops/configuration.md`.
- **configs/**: Example agent and proxy configs (JSON and YAML, including dev/prod variants) and README; `configs/plugins/` has Caddy, Traefik, KrakenD examples.
- **pkg/plugins (gateway adapters)**: **Done** — Caddy, Traefik, KrakenD have config shape, handler/middleware entrypoint, request→proxy translation, response mapping, docs, example configs, and compatibility matrix; all call the same proxy/policy/token runtime. Optional Caddy directive (build with `-tags caddy`).
- **schemas/**: Base runtime schema, per-binary config schemas, and per-plugin schemas under `schemas/plugins/**`; command-line tooling wired via `make schema`, `make validate-config`, `make print-schema`, and `make render-config-example`.

### What is stub or placeholder

- **pkg/policy**: Rego path is real; **wasm.go** is a TODO (no WASM runtime wired).
- **tools/**, **scripts/**: README-only placeholders.
- **packages/policy-bundles**: README-level; no real bundle layout or loader contract.

### Design decisions (from consolidated drafts)

| Topic | Decision |
|-------|----------|
| **Configuration** | Use **go-config** (external: `github.com/ArmanAvanesyan/go-config`). Do **not** implement a second config subsystem in `pkg/config`. Runtime consumes typed structs only. |
| **Cookie / token / session** | Implement in **Go** only. No Rust/WASM for these; bottlenecks are Redis and network, not crypto or serialization. |
| **Policy engine** | **WASM** for sandboxed policy execution. Use an existing runtime (e.g. Wasmtime) and consider OPA compiled to WASM or similar; finish `pkg/policy` WASM host and bundle loader. |
| **Plugins** | Built-in and manifest-based plugins first. Do **not** rely on Go `.so` plugins (ABI issues). Optional: WASM for untrusted/custom policy or filter plugins. |
| **Gateway adapters** | Caddy, Traefik, KrakenD adapters must all call the **same** proxy/policy/token runtime; no per-adapter duplication of auth logic. |

---

## Part 2: Product model and package rules (baseline)

### Frozen product model

- **authsentinel-agent**: Login, callback, refresh, logout, session bootstrap; browser redirects; CSRF/state/nonce; cookie-backed sessions.
- **authsentinel-proxy**: Gateway-agnostic request pipeline, token/session validation, policy evaluation, upstream selection, header mutation, deny/error responses.
- **Shared**: Embedded policy engine (WASM), plugin host, config (via go-config), SDKs, gateway adapters.

### Bounded package rules

| Boundary | Rule |
|----------|------|
| **cmd/** | Only wires applications; no business logic. |
| **internal/** | App-only orchestration (agent, proxy, store/redis). |
| **pkg/** | Reusable runtime libraries; no app-specific wiring. |
| **proto/** | Source of truth for RPC contracts. |
| **schemas/** | Source of truth for config JSON Schema (and plugin schemas). |
| **configs/** | Example and template config files consumed by go-config. |

**Config**: AuthSentinel does **not** implement config loading. It depends on **go-config** for: file, env, flags, defaults, validation, merging. No `pkg/config` loader; optional thin adapter in `internal/runtime` or `cmd` that calls go-config and passes structs into runtime.

---

## Part 3: Phased implementation plan

### Phase 0: Stabilize the architectural baseline

**Goal**: Make the stated architecture explicit and document it before adding features.

**Deliverables**:

- Freeze product model and package rules (as in Part 2).
- Write **docs/architecture/runtime-map.md**: exact responsibilities and allowed imports per package; which packages may depend on go-config (only `cmd` and/or `internal` bootstrap).
- **Done**: `pkg/config` removed; “config types” live in `internal/agent/config` and `internal/proxy/config`; loading is via go-config in `cmd`.

---

### Phase 1: Core runtime and configuration

**Goal**: Production-ready core with a single, external config system.

#### 1.1 Configuration (go-config)

- **Done**: go-config integrated; `cmd/agent` and `cmd/proxy` load from file (optional) + env; config types in `internal/agent/config` and `internal/proxy/config`; `configs/` has `agent.example.json`, `proxy.example.json`, `agent.example.yaml`, `proxy.example.yaml` and README. See [plan-replace-pkg-config-with-go-config.md](plan-replace-pkg-config-with-go-config.md) for the completed migration.
- **Do not** implement a full config loader in-tree; use **go-config** only (file → env → flags, defaults, validation as the library supports).
- **configs/** (remaining optional): e.g. `configs/agent.dev.yaml`, `configs/proxy.dev.yaml`, `configs/plugins/oidc.example.yaml`, `configs/plugins/krakend.example.yaml`.
- **schemas/** (optional but recommended): Base runtime and per-binary config JSON Schema; later per-plugin schemas under `schemas/plugins/...`.

#### 1.2 pkg/cookie

- Implement: encrypted codec, signing and rotation, same-site/secure/domain policy, chunking if needed, browser session helpers.
- Keep implementation in **Go** (no Rust/WASM).

#### 1.3 pkg/token

- Implement: JWT validation, JWKS resolver/cache, issuer/audience checks, claim normalization, principal model, token exchange helpers where relevant.
- Keep implementation in **Go** (existing `github.com/golang-jwt/jwt/v5` and JWKS usage).

#### 1.4 pkg/session

- Session model, cookie-backed browser session, Redis-backed store (align with `internal/store/redis`), refresh locks, revoke/invalidate flows.
- Keep implementation in **Go**.

#### 1.5 internal/store/redis

- Finish: session store, replay/state cache, JTI or revocation tracking, rate-limit counters if used. Keep as the concrete implementation of session/store interfaces used by `pkg/session` and agent/proxy.

#### 1.6 pkg/policy

- Normalized policy input, decision model (deny/reason/obligations), **WASM execution host** (integrate Wasmtime or similar; no custom WASM runtime), bundle loader and cache, fallback policy behavior.
- Use **WASM** for policy execution; consider OPA compiled to WASM or Rust/TinyGo policies. Do not move cookie/token/session into WASM.

#### 1.7 pkg/proxy

- Gateway-agnostic request pipeline, upstream selection, authn/authz pipeline stages, header mutation, error/deny responses, request context propagation. This remains the heart of the platform.

#### 1.8 pkg/agent

- Login start, callback handling, session bootstrap, refresh, logout, browser redirect flows, CSRF/state/nonce handling. Matches the flow: browser → agent → cookies → proxy → policy → upstream.

---

### Phase 2: Plugin platform

**Goal**: Move from “plugin-shaped folders” to a stable extension system.

- **Done**: pkg/pluginapi, pluginregistry, plugindiscovery, pluginconfig, pluginhost implemented; pipeline/provider/integration plugin interfaces; Caddy, Traefik, KrakenD adapters implement IntegrationPlugin; lifecycle state model (discovered → verified → registered → configured → initialized → started → healthy/degraded → stopped); schemas/plugins/** (provider/oidc, pipeline/ratelimit, integration/krakend); proxy and agent config extended (PipelinePlugins, PluginsManifestDir, ProviderPluginID); proxy startup wires registry and manifest discovery; v1 built-in and local manifest-based plugins only (no Go .so).

#### 2.1 Core plugin packages (under pkg/)

- **pkg/pluginapi**: Interfaces, plugin kinds, lifecycle, capability model, version compatibility.
- **pkg/pluginregistry**: Registration, enable/disable, resolution by capability, dependency graph.
- **pkg/plugindiscovery**: Built-in discovery, filesystem discovery, manifest parsing, optional signature/checksum verification.
- **pkg/pluginconfig**: Plugin config envelope, schema mapping, defaults, migrations (aligned with go-config where applicable).
- **pkg/pluginhost**: Logger, metrics, cache access, secret resolution, Redis abstraction, and other safe host services for plugins.

#### 2.2 Plugin classes

- **Pipeline plugins**: e.g. rate limit, headers, audit, request id, policy enforcement step.
- **Provider plugins**: e.g. OIDC, Keycloak, Authentik, generic IdP handler.
- **Integration plugins**: Caddy, Traefik, KrakenD adapters (same proxy/policy runtime).

#### 2.3 Lifecycle

- State model: discovered → verified → registered → configured → initialized → started → healthy/degraded → stopped.

#### 2.4 Plugin config schema

- Use **schemas/plugins/** (e.g. `schemas/plugins/provider/oidc.schema.json`, `schemas/plugins/pipeline/ratelimit.schema.json`, `schemas/plugins/integration/krakend.schema.json`).

#### 2.5 Packaging

- v1: built-in plugins and local manifest-based plugins. Avoid Go `.so` plugins as the primary path.

---

### Phase 3: Application binaries and internal orchestration

**Goal**: `cmd/agent` and `cmd/proxy` as small, deterministic entrypoints.

- **Done**: cmd/agent and cmd/proxy wire config via go-config, session store (agent), token/policy engine and plugin registry (proxy); health/readiness endpoints (`/healthz`, `/readyz`) on both; internal/agent has HTTP handlers, callback orchestration, session bootstrap, error mapping (`internal/agent/errormap`), and provider glue (`internal/agent/provider.IdP`, service uses IdP); internal/proxy has route assembly and pipeline in `routes()`, and `internal/proxy/response` for deny/API response shaping.

#### 3.1 cmd/agent

- Wire: config load via go-config, observability bootstrap, session store, provider plugins, agent HTTP router, health/readiness.

#### 3.2 cmd/proxy

- Wire: config load via go-config, token engine, policy engine, request pipeline, upstream router, plugin registry, health/readiness.

#### 3.3 internal/agent

- HTTP handlers, callback orchestration, session bootstrap flow, error mapping, provider glue.

#### 3.4 internal/proxy

- Route assembly, request pipeline assembly, upstream policies, response shaping, deny page / API error mapping.

---

### Phase 4: Protobuf and contract layer

**Goal**: Proto as the real contract system.

- **Done**: **agent/v1**: login, callback, refresh, logout, session introspection (messages and AgentService RPCs).
- **Done**: **proxy/v1**: request decision (Decide), principal introspection, route resolution, deny reason model (DenyReason, Route).
- **Done**: **policy/v1**: evaluation request/response (EvaluateRequest/Response, EvaluationInput, EvaluationDecision), obligations, trace/debug structures (TraceEvent).
- **Done**: **sdk/v1**: shared principal/session/auth context messages (Principal, Session, AuthContext).
- **Done**: Enforce generation via **buf**: lint, breaking checks (`make proto-lint`, `make proto-breaking`), code generation (`make proto-generate`). Policy for Go/TS/Flutter in `proto/README.md` (Go in-repo, TS under `packages/sdk`, Flutter aligned to same proto versioning).

---

### Phase 5: SDK products

**Goal**: Multi-platform SDKs as real deliverables.

- **Done**: **Go (pkg/sdk)**: Server-side entrypoint — HTTP middleware (PrincipalExtractor, Middleware), principal extraction (IdentityFromHTTPRequest, PrincipalFromContext), GraphQL context (GetPrincipalFromGraphQLContext), gRPC interceptor (UnaryServerInterceptor, JWTValidator), session-aware AgentClient (GetLoginURL, GetLogoutURL, GetSession, Refresh), proxy adapter (NewProxyPrincipalExtractor).
- **Done**: **JS/TS**: Under `packages/sdk/js` — core browser client (createClient, getSession/getLoginURL/getLogoutURL/refresh/login/logout), node middleware (createNodeMiddleware with agentBaseUrl/cookieName), React AuthProvider/useAuth/useSession, typed models (Principal, SessionUser, SessionInfo, etc.), typescript package re-exports.
- **Done**: **Flutter**: Under `packages/sdk/flutter` — auth/session wrappers (AuthClient, SessionClient), refresh-on-resume (AuthScope, WidgetsBindingObserver), principal/session models (Principal, SessionUser, SessionInfo, AuthContext), mobile/web adapter split (PlatformAdapter, UrlLauncherAdapter).

---

### Phase 6: Gateway adapters

**Goal**: Make `pkg/plugins/{caddy,traefik,krakend}` consumable.

- **Done**: For each adapter: config shape, plugin registration entrypoint (Handler/Middleware.Handler), translation from gateway request to common proxy engine input (`proxy.RequestFromHTTP`), response mapping (allow → headers + next/proxy, deny → `proxy.WriteResponse`), docs, example config, compatibility matrix. All adapters call the **same** proxy/policy/token runtime.
- **6.1 Caddy**: Config (`auth_url`, `upstream_url`, `require_auth`); `Handler(engine, next, upstreamURL)` and `Middleware.Handler(next)`; optional directive `authsentinel <url>` (build with `-tags caddy`); forward-auth example in `configs/plugins/caddy.Caddyfile`; `docs/integration/caddy.md`; `schemas/plugins/integration/caddy.schema.json`.
- **6.2 Traefik**: Config (`auth_url`, `upstream_url`, `auth_response_headers`); middleware plugin interface and `Handler(engine, next, upstreamURL, authResponseHeaders)`; header and deny response mapping; `configs/plugins/traefik.example.yaml`; `docs/integration/traefik.md`; `schemas/plugins/integration/traefik.schema.json`.
- **6.3 KrakenD**: Config (`endpoint`, `upstream_url`, `require_auth`); endpoint/auth middleware bridge and `Plugin.Handler(next)`; shared principal decision output (UpstreamHeaders); `configs/plugins/krakend.example.json`; `docs/integration/krakend.md`; existing `schemas/plugins/integration/krakend.schema.json`.
- **Compatibility matrix**: `docs/integration/compatibility-matrix.md`.

---

### Phase 7: Config and schema system (finish)

**Goal**: configs/ and schemas/ are first-class. **Done.**

- **configs/**: Done — agent and proxy example JSON/YAML (including dev/prod variants); `configs/plugins/` has Caddy, Traefik, KrakenD examples from Phase 6.
- **schemas/**: Done — base runtime schema, per-binary config schemas (`agent.schema.json`, `proxy.schema.json`), and per-plugin schemas under `schemas/plugins/**`.
- **Tooling**: Done — `make schema` (generate binary config schemas), `make validate-config` (load + `Validate()` for agent/proxy configs, JSON and YAML), `make print-schema` (print binary or plugin schema), `make render-config-example` (emit example configs from Go defaults).

---

### Phase 8: Observability and ops

**Goal**: Operable in production.

- **Observability**: Structured logging, metrics, tracing; per-plugin health; per-request auth decision metrics; JWKS/cache/session metrics. Optional shared **pkg/observability**.
- **Health**: `/healthz`, `/readyz`, `/livez` for both binaries.
- **Admin** (guarded): loaded plugins, plugin health, active config summary, policy bundle version, session store status.
- **Deployments**: Keep `deployments/docker/` (agent + proxy Dockerfiles, compose, sample env); add Helm/K8s later if needed.

---

### Phase 9: Testing strategy

- **Done**: Unit tests (pkg/pluginregistry, cookie codec error paths, JWT/claims in pkg/token, file-based config in internal/*/config, policy placeholder in pkg/policy); contract tests (test/contract: plugin config vs schemas, SDK session/principal shapes, agent JSON shape); integration tests (test/integration: agent with mock OIDC, proxy with mock agent/upstream, Redis session lifecycle, gateway Handler tests in pkg/plugins); E2E smoke playbook (test/e2e/playbook.sh), make e2e-docker, docs/ops/e2e.md and CI notes.
- **Unit**: cookie codecs, JWT validation, claims mapping, config validation (via go-config), policy evaluation, plugin registry logic.
- **Contract**: Proto compatibility, SDK expectations, plugin config schema validation.
- **Integration**: Agent login/callback/refresh/logout; proxy with mock upstream; Redis session lifecycle; policy bundle enforcement; gateway adapters.
- **E2E**: Runnable scenarios (browser + agent + proxy + upstream; SPA+BFF; API-only; Caddy/Traefik/KrakenD examples).

---

### Phase 10: Release engineering

- **Go**: GoReleaser for authsentinel-agent, authsentinel-proxy, checksums, archives.
- **JS/TS**: Changesets for versioning, changelogs, publish.
- **Version policy**: Monorepo release train; compatibility matrix (runtime, plugin API, proto, SDK versions).

---

### Phase 11: Documentation

**Goal**: Replace placeholders with buildable docs.

- **docs/architecture/runtime-map.md** (Phase 0).
- **docs/architecture/plugin-model.md**.
- **docs/runtime/agent-flows.md**, **docs/runtime/proxy-pipeline.md**, **docs/runtime/policy-engine.md**.
- **docs/integration/caddy.md**, **docs/integration/traefik.md**, **docs/integration/krakend.md** (done — Phase 6), **docs/integration/compatibility-matrix.md**.
- **docs/sdk/go.md**, **docs/sdk/js.md**, **docs/sdk/flutter.md** (extend existing).
- **docs/ops/configuration.md** (done — go-config, configs/, file vs env), **docs/ops/deployment-docker.md**, **docs/ops/health-metrics-tracing.md**.

---

## Part 4: Recommended implementation order (milestones)

Execute in this order to minimize rework and deliver value incrementally.

| Milestone | Scope |
|-----------|--------|
| **M1: Config and runtime core** | **Done**: go-config integrated; pkg/config removed; cmd/agent and cmd/proxy load via go-config; configs/ has agent and proxy example JSON/YAML. Remaining: finish pkg/cookie, pkg/token, pkg/session; finish internal/store/redis. |
| **M2: Agent and proxy** | Finish pkg/agent and pkg/proxy; wire cmd/agent and cmd/proxy with go-config and observability; keep internal/agent and internal/proxy as orchestration only. |
| **M3: Policy** | Finish pkg/policy: WASM host (Wasmtime or equivalent), bundle loader, decision model. |
| **M4: Plugin platform** | **Done**: pkg/pluginapi, pluginregistry, plugindiscovery, pluginconfig, pluginhost; gateway adapters (Caddy, Traefik, KrakenD) as IntegrationPlugin; schemas/plugins; proxy/agent config and startup wired to registry and discovery. |
| **M5: Contracts and SDKs** | **Done**: Proto definitions and buf workflow (Phase 4); Go/JS/Flutter SDKs implemented (Phase 5). |
| **M5b: Gateway adapters** | **Done**: Phase 6 — pkg/plugins/{caddy,traefik,krakend} consumable (config, handler, translation, response mapping, docs, example configs, compatibility matrix). |
| **M6: Ops and release** | configs + schemas tooling; deployments; health/metrics/tracing; release pipelines; documentation from Phase 11. |

---

## Part 5: Repository layout (target)

### Add under pkg/

- **pluginapi**, **pluginregistry**, **plugindiscovery**, **pluginconfig**, **pluginhost**
- Optionally **observability** (if shared logging/metrics/tracing is desired)

### Config

- Config loading is done only via **go-config** (external). Config types live in `internal/agent/config` and `internal/proxy/config`; `pkg/config` has been removed.

### Keep as-is (with implementation work as in phases)

- **pkg/**: agent, proxy, policy, token, cookie, session, graphql, grpc, sdk, testing, plugins/caddy, plugins/traefik, plugins/krakend
- **internal/**: agent, proxy, store/redis
- **cmd/**: agent, proxy
- **proto/**, **configs/**, **schemas/**, **docs/**, **deployments/**, **packages/**

---

## Part 6: Technology choices (recap)

| Layer | Technology | Notes |
|-------|------------|--------|
| Runtime | Go | HTTP, Redis, observability, simplicity. |
| Config | go-config (external) | Single config system; no in-tree loader. |
| Cookie / token / session | Go | No Rust/WASM; bottlenecks are I/O. |
| Policy execution | WASM | Sandbox; use Wasmtime (or similar) and consider OPA WASM. |
| Plugins | Built-in + manifest-based; optional WASM | Avoid Go .so plugins for v1. |
| Gateway adapters | Go | Same proxy/policy/token runtime for all. |

---

## Summary

The repo is already shaped as a serious platform. **Config and schema tooling are done**: go-config is the single config system; agent and proxy load from file + env; `configs/` has example and dev/prod files; `schemas/` has runtime, per-binary, and plugin JSON Schemas; `docs/ops/configuration.md`, `configs/README.md`, and `schemas/README.md` describe usage and tooling. **Proto and SDKs are done**: Phase 4 delivers proto as the contract system with buf lint/breaking/codegen; Phase 5 delivers Go (pkg/sdk), JS/TS (packages/sdk/js), and Flutter (packages/sdk/flutter) SDKs as real deliverables. **Gateway adapters are done**: Phase 6 makes pkg/plugins/{caddy,traefik,krakend} consumable with config, handlers, translation, response mapping, docs, example configs, and a compatibility matrix; all call the same proxy engine. The plan is to **finish the rest of the core runtime**, wire **observability and health**, and turn **deployments** and remaining **docs** from placeholders into working assets. One config system (go-config), one policy execution model (WASM), and one proxy engine shared by all gateway adapters.
