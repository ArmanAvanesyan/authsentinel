# Policy engine

The AuthSentinel policy engine evaluates **authorization decisions** for each proxy request. It is embedded in the proxy and shared by all gateway adapters. Input is normalized (protocol, method, path, principal, headers); output is allow/deny with status code, reason, headers, and obligations.

## Contract (pkg/policy)

- **Engine**: `Evaluate(ctx, input Input) (*Decision, error)`.
- **Input**: Protocol, Method, Path, GraphQLOperation, GRPCService, GRPCMethod, Principal (*token.Principal), Headers.
- **Decision**: Allow, StatusCode, Headers, Reason, Obligations (map[string]any). Obligations can drive header injection (e.g. `set_header_X_User` → `X-User`).

Optional **EngineWithStatus**: `Loaded() bool`, `BundlePath() string` for admin/observability.

## WASM runtime (default)

The proxy uses **policy.WASMRuntime** (wazero-based) as the policy engine:

- **ABI**: The WASM module must export **memory** and a function **evaluate(input_ptr, input_len) → (output_ptr, output_len)**. Input and output are JSON.
- **Input JSON**: Matches **policy.Input** (protocol, method, path, principal, headers, etc.).
- **Output JSON**: `allow`, `status_code`, `reason`, `obligations`, `headers`.

If no bundle is loaded or evaluation fails, the engine returns a **fallback** decision: **DefaultFallbackAllow** (allow, 200) or **DefaultFallbackDeny** (deny, 503). The proxy currently uses **DefaultFallbackAllow** when no WASM bundle is provided.

**EngineWithStatus**: WASMRuntime implements **Loaded()** and **BundlePath()** so `/admin` can report policy bundle status.

## Rego (OPA embedded)

The proxy can also run policies written in **Rego** using an embedded OPA evaluator (`pkg/policy.RegoEngine`).

### Contract

- **Package**: `package authsentinel`
- **Query**: `data.authsentinel.decision`
- **Result value**: a single object compatible with `pkg/policy.Decision`:
  - `allow` (bool)
  - `status_code` (number)
  - `reason` (string, optional)
  - `headers` (object string→string, optional)
  - `obligations` (object, optional)

### Input shape (important)

AuthSentinel passes the Go struct `policy.Input` as the Rego `input`. Because the struct has **no JSON tags**, fields are exported as **capitalized keys**:

- `input.Protocol`, `input.Method`, `input.Path`, `input.GraphQLOperation`, `input.GRPCService`, `input.GRPCMethod`, `input.Principal`, `input.Headers`

## Proxy configuration

Proxy config (`internal/proxy/config`) controls policy wiring:

- `policy_engine`: `"wasm"` (default) or `"rego"`
- `policy_bundle_path`: path to `.wasm` (WASM) or `.rego` (Rego)
- `policy_fallback_allow`: boolean; when true fallback is allow(200), else deny(503)

## Bundle loading

- **WASM**: `WASMRuntime.Load(path)` reads a WASM file and compiles/instantiates it; subsequent **Evaluate** calls use that module. Config can set a policy bundle path; proxy startup can load it and pass the engine to **DefaultEngine**.
- **BundleLoader** (e.g. in pkg/policy): Can pre-compile WASM and pass a **wazero.CompiledModule** into **NewWASMRuntimeWithRuntime** for reuse.
 - **Rego**: `RegoEngine.Load(path)` reads and compiles a `.rego` module; subsequent **Evaluate** calls execute `data.authsentinel.decision` with the normalized input.

## Security and sandboxing

WASM execution is sandboxed by wazero (no host access unless explicitly exported). Policy code cannot access the host filesystem or network unless the host exposes imports. AuthSentinel does not expose sensitive host APIs to the policy module; input is normalized and non-secret.

## References

- [pkg/policy](https://pkg.go.dev/github.com/ArmanAvanesyan/authsentinel/pkg/policy) — Engine, Input, Decision, WASMRuntime, RegoEngine.
- [Proxy pipeline](proxy-pipeline.md) — How the engine is used in the request flow.
- [Implementation plan](../architecture/implementation-plan.md) — Phase 1.6 policy, WASM.
