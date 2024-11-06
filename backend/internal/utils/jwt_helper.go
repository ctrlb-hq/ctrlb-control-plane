package utils

import (
	"time"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	"github.com/golang-jwt/jwt/v4"
)

// GenerateJWT generates a JWT token for a given email and expiration time
func GenerateJWT(email string, expiration time.Duration) (string, error) {
	expirationTime := time.Now().Add(expiration)
  
	// Set email as the subject claim
	claims := &jwt.RegisteredClaims{
		Subject:   email,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}

	// Create the token and sign it with the secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(constants.JWT_SECRET))
}

// GenerateAccessToken generates a short-lived access token
func GenerateAccessToken(email string) (string, error) {
	return GenerateJWT(email, 15*time.Minute) // 15 minutes
}

// GenerateRefreshToken generates a long-lived refresh token
func GenerateRefreshToken(email string) (string, error) {
	return GenerateJWT(email, 30*24*time.Hour) // 30 days
}

// RefreshToken generates a new access token using a valid refresh token
func RefreshToken(refreshToken string) (string, error) {
	email, err := ValidateJWT(refreshToken)
	if err != nil {
		return "", err // Invalid refresh token
	}
	// Generate a new access token
	return GenerateAccessToken(email)
}

// ValidateJWT parses and validates a JWT token and returns the email if valid
func ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(constants.JWT_SECRET), nil

	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims.Subject, nil // Subject now holds the user's email
	}

	return "", nil
}
