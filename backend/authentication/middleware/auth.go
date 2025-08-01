package middleware

import (
	"context"
	"net/http"
	"strings"

	"site-availability/authentication/session"
	"site-availability/config"
	"site-availability/logging"
)

// ContextKey is used to store user information in request context
type ContextKey string

const (
	UserContextKey        ContextKey = "user"
	SessionContextKey     ContextKey = "session"
	PermissionsContextKey ContextKey = "permissions"
)

// AuthMiddleware handles authentication for protected endpoints
type AuthMiddleware struct {
	config         *config.Config
	sessionManager *session.Manager
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(cfg *config.Config, sessionManager *session.Manager) *AuthMiddleware {
	return &AuthMiddleware{
		config:         cfg,
		sessionManager: sessionManager,
	}
}

// RequireAuth is middleware that requires authentication for protected endpoints
func (am *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if authentication is required
		if !am.isAuthRequired() {
			// Authentication disabled, allow request
			next.ServeHTTP(w, r)
			return
		}

		// Check if this endpoint should be excluded from authentication
		if am.isExcludedPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract session from cookie
		sessionID, err := am.extractSessionFromCookie(r)
		if err != nil {
			logging.Logger.WithError(err).Debug("Failed to extract session from cookie")
			am.sendUnauthorized(w, "Authentication required")
			return
		}

		// Validate session
		sessionInfo, valid := am.sessionManager.ValidateSession(sessionID)
		if !valid {
			logging.Logger.Debug("Invalid or expired session")
			am.sendUnauthorized(w, "Invalid or expired session")
			return
		}

		// Refresh session expiration
		am.sessionManager.RefreshSession(sessionID)

		// Add user and session info to request context
		ctx := context.WithValue(r.Context(), UserContextKey, sessionInfo)
		ctx = context.WithValue(ctx, SessionContextKey, sessionInfo)
		r = r.WithContext(ctx)

		logging.Logger.WithFields(map[string]interface{}{
			"username": sessionInfo.Username,
			"path":     r.URL.Path,
			"method":   r.Method,
		}).Debug("Authenticated request")

		// Continue to next handler
		next.ServeHTTP(w, r)
	}
}

// isAuthRequired checks if authentication is required based on configuration
func (am *AuthMiddleware) isAuthRequired() bool {
	return am.config.ServerSettings.LocalAdmin.Enabled || am.config.ServerSettings.OIDC.Enabled
}

// isExcludedPath checks if a path should be excluded from authentication
func (am *AuthMiddleware) isExcludedPath(path string) bool {
	excludedPaths := []string{
		"/",           // Login page
		"/sync",       // B2B endpoint with HMAC auth
		"/auth/login", // Login endpoint
		"/healthz",    // Health check
		"/readyz",     // Readiness check
		"/metrics",    // Metrics endpoint
	}

	// Also exclude static files (anything not starting with /api or /auth)
	if !strings.HasPrefix(path, "/api") && !strings.HasPrefix(path, "/auth") {
		return true
	}

	for _, excludedPath := range excludedPaths {
		if path == excludedPath {
			return true
		}
	}

	return false
}

// extractSessionFromCookie extracts the session ID from the session cookie
func (am *AuthMiddleware) extractSessionFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return "", err
	}

	if cookie.Value == "" {
		return "", http.ErrNoCookie
	}

	return cookie.Value, nil
}

// sendUnauthorized sends an unauthorized response
func (am *AuthMiddleware) sendUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	if _, err := w.Write([]byte(`{"error": "` + message + `"}`)); err != nil {
		logging.Logger.WithError(err).Error("Failed to write unauthorized response")
	}
}

// GetUserFromContext extracts user information from request context
func GetUserFromContext(r *http.Request) (*session.Session, bool) {
	user, ok := r.Context().Value(UserContextKey).(*session.Session)
	return user, ok
}

// GetSessionFromContext extracts session information from request context
func GetSessionFromContext(r *http.Request) (*session.Session, bool) {
	session, ok := r.Context().Value(SessionContextKey).(*session.Session)
	return session, ok
}

// IsSecureRequest determines if the request is secure, accounting for reverse proxies
func IsSecureRequest(r *http.Request, trustProxyHeaders bool) bool {
	// Check if TLS is directly available (direct HTTPS connection)
	if r.TLS != nil {
		return true
	}

	// Only check proxy headers if explicitly configured to trust them
	if !trustProxyHeaders {
		return false
	}

	// Check for X-Forwarded-Proto header (set by reverse proxies)
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}

	// Check for X-Forwarded-SSL header (alternative header used by some proxies)
	if r.Header.Get("X-Forwarded-SSL") == "on" {
		return true
	}

	// Check for X-Forwarded-Port header (if set to 443)
	if r.Header.Get("X-Forwarded-Port") == "443" {
		return true
	}

	return false
}

// CreateSessionCookie creates a secure session cookie
func CreateSessionCookie(sessionID string, maxAge int, r *http.Request, trustProxyHeaders bool) *http.Cookie {
	return &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		MaxAge:   maxAge,
		Path:     "/",
		HttpOnly: true,                                  // Prevent XSS attacks
		Secure:   IsSecureRequest(r, trustProxyHeaders), // Secure flag based on actual request security
		SameSite: http.SameSiteLaxMode,                  // CSRF protection
	}
}

// DeleteSessionCookie creates a cookie that deletes the session
func DeleteSessionCookie(r *http.Request, trustProxyHeaders bool) *http.Cookie {
	return &http.Cookie{
		Name:     "session_id",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   IsSecureRequest(r, trustProxyHeaders), // Secure flag based on actual request security
		SameSite: http.SameSiteLaxMode,
	}
}
