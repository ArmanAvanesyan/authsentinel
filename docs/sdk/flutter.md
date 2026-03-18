# Flutter SDK

The **authsentinel_sdk_flutter** package provides Flutter (and Dart) clients for web and mobile: session and auth APIs, principal/session models, and an **AuthScope** widget that refreshes session on app resume.

## Package layout

- **lib/src/auth/auth_client.dart** — Auth client (login/logout URLs, session, refresh).
- **lib/src/session/session_client.dart** — Session client (get session, refresh).
- **lib/src/models/** — **Principal**, **SessionUser**, **SessionInfo**, **AuthContext**.
- **lib/src/widgets/auth_scope.dart** — **AuthScope** widget and refresh-on-resume (e.g. `WidgetsBindingObserver`).
- **lib/src/platform/platform_adapter.dart** — **PlatformAdapter** for launching URLs (e.g. mobile vs web); **UrlLauncherAdapter** for opening login/logout in browser or in-app.

## Usage

1. **Configure the client** with the agent base URL (and cookie name if needed). The client calls the agent’s `/session`, `/login`, `/logout`, `/refresh` as appropriate.
2. **Wrap your app (or subtree) in AuthScope** so that session is refreshed when the app resumes (e.g. after background or deep link).
3. **Read session/principal** from the session client or from context provided by AuthScope (if exposed via InheritedWidget or similar).
4. **Platform**: On mobile, use a platform adapter that opens login/logout URLs in the system browser or in-app WebView; on web, same-origin or CORS-with-credentials to the agent.

## Models

- **Principal** — Subject, roles, claims, expiry, optional access token and tenant context.
- **SessionUser** — User payload from `/session` (sub, email, name, roles, etc.).
- **SessionInfo** / **AuthContext** — Session state and user for UI (authenticated vs unauthenticated).

## References

- [packages/sdk/flutter](../../packages/sdk/flutter) — Source and `pubspec.yaml`.
- [Agent flows](../runtime/agent-flows.md) — /session, /login, /logout, /refresh.
- [Configuration](../ops/configuration.md) — Cookie name, agent URL, CORS.
