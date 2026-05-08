package crawler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/musche/klp/internal/crawler/llm"
	"github.com/musche/klp/internal/storage"
)

// Crawler handles web scraping of kulturelle-landpartie.de
type Crawler struct {
	client          *http.Client
	rateLimiter     *time.Ticker
	userAgent       string
	baseURL         string
	geocoder        *Geocoder
	logger          *log.Logger
	useLLM          bool
	llmParser       *LLMParser
	crawlCache      *CrawlCache
	progressTracker *ProgressTracker
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
		useLLM:      false,
	}
}

// SetCrawlCache enables HTML caching for crawl operations
func (c *Crawler) SetCrawlCache(cacheDir string) {
	c.crawlCache = NewCrawlCache(cacheDir)
	if err := c.crawlCache.EnsureDir(); err != nil {
		c.logger.Printf("[WARN] Failed to create crawl cache directory: %v", err)
		c.crawlCache = nil
	} else {
		c.logger.Printf("[INFO] Crawl cache enabled: %s", cacheDir)
		if stats, size, err := c.crawlCache.Stats(); err == nil {
			c.logger.Printf("[INFO] Crawl cache contains %d entries (%.2f MB)", stats, float64(size)/(1024*1024))
		}
	}
}

// SetProgressTracker enables progress tracking for resumable crawls
func (c *Crawler) SetProgressTracker(trackerDir string) {
	c.progressTracker = NewProgressTracker(trackerDir)
	if err := c.progressTracker.Load(); err != nil {
		c.logger.Printf("[WARN] Failed to load progress tracker: %v", err)
	} else {
		completed := c.progressTracker.GetCompletedCount()
		if completed > 0 {
			c.logger.Printf("[INFO] Resuming from previous run: %d venues already crawled", completed)
		}
	}
}

// NewCrawlerWithLLM creates a new Crawler instance with LLM support
func NewCrawlerWithLLM(logger *log.Logger, apiKey, model string, cache *llm.Cache, batchSize int, dryRun bool) *Crawler {
	c := NewCrawler(logger)
	c.useLLM = true
	c.llmParser = NewLLMParser(apiKey, model, cache, batchSize, logger, dryRun)
	return c
}

// SetGoogleMapsGeocoder switches to Google Maps Geocoding API
func (c *Crawler) SetGoogleMapsGeocoder(apiKey string) {
	c.geocoder = NewGoogleMapsGeocoder(apiKey)
}

// Fetch fetches a URL and returns a goquery document
func (c *Crawler) Fetch(url string) (*goquery.Document, error) {
	// Check crawl cache first
	if c.crawlCache != nil {
		if html, ok := c.crawlCache.Get(url); ok {
			c.logger.Printf("[FETCH CACHE HIT] %s", url)
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				return nil, fmt.Errorf("failed to parse cached HTML: %w", err)
			}
			return doc, nil
		}
	}

	// Wait for rate limiter
	<-c.rateLimiter.C

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	c.logger.Printf("[FETCH] %s", url)

	startTime := time.Now()
	resp, err := c.client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("failed to execute request (after %v): %w", duration, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d (after %v)", resp.StatusCode, duration)
	}

	// Read the body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Printf("[FETCH COMPLETE] %s (%d bytes in %v)", url, len(body), duration)

	// Cache the HTML
	if c.crawlCache != nil {
		if err := c.crawlCache.Set(url, string(body)); err != nil {
			c.logger.Printf("[CACHE WARN] Failed to cache %s: %v", url, err)
		} else {
			c.logger.Printf("[CACHE STORED] %s", url)
		}
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return doc, nil
}

