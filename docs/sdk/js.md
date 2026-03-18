# JavaScript / TypeScript SDK

The AuthSentinel JS/TS SDK lives under **packages/sdk/js** and provides:

- **@authsentinel/sdk-core** — Browser client and shared types (session, principal, auth URLs).
- **@authsentinel/sdk-node** — Node.js middleware for BFFs (session resolution via agent).
- **@authsentinel/sdk-react** — React bindings (AuthProvider, useAuth, useSession).

Use the **core** package in any browser or Node context; add **node** for server-side session resolution and **react** for React apps.

## Core (browser)

**createClient(config)** returns an **AuthClient**:

- **config.agentBaseUrl** — Agent base URL (e.g. `https://auth.example.com`). If omitted, returns a stub client (getSession → status `"unknown"`).
- **getSession()** — Calls GET `/session` with `credentials: "include"`. Returns **SessionInfo** (`status`: `"authenticated"` | `"unauthenticated"` | `"unknown"`, optional **user**).
- **getLoginURL(returnUrl?)** — Returns agent login URL; optional `redirect_to` query.
- **getLogoutURL(redirectTo?)** — Returns agent logout URL.
- **refresh()** — Calls agent refresh; returns **RefreshResult** (`refreshed: boolean`).
- **login()** / **logout()** — Redirect to login/logout URL (e.g. `window.location.href = getLoginURL()`).

Types: **Principal**, **SessionUser**, **SessionInfo**, **SessionStatus**, **AuthClientConfig**, **RefreshResult**. Export from `@authsentinel/sdk-core` (or `@authsentinel/sdk-typescript` for TS re-exports).

## Node middleware

**createNodeMiddleware(options)** — For Express, Fastify, or any `(req, res, next)` style stack:

- **options.agentBaseUrl** — Agent base URL.
- **options.cookieName** — Session cookie name (must match agent).
- **options.forwardSetCookie** — If true (default), forwards agent `Set-Cookie` to the response.

Behavior: reads session cookie from `req.headers.cookie` (or `req.cookies`), calls agent GET `/session`, sets **req.authsentinel** = `{ session, sessionCookie }`. Use `req.authsentinel.session` to check auth and optionally set `Set-Cookie` on the response.

## React

See **[React SDK](react.md)** for **AuthProvider**, **useAuth**, **useSession**, and protected routes. The React package depends on core and uses the same **AuthClient** and types.

## Build and test

From repo root (pnpm workspace):

```bash
pnpm install
pnpm --filter @authsentinel/sdk-core build
pnpm --filter @authsentinel/sdk-node build
pnpm --filter @authsentinel/sdk-react build
```

Tests: vitest in each package (e.g. `packages/sdk/js/core`, `packages/sdk/js/node`).

## References

- [React SDK](react.md) — AuthProvider, useAuth, useSession.
- [Agent flows](../runtime/agent-flows.md) — /session, /login, /logout, /refresh.
- [Configuration](../ops/configuration.md) — Cookie name, agent URL.
