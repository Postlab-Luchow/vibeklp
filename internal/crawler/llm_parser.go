package crawler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/musche/klp/internal/crawler/llm"
	"github.com/musche/klp/internal/crawler/schemas"
	"github.com/musche/klp/internal/storage"
)

// LLMParser handles parsing HTML using LLM with structured outputs
type LLMParser struct {
	client      *llm.OpenRouterClient
	cache       *llm.Cache
	batchProc   *llm.BatchProcessor
	logger      *log.Logger
	dryRun      bool
	totalTokens int
}

// NewLLMParser creates a new LLM parser
func NewLLMParser(apiKey, model string, cache *llm.Cache, batchSize int, logger *log.Logger, dryRun bool) *LLMParser {
	return &LLMParser{
		client:      llm.NewClient(apiKey, model),
		cache:       cache,
		batchProc:   llm.NewBatchProcessor(batchSize),
		logger:      logger,
		dryRun:      dryRun,
		totalTokens: 0,
	}
}

// GetTotalTokens returns the total tokens used
func (p *LLMParser) GetTotalTokens() int {
	return p.totalTokens
}

// ParseVenue extracts venue data using LLM
func (p *LLMParser) ParseVenue(html string) (*storage.Venue, error) {
	cacheKey := p.generateCacheKey("venue", html)

	// Check cache
	if p.cache != nil {
		if cached, ok := p.cache.Get(cacheKey); ok {
			p.logger.Println("[LLM CACHE HIT] Venue")
			return p.unmarshalVenue(cached)
		}
	}

	if p.dryRun {
		p.logger.Println("[DRY RUN] Would extract venue with LLM")
		return nil, fmt.Errorf("dry run mode")
	}

	schema := schemas.VenueSchema()
	systemPrompt := `You are a data extraction specialist for a German cultural festival website. 
Extract venue information from the provided HTML content accurately.

Rules:
1. Extract the venue name from h1 tag
2. Extract address components (street, 5-digit postal code, city)
3. Decode any JavaScript-encoded email addresses (look for eval(decodeURIComponent(...)))
4. Clean phone numbers by removing "Fon" or "Tel" prefixes
5. Extract bike route number from "Fahrradtour: X" text
6. Return ONLY valid JSON matching the schema exactly`

	result, err := p.client.ExtractWithSchema(html, schema, systemPrompt, p.logger)
	if err != nil {
		return nil, fmt.Errorf("LLM extraction failed: %w", err)
	}

	p.totalTokens += result.PromptTokens + result.CompTokens

	// Cache response
	if p.cache != nil {
		if err := p.cache.Set(cacheKey, result.Content); err != nil {
			p.logger.Printf("[CACHE WARN] Failed to cache venue: %v", err)
		}
	}

	return p.unmarshalVenue(result.Content)
}

// ParseEvents extracts events using LLM with batch processing
func (p *LLMParser) ParseEvents(htmlBlocks []string, venueID, venueName string) ([]storage.Event, error) {
	if len(htmlBlocks) == 0 {
		return []storage.Event{}, nil
	}

	var allEvents []storage.Event

	// Process in batches
	batches := p.batchProc.CreateBatches(htmlBlocks)

	for i, batch := range batches {
		p.logger.Printf("[LLM] Processing events batch %d/%d for %s", i+1, len(batches), venueName)

		events, err := p.parseEventBatch(batch, venueID, venueName)
		if err != nil {
			p.logger.Printf("[LLM ERROR] Failed to parse event batch %d: %v", i+1, err)
			continue
		}

		allEvents = append(allEvents, events...)
	}

	return allEvents, nil
}

