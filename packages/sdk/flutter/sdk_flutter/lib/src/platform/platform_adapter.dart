import 'package:url_launcher/url_launcher.dart';

/// Platform adapter for opening login/logout URLs (browser or in-app WebView).
/// Default implementation uses [url_launcher]; override for custom behavior (e.g. in-app WebView).
abstract class PlatformAdapter {
  /// Open the given URL (e.g. login or logout redirect). Returns true if launched.
  Future<bool> openUrl(String url);
}

/// Uses [url_launcher] to open URLs in the system browser or in-app web view.
class UrlLauncherAdapter implements PlatformAdapter {
  @override
  Future<bool> openUrl(String url) => launchUrl(Uri.parse(url));
}
