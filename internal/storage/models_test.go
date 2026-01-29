package storage

import (
	"testing"
)

func TestAddress_String(t *testing.T) {
	addr := Address{
		Street:     "Hauptstraße 1",
		PostalCode: "29439",
		City:       "Lüchow",
	}

	expected := "Hauptstraße 1, 29439 Lüchow"
	if got := addr.String(); got != expected {
		t.Errorf("Address.String() = %q, want %q", got, expected)
	}
}

func TestVenue_Validate(t *testing.T) {
	tests := []struct {
		name    string
		venue   Venue
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid venue",
			venue: Venue{
				Name: "Test Venue",
				Address: Address{
					PostalCode: "29439",
					City:       "Lüchow",
				},
				Coordinates: Coordinates{
					Lat: 53.0,
					Lng: 11.2,
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			venue: Venue{
				Address: Address{
					PostalCode: "29439",
				},
				Coordinates: Coordinates{
					Lat: 53.0,
					Lng: 11.2,
				},
			},
			wantErr: true,
			errMsg:  "venue name is required",
		},
		{
			name: "missing address",
			venue: Venue{
				Name:    "Test Venue",
				Address: Address{},
				Coordinates: Coordinates{
					Lat: 53.0,
					Lng: 11.2,
				},
			},
			wantErr: true,
			errMsg:  "address (city or street) is required",
		},
		{
			name: "missing coordinates",
			venue: Venue{
				Name: "Test Venue",
				Address: Address{
					City: "Lüchow",
				},
				Coordinates: Coordinates{},
			},
			wantErr: true,
			errMsg:  "coordinates are required",
		},
		{
			name: "latitude out of range (too low)",
			venue: Venue{
				Name: "Test Venue",
				Address: Address{
					City: "Lüchow",
				},
				Coordinates: Coordinates{
					Lat: 50.0,
					Lng: 11.2,
				},
			},
			wantErr: true,
			errMsg:  "latitude 50.000000 out of expected range",
		},
		{
			name: "longitude out of range (too high)",
			venue: Venue{
				Name: "Test Venue",
				Address: Address{
					City: "Lüchow",
				},
				Coordinates: Coordinates{
					Lat: 53.0,
					Lng: 15.0,
				},
			},
			wantErr: true,
			errMsg:  "longitude 15.000000 out of expected range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.venue.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Venue.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg && !containsSubstring(err.Error(), tt.errMsg) {
					t.Errorf("Venue.Validate() error = %q, want substring %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   Event
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid event",
			event: Event{
				Title:   "Test Event",
				VenueID: "venue-1",
				Date:    "2026-05-29",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			event: Event{
				VenueID: "venue-1",
				Date:    "2026-05-29",
			},
			wantErr: true,
			errMsg:  "event title is required",
		},
		{
			name: "missing venue ID",
			event: Event{
				Title: "Test Event",
				Date:  "2026-05-29",
			},
			wantErr: true,
			errMsg:  "venue ID is required",
		},
		{
			name: "missing date",
			event: Event{
				Title:   "Test Event",
				VenueID: "venue-1",
			},
			wantErr: true,
			errMsg:  "date is required",
		},
		{
			name: "invalid date format",
			event: Event{
				Title:   "Test Event",
				VenueID: "venue-1",
				Date:    "29.05.2026",
			},
			wantErr: true,
			errMsg:  "invalid date format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Event.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !containsSubstring(err.Error(), tt.errMsg) {
					t.Errorf("Event.Validate() error = %q, want substring %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestExhibition_Validate(t *testing.T) {
	tests := []struct {
		name       string
		exhibition Exhibition
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid exhibition",
			exhibition: Exhibition{
				Title:   "Test Exhibition",
				VenueID: "venue-1",
				Artist:  "Test Artist",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			exhibition: Exhibition{
				VenueID: "venue-1",
			},
			wantErr: true,
			errMsg:  "exhibition title is required",
		},
		{
			name: "missing venue ID",
			exhibition: Exhibition{
				Title: "Test Exhibition",
			},
			wantErr: true,
			errMsg:  "venue ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.exhibition.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Exhibition.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("Exhibition.Validate() error = %q, want %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
