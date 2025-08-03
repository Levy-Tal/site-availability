package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"

	"site-availability/config"
	"site-availability/logging"
)

// MetricsAuthMiddleware handles authentication for the metrics endpoint
type MetricsAuthMiddleware struct {
	config *config.Config
}

// NewMetricsAuthMiddleware creates a new metrics authentication middleware
func NewMetricsAuthMiddleware(cfg *config.Config) *MetricsAuthMiddleware {
	return &MetricsAuthMiddleware{
		config: cfg,
	}
}

// RequireMetricsAuth is middleware that requires metrics authentication
func (mam *MetricsAuthMiddleware) RequireMetricsAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logging.Logger.WithFields(map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Debug("RequireMetricsAuth middleware called")

		// Check if metrics auth is enabled
		if !mam.config.ServerSettings.MetricsAuth.Enabled {
			logging.Logger.Debug("Metrics authentication not enabled, allowing request")
			next.ServeHTTP(w, r)
			return
		}

		// Authenticate based on the configured type
		authenticated := false
		switch mam.config.ServerSettings.MetricsAuth.Type {
		case "basic":
			authenticated = mam.authenticateBasicAuth(w, r)
		case "bearer":
			authenticated = mam.authenticateBearerToken(w, r)
		default:
			logging.Logger.WithField("type", mam.config.ServerSettings.MetricsAuth.Type).Error("Invalid metrics auth type")
			mam.sendUnauthorized(w, "Invalid authentication configuration")
			return
		}

		if !authenticated {
			logging.Logger.Debug("Metrics authentication failed")
			return
		}

		logging.Logger.Debug("Metrics authentication successful")
		next.ServeHTTP(w, r)
	}
}

// authenticateBasicAuth handles basic authentication for metrics
func (mam *MetricsAuthMiddleware) authenticateBasicAuth(w http.ResponseWriter, r *http.Request) bool {
	// Extract Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		logging.Logger.Debug("No Authorization header found")
		mam.sendBasicAuthChallenge(w)
		return false
	}

	// Check if it's a Basic auth header
	if !strings.HasPrefix(authHeader, "Basic ") {
		logging.Logger.Debug("Authorization header is not Basic auth")
		mam.sendBasicAuthChallenge(w)
		return false
	}

	// Extract and decode credentials
	encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")
	decodedCredentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		logging.Logger.WithError(err).Debug("Failed to decode basic auth credentials")
		mam.sendBasicAuthChallenge(w)
		return false
	}

	// Split username and password
	credentials := strings.SplitN(string(decodedCredentials), ":", 2)
	if len(credentials) != 2 {
		logging.Logger.Debug("Invalid basic auth credentials format")
		mam.sendBasicAuthChallenge(w)
		return false
	}

	username := credentials[0]
	password := credentials[1]

	// Validate credentials
	if username == mam.config.ServerSettings.MetricsAuth.Username &&
		password == mam.config.ServerSettings.MetricsAuth.Password {
		logging.Logger.WithField("username", username).Debug("Basic auth successful")
		return true
	}

	logging.Logger.WithField("username", username).Debug("Basic auth failed - invalid credentials")
	mam.sendBasicAuthChallenge(w)
	return false
}

// authenticateBearerToken handles bearer token authentication for metrics
func (mam *MetricsAuthMiddleware) authenticateBearerToken(w http.ResponseWriter, r *http.Request) bool {
	// Extract Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		logging.Logger.Debug("No Authorization header found")
		mam.sendBearerAuthChallenge(w)
		return false
	}

	// Check if it's a Bearer token header
	if !strings.HasPrefix(authHeader, "Bearer ") {
		logging.Logger.Debug("Authorization header is not Bearer token")
		mam.sendBearerAuthChallenge(w)
		return false
	}

	// Extract token
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate token
	if token == mam.config.ServerSettings.MetricsAuth.Token {
		logging.Logger.Debug("Bearer token authentication successful")
		return true
	}

	logging.Logger.Debug("Bearer token authentication failed - invalid token")
	mam.sendBearerAuthChallenge(w)
	return false
}

// sendBasicAuthChallenge sends a 401 response with Basic auth challenge
func (mam *MetricsAuthMiddleware) sendBasicAuthChallenge(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Metrics"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// sendBearerAuthChallenge sends a 401 response with Bearer auth challenge
func (mam *MetricsAuthMiddleware) sendBearerAuthChallenge(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="Metrics"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// sendUnauthorized sends a generic unauthorized response
func (mam *MetricsAuthMiddleware) sendUnauthorized(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusUnauthorized)
}