// discoverVenueLinks fetches /orte.html and the per-letter pages it links to,
// harvesting venue URLs from each page's .intnavi block. The site no longer
// lists every venue on /orte.html — that page only carries an alphabetical
// jump list, where each letter points to the first venue starting with that
// letter. The full sibling list for a letter lives in the .intnavi block on
// each individual venue page.
func (c *Crawler) discoverVenueLinks() (map[string]bool, error) {
	venueLinks := make(map[string]bool)

	indexURL := c.baseURL + "/orte.html"
	indexDoc, err := c.Fetch(indexURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", indexURL, err)
	}

	letterURLs := make(map[string]bool)
	indexDoc.Find(".listnavi a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		class, _ := s.Attr("class")
		if strings.Contains(class, "step") {
			return
		}
		if !strings.HasPrefix(href, "/orte/") || !strings.HasSuffix(href, ".html") {
			return
		}
		letterURLs[href] = true
	})

	if len(letterURLs) == 0 {
		return nil, fmt.Errorf("no letter-jump links found on %s — site structure may have changed", indexURL)
	}

	c.logger.Printf("[INFO] Found %d letter-jump pages on /orte.html", len(letterURLs))

	for href := range letterURLs {
		venueLinks[href] = true

		doc, err := c.Fetch(c.baseURL + href)
		if err != nil {
			c.logger.Printf("[ERROR] Failed to fetch letter page %s: %v", href, err)
			continue
		}

		before := len(venueLinks)
		doc.Find(".intnavi a").Each(func(i int, s *goquery.Selection) {
			link, exists := s.Attr("href")
			if !exists {
				return
			}
			if link == "/orte.html" {
				return
			}
			if strings.HasPrefix(link, "/orte/") && strings.HasSuffix(link, ".html") {
				venueLinks[link] = true
			}
		})
		c.logger.Printf("[INFO] Letter page %s added %d venues (total: %d)", href, len(venueLinks)-before, len(venueLinks))
	}

	return venueLinks, nil
}

// CrawlAll crawls all venues, events, and exhibitions
func (c *Crawler) CrawlAll() ([]storage.Venue, []storage.Event, []storage.Exhibition, error) {
	c.logger.Println("[INFO] Starting comprehensive crawl...")

	venueLinks, err := c.discoverVenueLinks()
	if err != nil {
		return nil, nil, nil, err
	}

	c.logger.Printf("[INFO] Found %d unique venue links", len(venueLinks))

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

// CrawlVenues crawls all venues (deprecated - use CrawlAll)
func (c *Crawler) CrawlVenues() ([]storage.Venue, error) {
	c.logger.Println("[INFO] Starting venue crawl...")

	venueLinks, err := c.discoverVenueLinks()
	if err != nil {
		return nil, err
	}

	c.logger.Printf("[INFO] Found %d unique venue links", len(venueLinks))

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
	// Check if already completed
	if c.progressTracker != nil && c.progressTracker.IsCompleted(url) {
		c.logger.Printf("[CACHE SKIP] Already crawled: %s", url)
		return nil, nil, nil, fmt.Errorf("already crawled")
	}

	doc, err := c.Fetch(url)
	if err != nil {
		return nil, nil, nil, err
	}

	var venue *storage.Venue
	var events []storage.Event
	var exhibitions []storage.Exhibition

	// Use LLM if enabled
	if c.useLLM {
		venue, events, exhibitions, err = c.crawlWithLLM(doc, url)
	} else {
		venue, events, exhibitions, err = c.crawlWithRegex(doc, url)
	}

	if err != nil {
		return nil, nil, nil, err
	}

	// Mark as completed
	if c.progressTracker != nil {
		c.progressTracker.MarkCompleted(url)
		if err := c.progressTracker.Save(); err != nil {
			c.logger.Printf("[WARN] Failed to save progress: %v", err)
		}
	}

	return venue, events, exhibitions, nil
}

// parseVenueCategories extracts the venue-level facility icons (Café, WC,
// Angebote für Kinder, …). The HTML structure is:
//
//	<div class="icons">
//	  <div><img title="Café (Kuchen, Getränke)" …>14.5.|15.5.|…</div>  <- date-restricted
//	  <img title="WC" …>                                             <- always available
//	</div>
//
// Wrapped icons (img inside a sub-div) carry a list of restriction dates in
// the trailing text node; bare icons (img directly under .icons) apply on all
// festival days.
func parseVenueCategories(doc *goquery.Document) []storage.VenueCategory {
	var categories []storage.VenueCategory
	doc.Find(".icons").First().Children().Each(func(i int, s *goquery.Selection) {
		var img *goquery.Selection
		var dates []string

		switch goquery.NodeName(s) {
		case "img":
			img = s
		case "div":
			img = s.Find("img").First()
			if img.Length() == 0 {
				return
			}
			dates = parseIconDates(s.Text())
		default:
			return
		}

		title, _ := img.Attr("title")
		title = CleanText(title)
		if title == "" {
			return
		}
		categories = append(categories, storage.VenueCategory{Name: title, Dates: dates})
	})
	return categories
}

// parseIconDates pulls the "DD.M." / "DD.MM." date stamps from the trailing
// text of a date-restricted icon entry and returns them as YYYY-MM-DD.
func parseIconDates(text string) []string {
	re := regexp.MustCompile(`(\d{1,2})\.(\d{1,2})\.`)
	matches := re.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil
	}
	dates := make([]string, 0, len(matches))
	for _, m := range matches {
		day, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		dates = append(dates, fmt.Sprintf("2026-%02d-%02d", month, day))
	}
	return dates
}

