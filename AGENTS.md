# Agent Development Guide - Kulturelle Landpartie (KLP)

This document provides comprehensive guidance for AI agents working on the Kulturelle Landpartie web application.

## Project Overview

**Kulturelle Landpartie** is a web application for a German cultural festival featuring:
- Interactive map with ~93 venue locations
- Calendar view (events grouped by day, sorted by start time)
- Search and filtering capabilities
- Favorites management with local storage
- Route planning between venues
- Dark mode via `prefers-color-scheme`

**Tech Stack:**
- **Backend:** Go 1.25 (see `go.mod` for exact toolchain)
- **Frontend:** Vanilla JavaScript with Leaflet.js maps
- **Styling:** Tailwind CSS v3 via the standalone CLI (no Node required)
- **Data:** JSON files (no database)
- **Router:** Gorilla Mux
- **Scraping:** goquery for web crawling, optional LLM parsing via OpenRouter

**Module:** `github.com/musche/klp` (repo lives at `Postlab-Luchow/vibeklp`)

## Essential Commands

### Running the Application

```bash
# 1. Install dependencies
go mod download

# 2. Crawl data (defaults to kulturelle-landpartie.de; -source switches feeds)
go run cmd/crawler/main.go

# Crawler options:
go run cmd/crawler/main.go -verbose              # Detailed logging
go run cmd/crawler/main.go -skip-geocoding       # Skip geocoding (faster, no coordinates)
go run cmd/crawler/main.go -output ./data        # Custom output directory
go run cmd/crawler/main.go -source all           # klp | wendlandpartie | landgang | all
go run cmd/crawler/main.go -use-llm              # Parse HTML with an LLM (needs OPENROUTER_API_KEY)
go run cmd/crawler/main.go -h                    # Full list (cache, LLM model, categorize, ...)

# 3. Build Tailwind CSS once (or run `make css-watch` while editing JS/HTML)
make css

# 4. Start web server
go run cmd/server/main.go
# Or: `make server` (rebuilds tailwind.css first, then starts the server)

# Server options:
go run cmd/server/main.go -port 8080             # Custom port (default: 8081)
go run cmd/server/main.go -data ./data           # Custom data directory
go run cmd/server/main.go -static ./web/static   # Custom static directory
go run cmd/server/main.go -templates ./web/templates
```

### Building

```bash
# Build crawler
go build -o bin/crawler cmd/crawler/main.go

# Build server
go build -o bin/server cmd/server/main.go

# Build both
go build -o bin/ ./cmd/...
```

### Testing

```bash
# Run all tests
./test.sh
go test ./...

# Verbose mode
./test.sh -v
go test ./... -v

# With coverage
./test.sh -c
go test ./... -cover -coverprofile=coverage.out

# View HTML coverage report
go tool cover -html=coverage.out

# Test specific package
./test.sh -p ./internal/api
go test ./internal/api/...

# Test specific function
go test ./internal/storage/... -run TestVenue_Validate

# Race detection
go test -race ./...
```

Coverage targets and per-package breakdown live in [`TESTING.md`](TESTING.md); regenerate with `./test.sh -c` for current numbers.

### Linting

```bash
# Run linter (if golangci-lint installed)
golangci-lint run

# Fix go.mod issues
go mod tidy
```

### Data Management

```bash
# View data files
ls -lh data/
cat data/venues.json | jq '.[0]'      # First venue
cat data/events.json | jq length      # Count events

# Test API endpoints (server must be running)
curl http://localhost:8081/api/venues | jq
curl http://localhost:8081/api/events?date=2026-05-29 | jq
curl http://localhost:8081/api/search?q=kunst | jq
```


## Task Management Workflow

This project uses `TASKS.md` to track issues and improvements. When working on tasks from that file:

### Task Execution Rules

1. **ONE TASK AT A TIME**: Only work on a single task per session
   - Complete the task fully before moving to the next
   - Don't combine multiple tasks unless explicitly requested

2. **Standard Task Workflow**:
   ```
   a. Read and understand the task from TASKS.md
   b. Implement the solution (code changes, new files, etc.)
   c. Run all relevant tests to verify the fix
   d. Commit changes with descriptive message
   e. Prompt user to test the changes
   ```

3. **Testing Requirements**:
   - Run unit tests if they exist
   - Run integration tests for API changes
   - Manually verify UI changes work correctly
   - Test edge cases mentioned in the task description

