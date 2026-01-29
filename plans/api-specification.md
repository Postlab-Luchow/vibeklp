# API-Spezifikation - Kulturelle Landpartie Web-App

## Base URL
```
http://localhost:8080/api
```

## Endpunkte

### Venues (Veranstaltungsorte)

#### GET /api/venues
Gibt alle Veranstaltungsorte zurück.

**Query-Parameter:**
- `search` (optional): Freitextsuche in Name und Beschreibung
- `amenity` (optional): Filter nach Ausstattung (z.B. "Fahrradroute")

**Response:**
```json
{
  "venues": [
    {
      "id": "venue-001",
      "name": "Bankewitz",
      "description": "Im vermeintlich wilden Garten",
      "address": {
        "street": "Zum Seinitz Moor 1",
        "postalCode": "29597",
        "city": "Stoetze OT Bankewitz"
      },
      "coordinates": {
        "lat": 53.0123,
        "lng": 11.0456
      },
      "contact": {
        "phone": "05872 986107",
        "email": "keramik@wandafulworld.de",
        "website": "www.wandafulworld.de"
      },
      "amenities": ["Fahrradroute"],
      "bikeRoute": "Fahrradtour: 🚴",
      "eventCount": 9,
      "exhibitionCount": 3
    }
  ],
  "total": 87
}
```

#### GET /api/venues/:id
Gibt Details zu einem spezifischen Veranstaltungsort zurück.

**Response:**
```json
{
  "id": "venue-001",
  "name": "Bankewitz",
  "description": "Im vermeintlich wilden Garten",
  "address": {
    "street": "Zum Seinitz Moor 1",
    "postalCode": "29597",
    "city": "Stoetze OT Bankewitz"
  },
  "coordinates": {
    "lat": 53.0123,
    "lng": 11.0456
  },
  "contact": {
    "phone": "05872 986107",
    "email": "keramik@wandafulworld.de",
    "website": "www.wandafulworld.de"
  },
  "amenities": ["Fahrradroute"],
  "bikeRoute": "Fahrradtour: 🚴",
  "events": [
    {
      "id": "event-001",
      "title": "Kaffeeröst Sängers!",
      "date": "2025-05-30",
      "startTime": "05:00",
      "category": "Musik"
    }
  ],
  "exhibitions": [
    {
      "id": "exhibition-001",
      "title": "Lebendigkeit und Heiterkeiten mit MOKKA im vermeintlich wilden GARTEN",
      "artist": "Gartennudist:innen",
      "category": "Kunst"
    }
  ]
}
```

### Events (Veranstaltungen)

#### GET /api/events
Gibt alle Veranstaltungen zurück.

**Query-Parameter:**
- `date` (optional): Filter nach Datum (ISO 8601: YYYY-MM-DD)
- `dateFrom` (optional): Von-Datum für Bereichsfilter
- `dateTo` (optional): Bis-Datum für Bereichsfilter
- `category` (optional): Filter nach Kategorie
- `venueId` (optional): Filter nach Veranstaltungsort
- `search` (optional): Freitextsuche

**Response:**
```json
{
  "events": [
    {
      "id": "event-001",
      "title": "Kaffeeröst Sängers!",
      "description": "Hingestreut im Maiengras, der morgenfrische Tau...",
      "venueId": "venue-001",
      "venueName": "Bankewitz",
      "date": "2025-05-30",
      "startTime": "05:00",
      "endTime": "08:30",
      "category": "Musik",
      "admission": "frei",
      "imageUrl": "/images/events/event-001.jpg"
    }
  ],
  "total": 250
}
```

#### GET /api/events/:id
Gibt Details zu einer spezifischen Veranstaltung zurück.

**Response:**
```json
{
  "id": "event-001",
  "title": "Kaffeeröst Sängers!",
  "description": "Hingestreut im Maiengras, der morgenfrische Tau, die Schönheit lange ich vergaß...",
  "venueId": "venue-001",
  "venue": {
    "id": "venue-001",
    "name": "Bankewitz",
    "address": {
      "street": "Zum Seinitz Moor 1",
      "postalCode": "29597",
      "city": "Stoetze OT Bankewitz"
    },
    "coordinates": {
      "lat": 53.0123,
      "lng": 11.0456
    }
  },
  "date": "2025-05-30",
  "startTime": "05:00",
  "endTime": "08:30",
  "category": "Musik",
  "admission": "frei",
  "imageUrl": "/images/events/event-001.jpg"
}
```

### Exhibitions (Ausstellungen)

#### GET /api/exhibitions
Gibt alle Ausstellungen zurück.

**Query-Parameter:**
- `category` (optional): Filter nach Kategorie
- `venueId` (optional): Filter nach Veranstaltungsort
- `search` (optional): Freitextsuche

**Response:**
```json
{
  "exhibitions": [
    {
      "id": "exhibition-001",
      "title": "Lebendigkeit und Heiterkeiten mit MOKKA",
      "description": "Gartennudist:innen gucken, dabei Mokka trinken...",
      "venueId": "venue-001",
      "venueName": "Bankewitz",
      "artist": "Wanda Sippl",
      "category": "Kunst",
      "imageUrl": "/images/exhibitions/exhibition-001.jpg"
    }
  ],
  "total": 150
}
```

#### GET /api/exhibitions/:id
Gibt Details zu einer spezifischen Ausstellung zurück.

