
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
