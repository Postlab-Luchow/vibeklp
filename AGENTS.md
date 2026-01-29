# Agent Development Notes

This file documents findings, insights, and troubleshooting notes for the Kulturelle Landpartie (KLP) crawler project.

## Issue: Events and Exhibitions Not Being Extracted (Fixed: 2026-01-28)

### Problem
The crawler was successfully finding venue pages but failing to extract any events or exhibitions, resulting in `null` in the `events.json` and `exhibitions.json` files.

### Root Cause Analysis

#### Investigation Steps
1. Ran the crawler with `-verbose` flag to observe behavior
2. Verified venue links were being discovered correctly (~44+ venues found)
3. Checked the output JSON files - found `null` instead of data arrays
4. Fetched actual HTML from a sample venue page (bankewitz.html)
5. Analyzed the DOM structure to identify the correct selectors

#### Key Finding
The CSS selectors in the code didn't match the actual HTML structure of kulturelle-landpartie.de:

**Incorrect selectors (before fix):**
```go
doc.Find(".veranstaltung, .event, [class*='event']")      // Events
doc.Find(".ausstellung, .exhibition, [class*='ausstellung']") // Exhibitions
```

**Actual HTML structure:**
```html
<!-- Exhibitions -->
<div class="slider aus">
  <div class="item">
    <div class="img">...</div>
    <div>
      <p>Artist Name</p>
      <p><b>Exhibition Title</b><br/><em>Description</em></p>
    </div>
  </div>
  <!-- more items... -->
</div>

<!-- Events -->
<div class="slider ver">
  <div class="item">
    <div class="img">...</div>
    <div>
      <p>Organizer Name</p>
      <p><b>Event Title</b><br/><em>Description</em></p>
      29.05. 17:00<br/>
      (Hutkasse)
    </div>
  </div>
  <!-- more items... -->
</div>
```

### Solution Implemented

#### 1. Updated Selectors (`internal/crawler/scraper.go:320-336`)
```go
// Events
doc.Find(".slider.ver .item")

// Exhibitions
doc.Find(".slider.aus .item")
```

#### 2. Rewrote Parsing Functions

**`parseEventFromVenue()` - Key parsing logic:**
- **Title**: Extract from `<b>` tag in second `<div>`
- **Description**: Extract from `<em>` tag
- **Artist/Organizer**: Extract from first `<p>` tag
- **Date/Time**: Parse using regex pattern `(\d{2}\.\d{2}\.)\s+(\d{2}:\d{2})`
- **Admission**: Extract from parentheses at end using regex `\(([^)]+)\)\s*$`

**`parseExhibitionFromVenue()` - Key parsing logic:**
- **Title**: Extract from `<b>` tag in second `<div>`
- **Description**: Extract from `<em>` tag
- **Artist**: Extract from first `<p>` tag

#### 3. Added Missing Import
Added `"regexp"` to imports in `scraper.go` for date/time parsing.

### Verification Results

Tested with sample venue (Bankewitz):
```
Before fix: 0 events, 0 exhibitions
After fix:  9 events, 3 exhibitions ✓
```

Sample extracted data:
- Event: "probt mit Anleitung zum Mitsingen" - Süttorfer Sängerei, 2026-05-29 17:00
- Exhibition: "KUNST & Gebrauchserfreulichkeiten mit MOKKA" - Wanda Sippl

All fields properly populated in JSON output.

## Website Structure Insights

### kulturelle-landpartie.de Organization

**URL Pattern:**
- Main venues page: `/orte.html`
- Individual venues: `/orte/{venue-name}.html`
- Pagination: Links to next page with class `.step` and title containing "nächste"

**Pagination System:**
- Main page has alphabetical navigation (a, b, d, f, g, h, etc.)
- Each letter page may span multiple paginated pages
- Navigation links have class `.step` with arrows (◄ ►)

**Content Sections on Venue Pages:**
- Venue info in `#comblock` div
- Contact details: phone (`a[href^='tel:']`), email (`a[href^='mailto:']`), website (`a[href^='http']`)
- Exhibitions: `.slider.aus .item` items
- Events: `.slider.ver .item` items

**German Terminology:**
- `aus` = Ausstellungen (exhibitions)
- `ver` = Veranstaltungen (events)
- `orte` = places/venues

### HTML/CSS Classes Reference

| Element | Classes | Purpose |
|---------|---------|---------|
| `.slider.aus` | Container | Holds all exhibitions |
| `.slider.ver` | Container | Holds all events |
| `.item` | Item wrapper | Individual event/exhibition |
| `.step` | Navigation | Pagination links |
| `#comblock` | Content block | Venue information section |

### Date Format
- Website uses: `DD.MM.` format (e.g., "29.05.")
- Crawler converts to: `YYYY-MM-DD` format (e.g., "2026-05-29")
- Year assumption: 2026 (hardcoded as next festival year)

## Crawler Configuration

