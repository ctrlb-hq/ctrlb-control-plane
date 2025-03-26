package models

import "github.com/golang-jwt/jwt/v5"

// UserRegisterRequest represents the data needed to register a new user.
type UserRegisterRequest struct {
	Name     string `json:"name"`     // User's name
	Email    string `json:"email"`    // User's email address
	Password string `json:"password"` // User's password
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CustomClaims struct {
	TokenUse string `json:"token_use"` // e.g., "access" or "refresh"
	jwt.RegisteredClaims
}
