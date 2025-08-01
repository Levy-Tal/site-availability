package authhandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"site-availability/authentication/local"
	"site-availability/authentication/middleware"
	"site-availability/authentication/oidc"
	"site-availability/authentication/session"
	"site-availability/config"
	"site-availability/logging"
)

// AuthHandlers contains all authentication-related HTTP handlers
type AuthHandlers struct {
	config         *config.Config
	sessionManager *session.Manager
	localAuth      *local.LocalAuthenticator
	oidcAuth       *oidc.OIDCAuthenticator
}

// NewAuthHandlers creates a new authentication handlers instance
func NewAuthHandlers(cfg *config.Config, sessionManager *session.Manager) (*AuthHandlers, error) {
	oidcAuth, err := oidc.NewOIDCAuthenticator(cfg)
	if err != nil {
		return nil, err
	}

	return &AuthHandlers{
		config:         cfg,
		sessionManager: sessionManager,
		localAuth:      local.NewLocalAuthenticator(cfg),
		oidcAuth:       oidcAuth,
	}, nil
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
		"local", // Local admin authentication
	)
	if err != nil {
		logging.Logger.WithError(err).Error("Failed to create session")
		ah.sendError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Set session cookie
	sessionTimeout, _ := session.ParseTimeout(ah.config.ServerSettings.SessionTimeout)
	maxAge := int(sessionTimeout.Seconds())
	cookie := middleware.CreateSessionCookie(sessionInfo.ID, maxAge, r, ah.config.ServerSettings.TrustProxyHeaders)
	http.SetCookie(w, cookie)

	logging.Logger.WithFields(map[string]interface{}{
		"username":   userInfo.Username,
		"session_id": "****", // Mask session ID for security
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

	// Check if any authentication is enabled (local admin OR OIDC)
	if !ah.localAuth.IsEnabled() && !ah.oidcAuth.IsEnabled() {
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
			AuthMethod: sessionInfo.AuthMethod, // Use the auth method stored in the session
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
			"session_id": "****", // Mask session ID for security
		}).Info("User logged out")
	}

	// Delete session cookie
	cookie := middleware.DeleteSessionCookie(r, ah.config.ServerSettings.TrustProxyHeaders)
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
	var authMethods []string

	if ah.localAuth.IsEnabled() {
		authMethods = append(authMethods, "local")
	}

	if ah.oidcAuth.IsEnabled() {
		authMethods = append(authMethods, "oidc")
	}

	response := map[string]interface{}{
		"auth_enabled": ah.localAuth.IsEnabled() || ah.oidcAuth.IsEnabled(),
		"auth_methods": authMethods,
	}

	// Add OIDC provider info if enabled
	if ah.oidcAuth.IsEnabled() {
		response["oidc_provider_name"] = ah.oidcAuth.GetProviderName()
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Logger.WithError(err).Error("Failed to encode auth config response")
	}
}

