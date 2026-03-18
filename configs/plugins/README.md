# Gateway adapter example configs

Example configurations for using AuthSentinel with Caddy, Traefik, and KrakenD. All adapters call the same proxy/policy/token runtime.

- **caddy.Caddyfile** — Caddy using `forward_auth` to authsentinel-proxy; optional `authsentinel` directive when building with `-tags caddy`.
- **caddy.example.json** — Config shape for the Caddy adapter (auth_url, upstream_url, require_auth).
- **traefik.example.yaml** — Traefik dynamic config with ForwardAuth middleware.
- **krakend.example.json** — KrakenD endpoint/auth plugin config (endpoint, upstream_url, require_auth).

See [docs/integration](../../docs/integration/) for full integration guides and the [compatibility matrix](../../docs/integration/compatibility-matrix.md).