// crawlWithLLM uses LLM for parsing
func (c *Crawler) crawlWithLLM(doc *goquery.Document, url string) (*storage.Venue, []storage.Event, []storage.Exhibition, error) {
	// Venue name lives in <h1> outside #comblock — extract deterministically so
	// the LLM can't mis-identify it (e.g. picking up the image's alt text).
	venueName := CleanText(doc.Find("h1").First().Text())
	if venueName == "" {
		c.logger.Printf("[LLM WARN] No venue name (h1) found, falling back to regex")
		return c.crawlWithRegex(doc, url)
	}

	// Extract venue HTML block
	comblockHTML, err := doc.Find("#comblock").Html()
	if err != nil {
		c.logger.Printf("[LLM WARN] Failed to extract venue HTML, falling back to regex: %v", err)
		return c.crawlWithRegex(doc, url)
	}

	// Prepend the h1 so the LLM has venue-name context when extracting
	// description/address/contact from comblock.
	venueHTML := fmt.Sprintf("<h1>%s</h1>\n%s", venueName, comblockHTML)

	// Parse venue with LLM
	venue, err := c.llmParser.ParseVenue(venueHTML)
	if err != nil {
		c.logger.Printf("[LLM WARN] Venue parsing failed, falling back to regex: %v", err)
		return c.crawlWithRegex(doc, url)
	}

	// Always override LLM's name/ID with the deterministic h1 value — the LLM
	// has been observed to substitute image alt text or address-line company
	// names for the venue name.
	venue.Name = venueName
	venue.ID = GenerateID("venue", venueName)

	// Categories are deterministic (icon titles) — parse them in Go regardless
	// of which extraction backend handled the rest.
	venue.Categories = parseVenueCategories(doc)

	// Extract events HTML blocks
	var eventHTMLs []string
	doc.Find(".slider.ver .item").Each(func(i int, s *goquery.Selection) {
		html, _ := s.Html()
		eventHTMLs = append(eventHTMLs, html)
	})

	// Parse events with LLM
	events, err := c.llmParser.ParseEvents(eventHTMLs, venue.ID, venue.Name)
	if err != nil {
		c.logger.Printf("[LLM WARN] Event parsing failed, falling back to regex: %v", err)
		// Fall back to regex for events
		var fallbackEvents []storage.Event
		doc.Find(".slider.ver .item").Each(func(i int, s *goquery.Selection) {
			event := c.parseEventFromVenue(s, venue.ID, venue.Name)
			if event != nil {
				fallbackEvents = append(fallbackEvents, *event)
			}
		})
		events = fallbackEvents
	}

	// Extract exhibitions HTML blocks
	var exhibitionHTMLs []string
	doc.Find(".slider.aus .item").Each(func(i int, s *goquery.Selection) {
		html, _ := s.Html()
		exhibitionHTMLs = append(exhibitionHTMLs, html)
	})

	// Parse exhibitions with LLM
	exhibitions, err := c.llmParser.ParseExhibitions(exhibitionHTMLs, venue.ID, venue.Name)
	if err != nil {
		c.logger.Printf("[LLM WARN] Exhibition parsing failed, falling back to regex: %v", err)
		// Fall back to regex for exhibitions
		var fallbackExhibitions []storage.Exhibition
		doc.Find(".slider.aus .item").Each(func(i int, s *goquery.Selection) {
			exhibition := c.parseExhibitionFromVenue(s, venue.ID, venue.Name)
			if exhibition != nil {
				fallbackExhibitions = append(fallbackExhibitions, *exhibition)
			}
		})
		exhibitions = fallbackExhibitions
	}

	venue.EventCount = len(events)
	venue.ExhibitionCount = len(exhibitions)

	return venue, events, exhibitions, nil
}

