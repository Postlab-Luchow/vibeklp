package crawler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/musche/klp/internal/storage"
)

const wendlandpartieEventsURL = "https://wendlandpartie.de/wp-json/tribe/events/v1/events"

// tribeEvent is the subset of fields we consume from the Tribe Events REST API.
type tribeEvent struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Excerpt     string `json:"excerpt"`
	URL         string `json:"url"`
	StartDate   string `json:"start_date"` // "2026-05-13 18:00:00"
	EndDate     string `json:"end_date"`
	AllDay      bool   `json:"all_day"`
	Cost        string `json:"cost"`
	// Image is a struct when present, or `false` when absent.
	Image json.RawMessage `json:"image"`
	// Venue is a single venue object normally, but can be an array when the
	// event is assigned to multiple venues. We unmarshal lazily.
	VenueRaw   json.RawMessage `json:"venue,omitempty"`
	Categories []tribeTaxonomy `json:"categories,omitempty"`
	Tags       []tribeTaxonomy `json:"tags,omitempty"`
}

// venue returns the first (or only) venue from the polymorphic venue field, or
// nil when no venue is associated with the event.
func (e tribeEvent) venue() *tribeVenue {
	raw := bytes.TrimSpace(e.VenueRaw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("false")) || bytes.Equal(raw, []byte("null")) || bytes.Equal(raw, []byte("[]")) {
		return nil
	}
	if raw[0] == '[' {
		var arr []tribeVenue
		if err := json.Unmarshal(raw, &arr); err != nil || len(arr) == 0 {
			return nil
		}
		return &arr[0]
	}
	var v tribeVenue
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil
	}
	return &v
}

// imageURL returns the URL inside the polymorphic image field, or "" if the
// field is missing/empty/false.
func (e tribeEvent) imageURL() string {
	if len(e.Image) == 0 || string(e.Image) == "false" || string(e.Image) == "null" {
		return ""
	}
	var img struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(e.Image, &img); err != nil {
		return ""
	}
	return img.URL
}

type tribeVenue struct {
	ID      int64   `json:"id"`
	Venue   string  `json:"venue"`
	Address string  `json:"address"`
	City    string  `json:"city"`
	ZIP     string  `json:"zip"`
	Phone   string  `json:"phone"`
	Website string  `json:"website"`
	GeoLat  float64 `json:"geo_lat"`
	GeoLng  float64 `json:"geo_lng"`
}

type tribeTaxonomy struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type tribeEventsPage struct {
	Events     []tribeEvent `json:"events"`
	TotalPages int          `json:"total_pages"`
	Total      int          `json:"total"`
}

// CrawlWendlandpartie pulls all events + venues from the wendlandpartie.de
// Tribe Events Calendar REST API. The site exposes complete venue records
// (with lat/lng), so no geocoding is needed.
func CrawlWendlandpartie(logger *log.Logger) ([]storage.Venue, []storage.Event, []storage.Exhibition, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	const perPage = 50
	page := 1
	venuesByTribeID := make(map[int64]*storage.Venue)
	var events []storage.Event
	var exhibitions []storage.Exhibition

	for {
		url := fmt.Sprintf("%s?per_page=%d&page=%d&status=publish", wendlandpartieEventsURL, perPage, page)
		logger.Printf("[wendlandpartie] GET %s", url)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("User-Agent", "KLP-Crawler/1.0 (+https://github.com/musche/klp)")

		resp, err := client.Do(req)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("fetch page %d: %w", page, err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("read page %d: %w", page, err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, nil, nil, fmt.Errorf("page %d status %d: %s", page, resp.StatusCode, string(body[:min(len(body), 200)]))
		}

		var data tribeEventsPage
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, nil, nil, fmt.Errorf("decode page %d: %w", page, err)
		}

		for _, e := range data.Events {
			tv := e.venue()
			if tv == nil || tv.Venue == "" {
				logger.Printf("[wendlandpartie] skip event %d (%q): no venue", e.ID, e.Title)
				continue
			}
			venue := buildWendlandpartieVenue(tv)
			if _, exists := venuesByTribeID[tv.ID]; !exists {
				venuesByTribeID[tv.ID] = &venue
			}
			// All-day, multi-day items on this site are standing exhibitions
			// (galleries, ongoing displays) — map them to Exhibition. Anything
			// with concrete start/end times is a regular Event.
			if e.AllDay {
				exhibitions = append(exhibitions, buildWendlandpartieExhibition(e, venuesByTribeID[tv.ID].ID))
			} else {
				events = append(events, buildWendlandpartieEvent(e, venuesByTribeID[tv.ID].ID))
			}
		}

		logger.Printf("[wendlandpartie] page %d: %d items parsed (events so far %d, exhibitions so far %d, totalPages=%d)",
			page, len(data.Events), len(events), len(exhibitions), data.TotalPages)

		if page >= data.TotalPages || len(data.Events) == 0 {
			break
		}
		page++
		// Be polite even though it's our own JSON.
		time.Sleep(500 * time.Millisecond)
	}

	// Materialize venues slice; backfill event + exhibition counts.
	venues := make([]storage.Venue, 0, len(venuesByTribeID))
	venueEventCount := make(map[string]int, len(venuesByTribeID))
	venueExhibitionCount := make(map[string]int, len(venuesByTribeID))
	for _, ev := range events {
		venueEventCount[ev.VenueID]++
	}
	for _, ex := range exhibitions {
		venueExhibitionCount[ex.VenueID]++
	}
	for _, v := range venuesByTribeID {
		v.EventCount = venueEventCount[v.ID]
		v.ExhibitionCount = venueExhibitionCount[v.ID]
		venues = append(venues, *v)
	}
	return venues, events, exhibitions, nil
}

