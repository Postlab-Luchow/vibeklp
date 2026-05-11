package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/musche/klp/internal/crawler"
	"github.com/musche/klp/internal/crawler/llm"
	"github.com/musche/klp/internal/storage"
)

func main() {
	// Parse command-line flags
	outputDir := flag.String("output", "./data", "Output directory for JSON files")
	verbose := flag.Bool("verbose", false, "Verbose logging")
	skipGeo := flag.Bool("skip-geocoding", false, "Skip geocoding step")
	googleAPIKey := flag.String("google-api-key", "", "Google Maps Geocoding API key (optional)")

	// LLM flags
	useLLM := flag.Bool("use-llm", false, "Use LLM for parsing HTML (requires OPENROUTER_API_KEY env var)")
	openRouterModel := flag.String("openrouter-model", "google/gemini-3-flash-preview", "OpenRouter model to use")
	llmCacheDir := flag.String("llm-cache-dir", "./.llm_cache", "Directory for LLM response cache")
	llmCacheTTL := flag.Duration("llm-cache-ttl", 24*time.Hour, "Cache TTL for LLM responses")
	llmBatchSize := flag.Int("llm-batch-size", 5, "Number of items to process per LLM API call")
	dryRun := flag.Bool("dry-run", false, "Show what would be extracted without making API calls")

	// Crawl caching flags
	crawlCacheDir := flag.String("crawl-cache-dir", "./.crawl_cache", "Directory for crawled HTML cache")
	useCrawlCache := flag.Bool("use-crawl-cache", true, "Cache crawled HTML to avoid re-fetching on retry")
	clearCrawlCache := flag.Bool("clear-crawl-cache", false, "Clear crawl cache before starting")
	progressDir := flag.String("progress-dir", "./.crawl_cache", "Directory for progress tracking")

	// Parse cached mode - parse existing cache without fetching
	parseCached := flag.Bool("parse-cached", false, "Parse cached HTML with LLM without fetching (requires --use-llm and existing cache)")
	fetchOnly := flag.Bool("fetch-only", false, "Only fetch and cache HTML, skip parsing")

	// Categorize mode - assign event categories to existing events.json/exhibitions.json
	categorize := flag.Bool("categorize", false, "Assign categories to existing events and exhibitions via LLM, then exit (no crawl)")
	categorizeBatch := flag.Int("categorize-batch-size", 25, "Items per LLM call in --categorize mode")

	// Source selection
	source := flag.String("source", "klp", "Source(s) to crawl: klp, wendlandpartie, landgang, all")

	flag.Parse()

	// Setup logger
	logger := log.New(os.Stdout, "", log.LstdFlags)
	if !*verbose {
		logger.SetOutput(os.Stderr)
	}

	logger.Println("=== Kulturelle Landpartie Crawler ===")
	logger.Printf("Output directory: %s", *outputDir)

	// Create storage
	store := storage.NewStorage(*outputDir)
	if err := store.EnsureDataDir(); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Create crawler
	var c *crawler.Crawler
	if *useLLM {
		apiKey := os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" && !*dryRun {
			log.Fatal("OPENROUTER_API_KEY environment variable is required when using --use-llm")
		}

		var cache *llm.Cache
		if *llmCacheDir != "" && *llmCacheTTL > 0 {
			cache = llm.NewCache(*llmCacheDir, *llmCacheTTL)
			if err := cache.EnsureDir(); err != nil {
				logger.Printf("[WARN] Failed to create LLM cache directory: %v", err)
				cache = nil
			}
		}

		c = crawler.NewCrawlerWithLLM(logger, apiKey, *openRouterModel, cache, *llmBatchSize, *dryRun)
		logger.Printf("Using LLM parsing with model: %s", *openRouterModel)
		if cache != nil {
			logger.Printf("LLM cache enabled: %s", *llmCacheDir)
		}
		if *dryRun {
			logger.Println("DRY RUN MODE: No API calls will be made")
		}
	} else {
		c = crawler.NewCrawler(logger)
	}

	if *googleAPIKey != "" {
		c.SetGoogleMapsGeocoder(*googleAPIKey)
		logger.Println("Using Google Maps Geocoding API")
	}

	// Setup crawl cache and progress tracking
	if *useCrawlCache {
		c.SetCrawlCache(*crawlCacheDir)
		c.SetProgressTracker(*progressDir)
	}

	if *clearCrawlCache && *crawlCacheDir != "" {
		logger.Printf("[INFO] Clearing crawl cache: %s", *crawlCacheDir)
		cache := crawler.NewCrawlCache(*crawlCacheDir)
		if err := cache.Clear(); err != nil {
			logger.Printf("[WARN] Failed to clear crawl cache: %v", err)
		} else {
			logger.Println("[INFO] Crawl cache cleared")
		}
	}

	defer c.Close()

	var venues []storage.Venue
	var events []storage.Event
	var exhibitions []storage.Exhibition
	var err error

	// Standalone post-processing: categorize already-crawled events/exhibitions.
	// Does not crawl or geocode — only reads/writes data JSON.
	if *categorize {
		apiKey := os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			log.Fatal("OPENROUTER_API_KEY environment variable is required for --categorize")
		}

		var cache *llm.Cache
		if *llmCacheDir != "" && *llmCacheTTL > 0 {
			cache = llm.NewCache(*llmCacheDir, *llmCacheTTL)
			if err := cache.EnsureDir(); err != nil {
				logger.Printf("[WARN] Failed to create LLM cache directory: %v", err)
				cache = nil
			}
		}

		categorizer := crawler.NewCategorizer(apiKey, *openRouterModel, cache, *categorizeBatch, logger)

		logger.Println("\n[MODE] Categorizing existing events and exhibitions...")
		loadedEvents, err := store.LoadEvents()
		if err != nil {
			log.Fatalf("Failed to load events.json: %v", err)
		}
		loadedExhibitions, err := store.LoadExhibitions()
		if err != nil {
			log.Fatalf("Failed to load exhibitions.json: %v", err)
		}

		if err := categorizer.CategorizeEvents(loadedEvents); err != nil {
			log.Fatalf("Categorizing events failed: %v", err)
		}
		if err := categorizer.CategorizeExhibitions(loadedExhibitions); err != nil {
			log.Fatalf("Categorizing exhibitions failed: %v", err)
		}

		if err := store.SaveEvents(loadedEvents); err != nil {
			log.Fatalf("Failed to save events: %v", err)
		}
		if err := store.SaveExhibitions(loadedExhibitions); err != nil {
			log.Fatalf("Failed to save exhibitions: %v", err)
		}

		logger.Println("\n=== Categorization Complete ===")
		logger.Printf("Events:      %d", len(loadedEvents))
		logger.Printf("Exhibitions: %d", len(loadedExhibitions))
		logger.Printf("LLM tokens used: %d", categorizer.TotalTokens())
		return
	}

	// Determine which sources to crawl. Each source produces its own
	// (venues, events, exhibitions) slice; results are concatenated below.
	sourcesToCrawl, err := parseSourceFlag(*source)
	if err != nil {
		log.Fatal(err)
	}
	logger.Printf("Sources: %v", sourcesToCrawl)

	// Source-specific modes (parseCached / fetchOnly) only make sense for KLP.
	if *parseCached || *fetchOnly {
		if !sliceContains(sourcesToCrawl, storage.SourceKLP) || len(sourcesToCrawl) != 1 {
			log.Fatal("--parse-cached and --fetch-only only work with --source klp")
		}
	}

	for _, src := range sourcesToCrawl {
		var sv []storage.Venue
		var se []storage.Event
		var sx []storage.Exhibition
		switch src {
		case storage.SourceKLP:
			if *parseCached {
				if !*useLLM {
					log.Fatal("--parse-cached requires --use-llm flag")
				}
				if !c.HasCrawlCache() {
					log.Fatal("--parse-cached requires crawl cache. Use --use-crawl-cache flag")
				}
				logger.Println("\n[KLP] Parsing cached HTML with LLM (no fetching)...")
				sv, se, sx, err = c.ParseCachedVenues()
			} else if *fetchOnly {
				logger.Println("\n[KLP] Fetch-only mode: caching HTML without parsing...")
				if !c.HasCrawlCache() {
					log.Fatal("--fetch-only requires crawl cache. Use --use-crawl-cache flag")
				}
				_, _, _, err = c.CrawlAll()
				if err != nil {
					log.Fatalf("Failed to fetch KLP data: %v", err)
				}
				logger.Println("\n=== Fetch Complete ===")
				logger.Println("HTML cached successfully. Run with --parse-cached to parse with LLM.")
				return
			} else {
				logger.Println("\n[KLP] Crawling kulturelle-landpartie.de ...")
				sv, se, sx, err = c.CrawlAll()
			}
		case storage.SourceWendlandpartie:
			logger.Println("\n[wendlandpartie] Crawling wendlandpartie.de via Tribe Events REST API ...")
			sv, se, sx, err = crawler.CrawlWendlandpartie(logger)
		case storage.SourceLandgang:
			logger.Println("\n[landgang] Crawling landgang-wendland.de venue pages ...")
			sv, se, sx, err = crawler.CrawlLandgang(c)
		default:
			log.Fatalf("Unknown source: %s", src)
		}
		if err != nil {
			log.Fatalf("Failed to crawl %s: %v", src, err)
		}
		tagSource(src, sv, se, sx)
		venues = append(venues, sv...)
		events = append(events, se...)
		exhibitions = append(exhibitions, sx...)
		logger.Printf("[%s] +%d venues, +%d events, +%d exhibitions", src, len(sv), len(se), len(sx))
	}

	logger.Printf("✓ Crawled %d venues, %d events, %d exhibitions (across %d source(s))",
		len(venues), len(events), len(exhibitions), len(sourcesToCrawl))

	// Geocode addresses
	if !*skipGeo {
		logger.Println("\n[STEP 2/2] Geocoding addresses...")
		if err := c.GeocodeVenues(venues); err != nil {
			log.Printf("Warning: Geocoding completed with errors: %v", err)
		}

		// Validate venues
		validVenues := 0
		for _, v := range venues {
			if err := v.Validate(); err == nil {
				validVenues++
			} else {
				logger.Printf("[WARN] Invalid venue %s: %v", v.Name, err)
			}
		}
		logger.Printf("✓ Geocoded %d/%d venues successfully", validVenues, len(venues))
	} else {
		logger.Println("\n[STEP 2/2] Skipping geocoding (--skip-geocoding flag set)")
	}

	// Merge with existing data on disk: drop any prior records belonging to
	// the source(s) we just crawled, then append the fresh records. This lets
	// `-source wendlandpartie` refresh only those records without touching
	// klp/landgang data, and vice versa.
	logger.Println("\n[SAVING] Merging into existing JSON files...")
	mergedVenues, mergedEvents, mergedExhibitions, err := mergeWithExisting(
		store, sourcesToCrawl, venues, events, exhibitions,
	)
	if err != nil {
		log.Fatalf("Failed to merge with existing data: %v", err)
	}

	if err := store.SaveVenues(mergedVenues); err != nil {
		log.Fatalf("Failed to save venues: %v", err)
	}
	logger.Printf("✓ Saved %d venues to %s/venues.json", len(mergedVenues), *outputDir)

	if err := store.SaveEvents(mergedEvents); err != nil {
		log.Fatalf("Failed to save events: %v", err)
	}
	logger.Printf("✓ Saved %d events to %s/events.json", len(mergedEvents), *outputDir)

	if err := store.SaveExhibitions(mergedExhibitions); err != nil {
		log.Fatalf("Failed to save exhibitions: %v", err)
	}
	logger.Printf("✓ Saved %d exhibitions to %s/exhibitions.json", len(mergedExhibitions), *outputDir)

	// Summary
	logger.Println("\n=== Crawl Complete ===")
	logger.Printf("This run added/updated for sources %v:", sourcesToCrawl)
	logger.Printf("  Venues:      %d", len(venues))
	logger.Printf("  Events:      %d", len(events))
	logger.Printf("  Exhibitions: %d", len(exhibitions))
	logger.Printf("On disk total: %d venues, %d events, %d exhibitions",
		len(mergedVenues), len(mergedEvents), len(mergedExhibitions))

	// Show LLM stats if used
	if *useLLM && c.GetLLMParser() != nil {
		totalTokens := c.GetLLMParser().GetTotalTokens()
		logger.Printf("\nLLM Usage:")
		logger.Printf("  Total tokens: %d", totalTokens)
		logger.Printf("  Est. cost: $%.4f", float64(totalTokens)*0.00000075) // gpt-4o-mini pricing
	}

	logger.Println("\nData saved successfully!")
	logger.Println("Run the server with: go run cmd/server/main.go")
}

