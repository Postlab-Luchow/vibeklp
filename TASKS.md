# Web Interface Tasks

This document tracks tasks and improvements needed for the Kulturelle Landpartie web interface.

## Critical Issues

### 1. ~~Missing API Response Fields~~ ✅ COMPLETED (2026-01-29)
- **Problem**: Frontend expects certain fields that may not be returned by API
  - ~~`event.venueName` expected but not provided by `/api/events` endpoint~~ ✅ Fixed
  - ~~`exhibition.venueName` expected but not provided by `/api/exhibitions` endpoint~~ ✅ Fixed
  - `venue.eventCount` and `venue.exhibitionCount` - Already exist in model, populated by crawler
- **Solution**: Added venue name enrichment to all relevant endpoints
  - `/api/events`, `/api/exhibitions`, `/api/calendar`, `/api/search` now populate venueName
  - Event and exhibition detail endpoints also return venueName
- **Files**: 
  - `internal/api/handlers.go` - added enrichment logic
  - `internal/api/validation.go` - new validation module
- **Priority**: ~~High~~ DONE

### 2. Missing Exhibition Details Functionality
- **Problem**: Frontend has `showEventDetails()` but exhibitions don't have a detail view
  - No `/api/exhibitions/:id` endpoint implementation
  - No JavaScript function to show exhibition details
  - Exhibition cards in favorites don't have click handlers
- **Files**: 
  - `internal/api/routes.go` - add exhibition detail route
  - `internal/api/handlers.go` - already has GetExhibition() at line 245
  - `web/static/js/app.js` - needs showExhibitionDetails() function
  - `web/static/js/favorites.js:228` - exhibition card missing click handler
- **Priority**: Medium

### 3. Error Handling in Frontend
- **Problem**: Very basic error handling using `alert()`
  - Line 402 in app.js uses simple alert for errors
  - No toast/notification system for better UX
  - No retry mechanism for failed API calls
- **Files**: `web/static/js/app.js`
- **Priority**: Medium

## Data & API Issues

### 4. Calendar API Response Structure Mismatch
- **Problem**: API returns calendar as object with date keys, frontend expects it to work with Object.keys()
  - Backend returns: `{ "calendar": { "2026-05-29": {...}, "2026-05-30": {...} } }`
  - Frontend correctly handles this but could benefit from a sorted array instead
- **Files**: 
  - `internal/api/handlers.go:364-410` - GetCalendar()
  - `web/static/js/calendar.js:6, 24`
- **Priority**: Low (currently works but not optimal)

### 5. Missing Venue Details Enrichment
- **Problem**: `/api/venues/:id` endpoint returns raw venue without counts
  - Should include event count, exhibition count
  - Should include full event and exhibition objects, not just IDs
- **Files**: `internal/api/handlers.go:88-100`
- **Priority**: Medium

### 6. ~~Search Results Don't Include venueName~~ ✅ COMPLETED (2026-01-29)
- **Problem**: Search returns IDs but UI needs venue names for events/exhibitions
  - ~~`VenueName` field missing on Event/Exhibition structs~~ - Actually already existed!
  - ~~Search() function wasn't populating the field~~ ✅ Fixed
- **Solution**: 
  - VenueName field already existed in storage models
  - Added venue name lookup in Search() handler for events and exhibitions
  - Exhibition results now fallback to venue name if artist is empty
- **Files**: 
  - `internal/api/handlers.go` - updated Search() function
  - `internal/storage/models.go` - VenueName field already present
- **Priority**: ~~High~~ DONE

## UI/UX Issues

### 7. No Loading States During Navigation
- **Problem**: Switching views doesn't show loading indicators
  - Calendar and favorites load data but don't show loading state
  - `switchView()` function doesn't trigger loading indicator
- **Files**: `web/static/js/app.js:134-160`
- **Priority**: Low

### 8. Favorites Badge Hidden by Default
- **Problem**: Badge starts with `display: none` and only shows when count > 0
  - This causes layout shift when favorites are added
- **Files**: `web/static/js/favorites.js:73`
- **Priority**: Low

### 9. Modal Close Handler Potentially Broken
- **Problem**: Modal close button uses optional chaining which might not work if modal doesn't exist yet
  - `document.querySelector('.modal-close')?.addEventListener`
- **Files**: `web/static/js/app.js:374-382`
- **Priority**: Low

### 10. No Empty State for Results List
- **Problem**: When no results found after filtering, results list is just empty
  - Should show "No results found" message
- **Files**: `web/static/js/app.js:208-227`
- **Priority**: Low

## Functionality Gaps

### 11. Route Planning Not Integrated with Current Location
- **Problem**: User location button exists but route planning doesn't use it as start point
  - `locateUser()` adds a marker but doesn't integrate with routing
  - No "Route from my location" feature
- **Files**: 
  - `web/static/js/map.js:143-170` - locateUser()
  - `web/static/js/routing.js` - routing module
- **Priority**: Medium

### 12. Bike Route Display Missing
- **Problem**: Venues have `bikeRoute` field but it's only shown in popups
  - No visualization of bike routes on map
  - No filter to show only venues on specific bike routes
  - Frontend expects `bikeRouteFilter` checkbox but functionality unclear
