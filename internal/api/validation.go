package api

import (
	"html"
	"net/http"
	"regexp"
	"strings"
)

var (
	// dateRegex validates YYYY-MM-DD format
	dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	// idRegex validates safe ID format (alphanumeric, dash, underscore)
	idRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// SanitizeString sanitizes user input to prevent XSS
func SanitizeString(s string) string {
	// Trim whitespace
	s = strings.TrimSpace(s)
	// Escape HTML special characters
	s = html.EscapeString(s)
	// Limit length
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}

// ValidateSearchQuery validates and sanitizes search query
func ValidateSearchQuery(q string) (string, bool) {
	q = strings.TrimSpace(q)
	if q == "" {
		return "", false
	}
	// Minimum 2 characters for search
	if len(q) < 2 {
		return "", false
	}
	// Maximum 100 characters
	if len(q) > 100 {
		q = q[:100]
	}
	return SanitizeString(q), true
}

// ValidateDate validates date format
func ValidateDate(date string) bool {
	if date == "" {
		return true // Empty date is valid (means no filter)
	}
	return dateRegex.MatchString(date)
}

// ValidateID validates ID format
func ValidateID(id string) bool {
	if id == "" {
		return false
	}
	if len(id) > 100 {
		return false
	}
	return idRegex.MatchString(id)
}

// ValidateSearchType validates search type parameter
func ValidateSearchType(searchType string) bool {
	if searchType == "" {
		return true // Empty means all types
	}
	validTypes := map[string]bool{
		"venues":      true,
		"events":      true,
		"exhibitions": true,
	}
	return validTypes[searchType]
}

// GetQueryParam safely extracts and sanitizes a query parameter
func GetQueryParam(r *http.Request, key string) string {
	value := r.URL.Query().Get(key)
	return SanitizeString(value)
}

// GetQueryParamRaw extracts a query parameter without sanitization
func GetQueryParamRaw(r *http.Request, key string) string {
	return strings.TrimSpace(r.URL.Query().Get(key))
}