### Rate Limiting
- Current setting: 2 seconds between requests (`time.NewTicker(2 * time.Second)`)
- User-Agent: `"KLP-Crawler/1.0 (kulturelle-landpartie)"`
- Timeout: 30 seconds per request

### Crawler Performance
- ~2 seconds per page fetch (rate limited)
- ~100+ venues total on the website
- Estimated full crawl time: ~5-10 minutes (with pagination discovery + venue details)

## Development Tips

### Testing the Crawler

**Quick test with single venue:**
```go
c := crawler.NewCrawler(logger)
venue, events, exhibitions, err := c.CrawlVenueDetails(
    "https://www.kulturelle-landpartie.de/orte/bankewitz.html"
)
```

**Run with verbose logging:**
```bash
go run cmd/crawler/main.go -verbose -skip-geocoding
```

**Check output quickly:**
```bash
ls -lh data/*.json
head -20 data/events.json
```

### Common Issues

1. **Empty/null JSON files**: Check if selectors match current website HTML
2. **Missing data fields**: Inspect actual HTML structure, don't assume class names
3. **Geocoding failures**: Use `-skip-geocoding` flag for faster testing
4. **Rate limiting**: Website may block if crawler runs too fast

### Debugging Workflow

1. Fetch sample page HTML: Use `fetch` tool or curl to get actual HTML
2. Inspect structure: Look at the raw HTML, not just DevTools (JavaScript may modify DOM)
3. Test selectors: Use goquery playground or write small test program
4. Verify extraction: Print parsed data before saving to JSON
5. Check edge cases: Empty descriptions, missing dates, multiple date formats

## Future Maintenance

### Potential Breaking Changes to Watch

The website structure could change in future years. Monitor these areas:

1. **CSS class names** - `.slider.aus`, `.slider.ver` could be renamed
2. **HTML structure** - Number of `<div>` levels, tag arrangement
3. **Date format** - Currently `DD.MM.`, might change
4. **Pagination** - Navigation class names or structure
5. **Contact info encoding** - Email addresses use JavaScript obfuscation

### Suggested Improvements

1. **Add integration tests** with snapshot HTML files for regression testing
2. **Make year configurable** instead of hardcoding 2026
3. **Better error messages** when parsing fails (log which field failed)
4. **Validate JSON schema** after crawling to catch structural issues early
5. **Add retry logic** for failed venue fetches
6. **Support multiple dates per event** (currently only captures first date)
7. **Extract all event dates** - some events span multiple days

### Testing After Website Updates

Before each festival season (likely annual updates):

```bash
# 1. Test single venue first
go run test_crawler.go

# 2. Check structure hasn't changed
curl https://www.kulturelle-landpartie.de/orte/bankewitz.html | grep -E '(slider|item|aus|ver)'

# 3. Run limited crawl
# (modify test_full.go to test ~5 venues)

# 4. Validate output
jq '.[0]' data/events.json
jq '.[0]' data/exhibitions.json

# 5. Full crawl
go run cmd/crawler/main.go -skip-geocoding
```

## Code Architecture Notes

### File Organization
```
internal/crawler/
├── scraper.go      # Main crawling logic, HTTP fetching, parsing
├── models.go       # Helper functions (ParseAddress, CleanText, etc.)
└── geocoder.go     # Geocoding via Nominatim API
```

### Data Flow
```
Website → Fetch HTML → Parse DOM → Extract Data → Storage
                ↓
          goquery.Document
                ↓
         Selection.Find()
                ↓
          Parse functions
                ↓
         storage.Venue/Event/Exhibition
                ↓
          JSON files
```

### Key Functions

- `CrawlAll()` - Main entry point, discovers all venues via pagination
- `CrawlVenueDetails(url)` - Fetches single venue, returns venue + events + exhibitions
- `parseEventFromVenue(selection)` - Extracts event data from DOM selection
- `parseExhibitionFromVenue(selection)` - Extracts exhibition data from DOM selection

## Git History Reference

**2026-01-28** - Fixed event/exhibition extraction
- Updated selectors from generic classes to actual `.slider.aus` and `.slider.ver`
- Rewrote parsing logic to match real HTML structure
- Added regex support for date/time extraction
- Files changed: `internal/crawler/scraper.go`

---

## Testing Infrastructure (Added: 2026-01-29)

### Overview

The project now has comprehensive test coverage for the backend components. Tests are written using Go's standard testing package and cover unit tests, integration tests, and API endpoint tests.

### Test Structure

```
internal/
├── api/
│   └── handlers_test.go      # API endpoint tests (78% coverage)
└── storage/
    ├── models_test.go         # Model validation tests
    └── json_test.go           # Storage I/O tests (87% coverage)
```

### Running Tests

#### Quick Test (All packages)
```bash
./test.sh
```

#### Verbose Output
```bash
./test.sh -v
# or
go test ./... -v
```

#### With Coverage Report
```bash
./test.sh -c
# Then view detailed HTML report:
go tool cover -html=coverage.out
```

