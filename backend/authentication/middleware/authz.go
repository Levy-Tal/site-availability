package middleware

import (
	"context"
	"net/http"

	"site-availability/authentication/authorization"
	"site-availability/config"
	"site-availability/logging"
)

// AuthzMiddleware handles authorization for protected endpoints
type AuthzMiddleware struct {
	config     *config.Config
	authorizer *authorization.Authorizer
}

// NewAuthzMiddleware creates a new authorization middleware
func NewAuthzMiddleware(cfg *config.Config) *AuthzMiddleware {
	return &AuthzMiddleware{
		config:     cfg,
		authorizer: authorization.NewAuthorizer(cfg),
	}
}

// RequireAuthz is middleware that adds user permissions to the request context
func (am *AuthzMiddleware) RequireAuthz(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (should be set by auth middleware)
		userSession, ok := GetUserFromContext(r)
		if !ok {
			// If no user in context, this endpoint might not require auth
			// Just continue without permissions
			next.ServeHTTP(w, r)
			return
		}

		// Get user permissions based on their roles
		permissions := am.authorizer.GetUserPermissions(userSession)

		// Add permissions to request context
		ctx := context.WithValue(r.Context(), PermissionsContextKey, permissions)
		r = r.WithContext(ctx)

		logging.Logger.WithFields(map[string]interface{}{
			"username":    userSession.Username,
			"roles":       userSession.Roles,
			"is_admin":    permissions.IsAdmin,
			"full_access": permissions.HasFullAccess,
			"label_count": len(permissions.AllowedLabels),
		}).Debug("Authorization permissions loaded")

		// Continue to next handler
		next.ServeHTTP(w, r)
	}
}

// GetPermissionsFromContext extracts user permissions from request context
func GetPermissionsFromContext(r *http.Request) (authorization.UserPermissions, bool) {
	permissions, ok := r.Context().Value(PermissionsContextKey).(authorization.UserPermissions)
	return permissions, ok
}

// GetAuthorizerFromContext provides access to the authorizer for complex operations
func (am *AuthzMiddleware) GetAuthorizer() *authorization.Authorizer {
	return am.authorizer
}
