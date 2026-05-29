import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Base URL injected at compile time via:
///   flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8081
/// Android emulator: 10.0.2.2 → host loopback.
/// iOS simulator + macOS: localhost works directly.
const _defaultBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://localhost:8081',
);

Dio buildApiClient({String? baseUrl}) {
  return Dio(
    BaseOptions(
      baseUrl: baseUrl ?? _defaultBaseUrl,
      connectTimeout: const Duration(seconds: 5),
      receiveTimeout: const Duration(seconds: 15),
      responseType: ResponseType.json,
      headers: {'Accept': 'application/json'},
    ),
  );
}

final apiClientProvider = Provider<Dio>((ref) => buildApiClient());
