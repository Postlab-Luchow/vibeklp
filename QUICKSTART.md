# Schnellstart-Anleitung

## Voraussetzungen

- Go 1.21 oder höher
- Git
- Webbrowser

## Installation und Start

### 1. Dependencies installieren

```bash
go mod download
```

### 2. Daten crawlen (optional)

**Hinweis:** Der Crawler benötigt eine aktive Internetverbindung und crawlt die echte Website. Dies kann einige Minuten dauern.

```bash
go run cmd/crawler/main.go
```

Optionen:
- `-output ./data` - Ausgabeverzeichnis (Standard: ./data)
- `-verbose` - Ausführliches Logging
- `-skip-geocoding` - Geocoding überspringen (schneller, aber ohne Koordinaten)

**Beispiel mit allen Optionen:**
```bash
go run cmd/crawler/main.go -output ./data -verbose
```

### 3. Server starten

```bash
go run cmd/server/main.go
```

Der Server startet auf Port 8080. Öffnen Sie Ihren Browser und navigieren Sie zu:

```
http://localhost:8080
```

## Verwendung

### Karte
- **Marker anklicken**: Details zum Veranstaltungsort anzeigen
- **Zoom**: Mausrad oder +/- Buttons
- **Standort**: Klicken Sie auf das Standort-Icon (oben rechts)

### Filter
- **Suche**: Freitextsuche in der Sidebar
- **Datum**: Filtern Sie Events nach Datum
- **Kategorie**: Filtern Sie nach Event-Kategorie
- **Fahrradroute**: Nur Orte mit Fahrradroute anzeigen

### Kalender
- Wechseln Sie zur Kalenderansicht über die Navigation
- Klicken Sie auf Events für Details

### Favoriten
- Klicken Sie auf das Herz-Icon bei Orten oder Events
- Favoriten werden im Browser gespeichert (LocalStorage)
- Wechseln Sie zur Favoriten-Ansicht über die Navigation

### Routenplanung
1. Klicken Sie auf das Routen-Icon (oben rechts auf der Karte)
2. Klicken Sie nacheinander auf Orte auf der Karte
3. Die Route wird automatisch berechnet
4. Klicken Sie auf "Route löschen" zum Zurücksetzen

## API-Endpunkte

Die REST-API ist verfügbar unter `http://localhost:8080/api`:

- `GET /api/venues` - Alle Veranstaltungsorte
- `GET /api/venues/:id` - Einzelner Veranstaltungsort
- `GET /api/events` - Alle Veranstaltungen
- `GET /api/events/:id` - Einzelne Veranstaltung
- `GET /api/exhibitions` - Alle Ausstellungen
- `GET /api/calendar` - Kalenderansicht
- `GET /api/search?q=...` - Suche
- `GET /api/categories` - Kategorien
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

```bash
go test ./...
```

### Code-Struktur

```
klp/
├── cmd/                    # Hauptprogramme
│   ├── crawler/           # Web-Crawler
│   └── server/            # Web-Server
├── internal/              # Interne Packages
│   ├── api/              # API-Handler und Routen
│   ├── crawler/          # Crawler-Logik
│   └── storage/          # Datenmodelle und JSON-I/O
├── web/                   # Frontend
│   ├── static/           # CSS, JS, Bilder
│   └── templates/        # HTML-Templates
├── data/                  # JSON-Daten
└── plans/                 # Dokumentation
```

## Troubleshooting

### Crawler-Fehler

**Problem:** `Failed to crawl venues`
- **Lösung:** Überprüfen Sie Ihre Internetverbindung
- **Lösung:** Die Website könnte temporär nicht erreichbar sein

**Problem:** `Failed to geocode`
- **Lösung:** Nominatim API hat Rate-Limits (1 Request/Sekunde)
- **Lösung:** Verwenden Sie `-skip-geocoding` für schnelleres Testen

### Server-Fehler

**Problem:** `Could not load venues`
- **Lösung:** Führen Sie zuerst den Crawler aus: `go run cmd/crawler/main.go`
- **Lösung:** Stellen Sie sicher, dass `data/venues.json` existiert

**Problem:** `Port already in use`
- **Lösung:** Ändern Sie den Port: `go run cmd/server/main.go -port 8081`

### Frontend-Fehler

**Problem:** Karte wird nicht angezeigt
- **Lösung:** Überprüfen Sie die Browser-Konsole (F12)
- **Lösung:** Stellen Sie sicher, dass Leaflet.js geladen wurde

**Problem:** Keine Daten sichtbar
- **Lösung:** Öffnen Sie die Browser-Konsole und prüfen Sie auf API-Fehler
- **Lösung:** Stellen Sie sicher, dass der Server läuft

## Nächste Schritte

1. **Daten aktualisieren**: Führen Sie den Crawler regelmäßig aus
2. **Deployment**: Siehe [`README.md`](README.md) für Deployment-Optionen
3. **Anpassungen**: Passen Sie Farben in [`web/static/css/styles.css`](web/static/css/styles.css) an

## Support

Bei Fragen oder Problemen:
- Lesen Sie die [Architektur-Dokumentation](plans/kulturelle-landpartie-webapp.md)
- Öffnen Sie ein Issue auf GitHub
- Überprüfen Sie die Browser-Konsole und Server-Logs

## Lizenz

MIT License - Siehe LICENSE-Datei für Details
