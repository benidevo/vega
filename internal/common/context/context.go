package context

import (
	"context"
)

// Define context keys as custom types to avoid collisions
type contextKey string

const (
	// UserRoleKey is the context key for user role
	UserRoleKey contextKey = "userRole"
)

// WithRole adds the user role to the context
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, UserRoleKey, role)
}

// GetRole retrieves the user role from the context
func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}
