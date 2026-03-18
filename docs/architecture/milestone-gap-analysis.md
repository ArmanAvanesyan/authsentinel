# AuthSentinel: Milestone Gap Analysis & Implementation Plan

This document is the output of a **deep dive scan** of the AuthSentinel codebase against the [implementation-plan.md](implementation-plan.md) Part 4 milestones. It defines gaps, corrects outdated plan statements, and provides a prioritized implementation plan.

---

## Executive Summary

| Milestone | Plan Status | Actual Status | Priority |
|-----------|-------------|---------------|----------|
| **M1** | Mostly done | ~90% done | Low — finish YAML loader, metrics wiring |
| **M2** | Finish agent/proxy | ~95% done | Medium — agent metrics, tracing |
| **M3** | Finish policy | **WASM done** (plan outdated) | Medium — wire bundle path, sample bundle |
| **M4** | Done | Done | — |
| **M5** | Done | Done | — |
| **M5b** | Done | Done | — |
| **M6** | Ops and release | Partial | High — Helm, tracing, proto-gen in CI |

**Key correction**: The implementation plan states *"Rego path is real; wasm.go is a TODO"*. **Reality**: WASM is implemented (wazero); Rego is a stub. The policy engine uses `pkg/policy/wasm.go` and `pkg/policy/bundle.go`; proxy uses `policy.NewWASMRuntime(DefaultFallbackAllow)` but does **not** load a bundle from config.

---

## M1: Config and Runtime Core

### What Exists ✓

