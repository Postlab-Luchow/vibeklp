package crawler

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/musche/klp/internal/storage"
)

// TestParseVenueDetailsFromSnapshot tests venue parsing using saved HTML snapshot
func TestParseVenueDetailsFromSnapshot(t *testing.T) {
	// Load HTML snapshot from testdata
	htmlContent, err := os.ReadFile("testdata/bankewitz.html")
	if err != nil {
		t.Fatalf("Failed to read testdata: %v", err)
	}

	// Parse HTML into goquery document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlContent)))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	// Create crawler instance
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	c := NewCrawler(logger)

	// Extract venue info
	venue := &storage.Venue{
		ID:   "test-venue-id",
		Name: "BANKEWITZ",
	}

	// Test venue name extraction
	t.Run("VenueName", func(t *testing.T) {
		name := strings.TrimSpace(doc.Find("h1.norm").First().Text())
		if name != "BANKEWITZ" {
			t.Errorf("Expected venue name 'BANKEWITZ', got '%s'", name)
		}
	})

	// Test venue subtitle extraction
	t.Run("VenueSubtitle", func(t *testing.T) {
		subtitle := strings.TrimSpace(doc.Find("h3").First().Text())
		if subtitle != "IM VERMEINTLICH WILDEN GARTEN" {
			t.Errorf("Expected subtitle 'IM VERMEINTLICH WILDEN GARTEN', got '%s'", subtitle)
		}
	})

	// Test address extraction from comblock
	t.Run("VenueAddress", func(t *testing.T) {
		addressText := doc.Find("#comblock p").Eq(1).Text()
		if !strings.Contains(addressText, "Zum Seinitz Moor 1") {
			t.Errorf("Expected address to contain 'Zum Seinitz Moor 1', got '%s'", addressText)
		}
		if !strings.Contains(addressText, "29597 Stoetze OT Bankewitz") {
			t.Errorf("Expected address to contain '29597 Stoetze OT Bankewitz', got '%s'", addressText)
		}
	})

	// Test contact info extraction
	t.Run("VenuePhone", func(t *testing.T) {
		phoneText := doc.Find("#comblock p").Eq(2).Text()
		if !strings.Contains(phoneText, "05872 986107") {
			t.Errorf("Expected phone to contain '05872 986107', got '%s'", phoneText)
		}
	})

	t.Run("VenueWebsite", func(t *testing.T) {
		website, exists := doc.Find("#comblock a[href^='http']").First().Attr("href")
		if !exists || website != "http://www.wandafulworld.de" {
			t.Errorf("Expected website 'http://www.wandafulworld.de', got '%s'", website)
		}
	})

	// Test exhibitions extraction
	t.Run("ExhibitionsCount", func(t *testing.T) {
		exhibitions := doc.Find(".slider.aus .item")
		count := exhibitions.Length()
		if count != 3 {
			t.Errorf("Expected 3 exhibitions, found %d", count)
		}
	})

	t.Run("ExhibitionsData", func(t *testing.T) {
		exhibitions := []storage.Exhibition{}
		doc.Find(".slider.aus .item").Each(func(i int, s *goquery.Selection) {
			exhibition := c.parseExhibitionFromVenue(s, venue.ID, venue.Name)
			if exhibition != nil {
				exhibitions = append(exhibitions, *exhibition)
			}
		})

		if len(exhibitions) != 3 {
			t.Fatalf("Expected 3 exhibitions, got %d", len(exhibitions))
		}

		// Verify first exhibition
		ex1 := exhibitions[0]
		if ex1.Title != "KUNST & Gebrauchserfreulichkeiten mit MOKKA im vermeintlich wilden GARTEN" {
			t.Errorf("Exhibition 1 title incorrect: %s", ex1.Title)
		}
		if ex1.Artist != "Wanda Sippl" {
			t.Errorf("Exhibition 1 artist incorrect: %s", ex1.Artist)
		}
		if !strings.Contains(ex1.Description, "Gartennudist:innen gucken") {
			t.Errorf("Exhibition 1 description incorrect: %s", ex1.Description)
		}

		// Verify second exhibition
		ex2 := exhibitions[1]
		if ex2.Title != "Volle Vrauen SKULPTUREN" {
			t.Errorf("Exhibition 2 title incorrect: %s", ex2.Title)
		}
		if ex2.Artist != "Kathrin Matzak" {
			t.Errorf("Exhibition 2 artist incorrect: %s", ex2.Artist)
		}

		// Verify third exhibition
		ex3 := exhibitions[2]
		if ex3.Title != "Wilde MÖWEN – Gemälde, die Nazis den Vogel zeigen" {
			t.Errorf("Exhibition 3 title incorrect: %s", ex3.Title)
		}
		if ex3.Artist != "Nicole Gläß" {
			t.Errorf("Exhibition 3 artist incorrect: %s", ex3.Artist)
		}
	})

	// Test events extraction
	t.Run("EventsCount", func(t *testing.T) {
		events := doc.Find(".slider.ver .item")
		count := events.Length()
		if count != 9 {
			t.Errorf("Expected 9 events, found %d", count)
		}
	})

	t.Run("EventsData", func(t *testing.T) {
		events := []storage.Event{}
		doc.Find(".slider.ver .item").Each(func(i int, s *goquery.Selection) {
			event := c.parseEventFromVenue(s, venue.ID, venue.Name)
			if event != nil {
				events = append(events, *event)
			}
		})

		if len(events) != 9 {
			t.Fatalf("Expected 9 events, got %d", len(events))
		}

		// Verify first event
		ev1 := events[0]
		if ev1.Title != "probt mit Anleitung zum Mitsingen" {
			t.Errorf("Event 1 title incorrect: %s", ev1.Title)
		}
		if ev1.Category != "Süttorfer Sängerei" {
			t.Errorf("Event 1 organizer/artist incorrect: %s", ev1.Category)
		}
		if ev1.Date != "2026-05-29" {
			t.Errorf("Event 1 date incorrect: %s", ev1.Date)
		}
		if ev1.StartTime != "17:00" {
			t.Errorf("Event 1 time incorrect: %s", ev1.StartTime)
		}
		if ev1.Admission != "Hutkasse" {
			t.Errorf("Event 1 admission incorrect: %s", ev1.Admission)
		}

		// Verify event with multi-day dates (only captures first date)
		ev2 := events[1]
		if ev2.Title != "4x Plastizieren mit Papier -Upcycling-" {
			t.Errorf("Event 2 title incorrect: %s", ev2.Title)
		}
		if ev2.Date != "2026-05-30" {
			t.Errorf("Event 2 date incorrect (should be first date): %s", ev2.Date)
		}
		if !strings.Contains(ev2.Admission, "Hutkasse") {
			t.Errorf("Event 2 admission should contain 'Hutkasse': %s", ev2.Admission)
		}

		// Verify event with no description
		ev3 := events[2]
		if ev3.Title != "Kakao-Zeremonie und Klang-Atmen" {
			t.Errorf("Event 3 title incorrect: %s", ev3.Title)
		}
		if ev3.Category != "Janela Fitz" {
			t.Errorf("Event 3 organizer incorrect: %s", ev3.Category)
		}
		if !strings.Contains(ev3.Description, "Rohkakao") {
			t.Errorf("Event 3 should have description: %s", ev3.Description)
		}

		// Verify event with free admission
		ev7 := events[6]
		if ev7.Title != "Sofaplausch mit K. Peschel, N. Peschel & B. Scharnhop" {
			t.Errorf("Event 7 title incorrect: %s", ev7.Title)
		}
		if ev7.Admission != "Eintritt frei" {
			t.Errorf("Event 7 admission incorrect: %s", ev7.Admission)
		}
	})
}

