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
	"site-availability/scraping"
	"syscall"
	"time"

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

	cfg *config.Config // Global configuration variable
)

func main() {
	// Attempt to initialize the logger, and fall back to Go's log package if it fails
	if err := logging.Init(); err != nil {
		log.Fatalf("Logger initialization failed: %v", err)
	}

	// Load custom CA certificates only if the env variable is set and not empty
	if caPath := os.Getenv("CUSTOM_CA_PATH"); caPath != "" {
		scraping.InitCertificate("CUSTOM_CA_PATH")
	} else {
		logging.Logger.Info("CUSTOM_CA_PATH is not set or empty, skipping custom CA loading")
	}

	var err error
	cfg, err = loadConfig()
	if err != nil {
		logging.Logger.Fatalf("Failed to load configuration: %v", err)
	}

	initPrometheusMetrics()
	startBackgroundStatusFetcher()
	setupRoutes()
	startServer()
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

// Initialize Prometheus metrics
func initPrometheusMetrics() {
	prometheus.MustRegister(serverUptime)
	logging.Logger.Info("Prometheus metrics registered")

	go func() {
		for range time.Tick(time.Second) {
			serverUptime.Inc()
		}
	}()
}

// Setup HTTP routes and handlers
func setupRoutes() {
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /api/status request")
		handlers.GetAppStatus(w, r, cfg)
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /healthz probe")
		livenessProbe(w, r)
	})
	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /readyz probe")
		readinessProbe(w, r)
	})
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", http.FileServer(http.Dir("./static")))

	logging.Logger.Info("HTTP routes configured")
}

// Start the background status fetcher
func startBackgroundStatusFetcher() {
	scrapeInterval, err := time.ParseDuration(cfg.ScrapeInterval)
	if err != nil {
		logging.Logger.Fatalf("Failed to parse ScrapeInterval: %v", err)
	}

	scrapeTimeout, err := time.ParseDuration(cfg.ScrapeTimeout)
	if err != nil {
		logging.Logger.Warnf("Invalid ScrapeTimeout, using default 10s: %v", err)
		scrapeTimeout = 10 * time.Second
	}

	// Pass scrapeTimeout to the scraping package
	scraping.SetScrapeTimeout(scrapeTimeout)

	logging.Logger.Infof("Starting background status fetcher with interval: %s", scrapeInterval)

	go func() {
		ticker := time.NewTicker(scrapeInterval)
		defer ticker.Stop()

		checker := &scraping.DefaultPrometheusChecker{}
		statusFetcher(checker)

		for range ticker.C {
			statusFetcher(checker)
		}
	}()
}

// Fetch the application statuses
func statusFetcher(checker *scraping.DefaultPrometheusChecker) {
	logging.Logger.Info("Running status fetcher...")
	var newStatuses []handlers.AppStatus

	for _, app := range cfg.Apps {
		logging.Logger.Debugf("Checking app status: name=%s, url=%s", app.Name, app.Prometheus)

		appStatus := scraping.CheckAppStatus(app, checker)

		logging.Logger.Debugf("App status fetched: name=%s, status=%s", appStatus.Name, appStatus.Status)

		newStatuses = append(newStatuses, appStatus)
	}

	handlers.UpdateAppStatus(newStatuses)
	logging.Logger.Info("App statuses updated.")
}

// Start the HTTP server and handle graceful shutdown
func startServer() {
	port := getServerPort()

	srv := &http.Server{Addr: port}
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

// Get server port from environment variable or default to ":8080"
func getServerPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	logging.Logger.Infof("Server will listen on %s", port)
	return port
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
	w.Write([]byte("OK"))
}

// Readiness probe handler
func readinessProbe(w http.ResponseWriter, r *http.Request) {
	if handlers.IsAppStatusCacheEmpty() {
		logging.Logger.Warn("Readiness probe failed: App status cache is empty")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("NOT READY"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("READY"))
}