4. **Commit Message Format**:
   ```
   Fix #<task-number>: <Brief description>
   
   - <Change 1>
   - <Change 2>
   - <Change 3>
   ```
   Example: `Fix #1: Add missing API response fields`

5. **User Testing Prompt**:
   After committing, always prompt the user to test:
   ```
   Changes committed. Please test:
   - <Specific thing to test 1>
   - <Specific thing to test 2>
   ```

### Task Selection

When user says "do task #X":
- Find task #X in TASKS.md
- Read the full description including files affected
- Check priority and dependencies
- Execute using the workflow above

### Task Prioritization

Refer to the "Task Prioritization Summary" section in TASKS.md:
- **High Priority**: Critical bugs, missing core functionality
- **Medium Priority**: Important features, UX improvements
- **Low Priority**: Nice-to-have features, code quality improvements

### After Task Completion

- DO NOT automatically start another task
- DO NOT ask "what's next?" - wait for user input
- User will either:
  - Report test results (fix if issues found)
  - Request the next task explicitly

### Example Session

```
User: do task #1
AI: [reads task, implements fix, runs tests, commits]
    Changes committed. Please test:
    - Check venue cards show eventCount/exhibitionCount
    - Verify event cards display venueName
    - Test exhibitions show venueName in UI

User: works, do task #6
AI: [next task workflow...]
```


## Project Structure

```
klp/
├── cmd/
│   ├── crawler/main.go          # Crawler entry point
│   └── server/main.go           # Web server entry point
├── internal/
│   ├── api/                     # REST API handlers
│   │   ├── handlers.go          # API endpoint implementations
│   │   ├── handlers_test.go     # API tests (78.1% coverage)
│   │   ├── middleware.go        # Logging & CORS middleware
│   │   └── routes.go            # Route definitions
│   ├── crawler/                 # Web scraping logic
│   │   ├── scraper.go           # Main crawler implementation
│   │   ├── scraper_test.go      # Crawler tests
│   │   ├── geocoder.go          # Nominatim geocoding
│   │   ├── models.go            # Helper functions
│   │   ├── AGENTS.md            # Crawler-specific notes (KEEP THIS)
│   │   └── testdata/            # Test fixtures
│   └── storage/                 # Data models & JSON I/O
│       ├── models.go            # Venue, Event, Exhibition structs
│       ├── models_test.go       # Model tests (87.5% coverage)
│       ├── json.go              # JSON file operations
│       └── json_test.go         # Storage tests
├── web/
│   ├── static/
│   │   ├── css/
│   │   │   ├── tailwind-input.css  # Tailwind source (semantic tokens, dark mode)
│   │   │   └── tailwind.css        # Generated bundle (committed)
│   │   ├── js/
│   │   │   ├── app.js           # Main application logic + modal stack
│   │   │   ├── map.js           # Leaflet map integration
│   │   │   ├── calendar.js      # Calendar view (client-side filtered)
│   │   │   ├── favorites.js     # Favorites management
│   │   │   ├── filters.js       # Search & filter logic
│   │   │   └── routing.js       # Route planning
│   │   └── images/              # Static images
│   └── templates/
│       └── index.html           # Main HTML template
├── data/                        # JSON data files (generated by crawler)
│   ├── venues.json
│   ├── events.json
│   └── exhibitions.json
├── deploy/                      # Deploy script + systemd unit
│   ├── deploy.sh
│   └── vibeklp.service
├── plans/                       # Architecture documentation
│   ├── kulturelle-landpartie-webapp.md
│   ├── api-specification.md
│   └── crawler-strategy.md
├── Makefile                     # Tailwind build, server/crawler shortcuts
├── tailwind.config.js           # Tailwind theme config
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── test.sh                      # Test runner script
├── README.md                    # User documentation
├── QUICKSTART.md                # Quick start guide
├── TESTING.md                   # Detailed testing guide
└── TASKS.md                     # Frontend task tracker
```

## Code Organization & Patterns

### Package Structure

**`internal/storage`** - Data models and persistence
- `models.go`: Defines `Venue`, `Event`, `Exhibition`, `Address`, `Coordinates`, `Contact` structs
- `json.go`: JSON file I/O operations
- All structs have JSON tags for serialization
- Validation methods: `Validate()` on models

**`internal/crawler`** - Web scraping
- `scraper.go`: Main crawler with rate limiting (2s between requests)
- User-Agent: `"KLP-Crawler/1.0 (kulturelle-landpartie)"`
- Base URL: `https://www.kulturelle-landpartie.de`
- Timeout: 30 seconds per request
- See `internal/crawler/AGENTS.md` for detailed crawler notes

