package crawler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/musche/klp/internal/crawler/llm"
	"github.com/musche/klp/internal/storage"
)

// CategorizerItem is a minimal payload describing one item to classify.
type CategorizerItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Artist      string `json:"artist,omitempty"`
}

// Categorizer assigns one of storage.EventCategories to each event/exhibition
// via the LLM. Results are cached by content hash so reruns are cheap.
type Categorizer struct {
	client    *llm.OpenRouterClient
	cache     *llm.Cache
	logger    *log.Logger
	batchSize int
	tokens    int
}

// NewCategorizer creates a categorizer using the same OpenRouter infrastructure
// as the main LLM parser.
func NewCategorizer(apiKey, model string, cache *llm.Cache, batchSize int, logger *log.Logger) *Categorizer {
	if batchSize <= 0 {
		batchSize = 25
	}
	return &Categorizer{
		client:    llm.NewClient(apiKey, model),
		cache:     cache,
		logger:    logger,
		batchSize: batchSize,
	}
}

// TotalTokens returns the cumulative token count across all batches.
func (c *Categorizer) TotalTokens() int { return c.tokens }

// CategorizeEvents fills the Category field on every event that does not
// already have a valid category. Mutates events in place.
//
// Cleanup: earlier versions of the LLM parser wrote the artist name into
// Event.Category. Any value in Category that is not a known category is moved
// into Artist (unless Artist is already set) and the Category is cleared
// before classification.
func (c *Categorizer) CategorizeEvents(events []storage.Event) error {
	seenIDs := make(map[string]bool, len(events))
	var pending []CategorizerItem
	migrated := 0
	for i := range events {
		ev := &events[i]
		if ev.Category != "" && !storage.IsValidEventCategory(ev.Category) {
			if ev.Artist == "" {
				ev.Artist = ev.Category
			}
			ev.Category = ""
			migrated++
		}
		if storage.IsValidEventCategory(ev.Category) {
			continue
		}
		// Only send each unique ID once to the LLM; assignments later are
		// applied to ALL slice entries sharing that ID (the data is not
		// strictly unique-by-ID).
		if seenIDs[ev.ID] {
			continue
		}
		seenIDs[ev.ID] = true
		pending = append(pending, CategorizerItem{
			ID:          ev.ID,
			Title:       ev.Title,
			Description: ev.Description,
			Artist:      ev.Artist,
		})
	}
	if migrated > 0 {
		c.logger.Printf("[CATEGORIZER] Migrated %d events: legacy artist-in-category moved to Artist", migrated)
	}

	c.logger.Printf("[CATEGORIZER] %d unique events need a category (skipping %d already classified)",
		len(pending), len(events)-len(pending))

	assignments, err := c.classifyBatches(pending)
	if err != nil {
		return err
	}

	for i := range events {
		ev := &events[i]
		if storage.IsValidEventCategory(ev.Category) {
			continue
		}
		if cat, ok := assignments[ev.ID]; ok {
			ev.Category = cat
		}
	}
	return nil
}

// CategorizeExhibitions fills the Category field on every exhibition that does
// not already have a valid category. Mutates exhibitions in place.
func (c *Categorizer) CategorizeExhibitions(exhibitions []storage.Exhibition) error {
	seenIDs := make(map[string]bool, len(exhibitions))
	var pending []CategorizerItem
	for i := range exhibitions {
		ex := &exhibitions[i]
		if storage.IsValidEventCategory(ex.Category) {
			continue
		}
		if seenIDs[ex.ID] {
			continue
		}
		seenIDs[ex.ID] = true
		pending = append(pending, CategorizerItem{
			ID:          ex.ID,
			Title:       ex.Title,
			Description: ex.Description,
			Artist:      ex.Artist,
		})
	}

	c.logger.Printf("[CATEGORIZER] %d unique exhibitions need a category (skipping %d already classified)",
		len(pending), len(exhibitions)-len(pending))

	assignments, err := c.classifyBatches(pending)
	if err != nil {
		return err
	}

	for i := range exhibitions {
		ex := &exhibitions[i]
		if storage.IsValidEventCategory(ex.Category) {
			continue
		}
		if cat, ok := assignments[ex.ID]; ok {
			ex.Category = cat
		}
	}
	return nil
}

// classifyBatches splits items into batches, sends each through the LLM, and
// merges the result map. Failed batches are logged but do not abort the run.
func (c *Categorizer) classifyBatches(items []CategorizerItem) (map[string]string, error) {
	out := make(map[string]string, len(items))
	if len(items) == 0 {
		return out, nil
	}

	for start := 0; start < len(items); start += c.batchSize {
		end := start + c.batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[start:end]
		c.logger.Printf("[CATEGORIZER] Batch %d-%d / %d", start+1, end, len(items))

		got, err := c.classifyBatch(batch)
		if err != nil {
			c.logger.Printf("[CATEGORIZER ERROR] Batch %d-%d failed: %v", start+1, end, err)
			continue
		}
		for k, v := range got {
			out[k] = v
		}
	}
	return out, nil
}

