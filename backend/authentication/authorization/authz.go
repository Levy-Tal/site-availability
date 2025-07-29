package authorization

import (
	"site-availability/authentication/session"
	"site-availability/config"
)

// Authorizer handles role-based access control
type Authorizer struct {
	config *config.Config
}

// NewAuthorizer creates a new authorization handler
func NewAuthorizer(cfg *config.Config) *Authorizer {
	return &Authorizer{
		config: cfg,
	}
}

// LabelPermission represents a user's permission for a specific label
type LabelPermission struct {
	Key           string
	AllowedValues []string
	AllowAll      bool // Admin can access all values for any label
}

// UserPermissions represents all permissions for a user
type UserPermissions struct {
	IsAdmin       bool
	AllowedLabels map[string]LabelPermission
	HasFullAccess bool
}

// GetUserPermissions returns the permissions for a user based on their roles
func (a *Authorizer) GetUserPermissions(userSession *session.Session) UserPermissions {
	// Admin users have full access
	if userSession.IsAdmin {
		return UserPermissions{
			IsAdmin:       true,
			AllowedLabels: make(map[string]LabelPermission),
			HasFullAccess: true,
		}
	}

	permissions := UserPermissions{
		IsAdmin:       false,
		AllowedLabels: make(map[string]LabelPermission),
		HasFullAccess: false,
	}

	// Process each role the user has
	for _, roleName := range userSession.Roles {
		if roleConfig, exists := a.config.ServerSettings.Roles[roleName]; exists {
			// Add all labels from this role to user's permissions
			for labelKey, labelValue := range roleConfig.Labels {
				if existing, hasLabel := permissions.AllowedLabels[labelKey]; hasLabel {
					// Merge values for this label
					existing.AllowedValues = append(existing.AllowedValues, labelValue)
					permissions.AllowedLabels[labelKey] = existing
				} else {
					// Add new label permission
					permissions.AllowedLabels[labelKey] = LabelPermission{
						Key:           labelKey,
						AllowedValues: []string{labelValue},
						AllowAll:      false,
					}
				}
			}
		}
	}

	return permissions
}

// CanAccessLabel checks if user can access a specific label value
func (a *Authorizer) CanAccessLabel(userPermissions UserPermissions, labelKey, labelValue string) bool {
	// Admin has access to everything
	if userPermissions.HasFullAccess {
		return true
	}

	// Check if user has permission for this label
	labelPerm, hasPermission := userPermissions.AllowedLabels[labelKey]
	if !hasPermission {
		return false
	}

	// Check if user can access this specific value
	for _, allowedValue := range labelPerm.AllowedValues {
		if allowedValue == labelValue {
			return true
		}
	}

	return false
}

// FilterLabels returns only the labels that the user can access
func (a *Authorizer) FilterLabels(userPermissions UserPermissions, allLabels map[string][]string) map[string][]string {
	// Admin sees all labels
	if userPermissions.HasFullAccess {
		return allLabels
	}

	filtered := make(map[string][]string)

	// Only include labels the user has permission for
	for labelKey, labelValues := range allLabels {
		if labelPerm, hasPermission := userPermissions.AllowedLabels[labelKey]; hasPermission {
			var allowedValues []string

			// Filter values to only those the user can access
			for _, value := range labelValues {
				for _, allowedValue := range labelPerm.AllowedValues {
					if value == allowedValue {
						allowedValues = append(allowedValues, value)
						break
					}
				}
			}

			if len(allowedValues) > 0 {
				filtered[labelKey] = allowedValues
			}
		}
	}

	return filtered
}

// BuildLabelFilters creates label filters based on user permissions
// This is used to filter apps and other resources
func (a *Authorizer) BuildLabelFilters(userPermissions UserPermissions) map[string]string {
	// Admin doesn't need filters
	if userPermissions.HasFullAccess {
		return nil
	}

	filters := make(map[string]string)

	// Build filters based on user's allowed labels
	for labelKey, labelPerm := range userPermissions.AllowedLabels {
		if len(labelPerm.AllowedValues) == 1 {
			// If user only has access to one value, filter by that
			filters[labelKey] = labelPerm.AllowedValues[0]
		}
		// If user has access to multiple values, we'll need to handle this in the filtering logic
		// For now, we don't add a filter (which means we'll filter in post-processing)
	}

	return filters
}

// CanAccessApp checks if user can access an app based on its labels
func (a *Authorizer) CanAccessApp(userPermissions UserPermissions, appLabels map[string]string) bool {
	// Admin can access everything
	if userPermissions.HasFullAccess {
		return true
	}

	// If user has no permissions, deny access
	if len(userPermissions.AllowedLabels) == 0 {
		return false
	}

	// Check if user has permission for at least one label on this app
	for labelKey, labelValue := range appLabels {
		if a.CanAccessLabel(userPermissions, labelKey, labelValue) {
			return true
		}
	}

	// If no matching labels found, deny access
	return false
}