// HandleOIDCLogin processes OIDC login requests
func (ah *AuthHandlers) HandleOIDCLogin(w http.ResponseWriter, r *http.Request) {
	logging.Logger.WithFields(map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
		"query":  r.URL.RawQuery,
	}).Debug("OIDC login request received")

	if !ah.oidcAuth.IsEnabled() {
		logging.Logger.Error("OIDC authentication is not enabled")
		ah.sendError(w, http.StatusBadRequest, "OIDC authentication is not enabled")
		return
	}

	// Get the redirect URL from the request
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		// Default redirect URL using configured host_url
		redirectURL = ah.config.ServerSettings.HostURL + "/auth/oidc/callback"
		logging.Logger.WithField("redirect_url", redirectURL).Debug("Using default redirect URL")
	} else {
		// Validate the redirect URL
		parsedURL, err := oidc.ParseRedirectURL(redirectURL)
		if err != nil {
			logging.Logger.WithError(err).WithField("redirect_url", redirectURL).Error("Invalid redirect URL")
			ah.sendError(w, http.StatusBadRequest, "Invalid redirect URL")
			return
		}
		redirectURL = parsedURL
		logging.Logger.WithField("redirect_url", redirectURL).Debug("Using custom redirect URL")
	}

	// Generate authorization URL
	authURL, state, err := ah.oidcAuth.GenerateAuthURL(redirectURL)
	if err != nil {
		logging.Logger.WithError(err).Error("Failed to generate OIDC auth URL")
		ah.sendError(w, http.StatusInternalServerError, "Failed to initiate OIDC login")
		return
	}

	logging.Logger.WithFields(map[string]interface{}{
		"auth_url": authURL,
		"state":    state,
	}).Debug("Generated OIDC authorization URL")

	// Store state in session/cookie for validation (simplified approach)
	http.SetCookie(w, &http.Cookie{
		Name:     "oidc_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   middleware.IsSecureRequest(r, ah.config.ServerSettings.TrustProxyHeaders),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})

	logging.Logger.Debug("Redirecting to OIDC provider")
	// Redirect to OIDC provider
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleOIDCCallback handles the OIDC callback after authentication
func (ah *AuthHandlers) HandleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	logging.Logger.WithFields(map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
		"query":  r.URL.RawQuery,
	}).Debug("OIDC callback request received")

	if !ah.oidcAuth.IsEnabled() {
		logging.Logger.Error("OIDC authentication is not enabled")
		ah.sendError(w, http.StatusBadRequest, "OIDC authentication is not enabled")
		return
	}

	// Check for OIDC error response first
	if errorParam := r.URL.Query().Get("error"); errorParam != "" {
		errorDescription := r.URL.Query().Get("error_description")
		logging.Logger.WithFields(map[string]interface{}{
			"error":             errorParam,
			"error_description": errorDescription,
		}).Error("OIDC authentication error from provider")

		ah.sendError(w, http.StatusBadRequest, fmt.Sprintf("OIDC authentication failed: %s - %s", errorParam, errorDescription))
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		logging.Logger.Error("Missing authorization code in OIDC callback")
		ah.sendError(w, http.StatusBadRequest, "Missing authorization code")
		return
	}

	logging.Logger.WithField("code_length", len(code)).Debug("Received authorization code (masked)")

	// Get and validate state
	receivedState := r.URL.Query().Get("state")
	stateCookie, err := r.Cookie("oidc_state")
	if err != nil {
		logging.Logger.WithError(err).Error("Failed to get OIDC state cookie")
		ah.sendError(w, http.StatusBadRequest, "Invalid state parameter")
		return
	}

	logging.Logger.WithFields(map[string]interface{}{
		"received_state": "****", // Mask state for security
		"cookie_state":   "****", // Mask state for security
	}).Debug("Validating OIDC state parameter")

	if !oidc.ValidateState(receivedState, stateCookie.Value) {
		logging.Logger.WithFields(map[string]interface{}{
			"received_state": "****", // Mask state for security
			"cookie_state":   "****", // Mask state for security
		}).Error("OIDC state validation failed")
		ah.sendError(w, http.StatusBadRequest, "Invalid state parameter")
		return
	}

	logging.Logger.Debug("OIDC state validation successful")

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oidc_state",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   middleware.IsSecureRequest(r, ah.config.ServerSettings.TrustProxyHeaders),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Delete cookie
	})

	// Exchange code for user info
	logging.Logger.Debug("Exchanging authorization code for user info")
	userInfo, err := ah.oidcAuth.HandleCallback(r.Context(), code)
	if err != nil {
		logging.Logger.WithError(err).Error("OIDC callback failed")
		ah.sendError(w, http.StatusUnauthorized, "OIDC authentication failed")
		return
	}

	logging.Logger.WithFields(map[string]interface{}{
		"username":    userInfo.Username,
		"is_admin":    userInfo.IsAdmin,
		"roles":       userInfo.Roles,
		"groups":      userInfo.Groups,
		"auth_method": userInfo.AuthMethod,
	}).Debug("OIDC user info retrieved successfully")

	// Create session
	logging.Logger.Debug("Creating session for OIDC user")
	sessionInfo, err := ah.sessionManager.CreateSession(userInfo.Username, userInfo.IsAdmin, userInfo.Roles, userInfo.Groups, "oidc")
	if err != nil {
		logging.Logger.WithError(err).Error("Failed to create session for OIDC user")
		ah.sendError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	logging.Logger.WithFields(map[string]interface{}{
		"session_id": "****", // Mask session ID for security
		"username":   sessionInfo.Username,
		"expires_at": sessionInfo.ExpiresAt,
	}).Debug("Session created successfully for OIDC user")

	// Set session cookie
	sessionTimeout, _ := session.ParseTimeout(ah.config.ServerSettings.SessionTimeout)
	maxAge := int(sessionTimeout.Seconds())
	cookie := middleware.CreateSessionCookie(sessionInfo.ID, maxAge, r, ah.config.ServerSettings.TrustProxyHeaders)
	http.SetCookie(w, cookie)

	logging.Logger.WithFields(map[string]interface{}{
		"session_id": "****", // Mask session ID for security
		"max_age":    maxAge,
		"secure":     cookie.Secure,
	}).Debug("Session cookie set successfully")

	// Redirect to application or return success
	redirectTo := r.URL.Query().Get("redirect_to")
	if redirectTo == "" {
		redirectTo = "/" // Default to home page
		logging.Logger.Debug("Using default redirect to home page")
	} else {
		logging.Logger.WithField("redirect_to", redirectTo).Debug("Using custom redirect URL")
	}

	// For security, validate redirect_to URL
	if parsedURL, err := url.Parse(redirectTo); err != nil || (parsedURL.Host != "" && parsedURL.Host != r.Host) {
		logging.Logger.WithField("redirect_to", redirectTo).Warn("Invalid redirect_to URL, falling back to home page")
		redirectTo = "/" // Fallback to safe default
	}

	logging.Logger.WithFields(map[string]interface{}{
		"final_redirect": redirectTo,
		"username":       userInfo.Username,
		"session_id":     "****", // Mask session ID for security
	}).Info("OIDC authentication completed successfully, redirecting user")

	http.Redirect(w, r, redirectTo, http.StatusFound)
}

// GetSessionCount returns the number of active sessions (for debugging/monitoring)
func (ah *AuthHandlers) GetSessionCount() int {
	return ah.sessionManager.GetSessionCount()
}
