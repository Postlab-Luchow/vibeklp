package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/musche/klp/internal/storage"
)

func setupTestStorage(t *testing.T) (*storage.Storage, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "klp-api-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	store := storage.NewStorage(tempDir)

	// Create test data
	venues := []storage.Venue{
		{
			ID:   "venue-1",
			Name: "Test Venue 1",
			Address: storage.Address{
				Street:     "Teststraße 1",
				PostalCode: "29439",
				City:       "Lüchow",
			},
			Coordinates: storage.Coordinates{
				Lat: 53.0,
				Lng: 11.2,
			},
			EventCount:      2,
			ExhibitionCount: 1,
			BikeRoute:       "Route A",
		},
		{
			ID:   "venue-2",
			Name: "Test Venue 2",
			Address: storage.Address{
				Street:     "Teststraße 2",
				PostalCode: "29429",
				City:       "Dannenberg",
			},
			Coordinates: storage.Coordinates{
				Lat: 53.1,
				Lng: 11.1,
			},
			EventCount: 1,
		},
	}

	events := []storage.Event{
		{
			ID:        "event-1",
			Title:     "Test Event 1",
			VenueID:   "venue-1",
			VenueName: "Test Venue 1",
			Date:      "2026-05-29",
			StartTime: "17:00",
			Category:  "Music",
			Admission: "Free",
		},
		{
			ID:        "event-2",
			Title:     "Test Event 2",
			VenueID:   "venue-1",
			VenueName: "Test Venue 1",
			Date:      "2026-05-30",
			StartTime: "19:00",
			Category:  "Theater",
		},
		{
			ID:        "event-3",
			Title:     "Test Event 3",
			VenueID:   "venue-2",
			VenueName: "Test Venue 2",
			Date:      "2026-05-29",
			StartTime: "18:00",
			Category:  "Music",
		},
	}

	exhibitions := []storage.Exhibition{
		{
			ID:        "exhibition-1",
			Title:     "Test Exhibition 1",
			VenueID:   "venue-1",
			VenueName: "Test Venue 1",
			Artist:    "Test Artist",
			Category:  "Art",
		},
	}

	store.SaveVenues(venues)
	store.SaveEvents(events)
	store.SaveExhibitions(exhibitions)

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return store, cleanup
}

func TestGetVenues(t *testing.T) {
	store, cleanup := setupTestStorage(t)
	defer cleanup()

	handler := NewHandler(store)

	t.Run("get all venues", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/venues", nil)
		w := httptest.NewRecorder()

		handler.GetVenues(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		venues := response["venues"].([]interface{})
		if len(venues) != 2 {
			t.Errorf("Got %d venues, want 2", len(venues))
		}
	})

	t.Run("search venues", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/venues?search=Venue+1", nil)
		w := httptest.NewRecorder()

		handler.GetVenues(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		venues := response["venues"].([]interface{})
		if len(venues) != 1 {
			t.Errorf("Got %d venues, want 1", len(venues))
		}
	})
}

func TestGetVenue(t *testing.T) {
	store, cleanup := setupTestStorage(t)
	defer cleanup()

	handler := NewHandler(store)
	router := mux.NewRouter()
	router.HandleFunc("/api/venues/{id}", handler.GetVenue)

	t.Run("existing venue", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/venues/venue-1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
		}

		var venue storage.VenueWithDetails
		if err := json.NewDecoder(w.Body).Decode(&venue); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if venue.ID != "venue-1" {
			t.Errorf("Venue ID = %q, want %q", venue.ID, "venue-1")
		}

		if len(venue.Events) != 2 {
			t.Errorf("Got %d events, want 2", len(venue.Events))
		}

		if len(venue.Exhibitions) != 1 {
			t.Errorf("Got %d exhibitions, want 1", len(venue.Exhibitions))
		}
	})

	t.Run("nonexistent venue", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/venues/nonexistent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Status code = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestGetEvents(t *testing.T) {
	store, cleanup := setupTestStorage(t)
	defer cleanup()

	handler := NewHandler(store)

	t.Run("get all events", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/events", nil)
		w := httptest.NewRecorder()

		handler.GetEvents(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		events := response["events"].([]interface{})
		if len(events) != 3 {
			t.Errorf("Got %d events, want 3", len(events))
		}
	})

	t.Run("filter by date", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/events?date=2026-05-29", nil)
		w := httptest.NewRecorder()

		handler.GetEvents(w, req)

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		events := response["events"].([]interface{})
		if len(events) != 2 {
			t.Errorf("Got %d events, want 2", len(events))
		}
	})

	t.Run("filter by category", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/events?category=Music", nil)
		w := httptest.NewRecorder()

		handler.GetEvents(w, req)

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		events := response["events"].([]interface{})
		if len(events) != 2 {
			t.Errorf("Got %d events, want 2 (category=Music)", len(events))
		}
	})

	t.Run("filter by venue", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/events?venueId=venue-1", nil)
		w := httptest.NewRecorder()

		handler.GetEvents(w, req)

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		events := response["events"].([]interface{})
		if len(events) != 2 {
			t.Errorf("Got %d events, want 2 (venueId=venue-1)", len(events))
		}
	})
}

func TestGetEvent(t *testing.T) {
	store, cleanup := setupTestStorage(t)
	defer cleanup()

	handler := NewHandler(store)
	router := mux.NewRouter()
	router.HandleFunc("/api/events/{id}", handler.GetEvent)

	t.Run("existing event", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/events/event-1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["id"] != "event-1" {
			t.Errorf("Event ID = %q, want %q", response["id"], "event-1")
		}

		// Check venue details are included
		if _, ok := response["venue"]; !ok {
			t.Error("Response missing 'venue' field")
		}
	})
}

