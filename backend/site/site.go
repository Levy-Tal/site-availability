package site

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"site-availability/authentication/hmac"
	"site-availability/config"
	"site-availability/handlers"
	"site-availability/logging"
	"site-availability/metrics"
)

type SiteSync struct {
	config       *config.Config
	metrics      *metrics.SiteSyncMetrics
	sites        map[string]*config.Site
	siteClients  map[string]*http.Client // Map of site name to HTTP client with custom CA
	httpClient   *http.Client
	mu           sync.RWMutex
	doneChannels map[string]chan struct{}
}

type SyncedAppStatus struct {
	handlers.AppStatus
	SourceSite  string    `json:"source_site"`
	LastSynced  time.Time `json:"last_synced"`
	LastSuccess time.Time `json:"last_success"`
	SyncEnabled bool      `json:"sync_enabled"`
	ErrorCount  int       `json:"error_count"`
	LastError   string    `json:"last_error,omitempty"`
}

func (s *SiteSync) createHTTPClient(site config.Site) (*http.Client, error) {
	timeout, err := time.ParseDuration(site.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout for site %s: %w", site.Name, err)
	}

	// Create base transport
	transport := &http.Transport{
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: timeout,
		IdleConnTimeout:       timeout,
	}

	// If custom CA is specified, configure TLS
	if site.CustomCAPath != "" {
		caCert, err := os.ReadFile(site.CustomCAPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert for site %s: %w", site.Name, err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA cert for site %s", site.Name)
		}

		transport.TLSClientConfig = &tls.Config{
			RootCAs: caCertPool,
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}

func NewSiteSync(cfg *config.Config) *SiteSync {
	sync := &SiteSync{
		config:      cfg,
		metrics:     metrics.NewSiteSyncMetrics(),
		sites:       make(map[string]*config.Site),
		siteClients: make(map[string]*http.Client),
		httpClient:  &http.Client{Timeout: time.Second * 10},
	}

	// Initialize HTTP clients for each site with their respective CA certs
	for _, site := range cfg.Sites {
		if site.Enabled {
			client, err := sync.createHTTPClient(site)
			if err != nil {
				logging.Logger.Errorf("Failed to create HTTP client for site %s: %v", site.Name, err)
				continue
			}
			sync.siteClients[site.Name] = client
		}
	}

	return sync
}

func (s *SiteSync) Start() error {
	if !s.config.ServerSettings.SyncEnable {
		logging.Logger.Info("Site sync is disabled")
		return nil
	}

	// Initialize sites map
	for _, site := range s.config.Sites {
		siteCopy := site // Create a copy to avoid using the loop variable directly
		s.sites[site.Name] = &siteCopy
	}

	// Start sync workers
	for _, site := range s.sites {
		if site.Enabled {
			go s.startSiteSync(site)
		}
	}

	return nil
}

func (s *SiteSync) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close all done channels
	for _, done := range s.doneChannels {
		close(done)
	}
	s.doneChannels = make(map[string]chan struct{})

	return nil
}

func (s *SiteSync) startSiteSync(site *config.Site) {
	interval, err := time.ParseDuration(site.CheckInterval)
	if err != nil {
		logging.Logger.Errorf("Invalid check interval for site %s: %v", site.Name, err)
		return
	}

	ticker := time.NewTicker(interval)
	done := make(chan struct{})

	// Store the done channel in a map for cleanup
	s.mu.Lock()
	if s.doneChannels == nil {
		s.doneChannels = make(map[string]chan struct{})
	}
	s.doneChannels[site.Name] = done
	s.mu.Unlock()

	for {
		select {
		case <-done:
			ticker.Stop()
			return
		case <-ticker.C:
			if err := s.syncSite(site); err != nil {
				s.metrics.SyncFailures.Inc()
				s.updateSiteStatus(site.Name, false, err.Error())
				logging.Logger.Errorf("Failed to sync with site %s: %v", site.Name, err)
			} else {
				s.metrics.SyncAttempts.Inc()
				s.updateSiteStatus(site.Name, true, "")
				logging.Logger.Infof("Successfully synced with site %s", site.Name)
			}
		}
	}
}

func (s *SiteSync) syncSite(site *config.Site) error {
	start := time.Now()
	defer func() {
		s.metrics.SyncLatency.Observe(time.Since(start).Seconds())
	}()

	req, err := s.prepareSyncRequest(site)
	if err != nil {
		s.updateSiteStatus(site.Name, false, err.Error())
		return err
	}

	// Use site-specific client with custom CA if available
	client := s.siteClients[site.Name]
	if client == nil {
		// Fallback to default client if no custom client is available
		client = &http.Client{
			Timeout: time.Second * 10,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		s.updateSiteStatus(site.Name, false, err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("sync failed with status: %d", resp.StatusCode)
		s.updateSiteStatus(site.Name, false, err.Error())
		return err
	}

	var remoteStatuses map[string]*SyncedAppStatus
	if err := json.NewDecoder(resp.Body).Decode(&remoteStatuses); err != nil {
		s.updateSiteStatus(site.Name, false, err.Error())
		return err
	}

	s.mergeStatuses(site.Name, remoteStatuses)
	s.updateSiteStatus(site.Name, true, "")
	return nil
}

func (s *SiteSync) prepareSyncRequest(site *config.Site) (*http.Request, error) {
	url := fmt.Sprintf("%s/sync", site.URL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add timestamp
	timestamp := time.Now().Format(time.RFC3339)
	req.Header.Set("X-Site-Sync-Timestamp", timestamp)

	// Add HMAC signature if token is available
	if site.Token != "" {
		validator := hmac.NewValidator(site.Token)
		signature := validator.GenerateSignature(timestamp, nil) // Empty body for GET request
		req.Header.Set("X-Site-Sync-Signature", signature)
	}

	return req, nil
}

func (s *SiteSync) mergeStatuses(sourceSite string, remoteStatuses map[string]*SyncedAppStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Convert remote statuses to AppStatus slice
	var newStatuses []handlers.AppStatus
	for _, status := range remoteStatuses {
		status.SourceSite = sourceSite
		status.LastSynced = time.Now()
		newStatuses = append(newStatuses, status.AppStatus)
	}

	// Update the global cache
	handlers.UpdateAppStatus(newStatuses)
}

func (s *SiteSync) updateSiteStatus(siteName string, success bool, errorMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	site, exists := s.sites[siteName]
	if !exists {
		return
	}

	if success {
		site.LastSuccess = time.Now()
		site.ErrorCount = 0
		site.LastError = ""
	} else {
		site.ErrorCount++
		site.LastError = errorMsg
	}
}
