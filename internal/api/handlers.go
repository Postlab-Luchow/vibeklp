package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/musche/klp/internal/storage"
)

// Handler handles API requests
type Handler struct {
	storage *storage.Storage
}

// NewHandler creates a new Handler
func NewHandler(store *storage.Storage) *Handler {
	return &Handler{
		storage: store,
	}
}

// respondJSON sends a JSON response
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    status,
			"message": message,
		},
	})
}

// GetVenues returns all venues
func (h *Handler) GetVenues(w http.ResponseWriter, r *http.Request) {
	venues, err := h.storage.LoadVenues()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to load venues")
		return
	}

	// Apply filters
	search := GetQueryParam(r, "search")
	amenity := GetQueryParam(r, "amenity")

	var filtered []storage.Venue
	for _, v := range venues {
		// Search filter
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(v.Name), searchLower) &&
				!strings.Contains(strings.ToLower(v.Description), searchLower) {
				continue
			}
		}

		// Amenity filter
		if amenity != "" {
			found := false
			for _, a := range v.Amenities {
				if strings.EqualFold(a, amenity) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		filtered = append(filtered, v)
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"venues": filtered,
		"total":  len(filtered),
	})
}

// GetVenue returns a single venue by ID
func (h *Handler) GetVenue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	venue, err := h.storage.GetVenueWithDetails(id)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Venue not found")
		return
	}

	h.respondJSON(w, http.StatusOK, venue)
}

// GetEvents returns all events
func (h *Handler) GetEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.storage.LoadEvents()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to load events")
		return
	}

	// Enrich events with venue names
	venues, _ := h.storage.LoadVenues()
	venueMap := make(map[string]string)
	for _, v := range venues {
		venueMap[v.ID] = v.Name
	}
	for i := range events {
		if venueName, ok := venueMap[events[i].VenueID]; ok {
			events[i].VenueName = venueName
		}
	}

	// Apply filters
	date := GetQueryParamRaw(r, "date")
	dateFrom := GetQueryParamRaw(r, "dateFrom")
	dateTo := GetQueryParamRaw(r, "dateTo")
	category := GetQueryParam(r, "category")
	venueID := GetQueryParamRaw(r, "venueId")
	search := GetQueryParam(r, "search")

	// Validate date formats
	if !ValidateDate(date) || !ValidateDate(dateFrom) || !ValidateDate(dateTo) {
		h.respondError(w, http.StatusBadRequest, "Invalid date format (expected YYYY-MM-DD)")
		return
	}
	if venueID != "" && !ValidateID(venueID) {
		h.respondError(w, http.StatusBadRequest, "Invalid venue ID")
		return
	}

	var filtered []storage.Event
	for _, e := range events {
		// Date filter
		if date != "" && e.Date != date {
			continue
		}

		// Date range filter
		if dateFrom != "" && e.Date < dateFrom {
			continue
		}
		if dateTo != "" && e.Date > dateTo {
			continue
		}

		// Category filter
		if category != "" && !strings.EqualFold(e.Category, category) {
			continue
		}

		// Venue filter
		if venueID != "" && e.VenueID != venueID {
			continue
		}

		// Search filter
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(e.Title), searchLower) &&
				!strings.Contains(strings.ToLower(e.Description), searchLower) {
				continue
			}
		}

		filtered = append(filtered, e)
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"events": filtered,
		"total":  len(filtered),
	})
}

// GetEvent returns a single event by ID
func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	event, err := h.storage.GetEventByID(id)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Get venue details
	venue, _ := h.storage.GetVenueByID(event.VenueID)
	if venue != nil {
		event.VenueName = venue.Name
	}

	response := map[string]interface{}{
		"id":          event.ID,
		"title":       event.Title,
		"description": event.Description,
		"venueId":     event.VenueID,
		"venueName":   event.VenueName,
		"date":        event.Date,
		"startTime":   event.StartTime,
		"endTime":     event.EndTime,
		"category":    event.Category,
		"admission":   event.Admission,
		"imageUrl":    event.ImageURL,
	}

	if venue != nil {
		response["venue"] = map[string]interface{}{
			"id":          venue.ID,
			"name":        venue.Name,
			"address":     venue.Address,
			"coordinates": venue.Coordinates,
		}
	}

	h.respondJSON(w, http.StatusOK, response)
}

