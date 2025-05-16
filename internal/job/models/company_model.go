package models

import (
	"time"
)

// Company represents a company entity with its basic details.
type Company struct {
	ID        int       `json:"id" db:"id" sql:"primary_key;auto_increment"`
	Name      string    `json:"name" db:"name" sql:"type:text;not null;unique;index"`
	CreatedAt time.Time `json:"created_at" db:"created_at" sql:"not null;default:current_timestamp"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" sql:"not null;default:current_timestamp"`
}

// NewCompany creates a new Company instance with the given name.
func NewCompany(name string) *Company {
	now := time.Now().UTC()
	return &Company{
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
