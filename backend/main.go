package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"site-availability/config"
	"site-availability/handlers"
	"site-availability/scraping"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	serverUptime = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "server_uptime_seconds",
			Help: "Total uptime of the server in seconds",
		},
	)
)

var cfg *config.Config // Global configuration variable

func main() {
	// Load configuration
	var err error
	cfg, err = loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Prometheus metrics
	initPrometheusMetrics()

	// Start background status fetcher
	startBackgroundStatusFetcher()

	// Setup HTTP routes and handlers
	setupRoutes()

	// Start server and handle shutdown gracefully
	startServer()
}

// Load configuration file
func loadConfig() (*config.Config, error) {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config.yaml"
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// Initialize Prometheus metrics
func initPrometheusMetrics() {
	prometheus.MustRegister(serverUptime)
	go func() {
		for range time.Tick(time.Second) {
			serverUptime.Inc()
		}
	}()
}

// Setup HTTP routes and handlers
func setupRoutes() {
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) { handlers.GetAppStatus(w, r, cfg) })
	http.HandleFunc("/healthz", livenessProbe)
	http.HandleFunc("/readyz", readinessProbe)
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", http.FileServer(http.Dir("./static")))
}
func statusFetcher(checker *scraping.DefaultPrometheusChecker) {
	log.Println("Fetching app statuses...")

	var newStatuses []handlers.AppStatus

	// Update app statuses
	for _, app := range cfg.Apps {
		// Get the status using the CheckAppStatus function, which returns an AppStatus struct
		appStatus := scraping.CheckAppStatus(app, checker)

		// Append the updated AppStatus to the newStatuses slice
		newStatuses = append(newStatuses, appStatus)
	}

	// Update the cache with the new statuses
	handlers.UpdateAppStatus(newStatuses)

	log.Println("App statuses updated.")
}

func startBackgroundStatusFetcher() {
	scrapeInterval, err := time.ParseDuration(cfg.ScrapeInterval)
	if err != nil {
		log.Fatalf("Failed to parse ScrapeInterval: %v", err)
	}
	go func() {
		// Create a ticker with the desired scrape interval
		ticker := time.NewTicker(scrapeInterval)
		defer ticker.Stop()

		checker := &scraping.DefaultPrometheusChecker{}
		statusFetcher(checker)
		// Use for-range to listen to the ticker channel
		for range ticker.C {
		  statusFetcher(checker)
		}
	}()
}

// Start the HTTP server and handle graceful shutdown
func startServer() {
	port := getServerPort()

	srv := &http.Server{Addr: port}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-sigChan
	gracefulShutdown(srv)
}

// Get server port from environment variable or default to ":8080"
func getServerPort() string {
	port := ":8080"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	return port
}

// Gracefully shut down the server
func gracefulShutdown(srv *http.Server) {
	log.Println("Shutdown signal received, shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// Liveness probe handler
func livenessProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func readinessProbe(w http.ResponseWriter, r *http.Request) {
	if handlers.IsAppStatusCacheEmpty() {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("NOT READY"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("READY"))
}