// GetExhibitions returns all exhibitions
func (h *Handler) GetExhibitions(w http.ResponseWriter, r *http.Request) {
	exhibitions, err := h.storage.LoadExhibitions()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to load exhibitions")
		return
	}

	// Enrich exhibitions with venue names
	venues, _ := h.storage.LoadVenues()
	venueMap := make(map[string]string)
	for _, v := range venues {
		venueMap[v.ID] = v.Name
	}
	for i := range exhibitions {
		if venueName, ok := venueMap[exhibitions[i].VenueID]; ok {
			exhibitions[i].VenueName = venueName
		}
	}

	// Apply filters
	category := GetQueryParam(r, "category")
	venueID := GetQueryParamRaw(r, "venueId")
	search := GetQueryParam(r, "search")

	// Validate venue ID
	if venueID != "" && !ValidateID(venueID) {
		h.respondError(w, http.StatusBadRequest, "Invalid venue ID")
		return
	}

	var filtered []storage.Exhibition
	for _, ex := range exhibitions {
		// Category filter
		if category != "" && !strings.EqualFold(ex.Category, category) {
			continue
		}

		// Venue filter
		if venueID != "" && ex.VenueID != venueID {
			continue
		}

		// Search filter
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(ex.Title), searchLower) &&
				!strings.Contains(strings.ToLower(ex.Description), searchLower) &&
				!strings.Contains(strings.ToLower(ex.Artist), searchLower) {
				continue
			}
		}

		filtered = append(filtered, ex)
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"exhibitions": filtered,
		"total":       len(filtered),
	})
}

// GetExhibition returns a single exhibition by ID
func (h *Handler) GetExhibition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	exhibition, err := h.storage.GetExhibitionByID(id)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Exhibition not found")
		return
	}

	// Get venue details
	venue, _ := h.storage.GetVenueByID(exhibition.VenueID)
	if venue != nil {
		exhibition.VenueName = venue.Name
	}

	response := map[string]interface{}{
		"id":          exhibition.ID,
		"title":       exhibition.Title,
		"description": exhibition.Description,
		"venueId":     exhibition.VenueID,
		"venueName":   exhibition.VenueName,
		"artist":      exhibition.Artist,
		"category":    exhibition.Category,
		"imageUrl":    exhibition.ImageURL,
	}

	if venue != nil {
		response["venue"] = map[string]interface{}{
			"id":          venue.ID,
			"name":        venue.Name,
			"address":     venue.Address,
			"coordinates": venue.Coordinates,
		}
	}

	h.respondJSON(w, http.StatusOK, response)
}

// Search performs a global search
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := GetQueryParamRaw(r, "q")
	query, valid := ValidateSearchQuery(query)
	if !valid {
		h.respondError(w, http.StatusBadRequest, "Query parameter 'q' is required (min 2 characters)")
		return
	}

	searchType := GetQueryParamRaw(r, "type")
	if !ValidateSearchType(searchType) {
		h.respondError(w, http.StatusBadRequest, "Invalid search type (must be: venues, events, exhibitions)")
		return
	}
	queryLower := strings.ToLower(query)

	results := map[string]interface{}{
		"venues":      []storage.SearchResult{},
		"events":      []storage.SearchResult{},
		"exhibitions": []storage.SearchResult{},
	}

	// Search venues
	if searchType == "" || searchType == "venues" {
		venues, _ := h.storage.LoadVenues()
		var venueResults []storage.SearchResult
		for _, v := range venues {
			if strings.Contains(strings.ToLower(v.Name), queryLower) ||
				strings.Contains(strings.ToLower(v.Description), queryLower) {
				venueResults = append(venueResults, storage.SearchResult{
					ID:       v.ID,
					Type:     "venue",
					Title:    v.Name,
					Subtitle: v.Address.City,
				})
			}
		}
		results["venues"] = venueResults
	}

	// Search events
	if searchType == "" || searchType == "events" {
		events, _ := h.storage.LoadEvents()
		// Enrich with venue names
		venues, _ := h.storage.LoadVenues()
		venueMap := make(map[string]string)
		for _, v := range venues {
			venueMap[v.ID] = v.Name
		}
		var eventResults []storage.SearchResult
		for _, e := range events {
			if strings.Contains(strings.ToLower(e.Title), queryLower) ||
				strings.Contains(strings.ToLower(e.Description), queryLower) {
				venueName := venueMap[e.VenueID]
				eventResults = append(eventResults, storage.SearchResult{
					ID:       e.ID,
					Type:     "event",
					Title:    e.Title,
					Subtitle: venueName,
					Date:     e.Date,
				})
			}
		}
		results["events"] = eventResults
	}

	// Search exhibitions
	if searchType == "" || searchType == "exhibitions" {
		exhibitions, _ := h.storage.LoadExhibitions()
		// Enrich with venue names for subtitle fallback
		venues, _ := h.storage.LoadVenues()
		venueMap := make(map[string]string)
		for _, v := range venues {
			venueMap[v.ID] = v.Name
		}
		var exhibitionResults []storage.SearchResult
		for _, ex := range exhibitions {
			if strings.Contains(strings.ToLower(ex.Title), queryLower) ||
				strings.Contains(strings.ToLower(ex.Description), queryLower) ||
				strings.Contains(strings.ToLower(ex.Artist), queryLower) {
				subtitle := ex.Artist
				if subtitle == "" {
					subtitle = venueMap[ex.VenueID]
				}
				exhibitionResults = append(exhibitionResults, storage.SearchResult{
					ID:       ex.ID,
					Type:     "exhibition",
					Title:    ex.Title,
					Subtitle: subtitle,
				})
			}
		}
		results["exhibitions"] = exhibitionResults
	}

	// Calculate total
	total := len(results["venues"].([]storage.SearchResult)) +
		len(results["events"].([]storage.SearchResult)) +
		len(results["exhibitions"].([]storage.SearchResult))

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"total":   total,
	})
}

