package models

import "time"

// AccountActivity represents the timestamps of the last login and account creation for a user.
type AccountActivity struct {
	LastLogin time.Time `json:"login_time"`
	CreatedAt time.Time `json:"created_at"`
}

// NewAccountActivity creates a new AccountActivity instance.
func NewAccountActivity(lastLogin, createdAt time.Time) *AccountActivity {
	return &AccountActivity{
		LastLogin: lastLogin,
		CreatedAt: createdAt,
	}
}

// SecuritySettings represents the security settings of a user
type SecuritySettings struct {
	Activity *AccountActivity
}

// NewSecuritySettings creates and returns a new SecuritySettings instance
func NewSecuritySettings(accountActivity *AccountActivity) *SecuritySettings {
	return &SecuritySettings{
		Activity: accountActivity,
	}
}
