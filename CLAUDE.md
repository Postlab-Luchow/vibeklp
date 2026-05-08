# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Kulturelle Landpartie (KLP) - a web app for a German cultural festival. Crawls venue/event/exhibition data from kulturelle-landpartie.de and serves it through an interactive map + calendar UI.

**Go module:** `github.com/musche/klp`

## Commands

```bash
# Build
go build -o bin/ ./cmd/...

# Run server (port 8081)
go run cmd/server/main.go

# Run crawler
go run cmd/crawler/main.go                    # full crawl with geocoding
go run cmd/crawler/main.go -skip-geocoding    # faster, no coordinate lookup
go run cmd/crawler/main.go -verbose           # detailed logging

# Test
./test.sh                          # all tests
./test.sh -v                       # verbose
./test.sh -c                       # with coverage report
./test.sh -p ./internal/api        # specific package
go test ./internal/storage/... -run TestVenue_Validate   # single test function
go test -race ./...                # race detection

# Lint
golangci-lint run
go mod tidy
```

## Architecture

**Data flow:** Crawler scrapes HTML -> parses + geocodes -> saves JSON to `data/` -> Server loads JSON + serves REST API -> Vanilla JS frontend consumes API

**Backend (Go):**
- `cmd/server/main.go` - HTTP server entry point
- `cmd/crawler/main.go` - Crawler entry point with many flags (LLM parsing, caching, geocoding)
- `internal/api/` - REST API: handlers, routes, middleware (CORS + logging), input validation
- `internal/storage/` - Data models (Venue, Event, Exhibition) + JSON file I/O. No database.
- `internal/crawler/` - Web scraper (goquery), geocoder (Nominatim/Google), optional LLM parser (OpenRouter)

**Frontend (Vanilla JS + Leaflet.js):**
- `web/static/js/app.js` - Main orchestrator, data loading, view switching, filtering
- `web/static/js/map.js` - Leaflet map with MarkerCluster
- `web/static/js/calendar.js` - Events grouped by date
- `web/static/js/favorites.js` - LocalStorage-based favorites
- `web/static/js/filters.js` - Search/date/category/bike-route filters
- `web/static/js/routing.js` - Route planning via Leaflet Routing Machine
- `web/templates/index.html` - Single-page app template

JS modules export functions to `window` (not ES6 modules) - don't refactor to ES6 imports without updating all cross-module references.

## Key Patterns and Gotchas

- **API responses** are always wrapped: `{"venues": [...], "total": N}` - never bare arrays
- **Search is venue-centric**: results show venues with match badges for events/exhibitions at that venue, not individual items
- **Null-safety in JS**: Many events lack descriptions. Always use `(field && field.toLowerCase())` pattern
- **Crawler rate limiting**: 2-second ticker between requests. Never remove this.
- **Geocoding rate limit**: Nominatim = 1 req/sec. Use `-skip-geocoding` for development
- **Hardcoded year**: Crawler hardcodes `2026` when parsing dates from DD.MM. format (`scraper.go`)
- **Coordinate validation**: Venues must be within Wendland region (lat 52.5-53.5, lng 10.5-12.0)
- **Venue IDs**: Slugified from venue name (lowercased, spaces to dashes)
- **Static assets**: Served at `/static/` prefix, mapped to `web/static/` on disk. Caching is disabled so JS/CSS edits land on next reload.
- **Calendar is filtered client-side**: `loadCalendar()` reads `App.data.events` and applies the active filter state (date/category/search/bike-route). `applyFilters()` re-renders the calendar when it's the active view. Don't reintroduce a server-side `/api/calendar` fetch.
- **Modal back-stack**: `App.state.modalStack` holds `{type, id}` entries. Entry points (`showVenueDetails`/`showEventDetails`/`showExhibitionDetails`) reset the stack. Use `pushModal(type, id)` for in-modal navigation (e.g. clicking an event in a venue modal), `closeModal()` to pop+restore, `hideModal()` to fully dismiss (used by `centerMapOnVenue`).
- **Mobile bottom sheet**: On `<=768px`, the sidebar is a fixed-position off-screen panel toggled via `body.sidebar-open`. Wired up in `initMobileSidebar()`. Auto-closes on tab switch and on any modal entry-point.
- **Routing minimize vs clear**: Closing the routing dialog (X or toggling the route button) calls `minimizeRoutingMode()` — this hides the dialog but keeps the route polyline + LRM box on the map. Only `clearRoute()` (the "Route löschen" button) tears down `routingControl` and `selectedWaypoints`.

## Task Workflow

Tasks are tracked in `TASKS.md`. When working on a task:
1. Work on ONE task at a time
2. Implement, test, commit with format: `Fix #<number>: <description>`
3. Prompt user to test specific behaviors
4. Do NOT start the next task automatically

## Deployment

- `deploy/deploy.sh` - cross-compiles `cmd/server` for linux/amd64, rsyncs `server` + `web/` + `data/` + the systemd unit to a staging dir on the remote, and runs a sudo-elevated install into `/opt/vibeklp` owned by a system user `vibeklp`.
- `deploy/vibeklp.service` - systemd unit (User=vibeklp, WorkingDirectory=/opt/vibeklp, Restart=on-failure).
- `deploy/deploy.env` (gitignored) - per-host settings sourced by `deploy.sh`. At minimum: `REMOTE=user@host`. Optional: `INSTALL_DIR`, `SERVICE_NAME`, `STAGING_DIR`. `REMOTE` is required (script aborts otherwise).
- The deploy uses `rsync --delete`, so it overwrites `/opt/vibeklp/data` with the local `data/` snapshot — re-crawl locally before deploying if the server has fresher data.

## Documentation

- `plans/api-specification.md` - Full REST API reference
- `internal/crawler/AGENTS.md` - Crawler-specific troubleshooting and known issues
- `TESTING.md` - Testing guide with coverage details
