package models

// UserRegisterRequest represents the data needed to register a new user.
type UserRegisterRequest struct {
	Name     string `json:"name"`     // User's name
	Email    string `json:"email"`    // User's email address
	Password string `json:"password"` // User's password
}
