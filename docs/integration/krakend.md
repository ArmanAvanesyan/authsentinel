# KrakenD integration

AuthSentinel can protect KrakenD endpoints by using the same proxy/policy/token runtime. The KrakenD adapter acts as an endpoint/auth middleware bridge and exposes the shared principal decision output to backends.

## Rule

All adapters call the **same** proxy/policy/token runtime. The KrakenD adapter does not duplicate auth logic; it delegates to `proxy.Engine` and maps the decision to headers and deny responses.

## Config shape

- **endpoint**: KrakenD endpoint pattern this plugin protects (e.g. `/api/*`).
- **upstream_url**: Upstream to forward allowed requests to.
- **require_auth**: When true, unauthenticated requests are denied with 401.

See [schemas/plugins/integration/krakend.schema.json](../../schemas/plugins/integration/krakend.schema.json) and [configs/plugins/krakend.example.json](../../configs/plugins/krakend.example.json).

## Endpoint/auth middleware bridge

- Use `pkg/plugins/krakend.NewPlugin(engine, descriptor, upstreamURL)` and call `Plugin.Handler(next)` to get an `http.Handler`. Wire this handler into KrakenD’s pipeline for the protected endpoint (e.g. via a custom KrakenD plugin or backend that calls AuthSentinel).
- Alternatively, run authsentinel-proxy as a separate service and configure KrakenD to use it as an auth backend (e.g. call the proxy before forwarding to your backends).

## Shared principal decision output

When the engine allows the request, the proxy response’s `UpstreamHeaders` are set on the response. These include the principal decision output that backends expect:

- **X-User-Id**, **X-Roles**, **X-User-Email**, **X-Tenant-Id**, **Authorization**, and any obligation-derived headers.

KrakenD backends receive these headers when the request is forwarded after a successful auth decision.

## Translation and response mapping

- **Request**: Built via `proxy.RequestFromHTTP(r)` (same common proxy input).
- **Allow**: Principal/obligation headers are set; then either `next` is called or the request is proxied to `upstream_url`.
- **Deny**: Proxy status and body are written with `proxy.WriteResponse`.

## Compatibility

See [Compatibility matrix](compatibility-matrix.md).
