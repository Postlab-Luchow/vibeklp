# Testing Guide

This document provides a comprehensive guide to testing the Kulturelle Landpartie (KLP) project.

## Quick Start

```bash
# Run all tests
./test.sh

# Run with verbose output
./test.sh -v

# Run with coverage report
./test.sh -c

# Test specific package
./test.sh -p ./internal/api
```

## Test Coverage

Current coverage as of 2026-01-29:

| Package | Coverage | Test Files |
|---------|----------|------------|
| `internal/api` | 78.1% | `handlers_test.go` |
| `internal/storage` | 87.5% | `models_test.go`, `json_test.go` |
| **Overall** | **82.8%** | - |

## Test Organization

```
internal/
├── api/
│   └── handlers_test.go         # 11 tests, 28 subtests
└── storage/
    ├── models_test.go            # 3 tests, 17 subtests  
    └── json_test.go              # 5 tests
```

## What's Tested

### ✅ Storage Package (`internal/storage`)

**Models (`models_test.go`):**
- `TestAddress_String` - Address formatting
- `TestVenue_Validate` - Venue validation rules
  - Valid venue
  - Missing name
  - Missing postal code
  - Missing coordinates
  - Latitude out of range
  - Longitude out of range
- `TestEvent_Validate` - Event validation rules
  - Valid event
  - Missing title
  - Missing venue ID
  - Missing date
  - Invalid date format
- `TestExhibition_Validate` - Exhibition validation rules
  - Valid exhibition
  - Missing title
  - Missing venue ID

**Storage (`json_test.go`):**
- `TestStorage_SaveAndLoad` - JSON file I/O
  - Save/load venues
  - Save/load events
  - Save/load exhibitions
- `TestStorage_GetByID` - Item retrieval
  - Get venue by ID (found & not found)
  - Get event by ID
  - Get exhibition by ID
- `TestStorage_GetVenueWithDetails` - Related data loading
- `TestStorage_LoadJSON_FileNotFound` - Error handling
- `TestStorage_EnsureDataDir` - Directory creation

### ✅ API Package (`internal/api`)

**Handlers (`handlers_test.go`):**
- `TestGetVenues` - Venue listing
  - Get all venues
  - Search venues by name
- `TestGetVenue` - Single venue retrieval
  - Existing venue
  - Nonexistent venue (404)
- `TestGetEvents` - Event listing with filters
  - Get all events
  - Filter by date
  - Filter by category
  - Filter by venue
- `TestGetEvent` - Single event retrieval
  - Existing event with venue details
- `TestGetExhibitions` - Exhibition listing
- `TestSearch` - Global search
  - Missing query parameter (400)
  - Search all types
  - Search specific type
- `TestGetCalendar` - Calendar view
  - Event grouping by date
  - German weekday names
- `TestGetCategories` - Category listing
  - Category counts
  - Color assignment
- `TestGetStats` - Statistics
  - Venue/event/exhibition counts
  - Bike route statistics
- `TestSetupRoutes` - Route registration
  - All endpoints registered

## Running Tests

### Basic Commands

```bash
# All tests
go test ./...

# Verbose mode
go test ./... -v

# With coverage
go test ./... -cover

# Coverage profile
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Race detector
go test -race ./...

# Specific package
go test ./internal/api/...

# Specific test
go test ./internal/storage/... -run TestVenue_Validate

# Specific subtest
go test ./internal/api/... -run TestGetEvents/filter_by_date
```

### Using test.sh Script

The `test.sh` script provides a convenient wrapper:

```bash
./test.sh              # Run all tests
./test.sh -v           # Verbose output
./test.sh -c           # With coverage report
./test.sh -p ./internal/api  # Test specific package
./test.sh -h           # Show help
```

**Script features:**
- Colored output (green ✅ for pass, red ❌ for fail)
- Coverage report generation
- HTML coverage viewer suggestion
- Easy command-line options

## Writing Tests

### Test File Naming

- Implementation: `foo.go`
- Test file: `foo_test.go`
- Same package: `package foo`

### Table-Driven Tests

