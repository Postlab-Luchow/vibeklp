# Flutter Mobile App — Design Spec

**Date:** 2026-05-28
**Status:** Design approved, awaiting implementation plan
**Author:** brainstormed with Claude (superpowers:brainstorming)

## Problem

After the May 2026 festival, user feedback called out that an offline-capable app would have been valuable — Wendland's mobile reception is patchy enough that the existing web PWA (which already has a service worker doing stale-while-revalidate) isn't a complete answer. The goal is to ship a real, installable mobile app to both stores in time for the May 2027 festival.

## Scope decisions (locked)

1. **Stack:** Flutter rewrite of the frontend
2. **Backend:** Existing Go API stays; the Flutter app fetches from it at runtime (same model as today's web client)
3. **Web app:** Stays alive in parallel — the Go API serves both clients
4. **v1 feature scope:** map + clustering, calendar, all existing filters (search/date/category/bike-route/time-of-day), favorites, dark mode, modal back-stack. **Route planning is deferred to v2.**
5. **Platforms:** Both iOS and Android from day one (one Flutter codebase)
6. **Offline strategy:** JSON data cached automatically on fetch; map tiles cached opportunistically as the user pans, plus an opt-in regional pre-warm
7. **Brand independence:** App is shipped as an unofficial / fan-made guide. Festival association is in contact with the maintainer

## Architecture summary

**Approach A (chosen):** Pragmatic Riverpod + thin repository layer + `flutter_map`. Rejected alternatives: B (Isar/Drift DB — overkill for a 1.6 MB dataset) and C (no repository — too thin for cross-screen reuse and testability).

```
UI widget
  ↓ ref.watch
Riverpod AsyncNotifierProvider (data) + Provider (derived/filtered)
  ↓ calls
Repository (VenueRepo / EventRepo / ExhibitionRepo)
  ↓ tries network with If-Modified-Since header
ApiClient (Dio) ──→ Go server (unchanged endpoints + new Last-Modified handling)
  ↓ on 200: write cache; on 304: keep cache; on error: read cache
JsonFileCache (path_provider) ──→ <documents>/venues.json, events.json, exhibitions.json
                                  <documents>/cache_meta.json (per-resource Last-Modified value)
```

## Project structure

A new `mobile/` directory at the repo root holds the Flutter project alongside `cmd/`, `internal/`, `web/`.

```
mobile/
├── pubspec.yaml
├── lib/
│   ├── main.dart
│   ├── app.dart                      # MaterialApp.router, theme wiring
│   ├── core/
│   │   ├── api_client.dart           # Dio + base URL injected via --dart-define
│   │   ├── cache.dart                # JsonFileCache (atomic writes via .tmp + rename)
│   │   ├── theme.dart                # ColorScheme lifted from tailwind-input.css
│   │   └── exceptions.dart           # OfflineNoCacheException, etc.
│   ├── models/                       # venue.dart, event.dart, exhibition.dart + .g.dart (freezed + json_serializable)
│   ├── repositories/                 # venue_repo.dart, event_repo.dart, exhibition_repo.dart
│   ├── providers/                    # Riverpod providers (data + filter state + favorites + modal stack)
│   ├── features/
│   │   ├── map/                      # MapScreen, ClusterBubble
│   │   ├── calendar/                 # CalendarScreen, DateSection
│   │   ├── filters/                  # FilterSheet, FilterState notifier
│   │   ├── favorites/                # FavoritesScreen, FavoriteButton
│   │   ├── venue_detail/             # VenueDetailSheet
│   │   ├── event_detail/             # EventDetailSheet
│   │   ├── exhibition_detail/        # ExhibitionDetailSheet
│   │   └── tile_cache/               # OfflineDownloadScreen
│   └── shared/                       # widgets/, utils/, l10n/
└── test/                             # unit + widget mirror of lib/
└── integration_test/                 # full-stack tests against a fixture Go server
```

## Dependencies (pubspec.yaml)

- `flutter_riverpod` + `riverpod_annotation` — state management
- `flutter_map` + `flutter_map_marker_cluster` — OSM map + clustering
- `flutter_map_tile_caching` (FMTC) — tile provider with disk cache + bulk pre-warm API
- `dio` — HTTP client (timeouts, interceptors, cancel tokens)
- `freezed` + `json_serializable` (dev deps) — immutable models with JSON round-trip
- `shared_preferences` — favorites persistence
- `path_provider` — cache directory lookup
- `go_router` — declarative routing, deep links, modal-stack-as-URL
- `flutter_localizations` + `intl` — German formatting (DateTime, NumberFormat) and i18n hook
- `geolocator` + permission_handler — optional "Mein Standort" button
- `url_launcher` — open external links (festival site, system maps, email/phone if shown)
- `flutter_sticky_header` — sticky date headers in the calendar
- Dev tooling: `flutter_launcher_icons`, `flutter_native_splash`, `build_runner`, `mocktail`, `fake_async`

## Data layer

### Repository contract

```dart
abstract class VenueRepository {
  Future<List<Venue>> fetchAll({bool forceRefresh = false});
  Future<Venue> fetchById(String id);
}
```

`fetchAll` behavior:
- Default path: try network with 5 s connect / 15 s receive timeout, sending `If-Modified-Since` from cache_meta
  - `200`: parse, write cache + meta, return fresh data
  - `304`: return cached data without re-parsing
  - timeout/error: return cached data silently
  - timeout/error + no cache: throw `OfflineNoCacheException`
- `forceRefresh: true`: skip the silent fallback so the user sees the network error (used by manual "Refresh data" action)

`fetchById` prefers the in-memory list (already loaded via `fetchAll`); only hits `/api/<resource>/{id}` on miss.

### Cache layer (`JsonFileCache`)

- One file per resource: `venues.json`, `events.json`, `exhibitions.json` in `getApplicationDocumentsDirectory()`
- Stores raw API body verbatim; decoded into typed models on read
- Atomic writes: write to `<name>.tmp`, then rename — survives mid-write crashes
- Sibling `cache_meta.json` tracks per-resource `lastModified` (verbatim server header value) and `fetchedAt` (UTC ISO timestamp for the freshness indicator)

### API client config

- Base URL via `--dart-define=API_BASE_URL=https://klp.example.com` (same code runs against local dev + prod)
- 5 s connect / 15 s receive timeouts; one automatic retry on 5xx
- No auth — API stays public

### Models

Generated with `freezed` + `json_serializable` from existing JSON shapes. Wrapper classes per list endpoint reflect the `{"venues": [...], "total": N}` envelope.

## 304 / Last-Modified handling (requires small Go change)

### Server side (`internal/api/handlers.go`)

For the three list endpoints (`/api/venues`, `/api/events`, `/api/exhibitions`):
- `os.Stat("data/<file>.json").ModTime()` → emit `Last-Modified: <RFC1123 GMT>` header
- Compare incoming `If-Modified-Since` to that mtime; if `<=`, return `304 Not Modified` (no body)

For `/api/calendar` and `/api/search` (derived from multiple files): use `max(mtime)` across the inputs.

Detail endpoints (`/api/venues/{id}` etc.) skip 304 — payloads are small, and the merged-source mtime is awkward.

CORS middleware needs `Last-Modified` added to `Access-Control-Expose-Headers` so the Flutter client (and any future browser fetch) can read it.

### Client side (Flutter)

- Dio interceptor reads `cache_meta.json` and adds `If-Modified-Since: <cached value>` to list-endpoint requests
- Repository handles `200` / `304` / error as above
- Pull-to-refresh sends `If-Modified-Since` too; a `304` updates the "Aktualisiert vor …" stamp without rebuilding the UI

### Why Last-Modified over ETag

File mtime is the natural truth source — no hashing, no state. Second-resolution is fine since crawl runs are minutes apart at most. One-line change per endpoint. Turns a cold app open from ~1.6 MB into ~300 bytes of response headers when the data hasn't changed.

## State management (Riverpod provider tree)

```
Data (long-lived):
  venuesProvider, eventsProvider, exhibitionsProvider — AsyncNotifier<List<X>>
  favoritesProvider                                   — AsyncNotifier<Set<String>>

Filter state:
  filterStateProvider — Notifier<FilterState> with internal 250 ms debounce on setSearch()
    FilterState { searchQuery, dateFilter, category, eventCategory, bikeRoute, timeOfDay }

Derived (computed, auto-memoized):
  filteredVenuesProvider       — venue-centric search (matches if venue OR its events/exhibitions match)
  filteredEventsProvider
  filteredExhibitionsProvider
  calendarSectionsProvider     — groups filtered events by date
  visibleMarkersProvider       — feeds the map cluster layer

UI state:
  activeViewProvider           — StateProvider<AppView> { map, calendar, favorites }
  modalStackProvider           — Notifier<List<ModalEntry>>  (mirrored to go_router URL)
  dataFreshnessProvider        — Provider<DateTime?> reading cache_meta
```

### Startup flow

1. `ProviderScope` wraps `MaterialApp.router`
2. A top-level `AppStartup` widget eagerly reads the three data providers in parallel — single cold-start fetch, not per-screen
3. While loading, show skeleton; on error with cache → silently use cache; on error with no cache → "Keine Verbindung — zum Festival einmal mit Internet laden"
4. `AppLifecycleListener` invalidates the three data providers after >1 h background → next foreground re-issues with `If-Modified-Since`

## UI / screens

### Navigation

`HomeScaffold` with `BottomNavigationBar` (Karte / Kalender / Favoriten), `AppBar` containing a debounced search field, a filter button (badge with active count), and overflow menu (Daten aktualisieren / Einstellungen).

### Screens

- **MapScreen** — full-bleed `FlutterMap` with OSM tiles via FMTC + marker clustering. Marker tap → `pushModal('venue', id)`. Top-right freshness chip. Bottom-right "Mein Standort" FAB (uses `geolocator` with runtime permission prompt).
- **CalendarScreen** — `CustomScrollView` with sticky date headers; `EventCard` per item. Tap → `pushModal('event', id)`.
- **FavoritesScreen** — list of favorited venues with next-up event time if today. Empty state explains how to favorite.
- **VenueDetailSheet** — `DraggableScrollableSheet`. Header (name, address tap → system maps, amenities, bike-route badge). Action row (favorite, "Auf Karte zeigen", "Teilen"). Sections for events + exhibitions at this venue.
- **EventDetailSheet / ExhibitionDetailSheet** — similar layout; tapping the venue name swaps the modal-stack top to that venue.
- **FilterSheet** — bottom sheet with date filter chips, time-of-day chips, category / event-category / bike-route dropdowns, "Filter zurücksetzen", live application.
- **OfflineDownloadScreen** — region preview, "≈ 55 MB" estimate, progress bar during bulk download, cancel button.
- **SettingsScreen** — freshness + manual refresh, offline maps entry, cache size + "Leeren" action, app version, "Web-App öffnen", Impressum.

### Routing (go_router)

```
/                          HomeScaffold (map by default)
/?view=calendar            HomeScaffold (calendar tab)
/venue/:id                 Home + VenueDetailSheet
/venue/:id/event/:eid      Home + VenueDetailSheet + EventDetailSheet (stacked)
/event/:id                 Home + EventDetailSheet
/exhibition/:id            Home + ExhibitionDetailSheet
/settings                  SettingsScreen
/settings/offline-maps     OfflineDownloadScreen
```

Routes parse directly into `modalStackProvider`. Deep links via custom URL scheme (`klp://`) and universal links (`https://klp.example.com/...`) work for sharing. OS back-button pops modal stack first, then nav stack.

## Map + offline tile cache

### Runtime map

```dart
FlutterMap(
  options: MapOptions(
    initialCenter: LatLng(53.0, 11.25),
    initialZoom: 10,
    minZoom: 8, maxZoom: 18,
    interactionOptions: InteractionOptions(
      flags: InteractiveFlag.all & ~InteractiveFlag.rotate,
    ),
  ),
  children: [
    TileLayer(
      urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
      tileProvider: FMTCStore('klp_tiles').getTileProvider(),
      userAgentPackageName: 'de.musche.klpguide', // placeholder; final ID set in phase 8
      maxNativeZoom: 19,
    ),
    MarkerClusterLayerWidget(
      options: MarkerClusterLayerOptions(
        markers: ref.watch(visibleMarkersProvider),
        maxClusterRadius: 60,
        size: Size(40, 40),
        builder: (ctx, markers) => ClusterBubble(count: markers.length),
      ),
    ),
    AttributionWidget.defaultWidget(source: '© OpenStreetMap-Mitwirkende'),
  ],
);
```

### OSM tile policy compliance

- Distinct `User-Agent` set via `userAgentPackageName`
- FMTC respects upstream `Cache-Control`; tiles expire after 30 days and are refetched on stale read
- Visible attribution overlay
- Bulk pre-warm rate-limited (see below)

### Opt-in regional pre-warm

```dart
const wendlandBounds = LatLngBounds(LatLng(52.5, 10.5), LatLng(53.5, 12.0));
// Matches crawler coordinate validation range
```

Zoom range 10–14 (z15+ pushes into hundreds of MB).

| Zoom | Tiles | Approx. size |
|------|-------|--------------|
| 10   | 12    | ~150 KB |
| 11   | 54    | ~700 KB |
| 12   | 200   | ~3 MB |
| 13   | 780   | ~10 MB |
| 14   | 3,100 | ~40 MB |
| **Total** | **~4,150** | **~55 MB** |

Bulk download via FMTC `bulkDownload` with `parallelThreads: 5`, `rateLimit: 8 req/sec` — ~9 minutes total. Polite within OSM ToS for occasional bulk seeding. Cancel mid-flight is safe — partial downloads stay cached.

Storage UI in settings: "Karten-Cache: 47 MB" with a "Leeren" action. After first pre-warm, button label changes to "Aktualisieren".

If usage ever scales beyond OSM's tolerance, the migration path is: keep runtime fetches on OSM, switch only the pre-warm endpoint to a paid tile host (Mapbox / Stadia / MapTiler). Not a v1 concern.

## Theming + dark mode

ColorScheme lifted verbatim from `web/static/css/tailwind-input.css` CSS variables (canvas → surface, accent → primary, etc.). Generated `lib/core/theme_tokens.dart` keeps the values in sync with the web app.

```dart
MaterialApp.router(
  themeMode: ThemeMode.system,
  theme: ThemeData(useMaterial3: true, colorScheme: lightScheme, fontFamily: 'Inter', ...),
  darkTheme: ThemeData(useMaterial3: true, colorScheme: darkScheme, ...),
);
```

System-driven brightness only (matches the web app's `media`-strategy dark mode). Inter bundled as a Flutter asset (no Google Fonts runtime fetch).

## Testing strategy

| Layer | Type | Coverage focus |
|-------|------|----------------|
| Models | Unit | JSON round-trip per fixture (lifted from `data/*.json`) |
| Repositories | Unit (mocked Dio + cache) | 200 / 304 / network-error+cache / network-error+no-cache / forceRefresh paths |
| Derived providers | Unit | Filter logic, venue-centric search semantics, match badges |
| Favorites | Unit | SharedPreferences round-trip |
| Filter debounce | Unit (`fakeAsync`) | 250 ms debounce |
| Screens | Widget | Loading / error+empty / error+cache / populated states |
| Theming | Golden (Linux CI only) | EventCard, VenueCard, ClusterBubble, FilterChip — light + dark |
| Full stack | Integration | Cold start online; cold start offline (no cache); warm start with 304s; favorites persistence; modal back-stack |
| Tile pre-warm | Unit | Mocked FMTC.bulkDownload — progress states, completion, cancel |

**CI gates** (`mobile-ci.yml`): `dart format --set-exit-if-changed`, `dart analyze`, `flutter test`, `flutter test --tags golden` (Linux only), `flutter test integration_test/` against a job-spun-up Go server.

**Coverage target:** repositories + providers >80%. UI layers lower — golden + widget tests are enough signal.

**Excluded from v1 testing:** real-device matrix tests (manual QA on one iPhone + one Android pre-release), real OSM tile-server end-to-end, performance benchmarks.

## Distribution (independent / community app)

### Naming + brand independence

The app cannot present itself as "Kulturelle Landpartie". Maintainer is in contact with the festival association — final name selection coordinates with that conversation.

- **Display name:** TBD with association, but must be clearly distinct (e.g. "KLP-Guide", "Landpartie-Karte Wendland", "Wendland Kunsttour"). Final name is chosen before store submission, not during the implementation phases.
- **Bundle/Package ID:** namespaced under the maintainer's own reverse-DNS (e.g. `de.musche.klpguide`) — avoid `de.kulturellelandpartie.*`
- **Launcher icon + splash:** original artwork only — no festival logo, typeface, or campaign imagery
- **Store descriptions disclose independence** (DE + EN):
  - DE: "Inoffizieller, von Fans erstellter Guide zur Kulturellen Landpartie im Wendland."
  - EN: "Unofficial, fan-made guide to the Kulturelle Landpartie festival in Germany's Wendland region."
  - Same disclosure in-app on the About screen
- **DPMA trademark register check** before final name selection

### Data source attribution

Every venue/event detail screen credits the source: "Daten von kulturelle-landpartie.de" with a tap-through. About screen documents the 2 s/request rate-limited crawler. Freshness indicator notes "kann von der offiziellen Webseite abweichen".

### Accounts + signing

- Apple Developer Program ($99/yr) and Google Play Console ($25 one-time) in the maintainer's name
- Android: upload keystore generated once, stored in a password manager + as a base64 CI secret
- iOS: distribution cert + provisioning profile via `fastlane match`; bundle ID matches the chosen one

### CI/CD

- `mobile-ci.yml` — runs on every PR touching `mobile/**`: format, analyze, unit + widget + integration tests
- `mobile-release.yml` — triggered by `mobile-v*` tags: parallel jobs for Android (.aab → Play Internal Testing) and iOS (.ipa → TestFlight) via fastlane

### Store metadata + assets

- Bundle ID: `de.musche.klpguide` as the working placeholder; final form follows the chosen name and is locked in phase 8
- Display name: per the naming discussion with the association
- Launcher icons + splashes generated from a source SVG via `flutter_launcher_icons` + `flutter_native_splash`
- Screenshots: 6.7"/5.5" iPhone, 12.9" iPad, Android phone + tablet — three each
- Privacy policy hosted on a maintainer-controlled domain (NOT `kulturelle-landpartie.de`); covers: no personal data, no analytics, optional location, locally cached data, network requests to maintainer's Go server + OSM tile servers
- iOS App Privacy + Play Data Safety: optional location only, no identifiers, no third-party sharing
- Age rating: 4+ (iOS) / Everyone (Play); Category: Travel / Lifestyle

### Permissions

- Location (when in use) — optional, only for "Mein Standort"
- No background, push, camera, contacts, or calendar permissions in v1

### Versioning + release cadence

- Semantic versioning `1.0.0+1`, build number auto-bumped by CI on tag
- Target: ship v1.0.0 ~4 weeks before the May 2027 festival
- Feature freeze 1 week before the festival; patches only after that

## Rollout phasing

Vertical slices, each ending in a runnable app:

| Phase | Slice | Approx. |
|-------|-------|---------|
| 1 | Walking skeleton — Flutter boots, fetches `/api/venues`, renders plain list | ~3 days |
| 2 | Map MVP — FlutterMap + clustering + placeholder venue sheet | ~5 days |
| 3 | Three resources + detail sheets — events/exhibitions, calendar, real sheets, modal stack, go_router | ~1 week |
| 4 | Filters — state + sheet + derived providers + debounced search | ~1 week |
| 5 | Favorites + theming — SharedPreferences favorites, Material 3 theme, dark mode | ~4 days |
| 6 | Offline robustness — server Last-Modified + 304, cache fallback, offline states | ~4 days |
| 7 | Tile pre-warm — OfflineDownloadScreen + FMTC bulk download | ~3 days |
| 8 | Distribution prep — icons, splash, store metadata, CI workflows, signing | ~1 week |
| 9 | Closed beta — TestFlight + Play Internal Testing, gather feedback | ~2 weeks calendar |
| 10 | Public launch — submit, await review, promote via association | target: ~4 weeks pre-festival |

Phases 1–7 commit directly to `master` under `mobile/` — additive, doesn't break web or Go server. Phases 8+ need a real iOS device + Mac for the first signing run.

## Deferred to v1.1 / v2

- Route planning (currently uses Leaflet Routing Machine + OSRM on web)
- Bike-route polyline overlays (API doesn't return geometry today)
- Push notifications ("Event in 30 min am favorisierten Ort")
- Background data sync (`workmanager`)
- Vector tiles + alternate tile providers
- Localization beyond German
- iPad-specific layouts
- Analytics / crash reporting (would require explicit consent UX given EU/DE privacy expectations)

## Open questions

None blocking implementation planning. The final display name and bundle ID are coordinated with the festival association before store submission (phase 8+), not during build-out — phases 1–7 use the placeholder bundle ID `de.musche.klpguide`.

## Acceptance criteria for v1.0.0

- App boots on iOS 15+ and Android 8+ and shows the map with markers within 3 s on warm start, 10 s on cold start with network
- All v1 filters (search, date, category, event-category, bike-route, time-of-day) reproduce the web app's results 1:1 on the same dataset
- Favoriting a venue persists across app restarts and across data refreshes
- Modal back-stack: opening venue → event-in-venue → back returns to venue
- App opened with no network and no cache shows a clear "connect once" state, not a generic error
- App opened with no network and a cache shows last-cached data with a "vor X" freshness indicator
- Tile pre-warm completes within ~10 min on a normal connection and downloads ~4,150 tiles (±5%)
- 304 responses on `/api/venues`, `/api/events`, `/api/exhibitions` correctly skip re-parsing and update only the freshness stamp
- Store listings carry the unofficial-app disclaimer in DE + EN; in-app About screen does the same
- CI is green on `master`; `mobile-release.yml` produces signed builds for both stores