**Response:**
```json
{
  "id": "exhibition-001",
  "title": "Lebendigkeit und Heiterkeiten mit MOKKA im vermeintlich wilden GARTEN",
  "description": "Gartennudist:innen gucken, dabei Mokka trinken, aus wilden& braven Mokkalingen www.wandafulworld.de",
  "venueId": "venue-001",
  "venue": {
    "id": "venue-001",
    "name": "Bankewitz",
    "address": {
      "street": "Zum Seinitz Moor 1",
      "postalCode": "29597",
      "city": "Stoetze OT Bankewitz"
    },
    "coordinates": {
      "lat": 53.0123,
      "lng": 11.0456
    }
  },
  "artist": "Wanda Sippl",
  "category": "Kunst",
  "imageUrl": "/images/exhibitions/exhibition-001.jpg"
}
```

### Search (Suche)

#### GET /api/search
Globale Suche über alle Datentypen.

**Query-Parameter:**
- `q` (required): Suchbegriff
- `type` (optional): Filter nach Typ (venues, events, exhibitions)

**Response:**
```json
{
  "results": {
    "venues": [
      {
        "id": "venue-001",
        "name": "Bankewitz",
        "type": "venue"
      }
    ],
    "events": [
      {
        "id": "event-001",
        "title": "Kaffeeröst Sängers!",
        "date": "2025-05-30",
        "type": "event"
      }
    ],
    "exhibitions": [
      {
        "id": "exhibition-001",
        "title": "Lebendigkeit und Heiterkeiten",
        "type": "exhibition"
      }
    ]
  },
  "total": 15
}
```

### Calendar (Kalender)

#### GET /api/calendar
Gibt Events gruppiert nach Datum zurück.

**Query-Parameter:**
- `month` (optional): Monat (1-12)
- `year` (optional): Jahr (YYYY)

**Response:**
```json
{
  "calendar": {
    "2025-05-29": {
      "date": "2025-05-29",
      "dayOfWeek": "Donnerstag",
      "eventCount": 5,
      "events": [
        {
          "id": "event-001",
          "title": "Eröffnung",
          "startTime": "11:00",
          "venueName": "Hitzacker"
        }
      ]
    },
    "2025-05-30": {
      "date": "2025-05-30",
      "dayOfWeek": "Freitag",
      "eventCount": 12,
      "events": []
    }
  },
  "totalDays": 12,
  "totalEvents": 250
}
```

### Categories (Kategorien)

#### GET /api/categories
Gibt alle verfügbaren Kategorien zurück.

**Response:**
```json
{
  "categories": [
    {
      "name": "Musik",
      "count": 45,
      "color": "#FF6B6B"
    },
    {
      "name": "Kunst",
      "count": 78,
      "color": "#4ECDC4"
    },
    {
      "name": "Theater",
      "count": 23,
      "color": "#45B7D1"
    },
    {
      "name": "Lesung",
      "count": 15,
      "color": "#FFA07A"
    },
    {
      "name": "Workshop",
      "count": 34,
      "color": "#98D8C8"
    }
  ]
}
```

### Statistics (Statistiken)

#### GET /api/stats
Gibt Statistiken über die Daten zurück.

**Response:**
```json
{
  "stats": {
    "totalVenues": 87,
    "totalEvents": 250,
    "totalExhibitions": 150,
    "festivalDates": {
      "start": "2025-05-29",
      "end": "2025-06-09"
    },
    "categoriesDistribution": {
      "Musik": 45,
      "Kunst": 78,
      "Theater": 23
    },
    "venuesWithBikeRoutes": 42
  }
}
```

## Fehlerbehandlung

Alle Endpunkte geben bei Fehlern folgendes Format zurück:

**4xx Client-Fehler:**
```json
{
  "error": {
    "code": 400,
    "message": "Invalid date format. Use YYYY-MM-DD",
    "details": "date parameter must be in ISO 8601 format"
  }
}
```

**5xx Server-Fehler:**
```json
{
  "error": {
    "code": 500,
    "message": "Internal server error",
    "details": "Failed to read data file"
  }
}
```

## HTTP-Statuscodes

- `200 OK` - Erfolgreiche Anfrage
- `400 Bad Request` - Ungültige Parameter
- `404 Not Found` - Ressource nicht gefunden
- `500 Internal Server Error` - Serverfehler

## CORS

Die API unterstützt CORS für alle Origins während der Entwicklung.
In Produktion sollte dies auf die Frontend-Domain beschränkt werden.

## Rate-Limiting

Aktuell kein Rate-Limiting implementiert.
Für Produktion empfohlen: 100 Requests pro Minute pro IP.

## Caching

- Statische Daten (venues, exhibitions): Cache-Control: max-age=3600
- Dynamische Daten (events mit Datumsfilter): Cache-Control: max-age=300

## Beispiel-Requests

### Alle Events am 30. Mai 2025
```bash
curl http://localhost:8080/api/events?date=2025-05-30
```

### Suche nach "Musik"
```bash
curl http://localhost:8080/api/search?q=Musik
```

### Alle Venues mit Fahrradroute
```bash
curl http://localhost:8080/api/venues?amenity=Fahrradroute
```

### Events in einem Datumsbereich
```bash
curl "http://localhost:8080/api/events?dateFrom=2025-05-29&dateTo=2025-06-02"
```

### Venue-Details mit Events und Exhibitions
```bash
curl http://localhost:8080/api/venues/venue-001
```