// parseSourceFlag resolves --source into a list of source identifiers.
func parseSourceFlag(flag string) ([]string, error) {
	switch flag {
	case "klp":
		return []string{storage.SourceKLP}, nil
	case "wendlandpartie":
		return []string{storage.SourceWendlandpartie}, nil
	case "landgang":
		return []string{storage.SourceLandgang}, nil
	case "all":
		return []string{storage.SourceKLP, storage.SourceWendlandpartie, storage.SourceLandgang}, nil
	default:
		return nil, fmt.Errorf("unknown --source %q (want one of: klp, wendlandpartie, landgang, all)", flag)
	}
}

func sliceContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// tagSource sets the Source field on items whose source is not already set.
// Crawlers can set Source themselves (e.g. CrawlWendlandpartie does); this is
// a safety net mostly for the legacy KLP crawler which knows nothing of sources.
func tagSource(src string, venues []storage.Venue, events []storage.Event, exhibitions []storage.Exhibition) {
	for i := range venues {
		if venues[i].Source == "" {
			venues[i].Source = src
		}
	}
	for i := range events {
		if events[i].Source == "" {
			events[i].Source = src
		}
	}
	for i := range exhibitions {
		if exhibitions[i].Source == "" {
			exhibitions[i].Source = src
		}
	}
}

