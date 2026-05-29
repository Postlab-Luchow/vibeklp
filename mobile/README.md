# KLP-Guide (Flutter)

Inoffizieller Mobile-Guide zur Kulturellen Landpartie im Wendland.

> Status: **Phase 1 — Walking Skeleton**. Fetches `/api/venues` from the local Go server and renders the list. See `../docs/superpowers/specs/2026-05-28-flutter-mobile-app-design.md` for the overall design and `../docs/superpowers/plans/` for the phased implementation plans.

## Toolchain (fvm)

This project pins its Flutter version with [fvm](https://fvm.app) via `.fvmrc` (currently the `stable` channel, Flutter 3.44 / Dart 3.12). **Prefix every Flutter/Dart command with `fvm`** so you use the pinned SDK:

```bash
fvm flutter <args>
fvm dart <args>
```

If you don't use fvm, plain `flutter`/`dart` work too — just make sure your SDK is Flutter ≥ 3.22 (Dart ≥ 3.10).

## Prerequisites

- fvm + a Flutter SDK on the `stable` channel (`fvm install`)
- iOS dev: Xcode + a Mac
- Android dev: Android Studio + an AVD or a physical device

## Setup

```bash
cd mobile
fvm flutter pub get
fvm dart run build_runner build --delete-conflicting-outputs
```

The last command writes the `*.freezed.dart` and `*.g.dart` files that the model classes depend on. Re-run it whenever you change a `@freezed` class. These generated files **are** committed.

## Run against the local Go server

Start the backend in one terminal (from the repo root):

```bash
go run cmd/server/main.go
```

The server listens on `localhost:8081`.

Then launch the Flutter app pointed at the backend. The base URL is injected via `--dart-define=API_BASE_URL=...`:

| Target | Command |
|--------|---------|
| Android emulator | `fvm flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8081` |
| iOS simulator (Mac) | `fvm flutter run --dart-define=API_BASE_URL=http://localhost:8081` |
| Physical device on LAN | `fvm flutter run --dart-define=API_BASE_URL=http://<host-lan-ip>:8081` |

`10.0.2.2` is the Android emulator's mapping to the host loopback. The iOS simulator can talk to `localhost` directly. Without `--dart-define`, the app defaults to `http://localhost:8081`.

## Tests & checks

```bash
fvm flutter test                          # unit + widget tests
fvm flutter analyze                       # static analysis
fvm dart format lib test                  # auto-format (Dart 3.12 "tall" style)
```

CI (`.github/workflows/mobile-ci.yml`) runs, on every push/PR touching `mobile/**`:

1. `build_runner` + a check that committed generated files are up to date
2. `dart format --set-exit-if-changed lib test`
3. `flutter analyze`
4. `flutter test`

If CI fails on step 1, run `fvm dart run build_runner build --delete-conflicting-outputs` locally and commit the result.

## Project layout (Phase 1)

```
lib/
├── main.dart                   # ProviderScope + app entry
├── app.dart                    # MaterialApp shell
├── core/api_client.dart        # Dio factory (base URL via --dart-define)
├── models/                     # Venue + VenuesResponse (freezed)
├── repositories/venue_repo.dart
├── providers/venues_provider.dart
└── features/home/home_screen.dart
```

Future phases add `flutter_map` + clustering, calendar, filters, favorites, an offline cache with `If-Modified-Since`/304 handling, and an opt-in regional map-tile download. See the design spec for the full roadmap.