// classifyBatch sends one batch through the LLM and returns id -> category.
func (c *Categorizer) classifyBatch(batch []CategorizerItem) (map[string]string, error) {
	payload, err := json.Marshal(batch)
	if err != nil {
		return nil, fmt.Errorf("marshal batch: %w", err)
	}

	cacheKey := c.cacheKey(payload)
	if c.cache != nil {
		if cached, ok := c.cache.Get(cacheKey); ok {
			c.logger.Println("[CATEGORIZER CACHE HIT]")
			return c.parseAssignments(cached, batch)
		}
	}

	schema := categorizerSchema()
	systemPrompt := buildCategorizerSystemPrompt()

	result, err := c.client.ExtractWithSchema(string(payload), schema, systemPrompt, c.logger)
	if err != nil {
		return nil, err
	}
	c.tokens += result.PromptTokens + result.CompTokens

	if c.cache != nil {
		if err := c.cache.Set(cacheKey, result.Content); err != nil {
			c.logger.Printf("[CATEGORIZER WARN] cache write failed: %v", err)
		}
	}
	return c.parseAssignments(result.Content, batch)
}

// parseAssignments validates the LLM response and falls back to Sonstiges when
// an assignment is missing or invalid.
func (c *Categorizer) parseAssignments(jsonStr string, batch []CategorizerItem) (map[string]string, error) {
	var resp struct {
		Assignments []struct {
			ID       string `json:"id"`
			Category string `json:"category"`
		} `json:"assignments"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, fmt.Errorf("unmarshal categorizer response: %w", err)
	}

	out := make(map[string]string, len(batch))
	for _, a := range resp.Assignments {
		if !storage.IsValidEventCategory(a.Category) {
			c.logger.Printf("[CATEGORIZER WARN] invalid category %q for %s — falling back to Sonstiges",
				a.Category, a.ID)
			out[a.ID] = "Sonstiges"
			continue
		}
		out[a.ID] = a.Category
	}
	// Fill in any item the LLM omitted so we don't loop on it next run.
	for _, item := range batch {
		if _, ok := out[item.ID]; !ok {
			c.logger.Printf("[CATEGORIZER WARN] missing assignment for %s — falling back to Sonstiges", item.ID)
			out[item.ID] = "Sonstiges"
		}
	}
	return out, nil
}

func (c *Categorizer) cacheKey(payload []byte) string {
	prefix := "categorize:v1:" + strings.Join(storage.EventCategories, ",") + ":"
	hash := sha256.Sum256(append([]byte(prefix), payload...))
	return hex.EncodeToString(hash[:16])
}

func buildCategorizerSystemPrompt() string {
	return fmt.Sprintf(`You classify cultural-festival items (events and exhibitions) into exactly ONE of these German categories:

%s

Guidelines:
- Musik: concerts, DJ sets, choirs, sing-alongs, instrumental music
- Theater & Performance: theater plays, comedy, clown, circus, cabaret, performance art, audiovisual shows
- Wort & Vortrag: readings (Lesungen), poetry slams, talks, lectures, discussions, debates
- Kunst & Workshop: visual-art making, painting, sculpture, crafts, hands-on creative workshops, exhibitions of art
- Tanz & Bewegung: dance performances, dance workshops, yoga, movement, embodiment
- Film: film screenings, cinema
- Kulinarisches: food-centric events, dinners, tastings
- Kinder & Familie: events explicitly aimed at children or families
- Sonstiges: anything that clearly fits none of the above

Pick the single best fit. If an event mixes categories, pick the dominant theme (e.g. "Konzertlesung" -> Wort & Vortrag if it's primarily a reading, Musik if it's primarily a concert). Use Sonstiges as a last resort only.

Return an object {"assignments": [{"id": "...", "category": "..."}]} covering every input id exactly once. Categories must match the list above EXACTLY (including capitalization and the ampersand).`,
		strings.Join(storage.EventCategories, "\n- "))
}

func categorizerSchema() map[string]interface{} {
	return map[string]interface{}{
		"name":   "category_assignment",
		"strict": true,
		"schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"assignments": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"id":       map[string]interface{}{"type": "string"},
							"category": map[string]interface{}{"type": "string", "enum": storage.EventCategories},
						},
						"required":             []string{"id", "category"},
						"additionalProperties": false,
					},
				},
			},
			"required":             []string{"assignments"},
			"additionalProperties": false,
		},
	}
}
