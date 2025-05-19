package models

import (
	"errors"
	"strings"
	"time"
)

// Role represents the role of a user in the system.
// It is an enumerated type with predefined constants for different roles.
//
// The available roles are:
//   - ADMIN: Represents an administrative user with elevated privileges.
//   - STANDARD: Represents a standard user with regular privileges.
//
// Role can be converted to and from its string representation using the
// RoleFromString and String methods, respectively.
type Role int

const (
	ADMIN Role = iota
	STANDARD
)

// RoleFromString converts a string representation of a role to its corresponding
// Role type.
func RoleFromString(role string) (Role, error) {
	switch strings.ToLower(role) {
	case "admin":
		return ADMIN, nil
	case "standard":
		return STANDARD, nil
	default:
		return -1, ErrInvalidRole
	}
}

// String returns the string representation of the Role.
func (r Role) String() string {
	switch r {
	case ADMIN:
		return "Admin"
	case STANDARD:
		return "Standard"
	default:
		return "Standard"
	}
}

// User represents an authenticated user in the system.
// It contains user identification, authentication details, permissions,
// and timestamps for tracking user-related events.
type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"password" db:"password"`
	Role      Role      `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	LastLogin time.Time `json:"last_login" db:"last_login"`
}

// NewUser creates a new User instance with the provided username, password, and role.
func NewUser(username, password string, role Role) (*User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	now := time.Now().UTC()

	return &User{
		Username:  username,
		Password:  password,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
