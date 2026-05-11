package crawler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/musche/klp/internal/storage"
)

const landgangBaseURL = "https://landgang-wendland.de"

var (
	landgangAddressRe = regexp.MustCompile(`(?i)Adresse:\s*(.+)`)
	// Fallback: any paragraph that contains "<5-digit PLZ> <City>".
	landgangPlzRe = regexp.MustCompile(`\b\d{5}\s+\p{L}+`)
)

// CrawlLandgang crawls the venue pages of landgang-wendland.de. The site does
// not (yet) publish a structured event programme — its Tribe Events API is
// empty and the /programm/ page is largely placeholder ("kommt noch") — so we
// only return venues and let the geocoder fill in coordinates.
func CrawlLandgang(c *Crawler) ([]storage.Venue, []storage.Event, []storage.Exhibition, error) {
	logger := c.logger
	client := &http.Client{Timeout: 30 * time.Second}

	venueURLs, err := discoverLandgangVenueURLs(client, logger)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("discover venues: %w", err)
	}
	logger.Printf("[landgang] %d venue pages discovered", len(venueURLs))

	venues := make([]storage.Venue, 0, len(venueURLs))
	for i, vu := range venueURLs {
		logger.Printf("[landgang] [%d/%d] %s", i+1, len(venueURLs), vu)
		v, err := parseLandgangVenue(client, vu, logger)
		if err != nil {
			logger.Printf("[landgang ERROR] %s: %v", vu, err)
			continue
		}
		venues = append(venues, v)
		// Be polite to the host.
		time.Sleep(800 * time.Millisecond)
	}

	return venues, nil, nil, nil
}

func discoverLandgangVenueURLs(client *http.Client, logger *log.Logger) ([]string, error) {
	doc, err := fetchLandgangDoc(client, landgangBaseURL+"/")
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var urls []string
	doc.Find(`a[href*="/ort/"]`).Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		href = strings.TrimSpace(href)
		// Normalize: only keep /ort/<slug>/ on the same host.
		if !strings.HasPrefix(href, landgangBaseURL+"/ort/") {
			return
		}
		// Trim query/fragments.
		if i := strings.IndexAny(href, "?#"); i != -1 {
			href = href[:i]
		}
		if !strings.HasSuffix(href, "/") {
			href += "/"
		}
		// Skip the index page itself ("/ort/").
		if strings.Count(strings.TrimPrefix(href, landgangBaseURL+"/ort/"), "/") < 1 {
			return
		}
		if seen[href] {
			return
		}
		seen[href] = true
		urls = append(urls, href)
	})
	if len(urls) == 0 {
		logger.Printf("[landgang WARN] no venue URLs discovered on home page")
	}
	return urls, nil
}

func parseLandgangVenue(client *http.Client, url string, logger *log.Logger) (storage.Venue, error) {
	doc, err := fetchLandgangDoc(client, url)
	if err != nil {
		return storage.Venue{}, err
	}

	name := strings.TrimSpace(doc.Find("h1").First().Text())
	if name == "" {
		// fall back to slug from URL
		name = slugToName(url)
	}

	var description string
	var addressLine string
	// Walk paragraphs in document order; capture description, an "Adresse: …"
	// line if present, or any paragraph that looks like a "<PLZ> <City>" line.
	doc.Find("p").Each(func(_ int, s *goquery.Selection) {
		txt := strings.TrimSpace(s.Text())
		if txt == "" {
			return
		}
		if m := landgangAddressRe.FindStringSubmatch(txt); m != nil {
			addressLine = strings.TrimSpace(m[1])
			return
		}
		if addressLine == "" && landgangPlzRe.MatchString(txt) && len(txt) < 200 {
			addressLine = txt
			return
		}
		if description == "" && len(txt) > 30 && !strings.HasPrefix(strings.ToLower(txt), "im wendland") {
			description = txt
		}
	})

	address := parseLandgangAddress(addressLine)
	if address.City == "" && address.Street == "" {
		logger.Printf("[landgang WARN] %s: no address found", url)
	}

	slug := strings.TrimSuffix(strings.TrimPrefix(url, landgangBaseURL+"/ort/"), "/")
	return storage.Venue{
		ID:          fmt.Sprintf("venue-lg-%s", slug),
		Name:        name,
		Description: description,
		Address:     address,
		Contact:     storage.Contact{Website: url},
		Source:      storage.SourceLandgang,
	}, nil
}

var landgangPlzCityRe = regexp.MustCompile(`(.*?)\s*(\d{5})\s+(.+)$`)

// parseLandgangAddress handles the typical "Beseland Nr. 9, 29459 Clenze" form
// as well as the comma-less "Grünewaldstraße 8 29456 Hitzacker" form some pages
// use.
func parseLandgangAddress(line string) storage.Address {
	line = strings.TrimSpace(line)
	if line == "" {
		return storage.Address{}
	}
	// Expand common German prefix abbreviations so Nominatim can match.
	line = strings.ReplaceAll(line, "Gr. ", "Groß ")
	line = strings.ReplaceAll(line, "Kl. ", "Klein ")
	line = strings.ReplaceAll(line, "St. ", "Sankt ")
	// Fast path: existing helper handles "Street, PLZ City".
	if addr := ParseAddress(line); addr.PostalCode != "" {
		return addr
	}
	// Comma-less fallback: "<Street> <PLZ> <City>" anywhere in the line.
	if m := landgangPlzCityRe.FindStringSubmatch(line); len(m) == 4 {
		return storage.Address{
			Street:     strings.TrimSpace(strings.TrimRight(m[1], ",")),
			PostalCode: m[2],
			City:       strings.TrimSpace(m[3]),
		}
	}
	// Last resort: keep the raw line in Street.
	return storage.Address{Street: line}
}

func slugToName(url string) string {
	slug := strings.TrimSuffix(strings.TrimPrefix(url, landgangBaseURL+"/ort/"), "/")
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

func fetchLandgangDoc(client *http.Client, url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "KLP-Crawler/1.0 (+https://github.com/musche/klp)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	return goquery.NewDocumentFromReader(resp.Body)
}
