# AuthSentinel Protobuf APIs

Protobuf definitions for AuthSentinel live under this directory, organized by domain and version.

- `authsentinel/agent/v1`: Agent login, callback, refresh, logout, and session introspection.
- `authsentinel/proxy/v1`: Request decision, principal introspection, route resolution, and deny reason model.
- `authsentinel/policy/v1`: Policy evaluation request/response, obligations, and trace/debug structures.
- `authsentinel/sdk/v1`: Shared principal, session, and auth context messages used by SDKs.

## Buf workflow

This repository treats proto as the **real contract system**. All generated clients must be produced via `buf`:

- Linting: `make proto-lint` (runs `buf lint` using `proto/buf.yaml`).
- Breaking changes: `make proto-breaking` (runs `buf breaking --against '.git#branch=main'`).
- Code generation: `make proto-generate` (runs `buf generate` using `buf.gen.yaml`).

### Generated code policy

- **Go**: Generated Go code lives under `proto/gen/go/...` and is imported by runtime packages as needed.
- **TypeScript**: Generated TS/ES modules live under `proto/gen/ts/...` and are consumed by JS SDK packages under `packages/sdk/js`.
- **Flutter/Dart**: Flutter SDKs (`packages/sdk/flutter`) must align to the same proto versioning but may use language-specific generators configured alongside `buf` in future revisions.

