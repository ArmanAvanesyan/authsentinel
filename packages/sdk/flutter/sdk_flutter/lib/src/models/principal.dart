/// Principal model aligned with AuthSentinel proto sdk/v1.
class Principal {
  const Principal({
    required this.subject,
    this.scopes,
    this.roles,
    this.claims,
    this.tenantContext,
    this.accessToken,
    this.expiresAt,
  });

  final String subject;
  final List<String>? scopes;
  final List<String>? roles;
  final Map<String, dynamic>? claims;
  final Map<String, dynamic>? tenantContext;
  final String? accessToken;
  final int? expiresAt;

  factory Principal.fromJson(Map<String, dynamic> json) {
    return Principal(
      subject: json['sub'] as String? ?? '',
      scopes: (json['scopes'] as List<dynamic>?)?.cast<String>(),
      roles: (json['roles'] as List<dynamic>?)?.cast<String>(),
      claims: json['claims'] as Map<String, dynamic>?,
      tenantContext: json['tenant_context'] as Map<String, dynamic>?,
      accessToken: json['access_token'] as String?,
      expiresAt: json['expires_at'] as int?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'sub': subject,
      if (scopes != null) 'scopes': scopes,
      if (roles != null) 'roles': roles,
      if (claims != null) 'claims': claims,
      if (tenantContext != null) 'tenant_context': tenantContext,
      if (accessToken != null) 'access_token': accessToken,
      if (expiresAt != null) 'expires_at': expiresAt,
    };
  }
}
