package models

import (
	"errors"
	"strings"
	"time"
)

// ValidateUsername validates a username
func ValidateUsername(username string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}

	if len(username) < 3 || len(username) > 30 {
		return errors.New("username must be between 3 and 30 characters")
	}

	return nil
}

// Role represents the user's quota tier in the system.
// It is an enumerated type that determines AI analysis limits.
//
// The available roles are:
//   - ADMIN: Users with unlimited AI analysis quota (no monthly limits).
//     Note: Despite the name, this does NOT grant system administration privileges.
//     It only removes the monthly quota limit for AI job analyses.
//   - STANDARD: Users with standard quotas (10 AI analyses per month in cloud mode).
//
// In self-hosted mode, all users have unlimited quotas regardless of role.
// The role distinction only affects cloud mode users.
//
// Role can be converted to and from its string representation using the
// RoleFromString and String methods, respectively.
type Role int

const (
	// ADMIN grants unlimited AI analysis quota (cloud mode only).
	// Does NOT provide system administration capabilities.
	ADMIN Role = iota

	// STANDARD provides 10 AI analyses per month (cloud mode only).
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

// IsAdmin returns true if the user has the ADMIN role.
func (u *User) IsAdmin() bool {
	return u.Role == ADMIN
}

// NewUser creates a new User instance with the provided username, password, and role.
func NewUser(username, password string, role Role) (*User, error) {
	if err := ValidateUsername(username); err != nil {
		return nil, err
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
