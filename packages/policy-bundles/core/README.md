# authsentinel-policy-bundle-core

Core policy bundle for AuthSentinel containing fundamental authorization rules shared across protocols.

Policy artifacts will live under the `policies/` directory.

## Included policies

- `policies/allow_all.rego`: always allow (200)
- `policies/deny_all.rego`: always deny (403)

## Output contract

AuthSentinel evaluates the query `data.authsentinel.decision` and expects an object shaped like:

- `allow` (bool)
- `status_code` (number)
- `reason` (string, optional)
- `headers` (object string→string, optional)
- `obligations` (object, optional)

These are mapped to `pkg/policy.Decision`.