// mergeWithExisting loads the current JSON files, drops everything tagged with
// a freshly-crawled source, and appends the new records. Records missing a
// Source field count as legacy KLP data so re-running `-source klp` cleans them
// up automatically.
func mergeWithExisting(
	store *storage.Storage,
	freshSources []string,
	venues []storage.Venue,
	events []storage.Event,
	exhibitions []storage.Exhibition,
) ([]storage.Venue, []storage.Event, []storage.Exhibition, error) {
	freshSet := make(map[string]bool, len(freshSources))
	for _, s := range freshSources {
		freshSet[s] = true
	}
	isFresh := func(s string) bool {
		if s == "" {
			s = storage.SourceKLP
		}
		return freshSet[s]
	}

	existingV, err := store.LoadVenues()
	if err != nil {
		// First run: file doesn't exist yet — just use the fresh data.
		existingV = nil
	}
	existingE, err := store.LoadEvents()
	if err != nil {
		existingE = nil
	}
	existingX, err := store.LoadExhibitions()
	if err != nil {
		existingX = nil
	}

	outV := make([]storage.Venue, 0, len(existingV)+len(venues))
	for _, v := range existingV {
		if !isFresh(v.Source) {
			outV = append(outV, v)
		}
	}
	outV = append(outV, venues...)

	outE := make([]storage.Event, 0, len(existingE)+len(events))
	for _, e := range existingE {
		if !isFresh(e.Source) {
			outE = append(outE, e)
		}
	}
	outE = append(outE, events...)

	outX := make([]storage.Exhibition, 0, len(existingX)+len(exhibitions))
	for _, x := range existingX {
		if !isFresh(x.Source) {
			outX = append(outX, x)
		}
	}
	outX = append(outX, exhibitions...)

	return outV, outE, outX, nil
}