func TestGetExhibitions(t *testing.T) {
	store, cleanup := setupTestStorage(t)
	defer cleanup()

	handler := NewHandler(store)

	req := httptest.NewRequest("GET", "/api/exhibitions", nil)
	w := httptest.NewRecorder()

	handler.GetExhibitions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	exhibitions := response["exhibitions"].([]interface{})
	if len(exhibitions) != 1 {
		t.Errorf("Got %d exhibitions, want 1", len(exhibitions))
	}
}

func TestSearch(t *testing.T) {
	store, cleanup := setupTestStorage(t)
	defer cleanup()

	handler := NewHandler(store)

	t.Run("missing query parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/search", nil)
		w := httptest.NewRecorder()

		handler.Search(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Status code = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("search all types", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/search?q=Test", nil)
		w := httptest.NewRecorder()

		handler.Search(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		results := response["results"].(map[string]interface{})
		total := response["total"].(float64)

		if total < 4 { // At least 2 venues + 3 events + 1 exhibition match "Test"
			t.Errorf("Total results = %.0f, want at least 4", total)
		}

		if results["venues"] == nil {
			t.Error("Search results missing 'venues'")
		}
		if results["events"] == nil {
			t.Error("Search results missing 'events'")
		}
		if results["exhibitions"] == nil {
			t.Error("Search results missing 'exhibitions'")
		}
	})

	t.Run("search specific type", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/search?q=Event&type=events", nil)
		w := httptest.NewRecorder()

		handler.Search(w, req)

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		results := response["results"].(map[string]interface{})
		events := results["events"].([]interface{})

		if len(events) != 3 {
			t.Errorf("Got %d events, want 3", len(events))
		}
	})
}

func TestGetCalendar(t *testing.T) {
	store, cleanup := setupTestStorage(t)
	defer cleanup()

	handler := NewHandler(store)

	req := httptest.NewRequest("GET", "/api/calendar", nil)
	w := httptest.NewRecorder()

	handler.GetCalendar(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	calendar := response["calendar"].(map[string]interface{})
	if len(calendar) != 2 { // 2 different dates in test data
		t.Errorf("Got %d days in calendar, want 2", len(calendar))
	}

	// Check structure of a calendar day
	for date, dayData := range calendar {
		day := dayData.(map[string]interface{})
		if day["date"] != date {
			t.Errorf("Calendar day date mismatch: %v != %v", day["date"], date)
		}
		if day["dayOfWeek"] == nil {
			t.Error("Calendar day missing dayOfWeek")
		}
		if day["eventCount"] == nil {
			t.Error("Calendar day missing eventCount")
		}
		if day["events"] == nil {
			t.Error("Calendar day missing events")
		}
	}
}

func TestGetCategories(t *testing.T) {
	store, cleanup := setupTestStorage(t)
	defer cleanup()

	handler := NewHandler(store)

	req := httptest.NewRequest("GET", "/api/categories", nil)
	w := httptest.NewRecorder()

	handler.GetCategories(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	categories := response["categories"].([]interface{})
	if len(categories) < 3 { // At least Music, Theater, Art
		t.Errorf("Got %d categories, want at least 3", len(categories))
	}

	// Check category structure
	category := categories[0].(map[string]interface{})
	if category["name"] == nil {
		t.Error("Category missing 'name'")
	}
	if category["count"] == nil {
		t.Error("Category missing 'count'")
	}
	if category["color"] == nil {
		t.Error("Category missing 'color'")
	}
}

func TestGetStats(t *testing.T) {
	store, cleanup := setupTestStorage(t)
	defer cleanup()

	handler := NewHandler(store)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()

	handler.GetStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	stats := response["stats"].(map[string]interface{})

	if stats["totalVenues"].(float64) != 2 {
		t.Errorf("totalVenues = %.0f, want 2", stats["totalVenues"].(float64))
	}
	if stats["totalEvents"].(float64) != 3 {
		t.Errorf("totalEvents = %.0f, want 3", stats["totalEvents"].(float64))
	}
	if stats["totalExhibitions"].(float64) != 1 {
		t.Errorf("totalExhibitions = %.0f, want 1", stats["totalExhibitions"].(float64))
	}
	if stats["venuesWithBikeRoutes"].(float64) != 1 {
		t.Errorf("venuesWithBikeRoutes = %.0f, want 1", stats["venuesWithBikeRoutes"].(float64))
	}
}

func TestSetupRoutes(t *testing.T) {
	store, cleanup := setupTestStorage(t)
	defer cleanup()

	router := SetupRoutes(store)

	// Test that routes are registered with valid IDs
	routes := []struct {
		method       string
		path         string
		expectStatus int // 0 means any non-404
	}{
		{"GET", "/api/venues", 0},
		{"GET", "/api/venues/venue-1", 0}, // Use valid ID
		{"GET", "/api/events", 0},
		{"GET", "/api/events/event-1", 0}, // Use valid ID
		{"GET", "/api/exhibitions", 0},
		{"GET", "/api/exhibitions/exhibition-1", 0}, // Use valid ID
		{"GET", "/api/search?q=test", 0},
		{"GET", "/api/calendar", 0},
		{"GET", "/api/categories", 0},
		{"GET", "/api/stats", 0},
	}

	for _, route := range routes {
		req := httptest.NewRequest(route.method, route.path, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Routes should be registered (not 404)
		// They may return other errors based on the request (400, 500, etc.)
		if w.Code == http.StatusNotFound {
			t.Errorf("Route %s %s returned 404, route may not be registered", route.method, route.path)
		}
	}
}
