import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/session.dart';

/// Low-level HTTP client for the AuthSentinel agent /session and /refresh endpoints.
/// Pass [sessionCookie] when you have it (e.g. from secure storage on mobile or from cookie on web).
class SessionClient {
  SessionClient({
    required this.agentBaseUrl,
    this.cookieName = 'sess',
    http.Client? httpClient,
  })  : _baseUrl = agentBaseUrl.replaceAll(RegExp(r'/$'), ''),
        _http = httpClient ?? http.Client();

  final String agentBaseUrl;
  final String cookieName;
  final String _baseUrl;
  final http.Client _http;

  /// GET /session. Returns [SessionInfo] or [SessionStatus.unknown] on error.
  Future<SessionInfo> getSession({String? sessionCookie}) async {
    try {
      final headers = <String, String>{};
      if (sessionCookie != null && sessionCookie.isNotEmpty) {
        headers['Cookie'] = '$cookieName=$sessionCookie';
      }
      final res = await _http.get(
        Uri.parse('$_baseUrl/session'),
        headers: headers.isEmpty ? null : headers,
      );
      if (res.statusCode != 200) {
        return const SessionInfo(status: SessionStatus.unauthenticated);
      }
      final data = jsonDecode(res.body) as Map<String, dynamic>;
      return SessionInfo.fromJson(data);
    } catch (_) {
      return const SessionInfo(status: SessionStatus.unknown);
    }
  }

  /// GET /refresh. Returns true if the agent returned 200 and optionally set a new cookie.
  Future<RefreshResult> refresh({String? sessionCookie}) async {
    try {
      final headers = <String, String>{};
      if (sessionCookie != null && sessionCookie.isNotEmpty) {
        headers['Cookie'] = '$cookieName=$sessionCookie';
      }
      final res = await _http.get(
        Uri.parse('$_baseUrl/refresh'),
        headers: headers.isEmpty ? null : headers,
      );
      final setCookie = res.headers['set-cookie'];
      return RefreshResult(
        refreshed: res.statusCode == 200,
        setCookie: setCookie,
      );
    } catch (_) {
      return const RefreshResult(refreshed: false);
    }
  }

  /// Builds the login URL; redirect the user here to start login.
  String getLoginUrl([String? returnUrl]) {
    if (returnUrl == null || returnUrl.isEmpty) {
      return '$_baseUrl/login';
    }
    return '$_baseUrl/login?redirect_to=${Uri.encodeComponent(returnUrl)}';
  }

  /// Builds the logout URL; redirect the user here to log out.
  String getLogoutUrl([String? redirectTo]) {
    if (redirectTo == null || redirectTo.isEmpty) {
      return '$_baseUrl/logout';
    }
    return '$_baseUrl/logout?redirect_to=${Uri.encodeComponent(redirectTo)}';
  }
}

/// Result of a refresh call.
class RefreshResult {
  const RefreshResult({required this.refreshed, this.setCookie});

  final bool refreshed;
  final String? setCookie;
}