// parseEventBatch processes multiple events in one API call
func (p *LLMParser) parseEventBatch(htmlBlocks []string, venueID, venueName string) ([]storage.Event, error) {
	// Generate cache key from combined HTML
	combinedHTML := p.batchProc.CombineHTMLBlocks(htmlBlocks)
	cacheKey := p.generateCacheKey("events", combinedHTML)

	// Check cache
	if p.cache != nil {
		if cached, ok := p.cache.Get(cacheKey); ok {
			p.logger.Println("[LLM CACHE HIT] Events batch")
			return p.unmarshalEventsBatch(cached, venueID, venueName)
		}
	}

	if p.dryRun {
		p.logger.Printf("[DRY RUN] Would extract %d events with LLM", len(htmlBlocks))
		return nil, fmt.Errorf("dry run mode")
	}

	batchSchema := schemas.EventsBatchSchema()

	systemPrompt := fmt.Sprintf(`You are a data extraction specialist for a German cultural festival website.
Extract ALL events from the provided HTML. The venue is "%s".

Important rules:
1. Extract ALL dates - events can have multiple dates listed on separate lines (e.g., "30.05. 11:00<br>31.05. 11:00")
2. Convert dates from DD.MM. format to YYYY-MM-DD format (year is always 2026)
3. Each "DD.MM. HH:MM" entry is a single occurrence whose time is ALWAYS the startTime. Never put a single time into endTime.
4. endTime is empty UNLESS the title or description explicitly contains a time range like "18:30 – 21:00 Uhr" — in that case use the second time as endTime.
5. If the same date appears with two different times (e.g. "19.05. 13:00" and "19.05. 20:30"), emit TWO separate date entries, each with its own startTime.
6. Times are in 24-hour HH:MM format
7. Admission info is usually in parentheses like (Hutkasse) or (Eintritt frei)
8. Return ALL events found in the HTML
9. Each item in the HTML is separated by "---EVENT_SEPARATOR---"`, venueName)

	result, err := p.client.ExtractWithSchema(combinedHTML, batchSchema, systemPrompt, p.logger)
	if err != nil {
		return nil, err
	}

	p.totalTokens += result.PromptTokens + result.CompTokens

	// Cache response
	if p.cache != nil {
		if err := p.cache.Set(cacheKey, result.Content); err != nil {
			p.logger.Printf("[CACHE WARN] Failed to cache events batch: %v", err)
		}
	}

	return p.unmarshalEventsBatch(result.Content, venueID, venueName)
}

// ParseExhibitions extracts exhibitions using LLM with batch processing
func (p *LLMParser) ParseExhibitions(htmlBlocks []string, venueID, venueName string) ([]storage.Exhibition, error) {
	if len(htmlBlocks) == 0 {
		return []storage.Exhibition{}, nil
	}

	var allExhibitions []storage.Exhibition

	// Process in batches
	batches := p.batchProc.CreateBatches(htmlBlocks)

	for i, batch := range batches {
		p.logger.Printf("[LLM] Processing exhibitions batch %d/%d for %s", i+1, len(batches), venueName)

		exhibitions, err := p.parseExhibitionBatch(batch, venueID, venueName)
		if err != nil {
			p.logger.Printf("[LLM ERROR] Failed to parse exhibition batch %d: %v", i+1, err)
			continue
		}

		allExhibitions = append(allExhibitions, exhibitions...)
	}

	return allExhibitions, nil
}

// parseExhibitionBatch processes multiple exhibitions in one API call
func (p *LLMParser) parseExhibitionBatch(htmlBlocks []string, venueID, venueName string) ([]storage.Exhibition, error) {
	// Generate cache key from combined HTML
	combinedHTML := p.batchProc.CombineHTMLBlocks(htmlBlocks)
	cacheKey := p.generateCacheKey("exhibitions", combinedHTML)

	// Check cache
	if p.cache != nil {
		if cached, ok := p.cache.Get(cacheKey); ok {
			p.logger.Println("[LLM CACHE HIT] Exhibitions batch")
			return p.unmarshalExhibitionsBatch(cached, venueID, venueName)
		}
	}

	if p.dryRun {
		p.logger.Printf("[DRY RUN] Would extract %d exhibitions with LLM", len(htmlBlocks))
		return nil, fmt.Errorf("dry run mode")
	}

	batchSchema := schemas.ExhibitionsBatchSchema()

	systemPrompt := fmt.Sprintf(`You are a data extraction specialist for a German cultural festival website.
Extract ALL exhibitions from the provided HTML. The venue is "%s".

Important rules:
1. Extract artist name from the first <p> tag
2. Extract title from <b> tag
3. Extract description from <em> tag
4. Return ALL exhibitions found in the HTML
5. Each item in the HTML is separated by "---EVENT_SEPARATOR---"`, venueName)

	result, err := p.client.ExtractWithSchema(combinedHTML, batchSchema, systemPrompt, p.logger)
	if err != nil {
		return nil, err
	}

	p.totalTokens += result.PromptTokens + result.CompTokens

	// Cache response
	if p.cache != nil {
		if err := p.cache.Set(cacheKey, result.Content); err != nil {
			p.logger.Printf("[CACHE WARN] Failed to cache exhibitions batch: %v", err)
		}
	}

	return p.unmarshalExhibitionsBatch(result.Content, venueID, venueName)
}

