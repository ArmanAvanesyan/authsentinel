# Plan: Replace pkg/config with go-config

This document is the actionable plan and to-do list for fully replacing the stub `pkg/config` with the external **go-config** library. See [implementation-plan.md](implementation-plan.md) for the overall architecture.

---

## To-dos

- [ ] **Add go-config dependency** — `go get github.com/ArmanAvanesyan/go-config`; decide loading pattern (direct load into internal structs vs adapter) and required struct tags.
- [ ] **Refactor internal/agent/config** — Add go-config struct tags; remove env-only `Load()`; keep `Validate()` and `KeyLayout()`; add loader hook or have cmd call go-config.
- [ ] **Refactor internal/proxy/config** — Add go-config struct tags; remove env-only `Load()`; keep `Validate()`.
- [ ] **Wire go-config in cmd/agent** — Load config via go-config (file + env), populate `*agentconfig.Config`, call `Validate()`, preserve env-only behavior when no file given.
- [ ] **Wire go-config in cmd/proxy** — Load config via go-config (file + env), populate `*proxyconfig.Config`, call `Validate()`.
- [ ] **Add example configs** — Create `configs/agent.example.yaml` and `configs/proxy.example.yaml` with all main fields and comments.
- [ ] **Remove pkg/config** — Delete `pkg/config/config.go` and `pkg/config/config_test.go`; remove directory; run `go build ./...`.
- [ ] **Update docs** — Edit `docs/architecture/monorepo-structure.md` (remove config from pkg list); update `docs/architecture/implementation-plan.md` (pkg/config removed); add `docs/ops/configuration.md` (how config is loaded, configs/ location, file vs env); update README if it mentions config.

---

## Current state

- **pkg/config**: Stub only (`Shared struct{}`, `Load() (*Shared, error)`). Not imported by any Go code.
- **internal/agent/config**: Real struct; `Load()` from env only; `Validate()`, `KeyLayout()` used by agent.
- **internal/proxy/config**: Real struct; `Load()` from env only; `Validate()` used by proxy.
- **cmd/agent** and **cmd/proxy** use only internal config packages.

## Target state

- Config loading only via **go-config** (file + env, optional flags). Same structs in internal/agent/config and internal/proxy/config; **Validate()** and **KeyLayout()** unchanged.
- **pkg/config** removed.
- **configs/** contains example YAML for agent and proxy.
- Docs updated: no reference to pkg/config; ops config doc added.

## Steps (summary)

1. Add go-config dependency and define loading pattern + struct tags.
2. Refactor internal/agent/config (tags, remove env-only Load, keep Validate/KeyLayout).
3. Refactor internal/proxy/config (tags, remove env-only Load, keep Validate).
4. Wire go-config in cmd/agent and cmd/proxy.
5. Add configs/agent.example.yaml and configs/proxy.example.yaml.
6. Delete pkg/config (config.go and config_test.go).
7. Update docs (monorepo-structure, implementation-plan, new docs/ops/configuration.md, README if needed).
