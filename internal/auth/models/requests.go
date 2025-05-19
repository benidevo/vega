package models

// LoginRequest represents login form data with validation rules
type LoginRequest struct {
	Username string `form:"username" json:"username" binding:"required,min=3,max=50"`
	Password string `form:"password" json:"password" binding:"required,min=8"`
}