**`internal/api`** - REST API
- `handlers.go`: HTTP handlers for all endpoints
- `routes.go`: Route registration with Gorilla Mux
- `middleware.go`: Logging and CORS middleware
- All responses use `respondJSON()` helper
- Error responses use `respondError()` helper

### Naming Conventions

**Files:**
- Implementation: `foo.go`
- Tests: `foo_test.go`
- Same package name in both

**Functions:**
- Exported: `PascalCase` (e.g., `GetVenues`, `LoadJSON`)
- Unexported: `camelCase` (e.g., `respondJSON`, `parseEventFromVenue`)
- Test functions: `TestFunctionName` or `TestType_Method`

**Variables:**
- Exported: `PascalCase`
- Unexported: `camelCase`
- Constants: `PascalCase` or `SCREAMING_SNAKE_CASE`

**Struct Tags:**
- Always use JSON tags: `json:"fieldName,omitempty"`
- Omit empty fields with `omitempty` for optional data

### Error Handling

**Standard pattern:**
```go
result, err := SomeFunction()
if err != nil {
    return fmt.Errorf("descriptive context: %w", err)
}
```

**API handlers:**
```go
if err != nil {
    h.respondError(w, http.StatusInternalServerError, "User-friendly message")
    return
}
```

**Never ignore errors** - always check and handle them.

### Testing Patterns

**Table-driven tests:**
```go
tests := []struct {
    name    string
    input   string
    want    string
    wantErr bool
}{
    {"description", "input", "expected", false},
    // ...
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := Function(tt.input)
        if (err != nil) != tt.wantErr {
            t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
        }
        if got != tt.want {
            t.Errorf("got %v, want %v", got, tt.want)
        }
    })
}
```

**HTTP testing:**
```go
req := httptest.NewRequest("GET", "/api/venues", nil)
w := httptest.NewRecorder()
handler.GetVenues(w, req)

if w.Code != http.StatusOK {
    t.Errorf("Status = %d, want 200", w.Code)
}

var response map[string]interface{}
json.NewDecoder(w.Body).Decode(&response)
```

**Cleanup:**
```go
tempDir, _ := os.MkdirTemp("", "test-*")
defer os.RemoveAll(tempDir)
```

## API Endpoints

All endpoints are prefixed with `/api`:

### Venues
- `GET /api/venues` - List all venues
  - Query params: `search`, `amenity`
  - Returns: `{"venues": [...], "total": N}`
- `GET /api/venues/:id` - Get single venue with related events/exhibitions

### Events
- `GET /api/events` - List all events
  - Query params: `date`, `dateFrom`, `dateTo`, `category`, `venueId`
  - Returns: `{"events": [...], "total": N}`
- `GET /api/events/:id` - Get single event with venue details

### Exhibitions
- `GET /api/exhibitions` - List all exhibitions
  - Returns: `{"exhibitions": [...], "total": N}`
- `GET /api/exhibitions/:id` - Get single exhibition

### Other
- `GET /api/search?q=query` - Global search across all types
  - Query params: `q` (required), `type` (optional: "venue", "event", "exhibition")
  - Returns: `{"results": [...], "total": N}`
- `GET /api/calendar` - Get events grouped by date with German weekday names
- `GET /api/categories` - Get all event categories with counts and colors
- `GET /api/stats` - Get statistics (venue count, event count, categories, bike routes)

See `plans/api-specification.md` for full API documentation.

## Data Models

### Venue
```go
type Venue struct {
    ID              string      `json:"id"`
    Name            string      `json:"name"`
    Description     string      `json:"description,omitempty"`
    Address         Address     `json:"address"`
    Coordinates     Coordinates `json:"coordinates"`
    Contact         Contact     `json:"contact"`
    Amenities       []string    `json:"amenities,omitempty"`
    BikeRoute       string      `json:"bikeRoute,omitempty"`
    EventIDs        []string    `json:"eventIds,omitempty"`
    ExhibitionIDs   []string    `json:"exhibitionIds,omitempty"`
    EventCount      int         `json:"eventCount"`
    ExhibitionCount int         `json:"exhibitionCount"`
}
```

**Validation rules:**
- Name is required
- Postal code is required
- Coordinates required (lat/lng != 0)
- Lat must be in range 52.5-53.5 (Wendland region)
- Lng must be in range 10.5-12.0

