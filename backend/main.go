package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/logging"
	"site-availability/metrics"
	"site-availability/scraping"
	"syscall"
	"time"
)

var cfg *config.Config // Global configuration variable

func main() {
	// Attempt to initialize the logger, and fall back to Go's log package if it fails
	if err := logging.Init(); err != nil {
		log.Fatalf("Logger initialization failed: %v", err)
	}

	var err error
	cfg, err = loadConfig()
	if err != nil {
		logging.Logger.Fatalf("Failed to load configuration: %v", err)
	}
	loadCustomCA()
	metrics.Init()
	startBackgroundStatusFetcher()
	setupRoutes()
	startServer(cfg.ServerSettings.Port)
}

func loadCustomCA() {
	caPath := cfg.ServerSettings.CustomCAPath
	if caPath != "" {
		scraping.InitCertificate(caPath)
	}
}

// Load configuration file
func loadConfig() (*config.Config, error) {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config.yaml"
	}
	logging.Logger.Infof("Loading configuration from %s", configFile)

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return nil, err
	}
	logging.Logger.Debugf("Configuration loaded: %+v", cfg)
	return cfg, nil
}

// Setup HTTP routes and handlers
func setupRoutes() {
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /api/status request")
		handlers.GetAppStatus(w, r, cfg)
	})
	http.HandleFunc("/api/scrape-interval", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /api/scrape-interval request")
		handlers.GetScrapeInterval(w, r, cfg)
	})
	http.HandleFunc("/api/docs", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /api/docs request")
		handlers.GetDocs(w, r, cfg)
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /healthz probe")
		livenessProbe(w, r)
	})
	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /readyz probe")
		readinessProbe(w, r)
	})
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.SetupMetricsHandler().ServeHTTP(w, r)
	})
	http.Handle("/", http.FileServer(http.Dir("./static")))

	logging.Logger.Info("HTTP routes configured")
}

// Start the background status fetcher
func startBackgroundStatusFetcher() {
	scrapeInterval, err := time.ParseDuration(cfg.Scraping.Interval)
	if err != nil {
		logging.Logger.Fatalf("Failed to parse ScrapeInterval: %v", err)
	}

	scrapeTimeout, err := time.ParseDuration(cfg.Scraping.Timeout)
	if err != nil {
		logging.Logger.Warnf("Invalid ScrapeTimeout, using default 10s: %v", err)
		scrapeTimeout = 10 * time.Second
	}

	scraping.SetScrapeTimeout(scrapeTimeout)

	maxParallelScrapes := cfg.Scraping.MaxParallel
	if maxParallelScrapes <= 0 {
		logging.Logger.Warnf("Invalid MaxParallelScrapes value, using default 5")
		maxParallelScrapes = 5
	}

	logging.Logger.Infof("Starting background status fetcher with interval: %s and max parallel scrapes: %d", scrapeInterval, maxParallelScrapes)

	go func() {
		ticker := time.NewTicker(scrapeInterval)
		defer ticker.Stop()

		checker := &scraping.DefaultPrometheusChecker{}
		for range ticker.C {
			statusFetcher(checker, maxParallelScrapes)
		}
	}()
}

// Fetch the application statuses with a worker pool
func statusFetcher(checker *scraping.DefaultPrometheusChecker, maxParallelScrapes int) {
	logging.Logger.Info("Running status fetcher...")

	apps := cfg.Applications
	newStatuses := scraping.ParallelScrapeAppStatuses(apps, cfg.PrometheusServers, checker, maxParallelScrapes)

	handlers.UpdateAppStatus(newStatuses)
	logging.Logger.Info("App statuses updated.")
}

// Start the HTTP server and handle graceful shutdown
func startServer(port string) {

	srv := &http.Server{Addr: ":" + port}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		logging.Logger.Infof("Server starting on %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Logger.Fatalf("Server failed: %v", err)
		}
	}()

	<-sigChan
	gracefulShutdown(srv)
}

// Gracefully shut down the server
func gracefulShutdown(srv *http.Server) {
	logging.Logger.Info("Shutdown signal received, shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logging.Logger.Errorf("Server forced to shutdown: %v", err)
	} else {
		logging.Logger.Info("Server exited gracefully")
	}
}

// Liveness probe handler
func livenessProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		logging.Logger.Errorf("Failed to write liveness probe response: %v", err)
	}
}

// Readiness probe handler
func readinessProbe(w http.ResponseWriter, r *http.Request) {
	if handlers.IsAppStatusCacheEmpty() {
		logging.Logger.Warn("Readiness probe failed: App status cache is empty")
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte("NOT READY")); err != nil {
			logging.Logger.Errorf("Failed to write readiness probe response: %v", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("READY")); err != nil {
		logging.Logger.Errorf("Failed to write readiness probe response: %v", err)
	}
}

// getEnv gets an environment variable value and returns a default value if empty
func getEnv(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}