func buildWendlandpartieVenue(v *tribeVenue) storage.Venue {
	name := strings.TrimSpace(html.UnescapeString(v.Venue))
	street := strings.TrimSpace(html.UnescapeString(v.Address))
	city := strings.TrimSpace(html.UnescapeString(v.City))

	return storage.Venue{
		ID:   fmt.Sprintf("venue-wp-%d", v.ID),
		Name: name,
		Address: storage.Address{
			Street:     street,
			PostalCode: strings.TrimSpace(v.ZIP),
			City:       city,
		},
		Coordinates: storage.Coordinates{Lat: v.GeoLat, Lng: v.GeoLng},
		Contact: storage.Contact{
			Phone:   strings.TrimSpace(v.Phone),
			Website: strings.TrimSpace(v.Website),
		},
		Source: storage.SourceWendlandpartie,
	}
}

func buildWendlandpartieExhibition(e tribeEvent, venueID string) storage.Exhibition {
	desc := strings.TrimSpace(stripHTML(html.UnescapeString(e.Description)))
	if desc == "" {
		desc = strings.TrimSpace(stripHTML(html.UnescapeString(e.Excerpt)))
	}
	return storage.Exhibition{
		ID:          fmt.Sprintf("exhibition-wp-%d", e.ID),
		Title:       strings.TrimSpace(html.UnescapeString(e.Title)),
		Description: desc,
		VenueID:     venueID,
		ImageURL:    e.imageURL(),
		Source:      storage.SourceWendlandpartie,
	}
}

func buildWendlandpartieEvent(e tribeEvent, venueID string) storage.Event {
	// Tribe returns "YYYY-MM-DD HH:MM:SS" in local time.
	date, startTime, endTime := splitTribeDateTime(e.StartDate, e.EndDate)

	desc := strings.TrimSpace(stripHTML(html.UnescapeString(e.Description)))
	if desc == "" {
		desc = strings.TrimSpace(stripHTML(html.UnescapeString(e.Excerpt)))
	}

	id := fmt.Sprintf("event-wp-%d", e.ID)
	return storage.Event{
		ID:          id,
		Title:       strings.TrimSpace(html.UnescapeString(e.Title)),
		Description: desc,
		VenueID:     venueID,
		Date:        date,
		StartTime:   startTime,
		EndTime:     endTime,
		Admission:   strings.TrimSpace(e.Cost),
		ImageURL:    e.imageURL(),
		Source:      storage.SourceWendlandpartie,
	}
}

// splitTribeDateTime turns "2026-05-13 18:00:00" into ("2026-05-13", "18:00", "")
// and includes endTime only when the event ends on the same day.
func splitTribeDateTime(start, end string) (date, startTime, endTime string) {
	startParts := strings.SplitN(start, " ", 2)
	if len(startParts) == 2 {
		date = startParts[0]
		startTime = trimSeconds(startParts[1])
	} else {
		date = start
	}
	if end != "" {
		endParts := strings.SplitN(end, " ", 2)
		if len(endParts) == 2 && endParts[0] == date {
			endTime = trimSeconds(endParts[1])
		}
	}
	return
}

func trimSeconds(t string) string {
	// "18:00:00" -> "18:00"
	parts := strings.Split(t, ":")
	if len(parts) >= 2 {
		return parts[0] + ":" + parts[1]
	}
	return t
}

// stripHTML strips tags using a very simple state machine — Tribe descriptions
// contain WordPress-emitted <p>/<br> markup we don't need.
func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
			b.WriteRune(' ')
		case !inTag:
			b.WriteRune(r)
		}
	}
	// collapse whitespace
	return strings.Join(strings.Fields(b.String()), " ")
}
