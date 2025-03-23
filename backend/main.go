package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"site-avilability/config"
	"site-avilability/handlers" // Make sure this import exists
	"site-avilability/prometheus"
)

var appStatuses map[string]handlers.AppStatus // Global variable to store app statuses

func main() {
	// Load configuration from the YAML file
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Parse ScrapeInterval into a time.Duration
	scrapeInterval, err := time.ParseDuration(cfg.ScrapeInterval)
	if err != nil {
		log.Fatalf("Failed to parse ScrapeInterval: %v", err)
	}

	// Initialize the app statuses map
	appStatuses = make(map[string]handlers.AppStatus)

	// Set up API routes
	http.HandleFunc("/api/status", handlers.GetAppStatus)

	// Start background task to fetch app status every ScrapeInterval
	go startBackgroundStatusFetcher(scrapeInterval, cfg)

	// Serve React build files
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Start the HTTP server
	port := ":8080"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	log.Printf("Server starting on %s", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// Fetch and update app statuses in the background
func startBackgroundStatusFetcher(scrapeInterval time.Duration, cfg *config.Config) {
	ticker := time.NewTicker(scrapeInterval)
	defer ticker.Stop()

	// Default Prometheus checker
	checker := &prometheus.DefaultPrometheusChecker{}

	for {
		select {
		case <-ticker.C:
			log.Println("Fetching app statuses...")

			// Update app statuses
			for _, app := range cfg.Apps {
				status := prometheus.CheckAppStatus(app, checker)
				handlers.UpdateAppStatusCache(app.Name, status)
			}

			log.Println("App statuses updated.")
		}
	}
}
