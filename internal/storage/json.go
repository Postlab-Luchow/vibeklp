package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Storage handles reading and writing JSON data files
type Storage struct {
	dataDir string
}

// NewStorage creates a new Storage instance
func NewStorage(dataDir string) *Storage {
	return &Storage{
		dataDir: dataDir,
	}
}

// LoadVenues loads all venues from JSON file
func (s *Storage) LoadVenues() ([]Venue, error) {
	var venues []Venue
	err := s.loadJSON("venues.json", &venues)
	return venues, err
}

// SaveVenues saves venues to JSON file
func (s *Storage) SaveVenues(venues []Venue) error {
	return s.saveJSON("venues.json", venues)
}

// LoadEvents loads all events from JSON file
func (s *Storage) LoadEvents() ([]Event, error) {
	var events []Event
	err := s.loadJSON("events.json", &events)
	return events, err
}

// SaveEvents saves events to JSON file
func (s *Storage) SaveEvents(events []Event) error {
	return s.saveJSON("events.json", events)
}

// LoadExhibitions loads all exhibitions from JSON file
func (s *Storage) LoadExhibitions() ([]Exhibition, error) {
	var exhibitions []Exhibition
	err := s.loadJSON("exhibitions.json", &exhibitions)
	return exhibitions, err
}

// SaveExhibitions saves exhibitions to JSON file
func (s *Storage) SaveExhibitions(exhibitions []Exhibition) error {
	return s.saveJSON("exhibitions.json", exhibitions)
}

// GetVenueByID finds a venue by ID
func (s *Storage) GetVenueByID(id string) (*Venue, error) {
	venues, err := s.LoadVenues()
	if err != nil {
		return nil, err
	}

	for _, v := range venues {
		if v.ID == id {
			return &v, nil
		}
	}

	return nil, fmt.Errorf("venue not found: %s", id)
}

// GetEventByID finds an event by ID
func (s *Storage) GetEventByID(id string) (*Event, error) {
	events, err := s.LoadEvents()
	if err != nil {
		return nil, err
	}

	for _, e := range events {
		if e.ID == id {
			return &e, nil
		}
	}

	return nil, fmt.Errorf("event not found: %s", id)
}

// GetExhibitionByID finds an exhibition by ID
func (s *Storage) GetExhibitionByID(id string) (*Exhibition, error) {
	exhibitions, err := s.LoadExhibitions()
	if err != nil {
		return nil, err
	}

	for _, ex := range exhibitions {
		if ex.ID == id {
			return &ex, nil
		}
	}

	return nil, fmt.Errorf("exhibition not found: %s", id)
}

// GetVenueWithDetails returns a venue with its events and exhibitions
func (s *Storage) GetVenueWithDetails(id string) (*VenueWithDetails, error) {
	venue, err := s.GetVenueByID(id)
	if err != nil {
		return nil, err
	}

	events, err := s.LoadEvents()
	if err != nil {
		return nil, err
	}

	exhibitions, err := s.LoadExhibitions()
	if err != nil {
		return nil, err
	}

	// Filter events and exhibitions for this venue
	var venueEvents []Event
	var venueExhibitions []Exhibition

	for _, e := range events {
		if e.VenueID == id {
			venueEvents = append(venueEvents, e)
		}
	}

	for _, ex := range exhibitions {
		if ex.VenueID == id {
			venueExhibitions = append(venueExhibitions, ex)
		}
	}

	return &VenueWithDetails{
		Venue:       *venue,
		Events:      venueEvents,
		Exhibitions: venueExhibitions,
	}, nil
}

// loadJSON loads data from a JSON file
func (s *Storage) loadJSON(filename string, v interface{}) error {
	path := filepath.Join(s.dataDir, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filename, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to parse %s: %w", filename, err)
	}

	return nil
}

// saveJSON saves data to a JSON file
func (s *Storage) saveJSON(filename string, v interface{}) error {
	path := filepath.Join(s.dataDir, filename)

	// Ensure directory exists
	if err := os.MkdirAll(s.dataDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", filename, err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}

	return nil
}

// EnsureDataDir creates the data directory if it doesn't exist
func (s *Storage) EnsureDataDir() error {
	return os.MkdirAll(s.dataDir, 0o755)
}
