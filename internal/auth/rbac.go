package auth

import (
	"context"
	"time"
)

// Role defines user access privileges within the Corridor system.
type Role string

const (
	// RolePublic represents unauthenticated public users (H3-r6 generalized coordinates, 30-day delay).
	RolePublic Role = "public"
	// RoleReporter represents field staff or citizens submitting roadkill reports.
	RoleReporter Role = "reporter"
	// RoleAnalyst represents park rangers, wildlife biologists, and DOT analysts (un-redacted raw coordinates).
	RoleAnalyst Role = "analyst"
	// RoleAdmin represents system administrators.
	RoleAdmin Role = "admin"
)

// User represents an authenticated identity.
type User struct {
	Subject string `json:"sub"`
	Role    Role   `json:"role"`
	Email   string `json:"email,omitempty"`
	Name    string `json:"name,omitempty"`
	Agency  string `json:"agency,omitempty"`
}

type userContextKey struct{}

// WithUser returns a new Context that carries the provided User.
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

// UserFromContext retrieves the User from the Context, or nil if unauthenticated.
func UserFromContext(ctx context.Context) *User {
	if val := ctx.Value(userContextKey{}); val != nil {
		if u, ok := val.(*User); ok {
			return u
		}
	}
	return nil
}

// RoleFromContext returns the Role of the user in Context, defaulting to RolePublic.
func RoleFromContext(ctx context.Context) Role {
	u := UserFromContext(ctx)
	if u == nil || u.Role == "" {
		return RolePublic
	}
	return u.Role
}

// HasRole checks if the role in context matches or exceeds required roles.
func HasRole(ctx context.Context, allowedRoles ...Role) bool {
	userRole := RoleFromContext(ctx)
	for _, r := range allowedRoles {
		if userRole == r {
			return true
		}
	}
	return false
}

// PrivacyLevel controls spatial and temporal generalization for sensitive species.
type PrivacyLevel string

const (
	// PrivacyGeneralized applies H3-r6 spatial blurring (~36 km2) and 30-day delay to sensitive species.
	PrivacyGeneralized PrivacyLevel = "generalized"
	// PrivacyFull grants full-precision GPS coordinates and immediate un-delayed access.
	PrivacyFull PrivacyLevel = "full"
)

// SensitiveFilter defines the spatial and temporal boundaries based on user authorization.
type SensitiveFilter struct {
	AllowRawCoordinates bool
	MaxH3Resolution     int           // 6 for generalized public, 15 for analyst/raw
	MinAgeDelay         time.Duration // 30 days for public, 0 for analyst
}

// PrivacyLevelForRole returns PrivacyFull for analysts and admins; PrivacyGeneralized otherwise.
func PrivacyLevelForRole(role Role) PrivacyLevel {
	switch role {
	case RoleAnalyst, RoleAdmin:
		return PrivacyFull
	default:
		return PrivacyGeneralized
	}
}

// SensitiveFilterForRole constructs the privacy filter configuration for a given role.
func SensitiveFilterForRole(role Role) SensitiveFilter {
	if PrivacyLevelForRole(role) == PrivacyFull {
		return SensitiveFilter{
			AllowRawCoordinates: true,
			MaxH3Resolution:     15,
			MinAgeDelay:         0,
		}
	}
	return SensitiveFilter{
		AllowRawCoordinates: false,
		MaxH3Resolution:     6,
		MinAgeDelay:         30 * 24 * time.Hour,
	}
}
