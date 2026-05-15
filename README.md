# Kulturelle Landpartie Web-App

Eine moderne Web-Anwendung für das Kulturfestival "Kulturelle Landpartie" mit interaktiver Karte, Kalenderansicht und umfangreichen Filterfunktionen.

## Features

- 🗺️ **Interaktive Karte** mit allen 93 Veranstaltungsorten (Cluster-Marker auf OpenStreetMap)
- 📅 **Kalenderansicht** mit pro Tag gruppierten und nach Uhrzeit sortierten Veranstaltungen
- 🔍 **Intelligente Suche** über Orte, Events und Ausstellungen
  - Venue-zentrierte Ergebnisse mit Match-Badges
  - Filter für Datum, Kategorie, Event-Kategorie und Fahrradroute
- ⭐ **Favoritenliste** mit LocalStorage-Persistierung
- 🚴 **Routenplanung** zwischen Veranstaltungsorten
- 🌓 **Dark Mode** automatisch über `prefers-color-scheme`
- 📱 **Responsive Design** mit Bottom-Sheet-Filter auf Mobilgeräten

## Technologie-Stack

- **Backend:** Go 1.25+ (Gorilla Mux Router)
- **Frontend:** Vanilla JavaScript, Leaflet.js, Leaflet Routing Machine
- **Styling:** Tailwind CSS v3 (Standalone CLI, kein Node-Toolchain nötig)
- **Karten:** OpenStreetMap
- **Crawler:** goquery + optional LLM-Parser über OpenRouter
- **Geocoding:** Nominatim (mit optionalem Google Maps Fallback)
- **Datenspeicherung:** JSON-Dateien

## Projektstruktur

```
klp/
├── cmd/
│   ├── crawler/          # Web-Crawler-Entry-Point
│   └── server/           # Web-Server-Entry-Point
├── internal/
│   ├── crawler/          # Scraper, Geocoder, optionaler LLM-Parser
│   ├── api/              # REST-API-Handler, Routen, Middleware
│   └── storage/          # Datenmodelle und JSON-I/O
├── web/
│   ├── static/           # JS, generierte CSS, Bilder
│   └── templates/        # HTML-Templates
├── data/                 # JSON-Daten (venues, events, exhibitions)
├── deploy/               # Deploy-Skript + systemd-Unit
└── plans/                # Architektur- und API-Dokumentation
```

## Installation

### Voraussetzungen

- Go 1.25 oder höher
- Git
- (optional) `curl` — wird vom Makefile genutzt, um die Tailwind-Standalone-CLI bei Bedarf nachzuladen

### Setup

1. Repository klonen:
```bash
git clone git@github.com:Postlab-Luchow/vibeklp.git
cd vibeklp
```

2. Dependencies installieren:
```bash
go mod download
```

## Verwendung

### 1. Daten crawlen

Crawlt alle Daten von den unterstützten Quellen:

```bash
go run cmd/crawler/main.go
```

Häufig genutzte Optionen:

- `-source` — Quelle wählen: `klp` (Standard), `wendlandpartie`, `landgang`, `all`
- `-output ./data` — Ausgabeverzeichnis für JSON-Dateien
- `-verbose` — Ausführliches Logging
- `-skip-geocoding` — Geocoding überspringen (deutlich schneller)
- `-use-llm` — HTML mit einem LLM parsen (benötigt `OPENROUTER_API_KEY`)
- `-dry-run` — nur anzeigen, was extrahiert würde

Weitere Flags (Cache, LLM-Modell, Kategorisierung etc.) siehe `go run cmd/crawler/main.go -h`.

### 2. Tailwind-CSS generieren

Die Stylesheet-Datei `web/static/css/tailwind.css` wird aus `tailwind-input.css` generiert und ist eingecheckt. Bei Änderungen an HTML/JS-Klassen muss sie neu gebaut werden:

```bash
make css         # einmaliger Build
make css-watch   # automatisch bei jeder Änderung neu bauen
```

