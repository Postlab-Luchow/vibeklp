package crawler

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"

	"github.com/musche/klp/internal/storage"
)

// ParseAddress parses an address string into structured components
func ParseAddress(addressStr string) storage.Address {
	// Clean up the address string
	addressStr = strings.TrimSpace(addressStr)

	// Split by comma
	parts := strings.Split(addressStr, ",")

	if len(parts) < 2 {
		return storage.Address{}
	}

	street := strings.TrimSpace(parts[0])
	cityPart := strings.TrimSpace(parts[1])

	// Extract postal code and city using regex
	// Pattern: "PLZ Ort" or "PLZ Ort OT Ortsteil"
	re := regexp.MustCompile(`^(\d{5})\s+(.+)$`)
	matches := re.FindStringSubmatch(cityPart)

	if len(matches) > 2 {
		return storage.Address{
			Street:     street,
			PostalCode: matches[1],
			City:       matches[2],
		}
	}

	return storage.Address{
		Street: street,
		City:   cityPart,
	}
}

// GenerateID generates a unique ID from a string
func GenerateID(prefix, text string) string {
	hash := md5.Sum([]byte(text))
	return fmt.Sprintf("%s-%x", prefix, hash[:8])
}

// CleanText cleans and normalizes text
func CleanText(text string) string {
	// Remove extra whitespace
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return text
}

// ExtractNumber extracts a number from a string like "VERANSTALTUNGEN (9)"
func ExtractNumber(text string) int {
	re := regexp.MustCompile(`\((\d+)\)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		var num int
		fmt.Sscanf(matches[1], "%d", &num)
		return num
	}
	return 0
}

// ParseDateTime parses date and time from strings like "30.05. — 05:00"
func ParseDateTime(dateStr, timeStr string) (date, startTime string) {
	// Parse date: "30.05." -> "2025-05-30"
	dateStr = strings.TrimSpace(dateStr)
	timeStr = strings.TrimSpace(timeStr)

	// Extract day and month
	re := regexp.MustCompile(`(\d{2})\.(\d{2})\.`)
	matches := re.FindStringSubmatch(dateStr)
	if len(matches) > 2 {
		day := matches[1]
		month := matches[2]
		// Assume year 2026 (next festival)
		date = fmt.Sprintf("2026-%s-%s", month, day)
	}

	// Parse time: "05:00" or "05:00 | 08:30"
	if strings.Contains(timeStr, "|") {
		parts := strings.Split(timeStr, "|")
		startTime = strings.TrimSpace(parts[0])
	} else {
		startTime = timeStr
	}

	return date, startTime
}

// NormalizeCategory normalizes category names
func NormalizeCategory(category string) string {
	category = strings.TrimSpace(category)
	category = strings.Title(strings.ToLower(category))
	return category
}

// IsValidEmail checks if an email address is valid
func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// IsValidURL checks if a URL is valid
func IsValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// CleanPhone cleans phone number
func CleanPhone(phone string) string {
	// Remove common prefixes and clean up
	phone = strings.TrimPrefix(phone, "Fon ")
	phone = strings.TrimPrefix(phone, "Tel ")
	phone = strings.TrimSpace(phone)
	return phone
}
