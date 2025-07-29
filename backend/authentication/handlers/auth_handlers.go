package handlers

import (
	"encoding/json"
	"net/http"

	"site-availability/authentication/local"
	"site-availability/authentication/middleware"
	"site-availability/authentication/session"
	"site-availability/config"
	"site-availability/logging"
)

// AuthHandlers contains all authentication-related HTTP handlers
type AuthHandlers struct {
	config         *config.Config
	sessionManager *session.Manager
	localAuth      *local.LocalAuthenticator
}

// NewAuthHandlers creates a new authentication handlers instance
func NewAuthHandlers(cfg *config.Config, sessionManager *session.Manager) *AuthHandlers {
	return &AuthHandlers{
		config:         cfg,
		sessionManager: sessionManager,
		localAuth:      local.NewLocalAuthenticator(cfg),
	}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UserResponse represents the user information response
type UserResponse struct {
	User    UserInfo    `json:"user"`
	Session SessionInfo `json:"session"`
}

// UserInfo represents user information
type UserInfo struct {
	Username   string   `json:"username"`
	Roles      []string `json:"roles"`
	Groups     []string `json:"groups"`
	IsAdmin    bool     `json:"is_admin"`
	AuthMethod string   `json:"auth_method"`
}

// SessionInfo represents session information
type SessionInfo struct {
	ExpiresAt string `json:"expires_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// HandleLogin processes login requests
func (ah *AuthHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ah.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Check if authentication is enabled
	if !ah.localAuth.IsEnabled() {
		ah.sendError(w, http.StatusForbidden, "Authentication is disabled")
		return
	}

	// Parse request
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		logging.Logger.WithError(err).Debug("Failed to parse login request")
		ah.sendError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate credentials
	if err := ah.localAuth.Authenticate(loginReq.Username, loginReq.Password); err != nil {
		logging.Logger.WithFields(map[string]interface{}{
			"username": loginReq.Username,
			"error":    err.Error(),
		}).Info("Login failed")
		ah.sendError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Get user info
	userInfo := ah.localAuth.GetUserInfo(loginReq.Username)

	// Create session
	sessionInfo, err := ah.sessionManager.CreateSession(
		userInfo.Username,
		userInfo.IsAdmin,
		userInfo.Roles,
		userInfo.Groups,
	)
	if err != nil {
		logging.Logger.WithError(err).Error("Failed to create session")
		ah.sendError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Set session cookie
	sessionTimeout, _ := session.ParseTimeout(ah.config.ServerSettings.SessionTimeout)
	maxAge := int(sessionTimeout.Seconds())
	cookie := middleware.CreateSessionCookie(sessionInfo.ID, maxAge)
	http.SetCookie(w, cookie)

	logging.Logger.WithFields(map[string]interface{}{
		"username":   userInfo.Username,
		"session_id": sessionInfo.ID,
	}).Info("User logged in successfully")

	// Send success response
	response := LoginResponse{
		Success: true,
		Message: "Login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode login response")
	}
}

// HandleUser returns current user information
func (ah *AuthHandlers) HandleUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ah.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Check if authentication is enabled
	if !ah.localAuth.IsEnabled() {
		ah.sendError(w, http.StatusForbidden, "Authentication is disabled")
		return
	}

	// Get user from context (set by middleware)
	sessionInfo, ok := middleware.GetUserFromContext(r)
	if !ok {
		ah.sendError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Build response
	response := UserResponse{
		User: UserInfo{
			Username:   sessionInfo.Username,
			Roles:      sessionInfo.Roles,
			Groups:     sessionInfo.Groups,
			IsAdmin:    sessionInfo.IsAdmin,
			AuthMethod: "local", // For local admin authentication
		},
		Session: SessionInfo{
			ExpiresAt: sessionInfo.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode user response")
	}
}

// HandleLogout processes logout requests
func (ah *AuthHandlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ah.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract session from cookie
	sessionID := ""
	if cookie, err := r.Cookie("session_id"); err == nil {
		sessionID = cookie.Value
	}

	// Delete session if it exists
	if sessionID != "" {
		ah.sessionManager.DeleteSession(sessionID)
		logging.Logger.WithFields(map[string]interface{}{
			"session_id": sessionID,
		}).Info("User logged out")
	}

	// Delete session cookie
	cookie := middleware.DeleteSessionCookie()
	http.SetCookie(w, cookie)

	// Send success response
	response := map[string]interface{}{
		"success": true,
		"message": "Logout successful",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode logout response")
	}
}

// sendError sends an error response
func (ah *AuthHandlers) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error: message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode error response")
	}
}

// HandleAuthConfig returns authentication configuration
func (ah *AuthHandlers) HandleAuthConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		ah.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	response := map[string]interface{}{
		"auth_enabled": ah.localAuth.IsEnabled(),
		"auth_methods": []string{},
	}

	if ah.localAuth.IsEnabled() {
		response["auth_methods"] = []string{"local"}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode auth config response")
	}
}

// GetSessionCount returns the number of active sessions (for debugging/monitoring)
func (ah *AuthHandlers) GetSessionCount() int {
	return ah.sessionManager.GetSessionCount()
}
