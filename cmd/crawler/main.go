package main

import (
	"flag"
	"log"
	"os"

	"github.com/musche/klp/internal/crawler"
	"github.com/musche/klp/internal/storage"
)

func main() {
	// Parse command-line flags
	outputDir := flag.String("output", "./data", "Output directory for JSON files")
	verbose := flag.Bool("verbose", false, "Verbose logging")
	skipGeo := flag.Bool("skip-geocoding", false, "Skip geocoding step")
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
	c := crawler.NewCrawler(logger)
	defer c.Close()

	// Crawl all data
	logger.Println("\n[STEP 1/2] Crawling all venues, events, and exhibitions...")
	venues, events, exhibitions, err := c.CrawlAll()
	if err != nil {
		log.Fatalf("Failed to crawl data: %v", err)
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
	logger.Println("\nData saved successfully!")
	logger.Println("Run the server with: go run cmd/server/main.go")
}
