package storage

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// Address represents a physical address
type Address struct {
	Street     string `json:"street"`
	PostalCode string `json:"postalCode"`
	City       string `json:"city"`
}

// String returns the full address as a string
func (a Address) String() string {
	return fmt.Sprintf("%s, %s %s", a.Street, a.PostalCode, a.City)
}

// Coordinates represents geographic coordinates
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Contact represents contact information
type Contact struct {
	Phone   string `json:"phone,omitempty"`
	Email   string `json:"email,omitempty"`
	Website string `json:"website,omitempty"`
}

// VenueCategory is a venue-level facility/offering (Café, WC, Kinder, …).
// Some categories are only available on specific dates; an empty Dates slice
// means the category is offered throughout the festival.
type VenueCategory struct {
	Name  string   `json:"name"`
	Dates []string `json:"dates,omitempty"` // YYYY-MM-DD; empty = always available
}

// Source identifies which website a record originated from.
const (
	SourceKLP            = "klp"            // kulturelle-landpartie.de
	SourceWendlandpartie = "wendlandpartie" // wendlandpartie.de
	SourceLandgang       = "landgang"       // landgang-wendland.de
)

// Venue represents a venue/location for the festival
type Venue struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description,omitempty"`
	Address         Address         `json:"address"`
	Coordinates     Coordinates     `json:"coordinates"`
	Contact         Contact         `json:"contact"`
	Amenities       []string        `json:"amenities,omitempty"`
	Categories      []VenueCategory `json:"categories,omitempty"`
	BikeRoute       string          `json:"bikeRoute,omitempty"`
	Source          string          `json:"source,omitempty"`
	EventIDs        []string        `json:"eventIds,omitempty"`
	ExhibitionIDs   []string        `json:"exhibitionIds,omitempty"`
	EventCount      int             `json:"eventCount"`
	ExhibitionCount int             `json:"exhibitionCount"`
}

// Validate checks if the venue data is valid
func (v *Venue) Validate() error {
	if v.Name == "" {
		return errors.New("venue name is required")
	}
	// Postal code is not strictly required anymore (Google Maps can geocode without it)
	// But we still need at least city or street
	if v.Address.City == "" && v.Address.Street == "" {
		return errors.New("address (city or street) is required")
	}
	if v.Coordinates.Lat == 0 || v.Coordinates.Lng == 0 {
		return errors.New("coordinates are required")
	}
	// Plausibility check for Wendland region
	if v.Coordinates.Lat < 52.5 || v.Coordinates.Lat > 53.5 {
		return fmt.Errorf("latitude %.6f out of expected range (52.5-53.5)", v.Coordinates.Lat)
	}
	if v.Coordinates.Lng < 10.5 || v.Coordinates.Lng > 12.0 {
		return fmt.Errorf("longitude %.6f out of expected range (10.5-12.0)", v.Coordinates.Lng)
	}
	return nil
}

// Event represents an event at the festival
type Event struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	VenueID     string `json:"venueId"`
	VenueName   string `json:"venueName,omitempty"`
	Date        string `json:"date"` // ISO 8601: YYYY-MM-DD
	StartTime   string `json:"startTime,omitempty"`
	EndTime     string `json:"endTime,omitempty"`
	Artist      string `json:"artist,omitempty"`
	Category    string `json:"category,omitempty"` // one of EventCategories (or empty if uncategorized)
	Admission   string `json:"admission,omitempty"`
	ImageURL    string `json:"imageUrl,omitempty"`
	Source      string `json:"source,omitempty"`
}

// Validate checks if the event data is valid
func (e *Event) Validate() error {
	if e.Title == "" {
		return errors.New("event title is required")
	}
	if e.VenueID == "" {
		return errors.New("venue ID is required")
	}
	if e.Date == "" {
		return errors.New("date is required")
	}
	// Validate date format
	_, err := time.Parse("2006-01-02", e.Date)
	if err != nil {
		return fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}
	return nil
}

// Exhibition represents an exhibition at a venue
type Exhibition struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	VenueID     string `json:"venueId"`
	VenueName   string `json:"venueName,omitempty"`
	Artist      string `json:"artist,omitempty"`
	Category    string `json:"category,omitempty"`
	ImageURL    string `json:"imageUrl,omitempty"`
	Source      string `json:"source,omitempty"`
}

// Validate checks if the exhibition data is valid
func (ex *Exhibition) Validate() error {
	if ex.Title == "" {
		return errors.New("exhibition title is required")
	}
	if ex.VenueID == "" {
		return errors.New("venue ID is required")
	}
	return nil
}

// VenueWithDetails represents a venue with its events and exhibitions
type VenueWithDetails struct {
	Venue
	Events      []Event      `json:"events,omitempty"`
	Exhibitions []Exhibition `json:"exhibitions,omitempty"`
}

// SearchResult represents a search result item
type SearchResult struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "venue", "event", "exhibition"
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
	Date     string `json:"date,omitempty"`
}

// CalendarDay represents events for a specific day
type CalendarDay struct {
	Date       string  `json:"date"`
	DayOfWeek  string  `json:"dayOfWeek"`
	EventCount int     `json:"eventCount"`
	Events     []Event `json:"events"`
}

// Category represents an event/exhibition category
type Category struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Color string `json:"color,omitempty"`
}

// Statistics represents overall festival statistics
type Statistics struct {
	TotalVenues            int            `json:"totalVenues"`
	TotalEvents            int            `json:"totalEvents"`
	TotalExhibitions       int            `json:"totalExhibitions"`
	FestivalDates          FestivalDates  `json:"festivalDates"`
	CategoriesDistribution map[string]int `json:"categoriesDistribution"`
	VenuesWithBikeRoutes   int            `json:"venuesWithBikeRoutes"`
}

// FestivalDates represents the festival date range
type FestivalDates struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// EventCategories is the closed taxonomy of event/exhibition categories.
// Order is also the display order in the UI. Keep stable — the
// categorizer LLM is constrained to this list.
var EventCategories = []string{
	"Musik",
	"Theater & Performance",
	"Wort & Vortrag",
	"Kunst & Workshop",
	"Tanz & Bewegung",
	"Film",
	"Kulinarisches",
	"Kinder & Familie",
	"Sonstiges",
}

// IsValidEventCategory reports whether s is one of the known categories.
func IsValidEventCategory(s string) bool {
	return slices.Contains(EventCategories, s)
}
