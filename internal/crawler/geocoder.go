package crawler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/musche/klp/internal/storage"
)

// NominatimResponse represents the response from Nominatim API
type NominatimResponse struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

// Geocoder handles geocoding of addresses
type Geocoder struct {
	client    *http.Client
	userAgent string
	baseURL   string
}

// NewGeocoder creates a new Geocoder instance
func NewGeocoder(userAgent string) *Geocoder {
	return &Geocoder{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: userAgent,
		baseURL:   "https://nominatim.openstreetmap.org/search",
	}
}

// Geocode converts an address to coordinates
func (g *Geocoder) Geocode(address storage.Address) (storage.Coordinates, error) {
	// Build query string
	query := address.String()

	params := url.Values{}
	params.Add("q", query)
	params.Add("format", "json")
	params.Add("limit", "1")
	params.Add("countrycodes", "de")

	reqURL := fmt.Sprintf("%s?%s", g.baseURL, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return storage.Coordinates{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", g.userAgent)

	// Rate limiting: wait 1 second between requests (Nominatim requirement)
	time.Sleep(1 * time.Second)

	resp, err := g.client.Do(req)
	if err != nil {
		return storage.Coordinates{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return storage.Coordinates{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var results []NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return storage.Coordinates{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(results) == 0 {
		return storage.Coordinates{}, fmt.Errorf("no results found for address: %s", query)
	}

	// Parse coordinates
	var lat, lng float64
	fmt.Sscanf(results[0].Lat, "%f", &lat)
	fmt.Sscanf(results[0].Lon, "%f", &lng)

	return storage.Coordinates{
		Lat: lat,
		Lng: lng,
	}, nil
}

// GeocodeWithRetry attempts to geocode with retries
func (g *Geocoder) GeocodeWithRetry(address storage.Address, maxRetries int) (storage.Coordinates, error) {
	var coords storage.Coordinates
	var err error

	for i := 0; i < maxRetries; i++ {
		coords, err = g.Geocode(address)
		if err == nil {
			return coords, nil
		}

		if i < maxRetries-1 {
			// Exponential backoff
			waitTime := time.Duration(1<<uint(i)) * time.Second
			time.Sleep(waitTime)
		}
	}

	return storage.Coordinates{}, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}
