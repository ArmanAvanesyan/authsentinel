# Gateway adapters – compatibility matrix

All gateway adapters call the same proxy/policy/token runtime. This matrix describes supported gateway versions and integration modes.

| Gateway   | Adapter package           | Supported versions | Integration modes                    | Config schema |
|----------|---------------------------|--------------------|-------------------------------------|----------------|
| **Caddy**   | `pkg/plugins/caddy`       | Caddy v2.x         | Forward-auth (built-in), optional `authsentinel` directive (build with `-tags caddy`) | [caddy.schema.json](../../schemas/plugins/integration/caddy.schema.json) |
| **Traefik** | `pkg/plugins/traefik`    | Traefik v2.x, v3.x | ForwardAuth middleware, in-process Handler | [traefik.schema.json](../../schemas/plugins/integration/traefik.schema.json) |
| **KrakenD** | `pkg/plugins/krakend`    | KrakenD v2.x       | In-process Handler (endpoint/auth bridge), proxy as auth backend | [krakend.schema.json](../../schemas/plugins/integration/krakend.schema.json) |

## Runtime

- **AuthSentinel proxy**: All adapters assume authsentinel-proxy (and authsentinel-agent for login flows) are run and configured as described in [Configuration](../ops/configuration.md) and [OAuth Proxy](../runtime/oauth-proxy.md).
- **Plugin API**: Adapters implement `pluginapi.IntegrationPlugin` and use `pluginapi.PluginDescriptor`; they are compatible with the plugin registry and discovery in `pkg/pluginregistry` and `pkg/plugindiscovery`.

## Response mapping summary

| Scenario | Caddy | Traefik | KrakenD |
|---------|--------|---------|---------|
| Allow   | Set upstream headers + next or proxy | Set filtered headers + next or proxy | Set principal headers + next or proxy |
| Deny   | Write proxy status/body | Write proxy status/body | Write proxy status/body |
| Cookies | Set-Cookie from proxy response | Set-Cookie from proxy response | Set-Cookie from proxy response |

## Example configs

- Caddy: [configs/plugins/caddy.Caddyfile](../../configs/plugins/caddy.Caddyfile), [caddy.example.json](../../configs/plugins/caddy.example.json)
- Traefik: [configs/plugins/traefik.example.yaml](../../configs/plugins/traefik.example.yaml)
- KrakenD: [configs/plugins/krakend.example.json](../../configs/plugins/krakend.example.json)
