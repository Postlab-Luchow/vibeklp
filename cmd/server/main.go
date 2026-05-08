package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/musche/klp/internal/api"
	"github.com/musche/klp/internal/storage"
)

func main() {
	// Parse command-line flags
	port := flag.Int("port", 8081, "Server port")
	dataDir := flag.String("data", "./data", "Data directory")
	staticDir := flag.String("static", "./web/static", "Static files directory")
	templatesDir := flag.String("templates", "./web/templates", "Templates directory")
	flag.Parse()

	log.Printf("=== Kulturelle Landpartie Server ===")
	log.Printf("Port: %d", *port)
	log.Printf("Data directory: %s", *dataDir)

	// Create storage
	store := storage.NewStorage(*dataDir)

	// Verify data files exist
	if _, err := store.LoadVenues(); err != nil {
		log.Printf("Warning: Could not load venues: %v", err)
		log.Printf("Run the crawler first: go run cmd/crawler/main.go")
	}

	// Setup routes
	router := api.SetupRoutes(store)

	// Add middleware
	router.Use(api.LoggingMiddleware)
	router.Use(api.CORSMiddleware)

	// Serve static files. Set Cache-Control: no-cache so browsers must
	// revalidate with If-Modified-Since on every request — content is still
	// served from the local disk cache when unchanged (304), but fresh JS/CSS
	// is picked up immediately after a redeploy without needing a hard reload.
	staticPath := filepath.Join(*staticDir)
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath)))
	router.PathPrefix("/static/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		staticHandler.ServeHTTP(w, r)
	}))

	// Serve index.html for root
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexPath := filepath.Join(*templatesDir, "index.html")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			// If index.html doesn't exist yet, show a simple message
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Kulturelle Landpartie</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px;">
    <h1>🎨 Kulturelle Landpartie</h1>
    <p>Der Server läuft erfolgreich!</p>
    <h2>API-Endpunkte:</h2>
    <ul>
        <li><a href="/api/venues">/api/venues</a> - Alle Veranstaltungsorte</li>
        <li><a href="/api/events">/api/events</a> - Alle Veranstaltungen</li>
        <li><a href="/api/exhibitions">/api/exhibitions</a> - Alle Ausstellungen</li>
        <li><a href="/api/calendar">/api/calendar</a> - Kalenderansicht</li>
        <li><a href="/api/categories">/api/categories</a> - Kategorien</li>
        <li><a href="/api/stats">/api/stats</a> - Statistiken</li>
    </ul>
    <p><em>Das Frontend wird in Kürze verfügbar sein.</em></p>
</body>
</html>
			`)
			return
		}
		http.ServeFile(w, r, indexPath)
	})

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server starting on http://localhost%s", addr)
	log.Printf("API available at http://localhost%s/api", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
