# Traefik integration

AuthSentinel can run behind Traefik so that Traefik uses the AuthSentinel proxy for authentication (ForwardAuth) or an in-process middleware that delegates to the same proxy/policy/token runtime.

## Rule

All adapters call the **same** proxy/policy/token runtime. The Traefik adapter does not duplicate auth logic; it either uses an in-process `proxy.Engine` or forwards to authsentinel-proxy (ForwardAuth).

## Config shape

- **auth_url**: AuthSentinel proxy base URL when using ForwardAuth.
- **upstream_url**: Upstream to proxy to when the adapter performs upstream handoff (in-process).
- **auth_response_headers**: Optional list of header names to copy from the auth response to the forwarded request. Empty means forward `X-*` and `Authorization`.

See [schemas/plugins/integration/traefik.schema.json](../../schemas/plugins/integration/traefik.schema.json) and [configs/plugins/traefik.example.yaml](../../configs/plugins/traefik.example.yaml).

## Middleware plugin interface

- **In-process**: Use `pkg/plugins/traefik.NewMiddleware(engine, descriptor, upstreamURL, authResponseHeaders)` and call `Middleware.Handler(next)` to get an `http.Handler`. Wire this as your Traefik plugin or custom middleware.
- **ForwardAuth**: Configure Traefik’s `forwardAuth` middleware to point at authsentinel-proxy. Traefik sends the request to the proxy; 2xx + headers = allow, and the listed headers are copied to the request passed to the backend.

## Request context handoff

- The adapter builds a `proxy.Request` from the incoming HTTP request via `proxy.RequestFromHTTP(r)`. Traefik passes the same `*http.Request`; no extra context is required beyond the request and response writer.

## Header and deny response mapping

- **Allow**: Headers from the proxy response (`UpstreamHeaders`) are filtered by `auth_response_headers` (or default X-* and Authorization) and set on the response so Traefik can forward them to the backend. Set-Cookies are written.
- **Deny**: The proxy response status code and body are written as the response (no backend call).

## Example config

See [configs/plugins/traefik.example.yaml](../../configs/plugins/traefik.example.yaml) for a dynamic configuration example using ForwardAuth.

## Compatibility

See [Compatibility matrix](compatibility-matrix.md).
