package auth

import (
	"errors"
	"time"
)

type Role int

const (
	ADMIN Role = iota
	STANDARD
)

func (r Role) String() string {
	switch r {
	case ADMIN:
		return "Admin"
	case STANDARD:
		return "Standard"
	default:
		return "Unknown"
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