Der erste Aufruf lädt die `tailwindcss`-Binary nach `build/tools/` (gitignored), Node ist nicht erforderlich.

### 3. Server starten

```bash
go run cmd/server/main.go
# oder mit vorherigem CSS-Rebuild:
make server
```

Optionen:

- `-port 8081` — Server-Port (Standard: 8081)
- `-data ./data` — Datenverzeichnis
- `-static ./web/static` — Verzeichnis für statische Assets
- `-templates ./web/templates` — Verzeichnis für HTML-Templates

Die Web-App ist dann unter http://localhost:8081 erreichbar.

## API-Endpunkte

### Venues
- `GET /api/venues` — alle Veranstaltungsorte (Query: `search`, `amenity`)
- `GET /api/venues/:id` — einzelner Ort inkl. zugeordneter Events und Ausstellungen

### Events
- `GET /api/events` — alle Veranstaltungen (Query: `search`, `date`, `dateFrom`, `dateTo`, `category`, `venueId`)
- `GET /api/events/:id` — einzelne Veranstaltung mit Venue-Details

### Exhibitions
- `GET /api/exhibitions` — alle Ausstellungen (Query: `search`, `venueId`)
- `GET /api/exhibitions/:id` — einzelne Ausstellung

### Weitere
- `GET /api/search?q=...` — globale Suche
- `GET /api/calendar` — Veranstaltungen nach Datum gruppiert
- `GET /api/categories` — Venue-Kategorien mit Zähler
- `GET /api/event-categories` — Event-Kategorien mit Zähler
- `GET /api/stats` — Statistiken (Venue-/Event-Zähler, Kategorien, Fahrradrouten)

Alle Antworten sind in einem Wrapper-Objekt verpackt (`{"venues": [...], "total": N}`), nie als nackte Arrays.

Vollständige API-Dokumentation: [`plans/api-specification.md`](plans/api-specification.md).

## Entwicklung

### Build

```bash
go build -o bin/ ./cmd/...
```

### Tests

```bash
./test.sh                          # alle Tests
./test.sh -v                       # verbose
./test.sh -c                       # mit Coverage-Report
./test.sh -p ./internal/api        # einzelnes Package

go test ./internal/storage/... -run TestVenue_Validate   # einzelner Test
go test -race ./...                                       # Race-Detection
```

Details zum Testaufbau: [`TESTING.md`](TESTING.md).

### Linting

```bash
golangci-lint run
go mod tidy
```

## Deployment

Das Repo enthält ein Deploy-Skript für einen Linux-Host mit systemd:

- [`deploy/deploy.sh`](deploy/deploy.sh) — baut Tailwind neu, cross-kompiliert `cmd/server` für `linux/amd64`, rsynct `server` + `web/` + `data/` + die systemd-Unit auf den Zielhost und installiert nach `/opt/vibeklp` (User `vibeklp`).
- [`deploy/vibeklp.service`](deploy/vibeklp.service) — systemd-Unit, `Restart=on-failure`.
- `deploy/deploy.env` (gitignored) — Hostspezifische Settings. Mindestens `REMOTE=user@host` ist erforderlich; optional `INSTALL_DIR`, `SERVICE_NAME`, `STAGING_DIR`.

```bash
# einmalig:
echo 'REMOTE=user@beispiel.host' > deploy/deploy.env

# Deploy ausführen:
./deploy/deploy.sh
```

> ⚠️ Das Skript verwendet `rsync --delete`, überschreibt also `/opt/vibeklp/data` auf dem Server mit dem lokalen Stand. Vor dem Deploy ggf. lokal neu crawlen.

## Dokumentation

- [Architekturplan](plans/kulturelle-landpartie-webapp.md)
- [API-Spezifikation](plans/api-specification.md)
- [Crawler-Strategie](plans/crawler-strategy.md)
- [Test-Guide](TESTING.md)
- [Aufgabenliste](TASKS.md)

## Lizenz

MIT License

## Kontakt

Für Fragen und Feedback: https://github.com/Postlab-Luchow/vibeklp/issues
