package crawler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/musche/klp/internal/storage"
)

// NominatimResponse represents the response from Nominatim API
type NominatimResponse struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

// GoogleMapsResponse represents the response from Google Maps Geocoding API
type GoogleMapsResponse struct {
	Results []struct {
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
		FormattedAddress string `json:"formatted_address"`
	} `json:"results"`
	Status string `json:"status"`
}

// Geocoder handles geocoding of addresses
type Geocoder struct {
	client    *http.Client
	userAgent string
	baseURL   string
	apiKey    string
	provider  string // "nominatim" or "google"
}

// NewGeocoder creates a new Geocoder instance using Nominatim (default)
func NewGeocoder(userAgent string) *Geocoder {
	return &Geocoder{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: userAgent,
		baseURL:   "https://nominatim.openstreetmap.org/search",
		provider:  "nominatim",
	}
}

// NewGoogleMapsGeocoder creates a new Geocoder instance using Google Maps API
func NewGoogleMapsGeocoder(apiKey string) *Geocoder {
	return &Geocoder{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:   apiKey,
		baseURL:  "https://maps.googleapis.com/maps/api/geocode/json",
		provider: "google",
	}
}

// Geocode converts an address to coordinates
func (g *Geocoder) Geocode(address storage.Address) (storage.Coordinates, error) {
	if g.provider == "google" {
		return g.geocodeGoogle(address)
	}
	return g.geocodeNominatim(address)
}

// geocodeNominatim uses Nominatim (OpenStreetMap) for geocoding
func (g *Geocoder) geocodeNominatim(address storage.Address) (storage.Coordinates, error) {
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

// geocodeGoogle uses Google Maps Geocoding API
func (g *Geocoder) geocodeGoogle(address storage.Address) (storage.Coordinates, error) {
	// Build address query - be more flexible with missing postal codes
	var queryParts []string
	if address.Street != "" {
		queryParts = append(queryParts, address.Street)
	}
	if address.PostalCode != "" {
		queryParts = append(queryParts, address.PostalCode)
	}
	if address.City != "" {
		queryParts = append(queryParts, address.City)
	}
	queryParts = append(queryParts, "Germany")
	
	query := strings.Join(queryParts, ", ")

	params := url.Values{}
	params.Add("address", query)
	params.Add("key", g.apiKey)
	params.Add("region", "de")

	reqURL := fmt.Sprintf("%s?%s", g.baseURL, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return storage.Coordinates{}, fmt.Errorf("failed to create request: %w", err)
	}

	// No rate limiting needed for Google Maps with API key
	resp, err := g.client.Do(req)
	if err != nil {
		return storage.Coordinates{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return storage.Coordinates{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result GoogleMapsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return storage.Coordinates{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != "OK" || len(result.Results) == 0 {
		return storage.Coordinates{}, fmt.Errorf("no results found for address: %s (status: %s)", query, result.Status)
	}

	location := result.Results[0].Geometry.Location
	return storage.Coordinates{
		Lat: location.Lat,
		Lng: location.Lng,
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
