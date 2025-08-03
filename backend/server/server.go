package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	authHandlers "site-availability/authentication/handlers"
	"site-availability/authentication/middleware"
	"site-availability/authentication/session"
	"site-availability/config"
	appHandlers "site-availability/handlers"
	"site-availability/logging"
	"site-availability/metrics"
	"site-availability/scraping"
	"syscall"
	"time"
)

// Server represents the web server instance
type Server struct {
	config                *config.Config
	mux                   *http.ServeMux
	sessionManager        *session.Manager
	authHandlers          *authHandlers.AuthHandlers
	authMiddleware        *middleware.AuthMiddleware
	authzMiddleware       *middleware.AuthzMiddleware
	metricsAuthMiddleware *middleware.MetricsAuthMiddleware
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
	// Initialize custom CA certificates if configured
	scraping.InitCertificateFromPath(s.config.ServerSettings.CustomCAPath)
	scraping.InitScrapers(s.config)
	metrics.Init()
	scraping.Start(s.config)

	// Initialize authentication components
	s.initAuthentication()

	s.setupRoutes()
	return s.startServer(s.config.ServerSettings.Port)
}

// initAuthentication initializes authentication components
func (s *Server) initAuthentication() {
	// Parse session timeout
	sessionTimeout, err := session.ParseTimeout(s.config.ServerSettings.SessionTimeout)
	if err != nil {
		logging.Logger.WithError(err).Warn("Invalid session timeout, using default")
		sessionTimeout = 12 * time.Hour
	}

	// Initialize session manager
	s.sessionManager = session.NewManager(sessionTimeout)

	// Initialize auth handlers
	s.authHandlers, err = authHandlers.NewAuthHandlers(s.config, s.sessionManager)
	if err != nil {
		logging.Logger.WithError(err).Fatal("Failed to initialize authentication handlers")
	}

	// Initialize auth middleware
	s.authMiddleware = middleware.NewAuthMiddleware(s.config, s.sessionManager)

	// Initialize authorization middleware
	s.authzMiddleware = middleware.NewAuthzMiddleware(s.config)

	// Initialize metrics auth middleware
	s.metricsAuthMiddleware = middleware.NewMetricsAuthMiddleware(s.config)

	logging.Logger.Info("Authentication and authorization components initialized")
}

// requireAuthAndAuthz chains authentication and authorization middleware
func (s *Server) requireAuthAndAuthz(handler http.HandlerFunc) http.HandlerFunc {
	return s.authMiddleware.RequireAuth(s.authzMiddleware.RequireAuthz(handler))
}

// Setup HTTP routes and handlers
func (s *Server) setupRoutes() {
	// Authentication endpoints
	s.mux.HandleFunc("/auth/config", s.authHandlers.HandleAuthConfig)
	s.mux.HandleFunc("/auth/login", s.authHandlers.HandleLogin)
	s.mux.HandleFunc("/auth/user", s.authMiddleware.RequireAuth(s.authHandlers.HandleUser))
	s.mux.HandleFunc("/auth/logout", s.authHandlers.HandleLogout)

	// OIDC endpoints
	s.mux.HandleFunc("/auth/oidc/login", s.authHandlers.HandleOIDCLogin)
	s.mux.HandleFunc("/auth/oidc/callback", s.authHandlers.HandleOIDCCallback)

	// Protected API endpoints
	s.mux.HandleFunc("/api/locations", s.requireAuthAndAuthz(func(w http.ResponseWriter, r *http.Request) {
		appHandlers.GetLocationsWithAuthz(w, r, s.config)
	}))
	s.mux.HandleFunc("/api/apps", s.requireAuthAndAuthz(func(w http.ResponseWriter, r *http.Request) {
		appHandlers.GetAppsWithAuthz(w, r, s.config)
	}))
	s.mux.HandleFunc("/api/labels", s.requireAuthAndAuthz(func(w http.ResponseWriter, r *http.Request) {
		appHandlers.GetLabelsWithAuthz(w, r, s.config)
	}))
	s.mux.HandleFunc("/api/scrape-interval", s.requireAuthAndAuthz(func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /api/scrape-interval request")
		appHandlers.GetScrapeInterval(w, r, s.config)
	}))
	s.mux.HandleFunc("/api/docs", s.requireAuthAndAuthz(func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /api/docs request")
		appHandlers.GetDocs(w, r, s.config)
	}))
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /healthz probe")
		s.livenessProbe(w, r)
	})
	s.mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.Debug("Handling /readyz probe")
		s.readinessProbe(w, r)
	})
	s.mux.HandleFunc("/metrics", s.metricsAuthMiddleware.RequireMetricsAuth(func(w http.ResponseWriter, r *http.Request) {
		metrics.SetupMetricsHandler().ServeHTTP(w, r)
	}))

	// Add sync endpoint if sync is enabled
	if s.config.ServerSettings.SyncEnable {
		s.mux.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
			logging.Logger.Debug("Handling /sync request")
			appHandlers.HandleSyncRequest(w, r, s.config)
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
	if appHandlers.IsAppStatusCacheEmpty() {
		logging.Logger.Warn("Readiness probe failed: App status cache is empty")
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte("NOT READY")); err != nil {
			logging.Logger.Errorf("Failed to write readiness probe response: %v", err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		logging.Logger.Errorf("Failed to write readiness probe response: %v", err)
	}
}
