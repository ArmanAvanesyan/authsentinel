
/// Session status returned by the agent /session endpoint.
enum SessionStatus {
  unknown,
  authenticated,
  unauthenticated,
}

/// User info returned in /session and /me (aligned with agent API).
class SessionUser {
  const SessionUser({
    required this.sub,
    this.email,
    this.preferredUsername,
    this.name,
    this.roles,
    this.groups,
    this.isAdmin,
    this.tenantContext,
    this.claims,
  });

  final String sub;
  final String? email;
  final String? preferredUsername;
  final String? name;
  final List<String>? roles;
  final List<String>? groups;
  final bool? isAdmin;
  final Map<String, dynamic>? tenantContext;
  final Map<String, dynamic>? claims;

  factory SessionUser.fromJson(Map<String, dynamic> json) {
    return SessionUser(
      sub: json['sub'] as String? ?? '',
      email: json['email'] as String?,
      preferredUsername: json['preferred_username'] as String?,
      name: json['name'] as String?,
      roles: (json['roles'] as List<dynamic>?)?.cast<String>(),
      groups: (json['groups'] as List<dynamic>?)?.cast<String>(),
      isAdmin: json['is_admin'] as bool?,
      tenantContext: json['tenant_context'] as Map<String, dynamic>?,
      claims: json['claims'] as Map<String, dynamic>?,
    );
  }
}

/// Session info returned by GET /session (is_authenticated, user).
class SessionInfo {
  const SessionInfo({
    required this.status,
    this.user,
  });

  final SessionStatus status;
  final SessionUser? user;

  bool get isAuthenticated => status == SessionStatus.authenticated;

  factory SessionInfo.fromJson(Map<String, dynamic> json) {
    final isAuth = json['is_authenticated'] as bool? ?? false;
    final status = isAuth
        ? SessionStatus.authenticated
        : SessionStatus.unauthenticated;
    SessionUser? user;
    final userJson = json['user'];
    if (userJson is Map<String, dynamic>) {
      user = SessionUser.fromJson(userJson);
    }
    return SessionInfo(status: status, user: user);
  }
}
