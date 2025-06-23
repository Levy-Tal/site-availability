package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/logging"
	"site-availability/metrics"
	"site-availability/scraping"
	"site-availability/site"
	"syscall"
	"time"
)

// Server represents the web server instance
type Server struct {
	config   *config.Config
	siteSync *site.SiteSync
	mux      *http.ServeMux
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
		mux:    http.NewServeMux(),
	}
}

// Start initializes and starts the server
func (s *Server) Start() error {
	// Initialize site sync if enabled
	if s.config.ServerSettings.SyncEnable {
		s.siteSync = site.NewSiteSync(s.config)
		if err := s.siteSync.Start(); err != nil {
			return err
		}
	}

	s.loadCustomCA()
	metrics.Init()
	s.startBackgroundStatusFetcher()
	s.setupRoutes()
	return s.startServer(s.config.ServerSettings.Port)
}

func (s *Server) loadCustomCA() {
	caPath := s.config.ServerSettings.CustomCAPath
	if caPath != "" {
		scraping.InitCertificate(caPath)
	}
}

// Setup HTTP routes and handlers
func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /api/status request")
		handlers.GetAppStatus(w, r, s.config)
	})
	s.mux.HandleFunc("/api/scrape-interval", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /api/scrape-interval request")
		handlers.GetScrapeInterval(w, r, s.config)
	})
	s.mux.HandleFunc("/api/docs", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /api/docs request")
		handlers.GetDocs(w, r, s.config)
	})
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /healthz probe")
		s.livenessProbe(w, r)
	})
	s.mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /readyz probe")
		s.readinessProbe(w, r)
	})
	s.mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.SetupMetricsHandler().ServeHTTP(w, r)
	})

	// Add sync endpoint if sync is enabled
	if s.config.ServerSettings.SyncEnable {
		s.mux.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
			logging.Logger.Debug("Handling /sync request")
			handlers.HandleSyncRequest(w, r, s.config)
		})
	}

	// Handle static files
	staticDir := "./static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		logging.Logger.Warnf("Static directory %s does not exist", staticDir)
		// Create a simple handler that returns 200 OK for the root path
		s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	} else {
		s.mux.Handle("/", http.FileServer(http.Dir(staticDir)))
	}

	logging.Logger.Info("HTTP routes configured")
}

// Start the background status fetcher
func (s *Server) startBackgroundStatusFetcher() {
	scrapeInterval, err := time.ParseDuration(s.config.Scraping.Interval)
	if err != nil {
		logging.Logger.Fatalf("Failed to parse ScrapeInterval: %v", err)
	}

	scrapeTimeout, err := time.ParseDuration(s.config.Scraping.Timeout)
	if err != nil {
		logging.Logger.Warnf("Invalid ScrapeTimeout, using default 10s: %v", err)
		scrapeTimeout = 10 * time.Second
	}

	scraping.SetScrapeTimeout(scrapeTimeout)

	maxParallelScrapes := s.config.Scraping.MaxParallel
	if maxParallelScrapes <= 0 {
		logging.Logger.Warnf("Invalid MaxParallelScrapes value, using default 5")
		maxParallelScrapes = 5
	}

	logging.Logger.Infof("Starting background status fetcher with interval: %s and max parallel scrapes: %d", scrapeInterval, maxParallelScrapes)

	go func() {
		ticker := time.NewTicker(scrapeInterval)
		defer ticker.Stop()

		checker := &scraping.PrometheusMetricChecker{
			PrometheusServers: s.config.PrometheusServers,
		}
		for range ticker.C {
			s.statusFetcher(checker, maxParallelScrapes)
		}
	}()
}

// Fetch the application statuses with a worker pool
func (s *Server) statusFetcher(checker *scraping.PrometheusMetricChecker, maxParallelScrapes int) {
	logging.Logger.Info("Running status fetcher...")

	apps := s.config.Applications
	newStatuses := scraping.ParallelScrapeAppStatuses(apps, s.config.PrometheusServers, checker, maxParallelScrapes)

	handlers.UpdateAppStatus(newStatuses)
	logging.Logger.Info("App statuses updated.")
}

// Start the HTTP server and handle graceful shutdown
func (s *Server) startServer(port string) error {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: s.mux,
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		logging.Logger.Infof("Server starting on %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Logger.Fatalf("Server failed: %v", err)
		}
	}()

	<-sigChan
	return s.gracefulShutdown(srv)
}

// Gracefully shut down the server
func (s *Server) gracefulShutdown(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a channel to receive the shutdown error
	shutdownErr := make(chan error, 1)

	// Start shutdown in a goroutine
	go func() {
		shutdownErr <- srv.Shutdown(ctx)
	}()

	// Wait for either shutdown to complete or timeout
	select {
	case err := <-shutdownErr:
		return err
	case <-ctx.Done():
		return fmt.Errorf("Server forced to shutdown")
	}
}

// Liveness probe handler
func (s *Server) livenessProbe(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		logging.Logger.Errorf("Failed to write liveness probe response: %v", err)
	}
}

// Readiness probe handler
func (s *Server) readinessProbe(w http.ResponseWriter, _ *http.Request) {
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

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	// Stop site sync if enabled
	if s.siteSync != nil {
		if err := s.siteSync.Stop(); err != nil {
			logging.Logger.Errorf("Failed to stop site sync: %v", err)
		}
	}
	return nil
}