### Event
```go
type Event struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description,omitempty"`
    VenueID     string `json:"venueId"`
    VenueName   string `json:"venueName,omitempty"`
    Date        string `json:"date"` // ISO 8601: YYYY-MM-DD
    StartTime   string `json:"startTime,omitempty"`
    EndTime     string `json:"endTime,omitempty"`
    Category    string `json:"category,omitempty"`
    Admission   string `json:"admission,omitempty"`
    ImageURL    string `json:"imageUrl,omitempty"`
}
```

**Validation rules:**
- Title is required
- VenueID is required
- Date is required and must be in format `YYYY-MM-DD`

### Exhibition
```go
type Exhibition struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description,omitempty"`
    Artist      string `json:"artist,omitempty"`
    VenueID     string `json:"venueId"`
    VenueName   string `json:"venueName,omitempty"`
    ImageURL    string `json:"imageUrl,omitempty"`
}
```

**Validation rules:**
- Title is required
- VenueID is required

## Frontend Architecture

### JavaScript Modules

**`app.js`** - Main application controller
- Initializes map, loads data, manages views
- Global functions: `switchView()`, `showVenueDetails()`, `showEventDetails()`
- Filter application and result rendering

**`map.js`** - Leaflet map integration
- Map initialization with OpenStreetMap tiles
- Marker clustering (MarkerClusterGroup)
- Popup rendering for venues
- Exports: `initMap()`, `clearMap()`, `locateUser()`

**`calendar.js`** - Calendar view
- Reads `App.data.events` and applies the active filter state — **no** server fetch (do not reintroduce one)
- Events grouped by date, then sorted by `startTime` (whole-day events float to top)
- Multi-day sections are collapsible `<details>` elements
- Exports: `loadCalendar()`

**`favorites.js`** - Favorites management
- LocalStorage persistence
- Add/remove/toggle favorites for venues and events
- Badge counter updates
- Exports: `addFavorite()`, `removeFavorite()`, `isFavorite()`, `loadFavorites()`, `toggleFavorite()`, `updateFavoritesBadge()`

**`filters.js`** - Search and filtering
- Search input handling with debounce
- Date and category filters
- Bike route filter
- Applies filters and triggers re-render

**`routing.js`** - Route planning
- OSRM routing API integration
- Bike route profile (hardcoded)
- Waypoint management
- Route visualization on map

### CSS Organization

Tailwind CSS v3 via the **standalone CLI** (no Node toolchain). Sources:

- `web/static/css/tailwind-input.css` — `@layer base` defines semantic CSS variables (`--canvas`, `--surface`, `--ink`, `--muted`, `--accent`, `--border`, …) and re-binds them inside a `@media (prefers-color-scheme: dark)` block.
- `web/static/css/tailwind.css` — generated bundle, **committed** to the repo. Regenerate with `make css` (or `make css-watch`).
- `tailwind.config.js` — `darkMode: 'media'`, custom colors mapped to the CSS variables (`bg-surface`, `text-ink`, `text-accent`, etc.).

**Conventions:**
- Always use semantic tokens (`bg-surface`, `text-ink`, `border-border`) — never raw palette colors (`bg-gray-100`, `text-zinc-900`). Dark mode works automatically.
- No `dark:` class variants; the strategy is `media`, not `class`. There is no theme toggle.
- A class change in HTML/JS won't appear until `tailwind.css` is regenerated.

**Responsive breakpoints:**
- `max-width: 768px` triggers the mobile bottom-sheet sidebar (`body.sidebar-open`, wired in `initMobileSidebar()` in `app.js`).

## Important Gotchas & Non-Obvious Patterns

### 1. Crawler Rate Limiting
The crawler uses a `time.Ticker` with 2-second intervals. **Never remove this** - it prevents being blocked by the website.

```go
c.rateLimiter = time.NewTicker(2 * time.Second)
// In Fetch():
<-c.rateLimiter.C
```

### 2. Geocoding Rate Limits
Nominatim API has strict rate limits (1 request/second). The geocoder already implements this. Use `-skip-geocoding` flag for testing.

### 3. Date Format Hardcoding
The crawler hardcodes the year 2026 when parsing dates from the website (which only shows `DD.MM.` format). Update this for future years.

```go
// In parseEventFromVenue():
year := 2026  // TODO: Make this configurable
```

### 4. Server Port Default
The server defaults to port **8081** (not 8080). This is intentional to avoid conflicts.

### 5. German Weekday Names
Calendar API returns German weekday names. Frontend expects this format:

```go
weekday := []string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}[t.Weekday()]
```

### 6. File Paths in Tests
Always use `os.MkdirTemp()` for test data directories. Never hardcode paths.

```go
tempDir, err := os.MkdirTemp("", "klp-test-*")
defer os.RemoveAll(tempDir)
```

### 7. CORS Middleware Order
CORS middleware must be applied to the router in `cmd/server/main.go`. It's already configured - don't modify unless necessary.

### 8. JSON File Overwrites
`storage.SaveVenues()`, `SaveEvents()`, etc. **overwrite** existing files. There's no append mode.

### 9. Venue ID Generation
Venue IDs are an MD5-hash prefix of the venue's source identifier (see `GenerateID` in `internal/crawler/models.go`). The format is `venue-<8-byte-hex>`, e.g. `venue-4044317c59e8f143`. The same scheme is used for `event-…` and `exhibition-…` IDs.

### 10. Frontend Global Functions
Several JS functions are exported to `window` for cross-module access:
- `map.js`: `initMap`, `clearMap`, `locateUser`
- `calendar.js`: `loadCalendar`
- `favorites.js`: `addFavorite`, `removeFavorite`, `isFavorite`, etc.

Don't change these to ES6 modules without refactoring all imports.

### 11. Static Asset Paths
In HTML/CSS, static assets use `/static/` prefix. The server strips this when serving:

```html
<link rel="stylesheet" href="/static/css/styles.css">
<!-- Served from: ./web/static/css/styles.css -->
```

### 12. API Response Wrapper
API responses are wrapped in objects:

```json
{
  "venues": [...],
  "total": 42
}
```

**Not:**
```json
[...]  // ❌ Never return bare arrays
```

### 13. Coordinates Validation
Venue coordinates are validated to be within the Wendland region:
- Latitude: 52.5 - 53.5
- Longitude: 10.5 - 12.0

If geocoding returns coordinates outside this range, validation fails.

### 14. Error vs Warning in Crawler
The crawler logs warnings for individual venue failures but continues processing. It only returns errors for catastrophic failures (can't fetch main page).

### 15. go.mod Warnings
You may see warnings about unused dependencies (`github.com/rs/cors`) or indirect dependencies that should be direct. Run `go mod tidy` to fix.

### 16. Search Implementation (Venue-Centric)
The search is **venue-centric** — results are venues, even when the match came from an event or exhibition.

**How it works:**
1. Results list shows only venues (~93 items), not the ~1k mixed event/exhibition records
2. Search checks: venue name/description/address + events/exhibitions AT that venue
3. Match badges display "3 Events, 2 Ausstellungen" when search matches content
4. Date/category filters show venues with matching events

**Key functions in app.js:**
- `applyFilters()` - Main filtering logic with null-safe property access
- `updateResults()` - Renders venue-only results list
- `createResultItem()` - Creates venue cards with match badges

**Performance optimization:**
- Uses Map lookups for events/exhibitions instead of repeated array scans
- Builds lookup maps once per filter operation

**Null-safety critical:**
Many events lack descriptions. Always use:
```javascript
// ❌ BAD - crashes if description is undefined
venue.description.toLowerCase()