// generateCacheKey creates a cache key from content
func (p *LLMParser) generateCacheKey(prefix, content string) string {
	hash := sha256.Sum256([]byte(prefix + ":" + content))
	return hex.EncodeToString(hash[:16])
}

// unmarshalVenue converts JSON response to storage.Venue
func (p *LLMParser) unmarshalVenue(jsonStr string) (*storage.Venue, error) {
	var result struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Address     struct {
			Street     string `json:"street"`
			PostalCode string `json:"postalCode"`
			City       string `json:"city"`
		} `json:"address"`
		Contact struct {
			Phone   string `json:"phone"`
			Email   string `json:"email"`
			Website string `json:"website"`
		} `json:"contact"`
		BikeRoute string `json:"bikeRoute"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal venue: %w", err)
	}

	// Convert to storage.Venue
	venue := &storage.Venue{
		ID:          GenerateID("venue", result.Name),
		Name:        result.Name,
		Description: result.Description,
		Address: storage.Address{
			Street:     result.Address.Street,
			PostalCode: result.Address.PostalCode,
			City:       result.Address.City,
		},
		Contact: storage.Contact{
			Phone:   result.Contact.Phone,
			Email:   result.Contact.Email,
			Website: result.Contact.Website,
		},
		BikeRoute: result.BikeRoute,
	}

	return venue, nil
}

// unmarshalEventsBatch converts JSON response to []storage.Event
func (p *LLMParser) unmarshalEventsBatch(jsonStr, venueID, venueName string) ([]storage.Event, error) {
	var result struct {
		Events []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Artist      string `json:"artist"`
			Dates       []struct {
				Date      string `json:"date"`
				StartTime string `json:"startTime"`
				EndTime   string `json:"endTime"`
			} `json:"dates"`
			Admission string `json:"admission"`
		} `json:"events"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}

	var events []storage.Event
	for _, e := range result.Events {
		// Create separate event for each date
		for _, d := range e.Dates {
			startTime, endTime := d.StartTime, d.EndTime
			if startTime == "" && endTime != "" {
				startTime, endTime = endTime, ""
			}
			event := storage.Event{
				ID:          GenerateID("event", e.Title+venueID+d.Date+startTime),
				Title:       e.Title,
				Description: e.Description,
				VenueID:     venueID,
				VenueName:   venueName,
				Date:        d.Date,
				StartTime:   startTime,
				EndTime:     endTime,
				Category:    e.Artist,
				Admission:   e.Admission,
			}
			events = append(events, event)
		}
	}

	return events, nil
}

// unmarshalExhibitionsBatch converts JSON response to []storage.Exhibition
func (p *LLMParser) unmarshalExhibitionsBatch(jsonStr, venueID, venueName string) ([]storage.Exhibition, error) {
	var result struct {
		Exhibitions []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Artist      string `json:"artist"`
		} `json:"exhibitions"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal exhibitions: %w", err)
	}

	var exhibitions []storage.Exhibition
	for _, e := range result.Exhibitions {
		exhibition := storage.Exhibition{
			ID:          GenerateID("exhibition", e.Title+venueID),
			Title:       e.Title,
			Description: e.Description,
			Artist:      e.Artist,
			VenueID:     venueID,
			VenueName:   venueName,
		}
		exhibitions = append(exhibitions, exhibition)
	}

	return exhibitions, nil
}

// Helper function to check if a string is empty or whitespace
func isEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}