#### Test Specific Package
```bash
./test.sh -p ./internal/api
# or
go test ./internal/api/... -v
```

#### Test Individual Function
```bash
go test ./internal/storage/... -run TestVenue_Validate -v
```

### Test Coverage Summary

As of 2026-01-29:
- **API package**: 78.1% coverage
- **Storage package**: 87.5% coverage
- **Overall backend**: 82.8% average coverage

### What's Tested

#### Storage Package Tests
- ✅ Model validation (Venue, Event, Exhibition)
- ✅ Address formatting
- ✅ JSON save/load operations
- ✅ GetByID functions (venue, event, exhibition)
- ✅ GetVenueWithDetails (with related entities)
- ✅ Error handling (missing files, invalid data)
- ✅ Directory creation

#### API Package Tests
- ✅ All GET endpoints (venues, events, exhibitions)
- ✅ Individual item retrieval by ID
- ✅ Query parameter filtering (date, category, venue, search)
- ✅ Search functionality (all types, specific types)
- ✅ Calendar endpoint (event grouping by date)
- ✅ Categories endpoint
- ✅ Statistics endpoint
- ✅ Route registration verification
- ✅ Error responses (404, 400)

### Writing New Tests

When adding new features or modifying existing code:

1. **Create test file** alongside the implementation:
   ```
   myfile.go       → myfile_test.go
   ```

2. **Use table-driven tests** for multiple scenarios:
   ```go
   func TestMyFunction(t *testing.T) {
       tests := []struct {
           name    string
           input   string
           want    string
           wantErr bool
       }{
           {"valid input", "test", "TEST", false},
           {"empty input", "", "", true},
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               got, err := MyFunction(tt.input)
               if (err != nil) != tt.wantErr {
                   t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
               }
               if got != tt.want {
                   t.Errorf("got %v, want %v", got, tt.want)
               }
           })
       }
   }
   ```

3. **Use subtests** for organization:
   ```go
   t.Run("success case", func(t *testing.T) { ... })
   t.Run("error case", func(t *testing.T) { ... })
   ```

4. **Clean up resources** with defer:
   ```go
   tempDir, _ := os.MkdirTemp("", "test-*")
   defer os.RemoveAll(tempDir)
   ```

5. **Run tests before commits**:
   ```bash
   ./test.sh
   ```

### Testing Checklist

Before committing changes:

- [ ] All tests pass: `./test.sh`
- [ ] New code has tests (maintain >75% coverage)
- [ ] Tests verify both success and error cases
- [ ] Integration tests pass for API changes
- [ ] No race conditions: `go test -race ./...`

### CI/CD Integration

Tests can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Run tests
  run: go test ./... -v -race -coverprofile=coverage.out

- name: Check coverage
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$coverage < 75" | bc -l) )); then
      echo "Coverage $coverage% is below 75%"
      exit 1
    fi
```

### Manual Testing Checklist

For frontend and end-to-end testing:

#### Server Startup
```bash
# 1. Start server
go run cmd/server/main.go

# 2. Verify server responds
curl http://localhost:8081/api/venues

# 3. Check frontend loads
open http://localhost:8081
```

#### API Endpoints
```bash
# Test each endpoint manually
curl http://localhost:8081/api/venues | jq
curl http://localhost:8081/api/events?date=2026-05-29 | jq
curl http://localhost:8081/api/exhibitions | jq
curl http://localhost:8081/api/calendar | jq
curl http://localhost:8081/api/categories | jq
curl http://localhost:8081/api/stats | jq
curl "http://localhost:8081/api/search?q=test" | jq
```

#### Frontend Features
- [ ] Map loads with venue markers
- [ ] Clicking marker shows popup
- [ ] Search filters results
- [ ] Date filter works
- [ ] Category filter works
- [ ] Calendar view displays events
- [ ] Favorites can be added/removed
- [ ] Modal opens for venue details
- [ ] Route planning works
- [ ] Mobile responsive design

### Troubleshooting Tests

**Tests fail with "no such file":**
- Tests create temporary directories automatically
- Check that test data setup is correct
- Ensure no hardcoded file paths

**Tests pass locally but fail in CI:**
- Check for OS-specific path issues (use `filepath.Join`)
- Verify all dependencies are available
- Check for timezone-related date parsing issues

**Flaky tests:**
- Avoid time-dependent assertions
- Use deterministic test data
- Check for race conditions: `go test -race`

**Coverage drops:**
- Run `./test.sh -c` to see what's not covered
- Add tests for new code paths
- Consider edge cases and error paths

### Best Practices

1. **Fast tests**: Tests should complete in milliseconds
2. **Isolated tests**: Each test creates its own temp data
3. **Clear assertions**: Use descriptive error messages
4. **No test dependencies**: Tests should run in any order
5. **Use test helpers**: Extract common setup code

---

*This document should be updated whenever significant crawler changes are made or website structure issues are discovered.*
