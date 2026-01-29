package crawler

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/musche/klp/internal/storage"
)

// Crawler handles web scraping of kulturelle-landpartie.de
type Crawler struct {
	client      *http.Client
	rateLimiter *time.Ticker
	userAgent   string
	baseURL     string
	geocoder    *Geocoder
	logger      *log.Logger
}

// NewCrawler creates a new Crawler instance
func NewCrawler(logger *log.Logger) *Crawler {
	userAgent := "KLP-Crawler/1.0 (kulturelle-landpartie)"
	return &Crawler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: time.NewTicker(2 * time.Second),
		userAgent:   userAgent,
		baseURL:     "https://www.kulturelle-landpartie.de",
		geocoder:    NewGeocoder(userAgent),
		logger:      logger,
	}
}

// SetGoogleMapsGeocoder switches to Google Maps Geocoding API
func (c *Crawler) SetGoogleMapsGeocoder(apiKey string) {
	c.geocoder = NewGoogleMapsGeocoder(apiKey)
}

// Fetch fetches a URL and returns a goquery document
func (c *Crawler) Fetch(url string) (*goquery.Document, error) {
	// Wait for rate limiter
	<-c.rateLimiter.C

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	c.logger.Printf("[FETCH] %s", url)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return doc, nil
}

// CrawlAll crawls all venues, events, and exhibitions
func (c *Crawler) CrawlAll() ([]storage.Venue, []storage.Event, []storage.Exhibition, error) {
	c.logger.Println("[INFO] Starting comprehensive crawl...")

	venueLinks := make(map[string]bool)
	currentPage := "/orte.html"
	visitedPages := make(map[string]bool)

	// Follow pagination to collect all venue links
	for currentPage != "" {
		if visitedPages[currentPage] {
			break
		}
		visitedPages[currentPage] = true

		url := c.baseURL + currentPage
		doc, err := c.Fetch(url)
		if err != nil {
			c.logger.Printf("[ERROR] Failed to fetch page %s: %v", currentPage, err)
			break
		}

		// Find all venue links on this page
		doc.Find("a[href*='/orte/']").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists {
				return
			}

			if href == "/orte.html" || href == "/orte" || href == "/orte/" {
				return
			}

			if strings.HasPrefix(href, "/orte/") && strings.HasSuffix(href, ".html") {
				class, _ := s.Attr("class")
				title, _ := s.Attr("title")
				if strings.Contains(class, "step") || strings.Contains(title, "Seite") {
					return
				}
				venueLinks[href] = true
			}
		})

		// Find next page link
		nextPage := ""
		doc.Find("a.step").Each(func(i int, s *goquery.Selection) {
			title, _ := s.Attr("title")
			if strings.Contains(title, "nächste") || s.Text() == "►" {
				if href, exists := s.Attr("href"); exists {
					nextPage = href
				}
			}
		})

		c.logger.Printf("[INFO] Page %s: found %d total unique venues so far", currentPage, len(venueLinks))
		currentPage = nextPage
	}

	c.logger.Printf("[INFO] Found %d unique venue links across all pages", len(venueLinks))

	// Now crawl each venue
	var allVenues []storage.Venue
	var allEvents []storage.Event
	var allExhibitions []storage.Exhibition

	i := 0
	for href := range venueLinks {
		i++
		venueURL := c.baseURL + href

		c.logger.Printf("[PROGRESS] %d/%d (%.1f%%) - Crawling: %s",
			i, len(venueLinks), float64(i)/float64(len(venueLinks))*100, href)

		venue, events, exhibitions, err := c.CrawlVenueDetails(venueURL)
		if err != nil {
			c.logger.Printf("[ERROR] Failed to crawl venue %s: %v", href, err)
			continue
		}

		allVenues = append(allVenues, *venue)
		allEvents = append(allEvents, events...)
		allExhibitions = append(allExhibitions, exhibitions...)
	}

	c.logger.Printf("[INFO] Crawled %d venues, %d events, %d exhibitions",
		len(allVenues), len(allEvents), len(allExhibitions))

	return allVenues, allEvents, allExhibitions, nil
}

