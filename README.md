# Kulturelle Landpartie Web-App

Eine moderne Web-Anwendung für das Kulturfestival "Kulturelle Landpartie" mit interaktiver Karte, Kalenderansicht und umfangreichen Filterfunktionen.

## Features

- 🗺️ **Interaktive Karte** mit allen 87 Veranstaltungsorten
- 📅 **Kalenderansicht** für alle Events und Ausstellungen
- 🔍 **Intelligente Suche** nach Orten, Events und Ausstellungen
  - Durchsucht Veranstaltungsorte, Events und Ausstellungen
  - Zeigt passende Treffer mit Markierungen an
  - Filterung nach Datum, Kategorie und Fahrradroute
- ⭐ **Favoritenliste** mit LocalStorage-Persistierung
- 🚴 **Routenplanung** zwischen Veranstaltungsorten
- 🔔 **Benutzerfreundliche Fehlerbehandlung** mit Inline-Nachrichten
- 📱 **Responsive Design** für alle Geräte

## Technologie-Stack

- **Backend:** Go 1.21+
- **Frontend:** Vanilla JavaScript, Leaflet.js
- **Karten:** OpenStreetMap
- **Datenspeicherung:** JSON-Dateien

## Projektstruktur

```
klp/
├── cmd/
│   ├── crawler/          # Web-Crawler für kulturelle-landpartie.de
│   └── server/           # Web-Server
├── internal/
│   ├── crawler/          # Crawler-Logik
│   ├── api/              # API-Handler
│   └── storage/          # Datenmodelle und JSON-I/O
├── web/
│   ├── static/           # CSS, JS, Bilder
│   └── templates/        # HTML-Templates
├── data/                 # JSON-Daten (venues, events, exhibitions)
└── plans/                # Architektur-Dokumentation
```

## Installation

### Voraussetzungen

- Go 1.21 oder höher
- Git

### Setup

1. Repository klonen:
```bash
git clone https://github.com/musche/klp.git
cd klp
```

2. Dependencies installieren:
```bash
go mod download
```

## Verwendung

### 1. Daten crawlen

Crawlt alle Daten von kulturelle-landpartie.de:

```bash
go run cmd/crawler/main.go
```

Optionen:
- `-output ./data` - Ausgabeverzeichnis für JSON-Dateien
- `-verbose` - Ausführliches Logging
- `-skip-geocoding` - Geocoding überspringen

### 2. Server starten

Startet den Web-Server:

```bash
go run cmd/server/main.go
```

Optionen:
- `-port 8080` - Server-Port (Standard: 8080)
- `-data ./data` - Datenverzeichnis

Die Web-App ist dann verfügbar unter: http://localhost:8080

## API-Endpunkte

### Venues
- `GET /api/venues` - Alle Veranstaltungsorte
- `GET /api/venues/:id` - Einzelner Veranstaltungsort

### Events
- `GET /api/events` - Alle Veranstaltungen
- `GET /api/events/:id` - Einzelne Veranstaltung
- Query-Parameter: `date`, `dateFrom`, `dateTo`, `category`, `venueId`

### Exhibitions
- `GET /api/exhibitions` - Alle Ausstellungen
- `GET /api/exhibitions/:id` - Einzelne Ausstellung

### Weitere
- `GET /api/search?q=...` - Globale Suche
- `GET /api/calendar` - Kalenderansicht
- `GET /api/categories` - Alle Kategorien
- `GET /api/stats` - Statistiken

Vollständige API-Dokumentation: [`plans/api-specification.md`](plans/api-specification.md)

## Entwicklung

### Build

```bash
# Crawler
go build -o bin/crawler cmd/crawler/main.go

# Server
go build -o bin/server cmd/server/main.go
```

### Tests

Run all tests:
```bash
./test.sh
```

Options:
```bash
./test.sh -v               # Verbose output
./test.sh -c               # With coverage report
./test.sh -p ./internal/api # Test specific package
```

Or use Go directly:
```bash
go test ./...              # All tests
go test ./... -v           # Verbose
go test ./... -cover       # With coverage
go test -race ./...        # Check for race conditions
```

**Current Test Coverage:**
- API package: 78.1%
- Storage package: 87.5%
- Overall: 82.8%

See [`AGENTS.md`](AGENTS.md) for detailed testing documentation.

### Linting

```bash
golangci-lint run
```

## Deployment

### Option 1: Systemd Service

1. Binary kompilieren:
```bash
go build -o /usr/local/bin/klp-server cmd/server/main.go
```

2. Systemd-Service erstellen (`/etc/systemd/system/klp.service`):
```ini
[Unit]
Description=Kulturelle Landpartie Web Server
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/www/klp
ExecStart=/usr/local/bin/klp-server -port 8080 -data /var/www/klp/data
Restart=always

[Install]
WantedBy=multi-user.target
```

3. Service aktivieren:
```bash
sudo systemctl enable klp
sudo systemctl start klp
```

### Option 2: Docker

```bash
docker build -t klp-webapp .
docker run -p 8080:8080 klp-webapp
```

### Option 3: Nginx Reverse Proxy

Nginx-Konfiguration:
```nginx
server {
    listen 80;
    server_name kulturelle-landpartie.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Dokumentation

- [Architekturplan](plans/kulturelle-landpartie-webapp.md)
- [API-Spezifikation](plans/api-specification.md)
- [Crawler-Strategie](plans/crawler-strategy.md)

## Lizenz

MIT License

## Kontakt

Für Fragen und Feedback: https://github.com/musche/klp/issues
