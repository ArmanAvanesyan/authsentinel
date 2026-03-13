# Coding standards

## Package rules

- Packages expose interfaces, not concrete chaos.
- No package imports from `apps/*`.
- Plugins depend on proxy core, never the reverse.
- SDK helpers depend on token/claims contracts, not runtime servers.
- GraphQL and gRPC adapters extend normalization, not policy logic.

## Naming rules

- Public type names are domain-oriented.
- Avoid `util`, `helper`, `common`.
- Every package has one sentence of purpose in its README.

## Security rules

- Cookie model centralized in `authsentinel-cookie`.
- Token parsing centralized in `authsentinel-token`.
- Policy input model centralized in `authsentinel-policy`.
- No raw JWT parsing duplicated in plugins or SDKs.

