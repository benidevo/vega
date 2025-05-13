package models

import "time"

// NotificationSettings represents user notification preferences
type NotificationSettings struct {
	ID                 int       `json:"id" db:"id" sql:"primary_key;auto_increment"`
	UserID             int       `json:"user_id" db:"user_id" sql:"type:integer;not null;unique;index;references:users(id)"`
	EmailNotifications bool      `json:"email_notifications" db:"email_notifications" sql:"type:boolean;default:true"`
	JobAlerts          bool      `json:"job_alerts" db:"job_alerts" sql:"type:boolean;default:true"`
	ApplicationUpdates bool      `json:"application_updates" db:"application_updates" sql:"type:boolean;default:true"`
	WeeklyDigest       bool      `json:"weekly_digest" db:"weekly_digest" sql:"type:boolean;default:true"`
	CreatedAt          time.Time `json:"created_at" db:"created_at" sql:"type:timestamp;not null;default:current_timestamp"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at" sql:"type:timestamp;not null;default:current_timestamp"`
}
