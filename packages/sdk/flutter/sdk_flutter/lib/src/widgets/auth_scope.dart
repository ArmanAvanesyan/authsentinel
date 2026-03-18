import 'package:flutter/widgets.dart';

import '../auth/auth_client.dart';
import '../models/session.dart';

/// Provides [AuthClient] and current [SessionInfo] to the widget tree.
/// Call [AuthScopeState.refresh] to reload session (e.g. on resume).
class AuthScope extends StatefulWidget {
  const AuthScope({
    super.key,
    required this.client,
    required this.child,
  });

  final AuthClient client;
  final Widget child;

  static AuthScopeState of(BuildContext context) {
    final state = context.findAncestorStateOfType<AuthScopeState>();
    if (state == null) {
      throw StateError('AuthScope not found. Wrap your app with AuthScope.');
    }
    return state;
  }

  @override
  State<AuthScope> createState() => AuthScopeState();
}

class AuthScopeState extends State<AuthScope> with WidgetsBindingObserver {
  SessionInfo? _session;
  bool _loading = true;

  SessionInfo? get session => _session;
  bool get loading => _loading;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _load();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      refresh();
    }
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final info = await widget.client.getSession();
      if (mounted) setState(() => _session = info);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  /// Reload session and optionally refresh tokens.
  Future<void> refresh() async {
    await widget.client.refresh();
    await _load();
  }

  @override
  Widget build(BuildContext context) {
    return InheritedAuthScope(
      client: widget.client,
      session: _session,
      loading: _loading,
      refresh: refresh,
      child: widget.child,
    );
  }
}

/// Inherited widget for reading auth state.
class InheritedAuthScope extends InheritedWidget {
  const InheritedAuthScope({
    super.key,
    required this.client,
    required this.session,
    required this.loading,
    required this.refresh,
    required super.child,
  });

  final AuthClient client;
  final SessionInfo? session;
  final bool loading;
  final Future<void> Function() refresh;

  static InheritedAuthScope? maybeOf(BuildContext context) {
    return context.dependOnInheritedWidgetOfExactType<InheritedAuthScope>();
  }

  static InheritedAuthScope of(BuildContext context) {
    final scope = maybeOf(context);
    if (scope == null) {
      throw StateError(
        'InheritedAuthScope not found. Wrap your app with AuthScope.',
      );
    }
    return scope;
  }

  @override
  bool updateShouldNotify(InheritedAuthScope oldWidget) {
    return client != oldWidget.client ||
        session != oldWidget.session ||
        loading != oldWidget.loading;
  }
}

/// Extension to get auth scope from context.
extension AuthScopeContext on BuildContext {
  AuthClient get authClient => InheritedAuthScope.of(this).client;
  SessionInfo? get authSession => InheritedAuthScope.of(this).session;
  bool get authLoading => InheritedAuthScope.of(this).loading;
  Future<void> authRefresh() => InheritedAuthScope.of(this).refresh();
}