// GetCalendar returns events grouped by date
func (h *Handler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	events, err := h.storage.LoadEvents()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to load events")
		return
	}

	// Enrich events with venue names
	venues, _ := h.storage.LoadVenues()
	venueMap := make(map[string]string)
	for _, v := range venues {
		venueMap[v.ID] = v.Name
	}
	for i := range events {
		if venueName, ok := venueMap[events[i].VenueID]; ok {
			events[i].VenueName = venueName
		}
	}

	// Group events by date
	calendar := make(map[string]*storage.CalendarDay)
	for _, e := range events {
		if _, exists := calendar[e.Date]; !exists {
			// Parse date to get day of week
			t, _ := time.Parse("2006-01-02", e.Date)
			dayOfWeek := t.Weekday().String()
			if t.Weekday() == time.Monday {
				dayOfWeek = "Montag"
			} else if t.Weekday() == time.Tuesday {
				dayOfWeek = "Dienstag"
			} else if t.Weekday() == time.Wednesday {
				dayOfWeek = "Mittwoch"
			} else if t.Weekday() == time.Thursday {
				dayOfWeek = "Donnerstag"
			} else if t.Weekday() == time.Friday {
				dayOfWeek = "Freitag"
			} else if t.Weekday() == time.Saturday {
				dayOfWeek = "Samstag"
			} else if t.Weekday() == time.Sunday {
				dayOfWeek = "Sonntag"
			}

			calendar[e.Date] = &storage.CalendarDay{
				Date:      e.Date,
				DayOfWeek: dayOfWeek,
				Events:    []storage.Event{},
			}
		}
		calendar[e.Date].Events = append(calendar[e.Date].Events, e)
		calendar[e.Date].EventCount++
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"calendar":    calendar,
		"totalDays":   len(calendar),
		"totalEvents": len(events),
	})
}

// GetCategories returns all categories
func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) {
	events, _ := h.storage.LoadEvents()
	exhibitions, _ := h.storage.LoadExhibitions()

	categoryMap := make(map[string]int)

	for _, e := range events {
		if e.Category != "" {
			categoryMap[e.Category]++
		}
	}

	for _, ex := range exhibitions {
		if ex.Category != "" {
			categoryMap[ex.Category]++
		}
	}

	var categories []storage.Category
	colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#FFA07A", "#98D8C8", "#F7DC6F", "#BB8FCE"}
	i := 0
	for name, count := range categoryMap {
		categories = append(categories, storage.Category{
			Name:  name,
			Count: count,
			Color: colors[i%len(colors)],
		})
		i++
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
	})
}

// GetStats returns statistics
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	venues, _ := h.storage.LoadVenues()
	events, _ := h.storage.LoadEvents()
	exhibitions, _ := h.storage.LoadExhibitions()

	categoryDist := make(map[string]int)
	for _, e := range events {
		if e.Category != "" {
			categoryDist[e.Category]++
		}
	}

	bikeRouteCount := 0
	for _, v := range venues {
		if v.BikeRoute != "" {
			bikeRouteCount++
		}
	}

	stats := storage.Statistics{
		TotalVenues:      len(venues),
		TotalEvents:      len(events),
		TotalExhibitions: len(exhibitions),
		FestivalDates: storage.FestivalDates{
			Start: "2026-05-14",
			End:   "2026-05-25",
		},
		CategoriesDistribution: categoryDist,
		VenuesWithBikeRoutes:   bikeRouteCount,
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"stats": stats,
	})
}