Use table-driven tests for multiple scenarios:

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
        {"special chars", "t@st", "T@ST", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Subtests

Organize related tests with subtests:

```go
func TestAPI(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        // test success case
    })
    
    t.Run("error", func(t *testing.T) {
        // test error case
    })
}
```

### Test Helpers

Extract common setup code:

```go
func setupTestStorage(t *testing.T) (*storage.Storage, func()) {
    t.Helper()
    
    tempDir, _ := os.MkdirTemp("", "test-*")
    store := storage.NewStorage(tempDir)
    
    // Setup test data...
    
    cleanup := func() {
        os.RemoveAll(tempDir)
    }
    
    return store, cleanup
}

func TestSomething(t *testing.T) {
    store, cleanup := setupTestStorage(t)
    defer cleanup()
    
    // Use store...
}
```

### HTTP Testing

Use `httptest` for API tests:

```go
func TestHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/api/venues", nil)
    w := httptest.NewRecorder()
    
    handler.GetVenues(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Status = %d, want 200", w.Code)
    }
    
    var response map[string]interface{}
    json.NewDecoder(w.Body).Decode(&response)
    // Assert response...
}
```

## Test Best Practices

### ✅ DO

- **Write tests before fixing bugs** - Reproduce the bug first
- **Use descriptive test names** - Clear what's being tested
- **Test both success and failure** - Cover happy path & errors
- **Use table-driven tests** - Multiple scenarios efficiently
- **Clean up resources** - Use `defer` for cleanup
- **Test public APIs** - Focus on exported functions
- **Keep tests fast** - Under 100ms per test
- **Make tests deterministic** - No random data or time dependencies
- **Use test helpers** - Extract common setup
- **Check coverage** - Aim for >75% coverage

### ❌ DON'T

- **Don't test private functions** - Focus on public API
- **Don't use real external services** - Mock or use test servers
- **Don't hardcode paths** - Use `os.MkdirTemp` and `filepath.Join`
- **Don't depend on test order** - Tests should be independent
- **Don't ignore errors** - Always check error returns
- **Don't use time.Sleep** - Use channels or mocks
- **Don't test implementation details** - Test behavior, not internals
- **Don't write flaky tests** - Must pass consistently

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...
      
      - name: Check coverage
        run: |
          total=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Total coverage: $total%"
          if (( $(echo "$total < 75" | bc -l) )); then
            echo "Coverage $total% is below 75%"
            exit 1
          fi
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

## Troubleshooting

### Tests Fail with "no such file"

**Problem:** Tests can't find data files.

**Solution:** Tests create temporary directories. Check test setup:

```go
tempDir, _ := os.MkdirTemp("", "test-*")
defer os.RemoveAll(tempDir)
store := storage.NewStorage(tempDir)
```

### Tests Pass Locally but Fail in CI

**Problem:** Environment differences.

**Solutions:**
- Use `filepath.Join` instead of hardcoded paths
- Don't depend on specific timezone
- Avoid OS-specific features
- Check for missing dependencies

### Flaky Tests

**Problem:** Tests pass sometimes, fail other times.

**Solutions:**
- Remove time-based logic (use mocks)
- Avoid race conditions (use `-race` to detect)
- Don't depend on external services
- Use deterministic test data

### Coverage Drops

**Problem:** New code reduces overall coverage.

**Solutions:**
- Run `./test.sh -c` to see what's not covered
- Add tests for new code paths
- Test error paths and edge cases
- Check `go tool cover -html=coverage.out` for visual report

## Manual Testing

Before major releases, manually test:

### API Endpoints

```bash
# Start server
go run cmd/server/main.go

# Test endpoints
curl http://localhost:8081/api/venues | jq
curl http://localhost:8081/api/events?date=2026-05-29 | jq
curl http://localhost:8081/api/search?q=test | jq
```

### Frontend

- [ ] Open http://localhost:8081
- [ ] Map loads with markers
- [ ] Click marker shows popup
- [ ] Search filters results
- [ ] Calendar view works
- [ ] Favorites add/remove works
- [ ] Modal opens/closes
- [ ] Mobile responsive

## Resources

- **Go Testing Package:** https://pkg.go.dev/testing
- **Table-Driven Tests:** https://go.dev/wiki/TableDrivenTests
- **httptest Package:** https://pkg.go.dev/net/http/httptest
- **Project Details:** See `AGENTS.md` for in-depth testing notes

---

**Last Updated:** 2026-01-29  
**Test Coverage:** 82.8% (API: 78.1%, Storage: 87.5%)
