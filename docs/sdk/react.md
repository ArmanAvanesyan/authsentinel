# React SDK

**@authsentinel/sdk-react** provides React bindings for AuthSentinel: **AuthProvider**, **useAuth**, and **useSession** so you can integrate login, logout, and session state into React apps (SPA or BFF-backed).

## Setup

Wrap your app (or the subtree that needs auth) with **AuthProvider** and pass the agent base URL (and optional config):

```tsx
import { AuthProvider } from "@authsentinel/sdk-react";

<AuthProvider agentBaseUrl="https://auth.example.com">
  <App />
</AuthProvider>
```

## Hooks

- **useAuth()** — Returns the **AuthClient** (getSession, getLoginURL, getLogoutURL, refresh, login, logout).
- **useSession()** — Returns current **SessionInfo** (status, user) and typically triggers a session fetch when mounted; use for “current user” and to show login vs logout.

Use **getLoginURL()** and **login()** to send the user to the agent login flow; **getLogoutURL()** and **logout()** for logout. Use **refresh()** after resume or before sensitive operations if needed.

## Protected routes

Check **useSession()** (e.g. `status === "unauthenticated"`) and redirect to **getLoginURL(returnUrl)** when unauthenticated. The agent handles callback and redirect back to your app.

## Types

Session and principal types match **@authsentinel/sdk-core**: **SessionInfo**, **SessionUser**, **Principal**, etc. Re-exported from the React package where applicable.

## References

- [JS/TS SDK](js.md) — Core and Node middleware.
- [Agent flows](../runtime/agent-flows.md) — Login, callback, session, refresh.
- [Configuration](../ops/configuration.md) — Cookie, agent URL, CORS.