// ✅ GOOD - safe property access
(venue.description && venue.description.toLowerCase())
```

### 17. Error Handling System
Replaced native `alert()` with inline message system:

**Functions:**
- `showError(message, options)` - Display error banner with optional retry button
- `showSuccess(message)` - Display success banner
- `dismissError(id)` - Close specific message
- `fetchWithRetry(url, options, maxRetries)` - Auto-retry failed requests

**Features:**
- Slide-in animations via CSS
- Auto-dismiss (10s errors, 5s success)
- Close buttons on all messages
- Retry buttons for retryable errors
- HTML escaping to prevent XSS

**Usage:**
```javascript
// Simple error
showError('Failed to load data');

// Error with retry
showError('Connection failed', { retry: 'location.reload()' });

// Success message
showSuccess('Favorites imported successfully');
```

**CSS classes:**
- `.error-message` - Red error banner
- `.success-message` - Green success banner
- `.error-container` - Fixed position container at top of page

## Development Workflow

### Adding a New API Endpoint

1. Define handler in `internal/api/handlers.go`:
```go
func (h *Handler) GetMyEndpoint(w http.ResponseWriter, r *http.Request) {
    // Implementation
    h.respondJSON(w, http.StatusOK, data)
}
```

2. Register route in `internal/api/routes.go`:
```go
api.HandleFunc("/myendpoint", handler.GetMyEndpoint).Methods("GET")
```

3. Add tests in `internal/api/handlers_test.go`:
```go
func TestGetMyEndpoint(t *testing.T) {
    // Test implementation
}
```

4. Update `plans/api-specification.md` with documentation

### Adding a New Data Field

1. Update struct in `internal/storage/models.go`:
```go
type Venue struct {
    // ...
    NewField string `json:"newField,omitempty"`
}
```

2. Update validation if needed:
```go
func (v *Venue) Validate() error {
    // Add validation rules
}
```

3. Update crawler parsing in `internal/crawler/scraper.go`

4. Add tests in `internal/storage/models_test.go`

5. Run tests: `./test.sh`

### Fixing Crawler Issues

1. Always check `internal/crawler/AGENTS.md` first for known issues

2. Fetch sample HTML to verify structure:
```bash
curl -s "https://www.kulturelle-landpartie.de/orte/bankewitz.html" > test.html
```

3. Test selectors with a quick Go script or inspect HTML manually

4. Update selectors in `scraper.go` and parsing functions

5. Run crawler in verbose mode:
```bash
go run cmd/crawler/main.go -verbose -skip-geocoding
```

6. Verify output:
```bash
cat data/events.json | jq length
cat data/events.json | jq '.[0]'
```

## Common Issues & Solutions

### Crawler Returns Empty Data

**Symptoms:** `events.json` and `exhibitions.json` contain `null` or empty arrays.

**Diagnosis:**
1. Run with `-verbose` flag to see what's being fetched
2. Fetch a sample page and inspect HTML structure
3. Check if selectors in `scraper.go` match actual HTML

**Solution:** Update CSS selectors in `parseEventFromVenue()` and `parseExhibitionFromVenue()`. See `internal/crawler/AGENTS.md` for examples of past fixes.

### Test Failures on Fresh Clone

**Symptoms:** Tests fail with "file not found" or "no such directory".

**Solution:** 
1. Run `go mod download` first
2. Ensure test data directory is created with `os.MkdirTemp()`
3. Check that tests aren't looking for hardcoded paths

### Server Won't Start - Port Already in Use

**Symptoms:** `listen tcp :8081: bind: address already in use`

**Solution:**
```bash
# Find process using port
lsof -i :8081
# Kill it or use different port
go run cmd/server/main.go -port 8082
```

### Frontend Shows No Data

**Symptoms:** Map loads but no markers appear.

**Diagnosis:**
1. Open browser console (F12) - check for API errors
2. Verify server is running: `curl http://localhost:8081/api/venues`
3. Check that data files exist: `ls -lh data/`