// CrawlVenues crawls all venues by following pagination (deprecated - use CrawlAll)
func (c *Crawler) CrawlVenues() ([]storage.Venue, error) {
	c.logger.Println("[INFO] Starting venue crawl...")

	venueLinks := make(map[string]bool) // Track unique venue URLs
	currentPage := "/orte.html"
	visitedPages := make(map[string]bool)

	// Follow pagination to collect all venue links
	for currentPage != "" {
		if visitedPages[currentPage] {
			break // Avoid infinite loops
		}
		visitedPages[currentPage] = true

		url := c.baseURL + currentPage
		doc, err := c.Fetch(url)
		if err != nil {
			c.logger.Printf("[ERROR] Failed to fetch page %s: %v", currentPage, err)
			break
		}

		// Find all venue links on this page
		doc.Find("a[href*='/orte/']").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists {
				return
			}

			// Skip navigation links and the index page itself
			if href == "/orte.html" || href == "/orte" || href == "/orte/" {
				return
			}

			// Only include actual venue pages (not navigation)
			if strings.HasPrefix(href, "/orte/") && strings.HasSuffix(href, ".html") {
				// Check if this is a navigation link (has class="step" or title with "Seite")
				class, _ := s.Attr("class")
				title, _ := s.Attr("title")
				if strings.Contains(class, "step") || strings.Contains(title, "Seite") {
					// This might be the next page link
					if strings.Contains(title, "nächste") || s.Text() == "►" {
						// Don't add as venue, but we'll use it for pagination below
					}
					return
				}

				venueLinks[href] = true
			}
		})

		// Find next page link
		nextPage := ""
		doc.Find("a.step").Each(func(i int, s *goquery.Selection) {
			title, _ := s.Attr("title")
			if strings.Contains(title, "nächste") || s.Text() == "►" {
				if href, exists := s.Attr("href"); exists {
					nextPage = href
				}
			}
		})

		c.logger.Printf("[INFO] Page %s: found %d total unique venues so far", currentPage, len(venueLinks))
		currentPage = nextPage
	}

	c.logger.Printf("[INFO] Found %d unique venue links across all pages", len(venueLinks))

	// Now crawl each venue
	var allVenues []storage.Venue
	var allEvents []storage.Event
	var allExhibitions []storage.Exhibition

	i := 0
	for href := range venueLinks {
		i++
		venueURL := c.baseURL + href

		c.logger.Printf("[PROGRESS] %d/%d (%.1f%%) - Crawling: %s",
			i, len(venueLinks), float64(i)/float64(len(venueLinks))*100, href)

		venue, events, exhibitions, err := c.CrawlVenueDetails(venueURL)
		if err != nil {
			c.logger.Printf("[ERROR] Failed to crawl venue %s: %v", href, err)
			continue
		}

		allVenues = append(allVenues, *venue)
		allEvents = append(allEvents, events...)
		allExhibitions = append(allExhibitions, exhibitions...)
	}

	c.logger.Printf("[INFO] Crawled %d venues, %d events, %d exhibitions",
		len(allVenues), len(allEvents), len(allExhibitions))

	// Store events and exhibitions for later saving
	c.logger.Println("[INFO] Storing events and exhibitions...")
	// We'll return these through the main function

	return allVenues, nil
}

// CrawlVenueDetails crawls a single venue page and extracts all data
func (c *Crawler) CrawlVenueDetails(url string) (*storage.Venue, []storage.Event, []storage.Exhibition, error) {
	doc, err := c.Fetch(url)
	if err != nil {
		return nil, nil, nil, err
	}

	// Extract venue name from h1
	venueName := CleanText(doc.Find("h1").First().Text())
	if venueName == "" {
		return nil, nil, nil, fmt.Errorf("no venue name found")
	}

	venue := &storage.Venue{
		ID:   GenerateID("venue", venueName),
		Name: venueName,
	}

	// Extract description (usually the first paragraph or subtitle)
	venue.Description = CleanText(doc.Find("h2, .subtitle, p").First().Text())

	// Extract contact information
	doc.Find("a[href^='tel:']").Each(func(i int, s *goquery.Selection) {
		if venue.Contact.Phone == "" {
			venue.Contact.Phone = CleanPhone(s.Text())
		}
	})

	doc.Find("a[href^='mailto:']").Each(func(i int, s *goquery.Selection) {
		if venue.Contact.Email == "" {
			email := s.Text()
			if IsValidEmail(email) {
				venue.Contact.Email = email
			}
		}
	})

	doc.Find("a[href^='http']").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if venue.Contact.Website == "" && IsValidURL(href) && !strings.Contains(href, "kulturelle-landpartie.de") {
			venue.Contact.Website = href
		}
	})

	// Try to extract address from various possible locations
	addressText := ""
	doc.Find("address, .address, .location").Each(func(i int, s *goquery.Selection) {
		if addressText == "" {
			addressText = CleanText(s.Text())
		}
	})

	if addressText != "" {
		venue.Address = ParseAddress(addressText)
	}

	// Extract events from this venue page
	var events []storage.Event
	doc.Find(".slider.ver .item").Each(func(i int, s *goquery.Selection) {
		event := c.parseEventFromVenue(s, venue.ID, venue.Name)
		if event != nil {
			events = append(events, *event)
		}
	})

	// Extract exhibitions
	var exhibitions []storage.Exhibition
	doc.Find(".slider.aus .item").Each(func(i int, s *goquery.Selection) {
		exhibition := c.parseExhibitionFromVenue(s, venue.ID, venue.Name)
		if exhibition != nil {
			exhibitions = append(exhibitions, *exhibition)
		}
	})

	venue.EventCount = len(events)
	venue.ExhibitionCount = len(exhibitions)

	return venue, events, exhibitions, nil
}

