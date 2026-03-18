import 'principal.dart';
import 'session.dart';

/// Combined auth context (principal + session) for use across the SDK.
class AuthContext {
  const AuthContext({
    this.principal,
    this.session,
  });

  final Principal? principal;
  final SessionInfo? session;

  bool get isAuthenticated =>
      session?.isAuthenticated == true || principal != null;
}
