package main

import (
	"flag"
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
	openRouterModel := flag.String("openrouter-model", "openai/gpt-4o-mini", "OpenRouter model to use")
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

	// Handle different modes
	if *parseCached {
		// Mode: Parse cached HTML with LLM (no fetching)
		if !*useLLM {
			log.Fatal("--parse-cached requires --use-llm flag")
		}
		if !c.HasCrawlCache() {
			log.Fatal("--parse-cached requires crawl cache. Use --use-crawl-cache flag")
		}

		logger.Println("\n[MODE] Parsing cached HTML with LLM (no fetching)...")
		venues, events, exhibitions, err = c.ParseCachedVenues()
		if err != nil {
			log.Fatalf("Failed to parse cached data: %v", err)
		}
	} else if *fetchOnly {
		// Mode: Only fetch and cache HTML, skip parsing
		logger.Println("\n[MODE] Fetch-only mode: caching HTML without parsing...")
		if !c.HasCrawlCache() {
			log.Fatal("--fetch-only requires crawl cache. Use --use-crawl-cache flag")
		}

		// Just run CrawlAll but don't save results
		_, _, _, err = c.CrawlAll()
		if err != nil {
			log.Fatalf("Failed to fetch data: %v", err)
		}

		logger.Println("\n=== Fetch Complete ===")
		logger.Println("HTML cached successfully. Run with --parse-cached to parse with LLM.")
		return
	} else {
		// Mode: Normal crawl (fetch + parse)
		logger.Println("\n[STEP 1/2] Crawling all venues, events, and exhibitions...")
		venues, events, exhibitions, err = c.CrawlAll()
		if err != nil {
			log.Fatalf("Failed to crawl data: %v", err)
		}
	}

	logger.Printf("✓ Crawled %d venues, %d events, %d exhibitions",
		len(venues), len(events), len(exhibitions))

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

	// Save data
	logger.Println("\n[SAVING] Writing data to JSON files...")

	if err := store.SaveVenues(venues); err != nil {
		log.Fatalf("Failed to save venues: %v", err)
	}
	logger.Printf("✓ Saved venues to %s/venues.json", *outputDir)

	if err := store.SaveEvents(events); err != nil {
		log.Fatalf("Failed to save events: %v", err)
	}
	logger.Printf("✓ Saved events to %s/events.json", *outputDir)

	if err := store.SaveExhibitions(exhibitions); err != nil {
		log.Fatalf("Failed to save exhibitions: %v", err)
	}
	logger.Printf("✓ Saved exhibitions to %s/exhibitions.json", *outputDir)

	// Summary
	logger.Println("\n=== Crawl Complete ===")
	logger.Printf("Venues:      %d", len(venues))
	logger.Printf("Events:      %d", len(events))
	logger.Printf("Exhibitions: %d", len(exhibitions))

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
