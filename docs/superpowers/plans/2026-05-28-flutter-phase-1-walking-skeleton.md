# Flutter Mobile App — Phase 1: Walking Skeleton

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stand up a `mobile/` Flutter project that boots, fetches `/api/venues` from the local Go server, and renders the result as a plain `ListView` — validating the full Dio + freezed + Riverpod pipeline end-to-end before any other feature is built.

**Architecture:** Riverpod `AsyncNotifierProvider` reads from a `VenueRepository`, which calls a Dio-based `ApiClient` configured via `--dart-define=API_BASE_URL=...`. Models are immutable, code-generated with `freezed` + `json_serializable`. No persistence, no map, no filters yet — those are later phases.

**Tech Stack:** Flutter (stable channel), Dart 3, `flutter_riverpod` 2.x, `dio` 5.x, `freezed` 2.x + `json_serializable`, `mocktail` for tests.

**Spec reference:** `docs/superpowers/specs/2026-05-28-flutter-mobile-app-design.md`

**Phase context:** This is Phase 1 of 10. Subsequent phases (Map MVP, Three resources + detail sheets, Filters, Favorites + theming, Offline robustness, Tile pre-warm, Distribution prep, Closed beta, Public launch) each get their own plan after this one is implemented and verified.

**Out of scope for Phase 1:**
- Map (`flutter_map`, FMTC) — Phase 2
- Events and exhibitions data + detail sheets — Phase 3
- Filters and search — Phase 4
- Favorites and theming — Phase 5
- `JsonFileCache`, `If-Modified-Since`/304 handling — Phase 6
- Tile pre-warm — Phase 7
- Icons, splash, signing, CI release pipeline, store metadata — Phase 8

---

## File Structure

Files created in this phase, with one-line responsibilities:

| Path | Responsibility |
|------|----------------|
| `mobile/pubspec.yaml` | Package metadata + dependency list |
| `mobile/analysis_options.yaml` | Lints (extends `flutter_lints` package) |
| `mobile/.gitignore` | Ignore Flutter build artefacts |
| `mobile/README.md` | How to run the app + tests locally |
| `mobile/lib/main.dart` | Entry point, wires up `ProviderScope` |
| `mobile/lib/app.dart` | `MaterialApp` shell (placeholder home for now) |
| `mobile/lib/core/api_client.dart` | Dio instance, base URL from `--dart-define` |
| `mobile/lib/models/venue.dart` | `Venue`, `Address`, `Coordinates`, `Contact`, `VenueCategory` freezed classes |
| `mobile/lib/models/venues_response.dart` | Wraps `{"venues": [...], "total": N}` response envelope |
| `mobile/lib/repositories/venue_repo.dart` | `VenueRepository.fetchAll()` — no caching yet |
| `mobile/lib/providers/venues_provider.dart` | `AsyncNotifier<List<Venue>>` exposing `venuesProvider` |
| `mobile/lib/features/home/home_screen.dart` | Plain `ListView` showing venue names; handles loading / error / data states |
| `mobile/test/fixtures/venues_response_sample.json` | One-venue fixture lifted from `data/venues.json` |
| `mobile/test/models/venue_test.dart` | JSON round-trip + missing-optional-field tests |
| `mobile/test/models/venues_response_test.dart` | Wrapper parsing test |
| `mobile/test/repositories/venue_repo_test.dart` | Mocked-Dio repository test |
| `mobile/test/providers/venues_provider_test.dart` | `ProviderContainer` test with mocked repository |
| `mobile/test/features/home/home_screen_test.dart` | Widget tests for loading / error / data states |
| `.github/workflows/mobile-ci.yml` | Format + analyze + test on every push that touches `mobile/**` |

Generated files (do NOT commit by hand — `build_runner` writes them):
- `mobile/lib/models/venue.freezed.dart`
- `mobile/lib/models/venue.g.dart`
- `mobile/lib/models/venues_response.freezed.dart`
- `mobile/lib/models/venues_response.g.dart`

