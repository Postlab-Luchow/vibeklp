package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStorage_SaveAndLoad(t *testing.T) {
	// Create temporary directory for test data
	tempDir, err := os.MkdirTemp("", "klp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := NewStorage(tempDir)

	t.Run("venues", func(t *testing.T) {
		venues := []Venue{
			{
				ID:   "venue-1",
				Name: "Test Venue 1",
				Address: Address{
					Street:     "Teststraße 1",
					PostalCode: "29439",
					City:       "Lüchow",
				},
				Coordinates: Coordinates{
					Lat: 53.0,
					Lng: 11.2,
				},
				EventCount:      3,
				ExhibitionCount: 2,
			},
			{
				ID:   "venue-2",
				Name: "Test Venue 2",
				Address: Address{
					Street:     "Teststraße 2",
					PostalCode: "29429",
					City:       "Dannenberg",
				},
				Coordinates: Coordinates{
					Lat: 53.1,
					Lng: 11.1,
				},
				EventCount:      1,
				ExhibitionCount: 0,
			},
		}

		// Save
		if err := store.SaveVenues(venues); err != nil {
			t.Fatalf("SaveVenues() error = %v", err)
		}

		// Load
		loaded, err := store.LoadVenues()
		if err != nil {
			t.Fatalf("LoadVenues() error = %v", err)
		}

		// Verify
		if len(loaded) != len(venues) {
			t.Errorf("LoadVenues() returned %d venues, want %d", len(loaded), len(venues))
		}

		for i, v := range venues {
			if loaded[i].ID != v.ID {
				t.Errorf("Venue %d: ID = %q, want %q", i, loaded[i].ID, v.ID)
			}
			if loaded[i].Name != v.Name {
				t.Errorf("Venue %d: Name = %q, want %q", i, loaded[i].Name, v.Name)
			}
		}
	})

	t.Run("events", func(t *testing.T) {
		events := []Event{
			{
				ID:        "event-1",
				Title:     "Test Event 1",
				VenueID:   "venue-1",
				VenueName: "Test Venue 1",
				Date:      "2026-05-29",
				StartTime: "17:00",
				Category:  "Music",
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
		}

		// Save
		if err := store.SaveEvents(events); err != nil {
			t.Fatalf("SaveEvents() error = %v", err)
		}

		// Load
		loaded, err := store.LoadEvents()
		if err != nil {
			t.Fatalf("LoadEvents() error = %v", err)
		}

		// Verify
		if len(loaded) != len(events) {
			t.Errorf("LoadEvents() returned %d events, want %d", len(loaded), len(events))
		}
	})

	t.Run("exhibitions", func(t *testing.T) {
		exhibitions := []Exhibition{
			{
				ID:        "exhibition-1",
				Title:     "Test Exhibition 1",
				VenueID:   "venue-1",
				VenueName: "Test Venue 1",
				Artist:    "Test Artist",
			},
		}

		// Save
		if err := store.SaveExhibitions(exhibitions); err != nil {
			t.Fatalf("SaveExhibitions() error = %v", err)
		}

		// Load
		loaded, err := store.LoadExhibitions()
		if err != nil {
			t.Fatalf("LoadExhibitions() error = %v", err)
		}

		// Verify
		if len(loaded) != len(exhibitions) {
			t.Errorf("LoadExhibitions() returned %d exhibitions, want %d", len(loaded), len(exhibitions))
		}
	})
}

func TestStorage_GetByID(t *testing.T) {
	// Create temporary directory for test data
	tempDir, err := os.MkdirTemp("", "klp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := NewStorage(tempDir)

	// Setup test data
	venues := []Venue{
		{
			ID:   "venue-1",
			Name: "Test Venue 1",
			Address: Address{
				PostalCode: "29439",
			},
			Coordinates: Coordinates{
				Lat: 53.0,
				Lng: 11.2,
			},
		},
	}
	store.SaveVenues(venues)

	events := []Event{
		{
			ID:      "event-1",
			Title:   "Test Event 1",
			VenueID: "venue-1",
			Date:    "2026-05-29",
		},
	}
	store.SaveEvents(events)

	exhibitions := []Exhibition{
		{
			ID:      "exhibition-1",
			Title:   "Test Exhibition 1",
			VenueID: "venue-1",
		},
	}
	store.SaveExhibitions(exhibitions)

	t.Run("GetVenueByID - found", func(t *testing.T) {
		venue, err := store.GetVenueByID("venue-1")
		if err != nil {
			t.Errorf("GetVenueByID() error = %v, want nil", err)
		}
		if venue == nil {
			t.Fatal("GetVenueByID() returned nil")
		}
		if venue.ID != "venue-1" {
			t.Errorf("GetVenueByID() ID = %q, want %q", venue.ID, "venue-1")
		}
	})

	t.Run("GetVenueByID - not found", func(t *testing.T) {
		_, err := store.GetVenueByID("nonexistent")
		if err == nil {
			t.Error("GetVenueByID() error = nil, want error")
		}
	})

	t.Run("GetEventByID - found", func(t *testing.T) {
		event, err := store.GetEventByID("event-1")
		if err != nil {
			t.Errorf("GetEventByID() error = %v, want nil", err)
		}
		if event == nil {
			t.Fatal("GetEventByID() returned nil")
		}
		if event.ID != "event-1" {
			t.Errorf("GetEventByID() ID = %q, want %q", event.ID, "event-1")
		}
	})

	t.Run("GetExhibitionByID - found", func(t *testing.T) {
		exhibition, err := store.GetExhibitionByID("exhibition-1")
		if err != nil {
			t.Errorf("GetExhibitionByID() error = %v, want nil", err)
		}
		if exhibition == nil {
			t.Fatal("GetExhibitionByID() returned nil")
		}
		if exhibition.ID != "exhibition-1" {
			t.Errorf("GetExhibitionByID() ID = %q, want %q", exhibition.ID, "exhibition-1")
		}
	})
}

func TestStorage_GetVenueWithDetails(t *testing.T) {
	// Create temporary directory for test data
	tempDir, err := os.MkdirTemp("", "klp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := NewStorage(tempDir)

	// Setup test data
	venues := []Venue{
		{
			ID:   "venue-1",
			Name: "Test Venue 1",
			Address: Address{
				PostalCode: "29439",
			},
			Coordinates: Coordinates{
				Lat: 53.0,
				Lng: 11.2,
			},
		},
	}
	store.SaveVenues(venues)

	events := []Event{
		{
			ID:      "event-1",
			Title:   "Event at Venue 1",
			VenueID: "venue-1",
			Date:    "2026-05-29",
		},
		{
			ID:      "event-2",
			Title:   "Another Event at Venue 1",
			VenueID: "venue-1",
			Date:    "2026-05-30",
		},
		{
			ID:      "event-3",
			Title:   "Event at Different Venue",
			VenueID: "venue-2",
			Date:    "2026-05-29",
		},
	}
	store.SaveEvents(events)

	exhibitions := []Exhibition{
		{
			ID:      "exhibition-1",
			Title:   "Exhibition at Venue 1",
			VenueID: "venue-1",
		},
	}
	store.SaveExhibitions(exhibitions)

	venueWithDetails, err := store.GetVenueWithDetails("venue-1")
	if err != nil {
		t.Fatalf("GetVenueWithDetails() error = %v", err)
	}

	if venueWithDetails.ID != "venue-1" {
		t.Errorf("Venue ID = %q, want %q", venueWithDetails.ID, "venue-1")
	}

	if len(venueWithDetails.Events) != 2 {
		t.Errorf("Events count = %d, want 2", len(venueWithDetails.Events))
	}

	if len(venueWithDetails.Exhibitions) != 1 {
		t.Errorf("Exhibitions count = %d, want 1", len(venueWithDetails.Exhibitions))
	}
}

func TestStorage_LoadJSON_FileNotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "klp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := NewStorage(tempDir)

	_, err = store.LoadVenues()
	if err == nil {
		t.Error("LoadVenues() error = nil, want error for missing file")
	}
}

func TestStorage_EnsureDataDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "klp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dataDir := filepath.Join(tempDir, "data", "nested")
	store := NewStorage(dataDir)

	if err := store.EnsureDataDir(); err != nil {
		t.Errorf("EnsureDataDir() error = %v", err)
	}

	// Check directory was created
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Error("EnsureDataDir() did not create directory")
	}
}