// crawlWithRegex uses regex-based parsing (original implementation)
func (c *Crawler) crawlWithRegex(doc *goquery.Document, url string) (*storage.Venue, []storage.Event, []storage.Exhibition, error) {
	// Extract venue name from h1
	venueName := CleanText(doc.Find("h1").First().Text())
	if venueName == "" {
		return nil, nil, nil, fmt.Errorf("no venue name found")
	}

	venue := &storage.Venue{
		ID:         GenerateID("venue", venueName),
		Name:       venueName,
		Categories: parseVenueCategories(doc),
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

	// Try to extract address from #comblock paragraphs
	// Address is typically in format: "Street<br/>PostalCode City"
	addressText := ""
	doc.Find("#comblock p").Each(func(i int, s *goquery.Selection) {
		html, _ := s.Html()
		text := CleanText(s.Text())

		// Skip if it's a phone number or email
		if strings.Contains(text, "Fon") || strings.Contains(text, "@") {
			return
		}

		// Look for 5-digit postal code (German postal codes)
		if regexp.MustCompile(`\d{5}`).MatchString(text) {
			// Replace <br/> with comma for parsing
			html = strings.ReplaceAll(html, "<br/>", ", ")
			html = strings.ReplaceAll(html, "<br>", ", ")
			// Clean up prefixes like "Eingang:" or "oder"
			html = regexp.MustCompile(`(?i)(Eingang:\s*|oder\s+)`).ReplaceAllString(html, "")
			addressText = CleanText(html)
			return // Stop after finding first address-like paragraph
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

// GetLLMParser returns the LLM parser (nil if not using LLM)
func (c *Crawler) GetLLMParser() *LLMParser {
	return c.llmParser
}

// HasCrawlCache returns true if crawl cache is configured
func (c *Crawler) HasCrawlCache() bool {
	return c.crawlCache != nil
}

// ParseCachedVenues parses all cached HTML files using LLM
// This allows re-parsing cached data without re-fetching from the website
func (c *Crawler) ParseCachedVenues() ([]storage.Venue, []storage.Event, []storage.Exhibition, error) {
	if !c.useLLM {
		return nil, nil, nil, fmt.Errorf("LLM parsing is not enabled. Use NewCrawlerWithLLM() to enable LLM parsing")
	}

	if c.crawlCache == nil {
		return nil, nil, nil, fmt.Errorf("no crawl cache configured")
	}

	urls, err := c.crawlCache.GetAllCachedURLs()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get cached URLs: %w", err)
	}

	// Filter to only venue URLs (not pagination pages)
	var venueURLs []string
	for _, url := range urls {
		if strings.Contains(url, "/orte/") && strings.HasSuffix(url, ".html") && !strings.Contains(url, "/orte.html") {
			venueURLs = append(venueURLs, url)
		}
	}

	c.logger.Printf("[INFO] Found %d cached venue URLs to parse", len(venueURLs))

	var allVenues []storage.Venue
	var allEvents []storage.Event
	var allExhibitions []storage.Exhibition

	for i, url := range venueURLs {
		c.logger.Printf("[PROGRESS] %d/%d (%.1f%%) - Parsing cached: %s",
			i+1, len(venueURLs), float64(i+1)/float64(len(venueURLs))*100, url)

		venue, events, exhibitions, err := c.ParseCachedVenue(url)
		if err != nil {
			c.logger.Printf("[ERROR] Failed to parse cached venue %s: %v", url, err)
			continue
		}

		allVenues = append(allVenues, *venue)
		allEvents = append(allEvents, events...)
		allExhibitions = append(allExhibitions, exhibitions...)
	}

	c.logger.Printf("[INFO] Parsed %d venues, %d events, %d exhibitions from cache",
		len(allVenues), len(allEvents), len(allExhibitions))

	return allVenues, allEvents, allExhibitions, nil
}

// ParseCachedVenue parses a single cached venue HTML using LLM
func (c *Crawler) ParseCachedVenue(url string) (*storage.Venue, []storage.Event, []storage.Exhibition, error) {
	if !c.useLLM {
		return nil, nil, nil, fmt.Errorf("LLM parsing is not enabled")
	}

	if c.crawlCache == nil {
		return nil, nil, nil, fmt.Errorf("no crawl cache configured")
	}

	entry, ok := c.crawlCache.GetCachedEntry(url)
	if !ok {
		return nil, nil, nil, fmt.Errorf("URL not found in cache: %s", url)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entry.HTML))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse cached HTML: %w", err)
	}

	c.logger.Printf("[CACHE PARSE] %s (cached at %s)", url, entry.Timestamp.Format(time.RFC3339))

	return c.crawlWithLLM(doc, url)
}