**Solution:**
1. Run crawler first: `go run cmd/crawler/main.go`
2. Restart server
3. Check frontend console for JavaScript errors

### Geocoding Fails

**Symptoms:** Venues have no coordinates or geocoding warnings.

**Causes:**
- Nominatim rate limits (too fast)
- Invalid addresses
- Network issues

**Solutions:**
1. Use `-skip-geocoding` for testing
2. Check address format in crawler output
3. Manually verify a few addresses on maps.google.com
4. The geocoder already delays 1 second between requests - don't reduce this

### Tests Pass Locally But Fail in CI

**Common causes:**
- Timezone differences (use UTC in tests)
- Path separators (use `filepath.Join()`)
- Missing environment variables
- Different Go version

**Best practices:**
- Use `filepath.Join()` for all paths
- Use `os.MkdirTemp()` for temp files
- Don't depend on specific locale/timezone
- Test on Docker to match CI environment

## Known Issues & TODOs

`TASKS.md` is the source of truth — completed items are struck through with a date; open items live under their priority section. Check there before re-implementing anything that looks broken.

## Documentation Files

- **README.md** - User-facing documentation, installation, deployment
- **QUICKSTART.md** - Quick start guide for users
- **TESTING.md** - Comprehensive testing guide with examples
- **TASKS.md** - Frontend issues and task tracker
- **AGENTS.md** (this file) - Developer/agent guide
- **internal/crawler/AGENTS.md** - Crawler-specific troubleshooting
- **plans/*.md** - Architecture and API specification docs

## Contact & Support

For issues or questions:
- Check existing documentation first (AGENTS.md, TESTING.md, TASKS.md)
- Review `internal/crawler/AGENTS.md` for crawler-specific issues
- Check browser console for frontend errors
- Review server logs for API errors
- GitHub issues: https://github.com/Postlab-Luchow/vibeklp/issues
