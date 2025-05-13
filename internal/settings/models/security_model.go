package models

import "time"

// SecuritySettings represents user security configuration
type SecuritySettings struct {
	ID                 int       `json:"id" db:"id" sql:"primary_key;auto_increment"`
	UserID             int       `json:"user_id" db:"user_id" sql:"type:integer;not null;unique;index;references:users(id)"`
	TwoFactorEnabled   bool      `json:"two_factor_enabled" db:"two_factor_enabled" sql:"type:boolean;default:false"`
	TwoFactorMethod    string    `json:"two_factor_method" db:"two_factor_method" sql:"type:text"`
	LastPasswordChange time.Time `json:"last_password_change" db:"last_password_change" sql:"type:timestamp"`
	LastLogin          time.Time `json:"last_login" db:"last_login" sql:"type:timestamp"`
	LoginAttempts      int       `json:"login_attempts" db:"login_attempts" sql:"type:integer;default:0"`
	CreatedAt          time.Time `json:"created_at" db:"created_at" sql:"type:timestamp;not null;default:current_timestamp"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at" sql:"type:timestamp;not null;default:current_timestamp"`
}

// PasswordChangeRequest represents a request to change a user's password
type PasswordChangeRequest struct {
	CurrentPassword    string `json:"current_password" form:"current_password"`
	NewPassword        string `json:"new_password" form:"new_password"`
	ConfirmNewPassword string `json:"confirm_new_password" form:"confirm_new_password"`
}

// Validate checks if the password change request is valid
func (p *PasswordChangeRequest) Validate() error {
	if p.CurrentPassword == "" || p.NewPassword == "" || p.ConfirmNewPassword == "" {
		return ErrFieldRequired
	}

	if p.NewPassword != p.ConfirmNewPassword {
		return ErrPasswordMismatch
	}

	if len(p.NewPassword) < 8 {
		return ErrPasswordTooShort
	}

	return nil
}