- **go-config** integrated in `cmd/agent`, `cmd/proxy`, `cmd/validateconfig`, `cmd/schema`
- **pkg/config** removed; config types in `internal/agent/config` and `internal/proxy/config`
- **configs/** has agent and proxy example JSON/YAML (dev/prod variants); `configs/plugins/` for Caddy, Traefik, KrakenD
- **pkg/cookie**: Codec, Encrypt/Decrypt, SignedManager, chunking, helpers, options, model
- **pkg/token**: JWKS (interface + HTTPJWKSSource with OIDC discovery + cache), parser, validator, claims, validate_idtoken
- **pkg/session**: Session model, KeyLayout, SessionStore/PKCEStore/RefreshLockStore, BrowserSessionManager, store.go
- **internal/store/redis**: SessionStore, PKCEStore, RefreshLockStore, SetRevoked/IsRevoked, RecordReplay/CheckReplay

### Gaps

| Gap | Severity | Description |
|-----|----------|-------------|
| **YAML loader in runtime** | Medium | `cmd/agent` and `cmd/proxy` use `json.New()` only. YAML configs in `configs/` cannot be loaded at runtime. `validateconfig` has a workaround (reads YAML, unmarshals to map, converts to JSON, then decodes). |
| **JWKS/session metrics** | Low | JWKSCacheHit/Miss and session store metrics exist in pkg/token but are not wired into agent/proxy observability. |
| **Cookie rotation** | Low | Same-site/secure/domain policy and chunking exist; rotation and full production hardening may need tests. |

### Implementation Plan (M1)

1. **Add YAML format support for runtime** (if go-config supports it):
   - Check go-config for `format/yaml` or equivalent
   - In `cmd/agent` and `cmd/proxy` `loadConfig()`, detect file extension and add `yaml.New()` when path ends in `.yaml`/`.yml`
   - Fallback: document that runtime expects JSON; use `validateconfig` for YAML validation only

2. **Wire JWKS and session metrics** (optional, can defer to M6):
   - Expose JWKSCacheHit/Miss from pkg/token
   - Add session store metrics to internal/store/redis
   - Register in agent/proxy Prometheus registry

3. **Add cookie rotation tests** (optional):
   - Unit tests for SignedManager rotation scenarios

---

## M2: Agent and Proxy

### What Exists ✓

- **pkg/agent**: login, logout, refresh, session, agent.go (Service interface)
- **pkg/proxy**: proxy.go (Engine), engine.go (DefaultEngine), middleware, request, decision, normalize, http
- **internal/agent**: httpserver, service, redirect, provider/idp, oidc, errormap, config
- **internal/proxy**: httpserver, agentclient, response, headers, resolver
- **Observability**: `pkg/observability` (Metrics, PrometheusMetrics, Logger, StdLogger, Tracer, Provider); proxy uses Prometheus; agent uses std log only

### Gaps

| Gap | Severity | Description |
|-----|----------|-------------|
| **Agent Prometheus metrics** | Medium | Agent has no /metrics; proxy has Prometheus. Agent should expose login/callback/refresh counts, session store health. |
| **Distributed tracing** | Low | Tracer interface exists; NopTracer default; OpenTelemetry not wired. |
| **Structured logging in agent** | Low | Agent uses `log.Logger`; proxy may use observability.Logger. Unify on pkg/observability.Logger. |

### Implementation Plan (M2)

1. **Add Prometheus metrics to agent**:
   - Create `pkg/observability` metrics for: `authsentinel_agent_login_total`, `authsentinel_agent_callback_total`, `authsentinel_agent_refresh_total`, `authsentinel_agent_logout_total`
   - Wire in `internal/agent/httpserver` and `internal/agent/service`
   - Register `/metrics` in `cmd/agent` (same pattern as proxy)

2. **Wire OpenTelemetry Tracer** (optional, can defer to M6):
   - Add OTLP exporter config to agent/proxy config
   - Use `pkg/observability.Tracer` with real provider when configured

3. **Replace agent std log with pkg/observability.Logger**:
   - Use `observability.StdLogger` or structured logger in `cmd/agent`

---

## M3: Policy

### What Exists ✓

- **WASM host**: `pkg/policy/wasm.go` — wazero runtime, Load/LoadBundle, Evaluate with `evaluate(input_ptr, input_len)` ABI, fallback decision
- **Bundle loader**: `pkg/policy/bundle.go` — BundleLoader with path+mtime cache, recompilation on change
- **Decision model**: `pkg/policy/input.go`, `result.go` — Input, Decision (Allow, StatusCode, Reason, Obligations, Headers)
- **Engine interface**: `pkg/policy/engine.go` — Engine, EngineWithStatus
- **Proxy integration**: `internal/proxy/httpserver` uses `policy.NewWASMRuntime(DefaultFallbackAllow)` — **no bundle loaded**

### Gaps

| Gap | Severity | Description |
|-----|----------|-------------|
| **Policy bundle path in config** | High | `internal/proxy/config` has no `PolicyBundlePath` field. Proxy never loads a bundle; always uses fallback allow. |
| **Sample WASM policy bundle** | Medium | `packages/policy-bundles/` has READMEs only; no real WASM bundle. |
| **Rego stub** | Low | `pkg/policy/rego.go` — Load() returns nil. Either implement or remove. |

### Implementation Plan (M3)

1. **Add PolicyBundlePath to proxy config**:
   - Add `PolicyBundlePath string` to `internal/proxy/config.Config`
   - In `internal/proxy/httpserver.New()`, if `cfg.PolicyBundlePath != ""`:
     - Use `policy.NewBundleLoader(policy.DefaultFallbackAllow)` and load via `LoadBundle(cfg.PolicyBundlePath)`
     - Else: keep current `NewWASMRuntime(DefaultFallbackAllow)` (no bundle)
   - Update `internal/proxy/httpserver` to accept `policy.Engine` (or EngineWithStatus) from caller so `cmd/proxy` can construct it

2. **Wire BundleLoader in cmd/proxy**:
   - `cmd/proxy` loads config, creates policy engine (with or without bundle), passes to httpserver.New

3. **Create sample WASM policy bundle**:
   - Add minimal WASM policy in `packages/policy-bundles/core/` (e.g. TinyGo or Rust compiled to WASM)
   - Document ABI: `evaluate(input_ptr, input_len) -> (output_ptr, output_len)`, JSON input/output shape

4. **Rego**: Either implement OPA integration or remove stub and document WASM-only for v1.

---

## M4: Plugin Platform

### What Exists ✓

- **pkg/pluginapi**, **pluginregistry**, **plugindiscovery**, **pluginconfig**, **pluginhost** — all implemented
- Gateway adapters (Caddy, Traefik, KrakenD) as IntegrationPlugin
- **schemas/plugins**: provider/oidc, pipeline/ratelimit, integration/{caddy,traefik,krakend}
- Proxy startup wires PluginsManifestDir + plugindiscovery + registry

### Gaps

| Gap | Severity | Description |
|-----|----------|-------------|
| **Pipeline plugin execution** | Medium | Registry and discovery exist; proxy engine does **not** use pipeline plugins in the request flow. |
| **Provider plugin resolution** | Medium | `internal/agent/provider` uses IdP interface; provider plugin registry not wired. Agent uses built-in OIDC only. |
| **Built-in plugin manifests** | Low | No built-in manifests for Caddy, Traefik, KrakenD in discovery. |

### Implementation Plan (M4)

1. **Add pipeline plugin execution in proxy DefaultEngine**:
   - Resolve pipeline plugins by capability from registry
   - Run pipeline stages (e.g. rate limit, headers, audit) before/after policy evaluation
   - Document pipeline plugin contract and execution order

2. **Wire provider plugin registry in agent**:
   - Allow `ProviderPluginID` in agent config to select a provider from registry
   - Fallback to built-in OIDC when not set

3. **Add built-in manifests** (optional):
   - Ship manifests for Caddy, Traefik, KrakenD in `schemas/plugins/` or embedded in binary

---

## M5 & M5b: Contracts, SDKs, Gateway Adapters

### Status: Done ✓

- Proto definitions, buf lint/breaking, proto-generate
- Go/JS/Flutter SDKs implemented
- Gateway adapters (Caddy, Traefik, KrakenD) with config, handler, translation, response mapping, docs, example configs, compatibility matrix

### Minor Gaps

| Gap | Severity | Description |
|-----|----------|-------------|
| **Proto gen in CI** | Low | CI now runs `make proto-generate` before Go build/test to keep generated contracts in sync. |
| **Flutter proto** | Low | Flutter uses Dart models; no buf-based Dart codegen. Alignment is manual. |
| **Caddy build tag** | Low | Docs mention `-tags caddy`; verify module.go builds correctly. |

---

## M6: Ops and Release

### What Exists ✓

- **configs + schemas**: `make schema`, `validate-config`, `print-schema`, `render-config-example`
- **deployments**: docker-compose (agent, proxy, redis, bff), Dockerfiles, .env.example
- **Health**: /healthz, /livez, /readyz on agent and proxy; agent readyz pings Redis
- **Metrics**: Proxy exposes /metrics (Prometheus)
- **Release**: .goreleaser.yaml (agent, proxy); ci.yaml runs GoReleaser on tag push
- **Docs**: docs/ops (configuration, docker, deployment-docker, health-metrics-tracing, release, helm, e2e)

### Gaps

| Gap | Severity | Description |
|-----|----------|-------------|
| **Helm charts** | Medium | docs/ops/helm.md exists; no Helm charts in repo. |
| **Tracing backend** | Low | Tracer interface; no real OpenTelemetry/OTLP wiring. |
| **E2E playbook** | Low | test/e2e/playbook.sh is minimal (health, proxy 401, agent /login). Expand for core flows. |
| **Proto gen in CI** | Low | Implemented: CI runs `make proto-generate` (buf generate) before Go build/test. |

### Implementation Plan (M6)

1. **Add Helm charts**:
   - Create `deployments/helm/authsentinel/` with Chart.yaml, values.yaml
   - Templates for agent, proxy, redis (optional), configmaps, secrets

2. **Wire OpenTelemetry tracing** (optional):
   - Add OTLP config to agent/proxy
   - Use pkg/observability.Tracer with OTLP exporter

3. **Expand E2E playbook**:
   - Add scenarios: full login flow (redirect to IdP, callback, session), refresh, logout
   - Document in docs/ops/e2e.md

4. **Add proto-generate to CI**:
   - In .github/workflows/ci.yaml, before `go build`, add: `make proto-generate` (or `buf generate`)
   - Ensure proto gen output is either committed or generated in CI

---

## Recommended Implementation Order

Execute in this order to minimize rework and deliver value incrementally:

| Order | Task | Milestone | Effort |
|-------|------|-----------|--------|
| 1 | Add PolicyBundlePath to proxy config; wire BundleLoader | M3 | Small |
| 2 | Add YAML loader to cmd/agent and cmd/proxy (or document JSON-only) | M1 | Small |
| 3 | Add Prometheus metrics to agent | M2 | Small |
| 4 | Create sample WASM policy bundle | M3 | Medium |
| 5 | Add pipeline plugin execution in proxy engine | M4 | Medium |
| 6 | Wire provider plugin in agent | M4 | Medium |
| 7 | Add Helm charts | M6 | Medium |
| 8 | Add proto-generate to CI | M6 | Small |
| 9 | Expand E2E playbook | M6 | Small |
| 10 | Wire OpenTelemetry tracing (optional) | M2/M6 | Medium |

---

## Plan Updates for implementation-plan.md

The following statements in [implementation-plan.md](implementation-plan.md) should be updated:

1. **Part 1, "What is stub or placeholder"** (line ~27): Change *"Rego path is real; wasm.go is a TODO"* to: *"WASM path is implemented (wazero); Rego path is stub. Policy bundle loading from config is not yet wired."*

2. **Part 4, M3** (line ~276): Change *"Finish pkg/policy: WASM host (Wasmtime or equivalent), bundle loader, decision model"* to: *"WASM host and bundle loader done. Remaining: wire PolicyBundlePath in proxy config; add sample WASM bundle; remove or implement Rego stub."*

3. **Part 4, M1** (line ~274): Add: *"Remaining: YAML loader in runtime (or document JSON-only); wire JWKS/session metrics."*

4. **Part 4, M2** (line ~275): Add: *"Remaining: agent Prometheus metrics; optional tracing."*

5. **Part 4, M6** (line ~279): Add: *"Remaining: Helm charts; proto-generate in CI; expand E2E; optional tracing."*

---

## Summary

The AuthSentinel repo is in strong shape. Config, proto, SDKs, and gateway adapters are done. The main gaps are:

1. **Policy bundle not loadable from config** — proxy always uses fallback allow
2. **Agent lacks Prometheus metrics** — observability asymmetry
3. **YAML config not loadable at runtime** — configs/ has YAML but runtime uses JSON only
4. **Pipeline/provider plugins not executed** — registry exists but not wired into request flow
5. **Helm charts missing** — docs reference them but none exist
6. **Proto gen in CI** — implemented (CI runs `make proto-generate` before Go build/test)

Addressing items 1–3 and 6 first will yield the highest impact with minimal effort.