// TestParseExhibitionFromVenue tests exhibition parsing in isolation
func TestParseExhibitionFromVenue(t *testing.T) {
	htmlContent, err := os.ReadFile("testdata/bankewitz.html")
	if err != nil {
		t.Fatalf("Failed to read testdata: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlContent)))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	c := NewCrawler(logger)

	// Parse first exhibition
	firstExhibition := doc.Find(".slider.aus .item").First()
	exhibition := c.parseExhibitionFromVenue(firstExhibition, "test-venue", "Test Venue")

	if exhibition == nil {
		t.Fatal("Expected exhibition to be parsed, got nil")
	}

	// Validate required fields
	if exhibition.Title == "" {
		t.Error("Exhibition title should not be empty")
	}
	if exhibition.Artist == "" {
		t.Error("Exhibition artist should not be empty")
	}
	if exhibition.VenueID != "test-venue" {
		t.Errorf("Expected VenueID 'test-venue', got '%s'", exhibition.VenueID)
	}
	if exhibition.VenueName != "Test Venue" {
		t.Errorf("Expected VenueName 'Test Venue', got '%s'", exhibition.VenueName)
	}
}

// TestParseEventFromVenue tests event parsing in isolation
func TestParseEventFromVenue(t *testing.T) {
	htmlContent, err := os.ReadFile("testdata/bankewitz.html")
	if err != nil {
		t.Fatalf("Failed to read testdata: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlContent)))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	c := NewCrawler(logger)

	// Parse first event
	firstEvent := doc.Find(".slider.ver .item").First()
	event := c.parseEventFromVenue(firstEvent, "test-venue", "Test Venue")

	if event == nil {
		t.Fatal("Expected event to be parsed, got nil")
	}

	// Validate required fields
	if event.Title == "" {
		t.Error("Event title should not be empty")
	}
	if event.Date == "" {
		t.Error("Event date should not be empty")
	}
	if event.VenueID != "test-venue" {
		t.Errorf("Expected VenueID 'test-venue', got '%s'", event.VenueID)
	}
	if event.VenueName != "Test Venue" {
		t.Errorf("Expected VenueName 'Test Venue', got '%s'", event.VenueName)
	}
}

// TestHTMLStructureRegression ensures the HTML selectors remain valid
func TestHTMLStructureRegression(t *testing.T) {
	htmlContent, err := os.ReadFile("testdata/bankewitz.html")
	if err != nil {
		t.Fatalf("Failed to read testdata: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlContent)))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	// Verify critical selectors exist
	tests := []struct {
		name     string
		selector string
		minCount int
	}{
		{"Exhibition container", ".slider.aus", 1},
		{"Event container", ".slider.ver", 1},
		{"Exhibition items", ".slider.aus .item", 1},
		{"Event items", ".slider.ver .item", 1},
		{"Venue header", "h1.norm", 1},
		{"Content block", "#comblock", 1},
		{"Pagination", ".listnavi", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selection := doc.Find(tt.selector)
			count := selection.Length()
			if count < tt.minCount {
				t.Errorf("Selector '%s' found %d elements, expected at least %d", tt.selector, count, tt.minCount)
			}
		})
	}
}
