package llm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Cache provides disk-based caching for LLM responses
type Cache struct {
	Dir string
	TTL time.Duration
}

// CacheEntry represents a cached response
type CacheEntry struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// NewCache creates a new cache instance
func NewCache(dir string, ttl time.Duration) *Cache {
	return &Cache{
		Dir: dir,
		TTL: ttl,
	}
}

// EnsureDir creates the cache directory if it doesn't exist
func (c *Cache) EnsureDir() error {
	return os.MkdirAll(c.Dir, 0755)
}

// Get retrieves cached response if not expired
func (c *Cache) Get(key string) (string, bool) {
	hash := sha256.Sum256([]byte(key))
	filename := hex.EncodeToString(hash[:16]) + ".json"
	path := filepath.Join(c.Dir, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return "", false
	}

	// Check if expired
	if c.TTL > 0 && time.Since(entry.Timestamp) > c.TTL {
		// Delete expired entry
		os.Remove(path)
		return "", false
	}

	return entry.Content, true
}

// Set stores response in cache
func (c *Cache) Set(key, response string) error {
	if err := c.EnsureDir(); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	hash := sha256.Sum256([]byte(key))
	filename := hex.EncodeToString(hash[:16]) + ".json"
	path := filepath.Join(c.Dir, filename)

	entry := CacheEntry{
		Content:   response,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Clear removes all cached entries
func (c *Cache) Clear() error {
	return os.RemoveAll(c.Dir)
}

// Stats returns cache statistics
func (c *Cache) Stats() (total int, size int64, err error) {
	entries, err := os.ReadDir(c.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			total++
			info, err := entry.Info()
			if err == nil {
				size += info.Size()
			}
		}
	}

	return total, size, nil
}