- **Files**: 
  - `web/static/js/filters.js:17-21` - bike route filter
  - `web/static/js/app.js:175` - filter logic checks venue.bikeRoute
  - `web/static/js/map.js:95` - shown in popup
- **Priority**: Low

### 13. Routing Profile Hardcoded to Bike
- **Problem**: Routing always uses bike profile
  - No option to switch between bike, car, walking
  - Line 102 in routing.js hardcodes `profile: 'bike'`
- **Files**: `web/static/js/routing.js:102`
- **Priority**: Low

### 14. Export/Import Favorites UI Missing
- **Problem**: Functions exist but no UI buttons to trigger them
  - `exportFavorites()` and `importFavorites()` defined but not accessible
- **Files**: 
  - `web/static/js/favorites.js:234-263` - functions exist
  - `web/templates/index.html` - no buttons in UI
- **Priority**: Low

## Code Quality Issues

### 15. Inconsistent Date Formatting
- **Problem**: Date formatting duplicated across files
  - `formatDate()` in app.js (line 385) uses one approach
  - `GetCalendar()` in handlers.go (line 377-393) duplicates German weekday logic
  - Different formats used in different places
- **Files**: 
  - `web/static/js/app.js:385-391`
  - `internal/api/handlers.go:377-393`
- **Priority**: Low

### 16. Inline Styles in JavaScript
- **Problem**: routing.js generates elements with inline styles
  - Lines 149-159, 203-209 in routing.js
  - Should use CSS classes instead
- **Files**: `web/static/js/routing.js`
- **Priority**: Low

### 17. Global Function Exports Not Consistent
- **Problem**: Some modules export functions to window, others don't
  - map.js exports 3 functions (line 226-228)
  - calendar.js exports 1 function (line 89)
  - favorites.js exports 6 functions (line 266-271)
  - No clear pattern for what should be global vs. module-scoped
- **Files**: All JS files in `web/static/js/`
- **Priority**: Low

### 18. ~~No Input Validation~~ ✅ COMPLETED (2026-01-29)
- **Problem**: No validation of user inputs
  - ~~Search input not sanitized~~ ✅ Fixed
  - ~~Filter values not validated~~ ✅ Fixed
  - ~~Could lead to XSS if server reflects inputs~~ ✅ Prevented
- **Solution**: Created comprehensive validation module
  - HTML escaping for all user inputs (XSS prevention)
  - Regex validation for IDs (alphanumeric + dash/underscore)
  - Date format validation (YYYY-MM-DD)
  - Search query length limits (min 2, max 100 chars)
  - Search type validation (venues/events/exhibitions)
  - All query parameters now sanitized via GetQueryParam() helpers
- **Files**: 
  - `internal/api/validation.go` - NEW validation module
  - `internal/api/handlers.go` - updated to use validation
- **Priority**: ~~Medium~~ DONE

## Missing Features (from plan but not implemented)

### 19. Statistics Dashboard Missing
- **Problem**: `/api/stats` endpoint exists but no UI to display it
  - Stats include: venue count, event count, category distribution, bike routes
  - Plan mentions analytics but no visualization
- **Files**: 
  - `internal/api/handlers.go:448-483` - GetStats() implemented
  - No corresponding frontend component
- **Priority**: Low

### 20. Image Support Incomplete
- **Problem**: Events and exhibitions have `imageUrl` field but no image display
  - No images shown in popups, modals, or cards
  - Empty `web/static/images/` directory
  - No fallback images for missing content
- **Files**: 
  - `internal/storage/models.go` - ImageURL field exists
  - `web/static/images/` - empty directory
  - UI templates don't show images
- **Priority**: Medium

### 21. No Category Color Coding
- **Problem**: Categories have colors in API but not used consistently
  - `/api/categories` returns colors (line 432 in handlers.go)
  - Event cards have category badges but no color customization
- **Files**: 
  - `internal/api/handlers.go:432` - color assignment
  - `web/static/css/styles.css:419-426` - .category class static color
- **Priority**: Low

## Performance Issues

### 22. No Marker Clustering Configuration
- **Problem**: Marker clustering enabled but hardcoded settings
  - `maxClusterRadius: 50` might not be optimal for different zoom levels
  - No configuration options exposed
- **Files**: `web/static/js/map.js:15-20`
- **Priority**: Low

### 23. All Data Loaded on Startup
- **Problem**: App loads all venues, events, exhibitions on page load
  - No pagination or lazy loading
  - Could be slow with hundreds of items
  - `loadData()` fetches everything at once (app.js:62-96)
- **Files**: `web/static/js/app.js:62-96`
- **Priority**: Medium

### 24. No Request Caching
- **Problem**: Every venue/event detail click triggers new API call
  - No in-memory cache for previously loaded details
  - `showVenueDetails()` always fetches (line 262 app.js)
- **Files**: `web/static/js/app.js:262-325, 328-371`
- **Priority**: Low

## Mobile/Responsive Issues

