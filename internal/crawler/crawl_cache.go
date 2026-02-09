package crawler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CrawlCache provides disk-based caching for crawled HTML content
type CrawlCache struct {
	Dir string
}

// CrawlCacheEntry represents a cached crawl result
type CrawlCacheEntry struct {
	URL       string    `json:"url"`
	HTML      string    `json:"html"`
	Timestamp time.Time `json:"timestamp"`
}

// NewCrawlCache creates a new crawl cache instance
func NewCrawlCache(dir string) *CrawlCache {
	return &CrawlCache{
		Dir: dir,
	}
}

// EnsureDir creates the cache directory if it doesn't exist
func (c *CrawlCache) EnsureDir() error {
	return os.MkdirAll(c.Dir, 0755)
}

// generateCacheKey creates a cache key from URL
func (c *CrawlCache) generateCacheKey(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:16])
}

// Get retrieves cached HTML if it exists
func (c *CrawlCache) Get(url string) (string, bool) {
	key := c.generateCacheKey(url)
	path := filepath.Join(c.Dir, key+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	var entry CrawlCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return "", false
	}

	return entry.HTML, true
}

// Set stores HTML in cache
func (c *CrawlCache) Set(url, html string) error {
	if err := c.EnsureDir(); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	key := c.generateCacheKey(url)
	path := filepath.Join(c.Dir, key+".json")

	entry := CrawlCacheEntry{
		URL:       url,
		HTML:      html,
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

// Has checks if URL is cached
func (c *CrawlCache) Has(url string) bool {
	_, ok := c.Get(url)
	return ok
}

// Stats returns cache statistics
func (c *CrawlCache) Stats() (total int, size int64, err error) {
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

// Clear removes all cached entries
func (c *CrawlCache) Clear() error {
	return os.RemoveAll(c.Dir)
}

// ProgressTracker tracks crawl progress for resumption
type ProgressTracker struct {
	Dir           string
	completedURLs map[string]bool
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(dir string) *ProgressTracker {
	return &ProgressTracker{
		Dir:           dir,
		completedURLs: make(map[string]bool),
	}
}

// Load loads progress from disk
func (p *ProgressTracker) Load() error {
	path := filepath.Join(p.Dir, "progress.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var urls []string
	if err := json.Unmarshal(data, &urls); err != nil {
		return err
	}

	for _, url := range urls {
		p.completedURLs[url] = true
	}
	return nil
}

// Save saves progress to disk
func (p *ProgressTracker) Save() error {
	if err := os.MkdirAll(p.Dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(p.Dir, "progress.json")
	urls := make([]string, 0, len(p.completedURLs))
	for url := range p.completedURLs {
		urls = append(urls, url)
	}

	data, err := json.MarshalIndent(urls, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// MarkCompleted marks a URL as completed
func (p *ProgressTracker) MarkCompleted(url string) {
	p.completedURLs[url] = true
}

// IsCompleted checks if URL was already crawled
func (p *ProgressTracker) IsCompleted(url string) bool {
	return p.completedURLs[url]
}

// GetCompletedCount returns number of completed URLs
func (p *ProgressTracker) GetCompletedCount() int {
	return len(p.completedURLs)
}

// GetAllCachedURLs returns all URLs from cached entries
func (c *CrawlCache) GetAllCachedURLs() ([]string, error) {
	entries, err := os.ReadDir(c.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var urls []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(c.Dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var cacheEntry CrawlCacheEntry
		if err := json.Unmarshal(data, &cacheEntry); err != nil {
			continue
		}

		urls = append(urls, cacheEntry.URL)
	}

	return urls, nil
}

// GetCachedEntry retrieves the full cache entry for a URL
func (c *CrawlCache) GetCachedEntry(url string) (*CrawlCacheEntry, bool) {
	key := c.generateCacheKey(url)
	path := filepath.Join(c.Dir, key+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry CrawlCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	return &entry, true
}
