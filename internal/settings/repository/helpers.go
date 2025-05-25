package repository

import (
	"database/sql"
	"time"
)

// toNullTime converts a *time.Time to sql.NullTime
func toNullTime(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{Time: *t, Valid: true}
	}
	return sql.NullTime{}
}

// fromNullTime converts sql.NullTime to *time.Time
func fromNullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}