### 25. Map Height Issues on Mobile
- **Problem**: Fixed height breakpoints may not work well on all devices
  - `.main` uses `calc(100vh - 80px)` which might not account for mobile browser chrome
  - `.content { height: 60vh }` on mobile could be too small
- **Files**: `web/static/css/styles.css:96, 543`
- **Priority**: Medium

### 26. Touch Gestures for Modal Close
- **Problem**: Modal only closes with click, no swipe gesture
  - Mobile users might expect swipe-to-dismiss
- **Files**: `web/static/js/app.js:374-382`
- **Priority**: Low

### 27. Small Touch Targets
- **Problem**: Map control buttons may be too small for touch
  - 40x40px buttons (line 278 styles.css)
  - iOS recommendations suggest 44x44px minimum
- **Files**: `web/static/css/styles.css:276-289`
- **Priority**: Low

## Accessibility Issues

### 28. Missing ARIA Labels
- **Problem**: Interactive elements lack accessibility labels
  - Map control buttons have no aria-label
  - Modal close button has no aria-label
  - Filter controls missing labels for screen readers
- **Files**: 
  - `web/templates/index.html` - throughout
  - `web/static/css/styles.css` - decorative elements not hidden
- **Priority**: Medium

### 29. Keyboard Navigation Not Implemented
- **Problem**: Map and interactive elements not keyboard accessible
  - Can't navigate markers with keyboard
  - Modal close requires mouse
  - No focus management when opening/closing modal
- **Files**: All interactive components
- **Priority**: Medium

### 30. Color Contrast Issues
- **Problem**: Some color combinations may not meet WCAG standards
  - `var(--accent-color): #FFA07A` on white background
  - Badge text colors need verification
- **Files**: `web/static/css/styles.css:8-17`
- **Priority**: Low

## Documentation Issues

### 31. No Frontend Documentation
- **Problem**: JavaScript files have no JSDoc comments
  - No function parameter documentation
  - No return type information
  - No usage examples
- **Files**: All JS files
- **Priority**: Low

### 32. No API Documentation
- **Problem**: API endpoints not documented
  - No OpenAPI/Swagger spec
  - No example requests/responses
  - README doesn't list available endpoints
- **Files**: `internal/api/handlers.go`, `README.md`
- **Priority**: Low

### 33. Environment Configuration Undocumented
- **Problem**: No documentation for configuration options
  - Port, data directory, static directory flags exist but not documented
  - No example .env file
- **Files**: `cmd/server/main.go:17-21`
- **Priority**: Low

## DevOps/Build Issues

### 34. No Build Process
- **Problem**: No asset minification or bundling
  - CSS and JS served unminified
  - No cache busting for static assets
  - No build script mentioned
- **Files**: Project structure
- **Priority**: Low

### 35. ~~Go Module Warnings~~ ✅ COMPLETED (2026-01-29)
- **Problem**: go.mod has unnecessary dependencies
  - ~~`github.com/rs/cors` imported but not used~~ ✅ Removed
  - ~~`github.com/PuerkitoBio/goquery` and `github.com/gorilla/mux` should be direct~~ ✅ Fixed
- **Solution**: Ran `go mod tidy` to clean up dependencies
- **Result**: All warnings resolved, go.mod now properly structured
- **Files**: `go.mod`, `go.sum`
- **Priority**: ~~Low~~ DONE

---

## Task Prioritization Summary

### ✅ Completed (2026-01-29)
1. ~~Missing API Response Fields (#1)~~ - venueName now populated in all endpoints
6. ~~Search Results Missing venueName (#6)~~ - Search now includes venue names
18. ~~No Input Validation (#18)~~ - Comprehensive validation added
35. ~~Go Module Warnings (#35)~~ - Dependencies cleaned up

### High Priority (Do First)
_None remaining - all high-priority tasks completed!_

### Medium Priority (Do Next)
2. Missing Exhibition Details (#2)
3. Error Handling (#3)
5. Missing Venue Details Enrichment (#5)
11. Route Planning Integration (#11)
20. Image Support (#20)
23. All Data Loaded on Startup (#23)
25. Map Height on Mobile (#25)
28. Missing ARIA Labels (#28)
29. Keyboard Navigation (#29)

### Low Priority (Nice to Have)
All remaining tasks (4, 7-10, 12-17, 19, 21-22, 24, 26-27, 30-35)

---

## Notes for Implementation

- ✅ Tasks #1 and #6 completed together (API response enrichment)
- Task #5 (venue details enrichment) partially done - counts already exist, full objects could be added
- Tasks #28 and #29 can be tackled together as accessibility improvements
- ✅ Task #35 completed - `go mod tidy` resolved all warnings
- Consider implementing #23 (pagination/lazy loading) if dataset grows beyond 200+ venues

## Recent Changes (2026-01-29)

**Commit: Add input validation and API response enrichment**
- Created `internal/api/validation.go` with XSS protection
- All events and exhibitions now include `venueName` field
- Added validation for dates, IDs, search queries, and search types
- Fixed go.mod dependency warnings
- All tests passing (82.8% coverage)