// parseEventFromVenue parses an event from a venue page
func (c *Crawler) parseEventFromVenue(s *goquery.Selection, venueID, venueName string) *storage.Event {
	// Title is in the <b> tag within the first <p> in the second div
	divs := s.Find("div")
	if divs.Length() < 2 {
		return nil
	}

	contentDiv := divs.Eq(1)
	title := CleanText(contentDiv.Find("b").First().Text())
	if title == "" {
		return nil
	}

	// Description is in the <em> tag
	description := CleanText(contentDiv.Find("em").First().Text())

	// Artist/organizer is in the first <p>
	artist := CleanText(contentDiv.Find("p").First().Text())

	event := &storage.Event{
		ID:          GenerateID("event", title+venueID),
		Title:       title,
		Description: description,
		VenueID:     venueID,
		VenueName:   venueName,
		Category:    artist, // Store artist in category for now
	}

	// Extract date and time - they are in text nodes after the <p> tags
	// Format: "29.05. 17:00" or "30.05. 11:00<br/>31.05. 11:00"
	contentText := contentDiv.Text()
	// Find date/time pattern
	re := regexp.MustCompile(`(\d{2}\.\d{2}\.)\s+(\d{2}:\d{2})`)
	matches := re.FindStringSubmatch(contentText)
	if len(matches) >= 3 {
		date, startTime := ParseDateTime(matches[1], matches[2])
		event.Date = date
		event.StartTime = startTime
	}

	// Extract admission - it's in parentheses at the end
	re2 := regexp.MustCompile(`\(([^)]+)\)\s*$`)
	matches2 := re2.FindStringSubmatch(contentText)
	if len(matches2) >= 2 {
		event.Admission = matches2[1]
	}

	return event
}

// parseExhibitionFromVenue parses an exhibition from a venue page
func (c *Crawler) parseExhibitionFromVenue(s *goquery.Selection, venueID, venueName string) *storage.Exhibition {
	// Title is in the <b> tag within the first <p> in the second div
	divs := s.Find("div")
	if divs.Length() < 2 {
		return nil
	}

	contentDiv := divs.Eq(1)
	title := CleanText(contentDiv.Find("b").First().Text())
	if title == "" {
		return nil
	}

	// Description is in the <em> tag
	description := CleanText(contentDiv.Find("em").First().Text())

	// Artist is in the first <p>
	artist := CleanText(contentDiv.Find("p").First().Text())

	exhibition := &storage.Exhibition{
		ID:          GenerateID("exhibition", title+venueID),
		Title:       title,
		Description: description,
		Artist:      artist,
		VenueID:     venueID,
		VenueName:   venueName,
	}

	return exhibition
}

// GeocodeVenues adds coordinates to venues
func (c *Crawler) GeocodeVenues(venues []storage.Venue) error {
	c.logger.Println("[INFO] Starting geocoding...")

	for i := range venues {
		// For Google Maps, we can attempt geocoding even without postal code
		// For Nominatim, postal code is required
		if venues[i].Address.PostalCode == "" && c.geocoder.provider == "nominatim" {
			c.logger.Printf("[WARN] Skipping geocoding for %s (no postal code, Nominatim requires it)", venues[i].Name)
			continue
		}

		if venues[i].Address.City == "" && venues[i].Address.Street == "" {
			c.logger.Printf("[WARN] Skipping geocoding for %s (insufficient address data)", venues[i].Name)
			continue
		}

		c.logger.Printf("[PROGRESS] %d/%d (%.1f%%) - Geocoding: %s",
			i+1, len(venues), float64(i+1)/float64(len(venues))*100, venues[i].Name)

		coords, err := c.geocoder.GeocodeWithRetry(venues[i].Address, 3)
		if err != nil {
			c.logger.Printf("[ERROR] Failed to geocode %s: %v", venues[i].Name, err)
			continue
		}

		venues[i].Coordinates = coords
		c.logger.Printf("[INFO] Geocoded %s: %.6f, %.6f", venues[i].Name, coords.Lat, coords.Lng)
	}

	return nil
}

// Close closes the crawler and releases resources
func (c *Crawler) Close() {
	c.rateLimiter.Stop()
}

// CrawlEvents is deprecated - events are now crawled from venue pages
func (c *Crawler) CrawlEvents() ([]storage.Event, error) {
	c.logger.Println("[WARN] CrawlEvents is deprecated - events are crawled from venue pages")
	return []storage.Event{}, nil
}

// CrawlExhibitions is deprecated - exhibitions are now crawled from venue pages
func (c *Crawler) CrawlExhibitions(venues []storage.Venue) ([]storage.Exhibition, error) {
	c.logger.Println("[WARN] CrawlExhibitions is deprecated - exhibitions are crawled from venue pages")
	return []storage.Exhibition{}, nil
}
