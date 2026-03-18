import '../models/session.dart';
import '../platform/platform_adapter.dart';
import '../session/session_client.dart';

/// High-level auth client: session + refresh + login/logout URL opening.
/// Use [sessionCookie] from your storage (e.g. after WebView login on mobile, or from browser cookie on web).
class AuthClient {
  AuthClient({
    required this.agentBaseUrl,
    this.cookieName = 'sess',
    PlatformAdapter? platformAdapter,
    http.Client? httpClient,
  })  : _sessionClient = SessionClient(
          agentBaseUrl: agentBaseUrl,
          cookieName: cookieName,
          httpClient: httpClient,
        ),
        _adapter = platformAdapter ?? UrlLauncherAdapter();

  final String agentBaseUrl;
  final String cookieName;
  final SessionClient _sessionClient;
  final PlatformAdapter _adapter;

  /// Current session cookie (set by app after login; used for getSession/refresh).
  String? sessionCookie;

  /// Fetches session state. Uses [sessionCookie] if set.
  Future<SessionInfo> getSession() =>
      _sessionClient.getSession(sessionCookie: sessionCookie);

  /// Refreshes the session. If the agent returns a new Set-Cookie, you should persist it and set [sessionCookie].
  Future<RefreshResult> refresh() async {
    final result = await _sessionClient.refresh(sessionCookie: sessionCookie);
    // Caller can read result.setCookie and update sessionCookie / storage.
    return result;
  }

  /// Returns the agent login URL.
  String getLoginUrl([String? returnUrl]) =>
      _sessionClient.getLoginUrl(returnUrl);

  /// Returns the agent logout URL.
  String getLogoutUrl([String? redirectTo]) =>
      _sessionClient.getLogoutUrl(redirectTo);

  /// Opens the login URL (redirects user to agent). Returns true if launched.
  Future<bool> login([String? returnUrl]) =>
      _adapter.openUrl(getLoginUrl(returnUrl));

  /// Opens the logout URL. Returns true if launched.
  Future<bool> logout([String? redirectTo]) =>
      _adapter.openUrl(getLogoutUrl(redirectTo));
}
