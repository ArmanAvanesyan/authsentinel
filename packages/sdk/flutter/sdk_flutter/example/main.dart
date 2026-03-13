import 'package:flutter/material.dart';
import 'package:authsentinel_sdk_flutter/authsentinel_sdk_flutter.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    final client = SessionClient();
    return MaterialApp(
      home: Scaffold(
        appBar: AppBar(title: const Text('AuthSentinel Flutter SDK Example')),
        body: Center(
          child: FutureBuilder<String>(
            future: client.getStatus(),
            builder: (context, snapshot) {
              return Text('Session status: ${snapshot.data ?? 'loading'}');
            },
          ),
        ),
      ),
    );
  }
}

