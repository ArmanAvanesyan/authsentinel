# Package boundaries

Core Go packages live under `pkg/`:

- `pkg/cookie`: cookie model, codecs, and manager.
- `pkg/token`: token validation, JWKS, claims, and principal model.
- `pkg/policy`: embedded policy engine (WASM), input/decision contracts.
- `pkg/proxy`: gateway-agnostic proxy engine and middleware.
- `pkg/agent`: OAuth Agent runtime (session, login, refresh, logout).
- `pkg/graphql` / `pkg/grpc`: protocol adapters.
- `pkg/sdk`: server-side SDK helpers (identity, http, graphql, grpc).
- `pkg/testing`: test fixtures and helpers.
- `pkg/plugins/{caddy,traefik,krakend}`: gateway plugins.

Executables are in `cmd/agent` and `cmd/proxy`. App-only code is in `internal/agent` and `internal/proxy`. Shared config is in `pkg/config`.