(These ARE checked in once generated, per the `freezed` project's recommendation for app code.)

---

## Task 1: Scaffold the `mobile/` Flutter project

**Files:**
- Create: `mobile/` (via `flutter create`)
- Create: `mobile/pubspec.yaml` (overwrites the scaffold default)
- Create: `mobile/analysis_options.yaml` (overwrites the scaffold default)
- Modify: `.gitignore` at repo root (add `mobile/build/`, `mobile/.dart_tool/`, etc.)

- [ ] **Step 1: Confirm Flutter is installed**

Run: `flutter --version`
Expected: prints a Flutter version on the stable channel (any 3.x is fine). If not installed, install via https://docs.flutter.dev/get-started/install and re-run.

- [ ] **Step 2: Scaffold the project**

Run from repo root:
```bash
flutter create \
  --project-name klp_guide \
  --org de.musche \
  --platforms ios,android \
  --description "Inoffizieller Guide zur Kulturellen Landpartie im Wendland." \
  mobile
```
Expected: creates `mobile/` with `lib/`, `ios/`, `android/`, `pubspec.yaml`, `test/`, etc. Last printed line should be `All done!`.

- [ ] **Step 3: Replace `mobile/pubspec.yaml` with the Phase 1 dependency set**

Write `mobile/pubspec.yaml`:
```yaml
name: klp_guide
description: "Inoffizieller Guide zur Kulturellen Landpartie im Wendland."
publish_to: "none"
version: 0.1.0+1

environment:
  sdk: ">=3.3.0 <4.0.0"
  flutter: ">=3.22.0"

dependencies:
  flutter:
    sdk: flutter
  flutter_riverpod: ^2.5.1
  dio: ^5.4.0
  freezed_annotation: ^2.4.1
  json_annotation: ^4.8.1

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^3.0.1
  build_runner: ^2.4.8
  freezed: ^2.4.7
  json_serializable: ^6.7.1
  mocktail: ^1.0.3

flutter:
  uses-material-design: true
```

- [ ] **Step 4: Replace `mobile/analysis_options.yaml` with stricter lints**

Write `mobile/analysis_options.yaml`:
```yaml
include: package:flutter_lints/flutter.yaml

analyzer:
  exclude:
    - "**/*.freezed.dart"
    - "**/*.g.dart"
  language:
    strict-casts: true
    strict-inference: true
    strict-raw-types: true

linter:
  rules:
    - prefer_single_quotes
    - prefer_const_constructors
    - prefer_const_declarations
    - sort_pub_dependencies
    - unawaited_futures
    - require_trailing_commas
```

- [ ] **Step 5: Append Flutter build artefacts to the repo-root `.gitignore`**

Append to `.gitignore` at repo root:
```
# Flutter (mobile/)
mobile/.dart_tool/
mobile/.flutter-plugins
mobile/.flutter-plugins-dependencies
mobile/.packages
mobile/build/
mobile/ios/Pods/
mobile/ios/.symlinks/
mobile/ios/Flutter/Flutter.framework
mobile/ios/Flutter/Flutter.podspec
mobile/ios/Runner.xcworkspace/xcuserdata/
mobile/ios/Runner.xcodeproj/xcuserdata/
mobile/android/.gradle/
mobile/android/local.properties
mobile/android/app/debug/
mobile/android/app/profile/
mobile/android/app/release/
```

- [ ] **Step 6: Resolve dependencies**

Run: `cd mobile && flutter pub get`
Expected: prints `Got dependencies!` with no errors. If any version conflicts, adjust constraints in pubspec.yaml until resolved.

- [ ] **Step 7: Sanity-check the scaffold builds**

Run: `cd mobile && flutter analyze`
Expected: `No issues found!` (the default scaffold passes).

Run: `cd mobile && flutter test`
Expected: `All tests passed!` (the default scaffold ships one passing test in `test/widget_test.dart`).

- [ ] **Step 8: Delete the default scaffold test (we'll write our own)**

Run: `rm mobile/test/widget_test.dart`

- [ ] **Step 9: Commit**

```bash
git add mobile/ .gitignore
git commit -m "Scaffold mobile/ Flutter project (Phase 1, walking skeleton)"
```

---

## Task 2: Bootstrap the Riverpod app shell

**Files:**
- Modify: `mobile/lib/main.dart`
- Create: `mobile/lib/app.dart`

- [ ] **Step 1: Replace `mobile/lib/main.dart` with a Riverpod-wrapped entry point**

Write `mobile/lib/main.dart`:
```dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'app.dart';

void main() {
  runApp(const ProviderScope(child: KlpGuideApp()));
}
```

- [ ] **Step 2: Create `mobile/lib/app.dart` with a placeholder home**

Write `mobile/lib/app.dart`:
```dart
import 'package:flutter/material.dart';

class KlpGuideApp extends StatelessWidget {
  const KlpGuideApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'KLP-Guide',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(useMaterial3: true, colorSchemeSeed: Colors.deepOrange),
      home: const Scaffold(
        body: Center(child: Text('KLP-Guide bootstraps.')),
      ),
    );
  }
}
```

- [ ] **Step 3: Sanity check**

Run: `cd mobile && flutter analyze`
Expected: `No issues found!`

Run: `cd mobile && flutter test`
Expected: `All tests passed!` (no tests yet, but the framework should report success or "No tests ran" — neither is a failure).

- [ ] **Step 4: Commit**

```bash
git add mobile/lib/
git commit -m "Bootstrap Riverpod ProviderScope + MaterialApp shell"
```

---

## Task 3: Define the `Venue` model with freezed/json_serializable (TDD)

**Files:**
- Create: `mobile/lib/models/venue.dart`
- Create: `mobile/test/fixtures/venues_response_sample.json`
- Create: `mobile/test/models/venue_test.dart`

The shape is mirrored from `internal/storage/models.go`. Note that the Go API strips `phone` and `email` from `Contact` (see `internal/api/handlers.go` — issue #14 fix), so in practice those fields arrive empty. We still model them for correctness but they default to empty.

- [ ] **Step 1: Create the test fixture from real data**

Write `mobile/test/fixtures/venues_response_sample.json` — a single-venue envelope copied from the first record in `data/venues.json`:
```json
{
  "venues": [
    {
      "id": "venue-4044317c59e8f143",
      "name": "SCHLANZE",
      "description": "Jörn (Bif) Ebersbach",
      "address": {
        "street": "Schlanze 2",
        "postalCode": "29496",
        "city": "Waddeweitz"
      },
      "coordinates": {
        "lat": 52.979339,
        "lng": 10.9805823
      },
      "contact": {},
      "categories": [
        {
          "name": "Café (Kuchen, Getränke)",
          "dates": ["2026-05-14", "2026-05-15"]
        },
        {
          "name": "WC"
        }
      ],
      "bikeRoute": "5",
      "source": "klp",
      "eventCount": 14,
      "exhibitionCount": 5
    }
  ],
  "total": 1
}
```

(The `contact: {}` matches what the API actually returns — phone/email stripped server-side.)

- [ ] **Step 2: Write the failing test**

Write `mobile/test/models/venue_test.dart`:
```dart
import 'dart:convert';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:klp_guide/models/venue.dart';

void main() {
  group('Venue.fromJson', () {
    late Map<String, dynamic> sampleVenueJson;

    setUpAll(() {
      final raw = File('test/fixtures/venues_response_sample.json').readAsStringSync();
      final envelope = jsonDecode(raw) as Map<String, dynamic>;
      sampleVenueJson = (envelope['venues'] as List).first as Map<String, dynamic>;
    });

    test('parses required scalar fields', () {
      final venue = Venue.fromJson(sampleVenueJson);
      expect(venue.id, 'venue-4044317c59e8f143');
      expect(venue.name, 'SCHLANZE');
      expect(venue.description, 'Jörn (Bif) Ebersbach');
      expect(venue.bikeRoute, '5');
      expect(venue.source, 'klp');
      expect(venue.eventCount, 14);
      expect(venue.exhibitionCount, 5);
    });

    test('parses nested Address', () {
      final venue = Venue.fromJson(sampleVenueJson);
      expect(venue.address.street, 'Schlanze 2');
      expect(venue.address.postalCode, '29496');
      expect(venue.address.city, 'Waddeweitz');
    });

    test('parses nested Coordinates', () {
      final venue = Venue.fromJson(sampleVenueJson);
      expect(venue.coordinates.lat, closeTo(52.979339, 1e-6));
      expect(venue.coordinates.lng, closeTo(10.9805823, 1e-6));
    });

    test('parses empty Contact when API strips phone/email', () {
      final venue = Venue.fromJson(sampleVenueJson);
      expect(venue.contact.phone, isEmpty);
      expect(venue.contact.email, isEmpty);
      expect(venue.contact.website, isEmpty);
    });

    test('parses categories with optional dates', () {
      final venue = Venue.fromJson(sampleVenueJson);
      expect(venue.categories, hasLength(2));
      expect(venue.categories[0].name, 'Café (Kuchen, Getränke)');
      expect(venue.categories[0].dates, ['2026-05-14', '2026-05-15']);
      expect(venue.categories[1].name, 'WC');
      expect(venue.categories[1].dates, isEmpty);
    });

    test('round-trips back to the original JSON via toJson', () {
      final venue = Venue.fromJson(sampleVenueJson);
      final round = venue.toJson();
      // Required fields preserved
      expect(round['id'], 'venue-4044317c59e8f143');
      expect(round['name'], 'SCHLANZE');
      expect(round['coordinates'], {'lat': 52.979339, 'lng': 10.9805823});
    });

    test('tolerates missing optional fields', () {
      final minimal = {
        'id': 'v1',
        'name': 'Test',
        'address': {'street': '', 'postalCode': '', 'city': 'X'},
        'coordinates': {'lat': 53.0, 'lng': 11.0},
        'contact': <String, dynamic>{},
        'eventCount': 0,
        'exhibitionCount': 0,
      };
      final venue = Venue.fromJson(minimal);
      expect(venue.description, isEmpty);
      expect(venue.categories, isEmpty);
      expect(venue.bikeRoute, isEmpty);
      expect(venue.source, isEmpty);
    });
  });
}
```

- [ ] **Step 3: Verify the test fails (model doesn't exist yet)**

Run: `cd mobile && flutter test test/models/venue_test.dart`
Expected: FAIL — error about `'package:klp_guide/models/venue.dart' not found` or similar.

- [ ] **Step 4: Write the model**

Write `mobile/lib/models/venue.dart`:
```dart
import 'package:freezed_annotation/freezed_annotation.dart';

part 'venue.freezed.dart';
part 'venue.g.dart';

@freezed
class Venue with _$Venue {
  const factory Venue({
    required String id,
    required String name,
    @Default('') String description,
    required Address address,
    required Coordinates coordinates,
    @Default(Contact()) Contact contact,
    @Default(<String>[]) List<String> amenities,
    @Default(<VenueCategory>[]) List<VenueCategory> categories,
    @Default('') String bikeRoute,
    @Default('') String source,
    @Default(<String>[]) List<String> eventIds,
    @Default(<String>[]) List<String> exhibitionIds,
    @Default(0) int eventCount,
    @Default(0) int exhibitionCount,
  }) = _Venue;

  factory Venue.fromJson(Map<String, dynamic> json) => _$VenueFromJson(json);
}

@freezed
class Address with _$Address {
  const factory Address({
    @Default('') String street,
    @Default('') String postalCode,
    @Default('') String city,
  }) = _Address;

  factory Address.fromJson(Map<String, dynamic> json) => _$AddressFromJson(json);
}

@freezed
class Coordinates with _$Coordinates {
  const factory Coordinates({
    required double lat,
    required double lng,
  }) = _Coordinates;

  factory Coordinates.fromJson(Map<String, dynamic> json) => _$CoordinatesFromJson(json);
}

@freezed
class Contact with _$Contact {
  const factory Contact({
    @Default('') String phone,
    @Default('') String email,
    @Default('') String website,
  }) = _Contact;

  factory Contact.fromJson(Map<String, dynamic> json) => _$ContactFromJson(json);
}

@freezed
class VenueCategory with _$VenueCategory {
  const factory VenueCategory({
    required String name,
    @Default(<String>[]) List<String> dates,
  }) = _VenueCategory;

  factory VenueCategory.fromJson(Map<String, dynamic> json) => _$VenueCategoryFromJson(json);
}
```

- [ ] **Step 5: Generate freezed + json_serializable code**

Run: `cd mobile && dart run build_runner build --delete-conflicting-outputs`
Expected: writes `venue.freezed.dart` and `venue.g.dart` next to `venue.dart`. Final line: `Succeeded after Ns with N outputs`.

- [ ] **Step 6: Run the test**

Run: `cd mobile && flutter test test/models/venue_test.dart`
Expected: all 7 tests pass.

- [ ] **Step 7: Commit**

```bash
git add mobile/lib/models/ mobile/test/
git commit -m "Add Venue model (freezed + json_serializable) with round-trip tests"
```

---

## Task 4: Define the `VenuesResponse` wrapper (TDD)

**Files:**
- Create: `mobile/lib/models/venues_response.dart`
- Create: `mobile/test/models/venues_response_test.dart`

The Go API wraps every list response in `{"venues": [...], "total": N}` per the spec at `plans/api-specification.md`. The wrapper exists so callers don't have to remember the envelope shape.

- [ ] **Step 1: Write the failing test**

Write `mobile/test/models/venues_response_test.dart`:
```dart
import 'dart:convert';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:klp_guide/models/venues_response.dart';

void main() {
  test('VenuesResponse.fromJson parses envelope', () {
    final raw = File('test/fixtures/venues_response_sample.json').readAsStringSync();
    final response = VenuesResponse.fromJson(jsonDecode(raw) as Map<String, dynamic>);
    expect(response.total, 1);
    expect(response.venues, hasLength(1));
    expect(response.venues.first.name, 'SCHLANZE');
  });

  test('VenuesResponse.fromJson handles empty list', () {
    final response = VenuesResponse.fromJson({'venues': <dynamic>[], 'total': 0});
    expect(response.total, 0);
    expect(response.venues, isEmpty);
  });
}
```

- [ ] **Step 2: Verify the test fails**

Run: `cd mobile && flutter test test/models/venues_response_test.dart`
Expected: FAIL — `'package:klp_guide/models/venues_response.dart' not found`.

- [ ] **Step 3: Write the wrapper**

Write `mobile/lib/models/venues_response.dart`:
```dart
import 'package:freezed_annotation/freezed_annotation.dart';

import 'venue.dart';

part 'venues_response.freezed.dart';
part 'venues_response.g.dart';

@freezed
class VenuesResponse with _$VenuesResponse {
  const factory VenuesResponse({
    required List<Venue> venues,
    required int total,
  }) = _VenuesResponse;

  factory VenuesResponse.fromJson(Map<String, dynamic> json) =>
      _$VenuesResponseFromJson(json);
}
```

- [ ] **Step 4: Regenerate**

Run: `cd mobile && dart run build_runner build --delete-conflicting-outputs`
Expected: writes `venues_response.freezed.dart` and `venues_response.g.dart`.

- [ ] **Step 5: Run the test**

Run: `cd mobile && flutter test test/models/venues_response_test.dart`
Expected: 2 tests pass.

- [ ] **Step 6: Commit**

```bash
git add mobile/lib/models/venues_response.dart mobile/lib/models/venues_response.freezed.dart mobile/lib/models/venues_response.g.dart mobile/test/models/venues_response_test.dart
git commit -m "Add VenuesResponse envelope wrapper"
```

---

## Task 5: Create the `ApiClient`

**Files:**
- Create: `mobile/lib/core/api_client.dart`

No test for this one — it's a thin Dio factory + a Riverpod provider. Real coverage happens via the repository test in Task 6.

- [ ] **Step 1: Write the api client**

Write `mobile/lib/core/api_client.dart`:
```dart
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
```

- [ ] **Step 2: Sanity check**

Run: `cd mobile && flutter analyze`
Expected: `No issues found!`

- [ ] **Step 3: Commit**

```bash
git add mobile/lib/core/
git commit -m "Add Dio ApiClient with base URL from --dart-define"
```

---

## Task 6: Create `VenueRepository.fetchAll()` (TDD)

**Files:**
- Create: `mobile/lib/repositories/venue_repo.dart`
- Create: `mobile/test/repositories/venue_repo_test.dart`

Phase 1 repo has no caching, no 304 handling, no error swallowing — just network + parse. Those land in Phase 6.

- [ ] **Step 1: Write the failing test**

Write `mobile/test/repositories/venue_repo_test.dart`:
```dart
import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:klp_guide/repositories/venue_repo.dart';

class MockDio extends Mock implements Dio {}

void main() {
  late MockDio dio;
  late VenueRepository repo;

  setUpAll(() {
    registerFallbackValue(RequestOptions(path: ''));
  });

  setUp(() {
    dio = MockDio();
    repo = VenueRepository(dio);
  });

  test('fetchAll returns parsed venues on 200', () async {
    final fixturePath = 'test/fixtures/venues_response_sample.json';
    final body = jsonDecode(File(fixturePath).readAsStringSync()) as Map<String, dynamic>;

    when(() => dio.get<Map<String, dynamic>>('/api/venues')).thenAnswer(
      (_) async => Response<Map<String, dynamic>>(
        requestOptions: RequestOptions(path: '/api/venues'),
        statusCode: 200,
        data: body,
      ),
    );

    final venues = await repo.fetchAll();
    expect(venues, hasLength(1));
    expect(venues.first.id, 'venue-4044317c59e8f143');
    expect(venues.first.name, 'SCHLANZE');
  });

  test('fetchAll rethrows DioException on network failure', () async {
    when(() => dio.get<Map<String, dynamic>>('/api/venues')).thenThrow(
      DioException(
        requestOptions: RequestOptions(path: '/api/venues'),
        type: DioExceptionType.connectionTimeout,
      ),
    );

    await expectLater(repo.fetchAll(), throwsA(isA<DioException>()));
  });
}
```

- [ ] **Step 2: Verify the test fails**

Run: `cd mobile && flutter test test/repositories/venue_repo_test.dart`
Expected: FAIL — `'package:klp_guide/repositories/venue_repo.dart' not found`.

- [ ] **Step 3: Write the repository**

Write `mobile/lib/repositories/venue_repo.dart`:
```dart
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../core/api_client.dart';
import '../models/venue.dart';
import '../models/venues_response.dart';

class VenueRepository {
  VenueRepository(this._dio);

  final Dio _dio;

  Future<List<Venue>> fetchAll() async {
    final response = await _dio.get<Map<String, dynamic>>('/api/venues');
    final body = response.data;
    if (body == null) {
      throw StateError('GET /api/venues returned empty body');
    }
    return VenuesResponse.fromJson(body).venues;
  }
}

final venueRepositoryProvider = Provider<VenueRepository>(
  (ref) => VenueRepository(ref.watch(apiClientProvider)),
);
```

- [ ] **Step 4: Run the test**

Run: `cd mobile && flutter test test/repositories/venue_repo_test.dart`
Expected: 2 tests pass.

- [ ] **Step 5: Commit**

```bash
git add mobile/lib/repositories/ mobile/test/repositories/
git commit -m "Add VenueRepository.fetchAll with mocked-Dio test"
```

---

## Task 7: Create `venuesProvider` (TDD)

**Files:**
- Create: `mobile/lib/providers/venues_provider.dart`
- Create: `mobile/test/providers/venues_provider_test.dart`

- [ ] **Step 1: Write the failing test**

Write `mobile/test/providers/venues_provider_test.dart`:
```dart
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:klp_guide/models/venue.dart';
import 'package:klp_guide/providers/venues_provider.dart';
import 'package:klp_guide/repositories/venue_repo.dart';

class MockVenueRepository extends Mock implements VenueRepository {}

final _sampleVenues = [
  const Venue(
    id: 'v1',
    name: 'Alpha',
    address: Address(city: 'Lüchow'),
    coordinates: Coordinates(lat: 53.0, lng: 11.1),
  ),
  const Venue(
    id: 'v2',
    name: 'Beta',
    address: Address(city: 'Dannenberg'),
    coordinates: Coordinates(lat: 53.1, lng: 11.2),
  ),
];

void main() {
  test('venuesProvider exposes data from the repository', () async {
    final repo = MockVenueRepository();
    when(repo.fetchAll).thenAnswer((_) async => _sampleVenues);

    final container = ProviderContainer(
      overrides: [venueRepositoryProvider.overrideWithValue(repo)],
    );
    addTearDown(container.dispose);

    final result = await container.read(venuesProvider.future);
    expect(result, hasLength(2));
    expect(result.first.name, 'Alpha');
  });

  test('venuesProvider surfaces repository errors as AsyncError', () async {
    final repo = MockVenueRepository();
    when(repo.fetchAll).thenThrow(StateError('boom'));

    final container = ProviderContainer(
      overrides: [venueRepositoryProvider.overrideWithValue(repo)],
    );
    addTearDown(container.dispose);

    final result = container.read(venuesProvider);
    expect(result.isLoading, isTrue);

    await Future<void>.delayed(Duration.zero);

    final settled = container.read(venuesProvider);
    expect(settled.hasError, isTrue);
    expect(settled.error, isA<StateError>());
  });
}
```

- [ ] **Step 2: Verify the test fails**

Run: `cd mobile && flutter test test/providers/venues_provider_test.dart`
Expected: FAIL — `'package:klp_guide/providers/venues_provider.dart' not found`.

- [ ] **Step 3: Write the provider**

Write `mobile/lib/providers/venues_provider.dart`:
```dart
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/venue.dart';
import '../repositories/venue_repo.dart';

class VenuesNotifier extends AsyncNotifier<List<Venue>> {
  @override
  Future<List<Venue>> build() {
    final repo = ref.watch(venueRepositoryProvider);
    return repo.fetchAll();
  }
}

final venuesProvider = AsyncNotifierProvider<VenuesNotifier, List<Venue>>(
  VenuesNotifier.new,
);
```

- [ ] **Step 4: Run the test**

Run: `cd mobile && flutter test test/providers/venues_provider_test.dart`
Expected: 2 tests pass.

- [ ] **Step 5: Commit**

```bash
git add mobile/lib/providers/ mobile/test/providers/
git commit -m "Add venuesProvider (AsyncNotifier) with ProviderContainer tests"
```

---

## Task 8: Build `HomeScreen` with loading / error / data states (TDD)

**Files:**
- Create: `mobile/lib/features/home/home_screen.dart`
- Create: `mobile/test/features/home/home_screen_test.dart`

- [ ] **Step 1: Write the failing widget test**

Write `mobile/test/features/home/home_screen_test.dart`:
```dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:klp_guide/features/home/home_screen.dart';
import 'package:klp_guide/models/venue.dart';
import 'package:klp_guide/providers/venues_provider.dart';
import 'package:klp_guide/repositories/venue_repo.dart';

class MockVenueRepository extends Mock implements VenueRepository {}

Future<void> _pumpHome(WidgetTester tester, VenueRepository repo) {
  return tester.pumpWidget(
    ProviderScope(
      overrides: [venueRepositoryProvider.overrideWithValue(repo)],
      child: const MaterialApp(home: HomeScreen()),
    ),
  );
}

void main() {
  testWidgets('shows progress indicator while loading', (tester) async {
    final repo = MockVenueRepository();
    when(repo.fetchAll).thenAnswer(
      (_) => Future<List<Venue>>.delayed(const Duration(milliseconds: 200), () => []),
    );

    await _pumpHome(tester, repo);
    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    // Drain the pending future so the test exits cleanly
    await tester.pump(const Duration(milliseconds: 250));
  });

  testWidgets('shows venue names when data loads', (tester) async {
    final repo = MockVenueRepository();
    when(repo.fetchAll).thenAnswer(
      (_) async => const [
        Venue(
          id: 'v1',
          name: 'SCHLANZE',
          address: Address(city: 'Waddeweitz'),
          coordinates: Coordinates(lat: 53.0, lng: 11.0),
        ),
        Venue(
          id: 'v2',
          name: 'LÜBELNER MÜHLE',
          address: Address(city: 'Küsten'),
          coordinates: Coordinates(lat: 53.0, lng: 11.1),
        ),
      ],
    );

    await _pumpHome(tester, repo);
    await tester.pumpAndSettle();

    expect(find.text('SCHLANZE'), findsOneWidget);
    expect(find.text('LÜBELNER MÜHLE'), findsOneWidget);
    expect(find.text('Waddeweitz'), findsOneWidget);
  });

  testWidgets('shows error message when fetch fails', (tester) async {
    final repo = MockVenueRepository();
    when(repo.fetchAll).thenThrow(StateError('network down'));

    await _pumpHome(tester, repo);
    await tester.pumpAndSettle();

    expect(find.textContaining('Fehler'), findsOneWidget);
  });

  testWidgets('shows empty-state message when list is empty', (tester) async {
    final repo = MockVenueRepository();
    when(repo.fetchAll).thenAnswer((_) async => <Venue>[]);

    await _pumpHome(tester, repo);
    await tester.pumpAndSettle();

    expect(find.textContaining('Keine Veranstaltungsorte'), findsOneWidget);
  });
}
```

- [ ] **Step 2: Verify the test fails**

Run: `cd mobile && flutter test test/features/home/home_screen_test.dart`
Expected: FAIL — `'package:klp_guide/features/home/home_screen.dart' not found`.

- [ ] **Step 3: Build the HomeScreen**

Write `mobile/lib/features/home/home_screen.dart`:
```dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../models/venue.dart';
import '../../providers/venues_provider.dart';

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final venuesAsync = ref.watch(venuesProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('KLP-Guide')),
      body: venuesAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => _ErrorView(error: err),
        data: (venues) => venues.isEmpty
            ? const _EmptyView()
            : _VenueList(venues: venues),
      ),
    );
  }
}

class _VenueList extends StatelessWidget {
  const _VenueList({required this.venues});
  final List<Venue> venues;

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      itemCount: venues.length,
      itemBuilder: (context, index) {
        final v = venues[index];
        return ListTile(
          title: Text(v.name),
          subtitle: Text(v.address.city),
        );
      },
    );
  }
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({required this.error});
  final Object error;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Text(
          'Fehler beim Laden der Daten:\n$error',
          textAlign: TextAlign.center,
        ),
      ),
    );
  }
}

class _EmptyView extends StatelessWidget {
  const _EmptyView();

  @override
  Widget build(BuildContext context) {
    return const Center(child: Text('Keine Veranstaltungsorte gefunden.'));
  }
}
```

- [ ] **Step 4: Run the widget test**

Run: `cd mobile && flutter test test/features/home/home_screen_test.dart`
Expected: 4 tests pass.

- [ ] **Step 5: Commit**

```bash
git add mobile/lib/features/ mobile/test/features/
git commit -m "Add HomeScreen ListView with loading/error/empty/data states"
```

---

## Task 9: Wire `HomeScreen` into the app + manual smoke test

**Files:**
- Modify: `mobile/lib/app.dart`

- [ ] **Step 1: Replace the placeholder home with `HomeScreen`**

Edit `mobile/lib/app.dart` — replace the body of `MaterialApp.home`:
```dart
import 'package:flutter/material.dart';

import 'features/home/home_screen.dart';

class KlpGuideApp extends StatelessWidget {
  const KlpGuideApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'KLP-Guide',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(useMaterial3: true, colorSchemeSeed: Colors.deepOrange),
      home: const HomeScreen(),
    );
  }
}
```

- [ ] **Step 2: Run the full test suite**

Run: `cd mobile && flutter test`
Expected: all tests pass (model + wrapper + repo + provider + widget). Count: 16 tests.

- [ ] **Step 3: Run analyze**

Run: `cd mobile && flutter analyze`
Expected: `No issues found!`

- [ ] **Step 4: Start the Go server in another terminal**

Run from repo root (separate terminal): `go run cmd/server/main.go`
Expected: server boots on port 8081 and prints venue/event/exhibition counts.

Verify the API works directly:
```bash
curl -s http://localhost:8081/api/venues | head -c 200
```
Expected: JSON starting with `{"venues":[...`

- [ ] **Step 5: Launch the Flutter app against the local Go server**

Pick a target device:
- **Android emulator:** `flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8081` (Android emulator maps `10.0.2.2` to the host loopback).
- **iOS simulator (Mac only):** `flutter run --dart-define=API_BASE_URL=http://localhost:8081`
- **Physical device on same Wi-Fi:** use your host machine's LAN IP, e.g. `--dart-define=API_BASE_URL=http://192.168.x.x:8081`.

Expected: app launches; loading indicator briefly visible; then a `ListView` of venue names appears (SCHLANZE, LÜBELNER MÜHLE, etc.). Approx 87 entries.

- [ ] **Step 6: Verify error state manually**

Stop the Go server (Ctrl-C). In the Flutter app, perform a hot restart (press `R` in the `flutter run` terminal — not lowercase `r`, which is just hot reload).
Expected: "Fehler beim Laden der Daten:" message appears.

Restart the Go server and hot-restart the Flutter app to confirm recovery.

- [ ] **Step 7: Commit**

```bash
git add mobile/lib/app.dart
git commit -m "Wire HomeScreen into MaterialApp; Phase 1 walking skeleton end-to-end"
```

---

## Task 10: Add `mobile-ci.yml` GitHub Actions workflow

**Files:**
- Create: `.github/workflows/mobile-ci.yml`

- [ ] **Step 1: Ensure the directory exists**

Run: `mkdir -p .github/workflows`

- [ ] **Step 2: Write the workflow**

Write `.github/workflows/mobile-ci.yml`:
```yaml
name: mobile-ci

on:
  push:
    branches: [master]
    paths:
      - 'mobile/**'
      - '.github/workflows/mobile-ci.yml'
  pull_request:
    paths:
      - 'mobile/**'
      - '.github/workflows/mobile-ci.yml'

jobs:
  test:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: mobile
    steps:
      - uses: actions/checkout@v4

      - uses: subosito/flutter-action@v2
        with:
          channel: stable
          cache: true

      - name: Install dependencies
        run: flutter pub get

      - name: Verify generated files are up to date
        run: |
          dart run build_runner build --delete-conflicting-outputs
          if ! git diff --quiet; then
            echo "Generated files are out of date. Run 'dart run build_runner build --delete-conflicting-outputs' locally and commit."
            git --no-pager diff
            exit 1
          fi

      - name: Format check
        run: dart format --set-exit-if-changed lib test

      - name: Analyze
        run: flutter analyze

      - name: Test
        run: flutter test --reporter expanded
```

- [ ] **Step 3: Verify locally that each step would pass**

Run from repo root:
```bash
cd mobile
dart run build_runner build --delete-conflicting-outputs
git diff --quiet || echo "Uncommitted generated changes — investigate"
dart format --set-exit-if-changed lib test
flutter analyze
flutter test
```
Expected: each command exits cleanly; the `git diff --quiet` returns 0 (nothing to commit).

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/mobile-ci.yml
git commit -m "Add mobile-ci workflow: format, analyze, test, codegen-up-to-date check"
```

---

## Task 11: Add the `mobile/README.md`

**Files:**
- Create: `mobile/README.md`

- [ ] **Step 1: Write the README**

Write `mobile/README.md`:
````markdown
# KLP-Guide (Flutter)

Inoffizieller Mobile-Guide zur Kulturellen Landpartie im Wendland.

> Status: **Phase 1 — Walking Skeleton**. Fetches `/api/venues` from the local Go server and renders the list. See `docs/superpowers/specs/2026-05-28-flutter-mobile-app-design.md` for the overall design.

## Prerequisites

- Flutter SDK ≥ 3.22 (stable channel) — install via https://docs.flutter.dev/get-started/install
- iOS dev: Xcode + a Mac
- Android dev: Android Studio + an AVD or device

## Setup

```bash
cd mobile
flutter pub get
dart run build_runner build --delete-conflicting-outputs
```

The last command writes the `*.freezed.dart` and `*.g.dart` files that the model classes depend on. Re-run it whenever you change a `@freezed` class.

## Run against the local Go server

Start the backend in one terminal:
```bash
# repo root
go run cmd/server/main.go
```
The server listens on `localhost:8081`.

Then launch the Flutter app pointed at the backend. The base URL is injected via `--dart-define=API_BASE_URL=...`:

| Target | Command |
|--------|---------|
| Android emulator | `flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8081` |
| iOS simulator (Mac) | `flutter run --dart-define=API_BASE_URL=http://localhost:8081` |
| Physical device on LAN | `flutter run --dart-define=API_BASE_URL=http://<host-lan-ip>:8081` |

`10.0.2.2` is the Android emulator's mapping to the host loopback. iOS simulator can talk to `localhost` directly.

## Tests

```bash
flutter test           # unit + widget tests
flutter analyze        # static analysis
dart format lib test   # auto-format
```

CI runs all of these on every push to `mobile/**` (see `.github/workflows/mobile-ci.yml`).

## Project layout (Phase 1)

```
lib/
├── main.dart                   # ProviderScope + app entry
├── app.dart                    # MaterialApp shell
├── core/api_client.dart        # Dio factory
├── models/                     # Venue + VenuesResponse (freezed)
├── repositories/venue_repo.dart
├── providers/venues_provider.dart
└── features/home/home_screen.dart
```

Future phases add `flutter_map`, calendar, filters, favorites, offline cache, etc. See the design spec for the full roadmap.
````

- [ ] **Step 2: Commit**

```bash
git add mobile/README.md
git commit -m "Add mobile/ README covering Phase 1 setup, run, test"
```

---

## Phase 1 acceptance criteria

After the final commit:

- [ ] `cd mobile && flutter test` reports **16 tests passing** (7 venue model + 2 wrapper + 2 repo + 2 provider + 4 widget)
- [ ] `cd mobile && flutter analyze` reports `No issues found!`
- [ ] `dart format --set-exit-if-changed lib test` exits 0
- [ ] With the Go server running, `flutter run --dart-define=API_BASE_URL=<base>` boots the app and renders the live venue list
- [ ] Killing the Go server and hot-restarting the app shows the localized error view
- [ ] `mobile-ci.yml` workflow appears in `.github/workflows/`; pushing the branch triggers it and it passes
- [ ] All generated `*.freezed.dart` / `*.g.dart` files are committed

When all of the above hold, Phase 1 is done. The next plan to write is **Phase 2: Map MVP** — swap the `ListView` for `FlutterMap` + clustering and a placeholder venue sheet.
