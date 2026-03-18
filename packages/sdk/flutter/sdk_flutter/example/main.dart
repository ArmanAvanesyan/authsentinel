import 'package:flutter/material.dart';
import 'package:authsentinel_sdk_flutter/authsentinel_sdk_flutter.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    final client = AuthClient(agentBaseUrl: 'https://auth.example.com');
    return MaterialApp(
      home: AuthScope(
        client: client,
        child: Scaffold(
          appBar: AppBar(title: const Text('AuthSentinel Flutter SDK Example')),
          body: const Center(child: SessionStatusView()),
        ),
      ),
    );
  }
}

class SessionStatusView extends StatelessWidget {
  const SessionStatusView({super.key});

  @override
  Widget build(BuildContext context) {
    final loading = context.authLoading;
    final session = context.authSession;
    if (loading) {
      return const CircularProgressIndicator();
    }
    final status = session?.status.name ?? 'unknown';
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text('Session status: $status'),
        const SizedBox(height: 16),
        if (session?.isAuthenticated == true && session?.user != null)
          Text('User: ${session!.user!.sub}'),
        const SizedBox(height: 16),
        Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            FilledButton(
              onPressed: () => context.authClient.login(),
              child: const Text('Login'),
            ),
            const SizedBox(width: 8),
            FilledButton(
              onPressed: () => context.authClient.logout(),
              child: const Text('Logout'),
            ),
          ],
        ),
      ],
    );
  }
}

