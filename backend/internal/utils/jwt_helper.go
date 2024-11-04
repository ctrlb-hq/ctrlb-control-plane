package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("your_secret_key") // Store securely in env or config

// GenerateJWT generates a JWT token for a given user ID
func GenerateJWT(email string) (string, error) {
	expirationTime := time.Now().Add(15 * 24 * time.Hour) // Token valid for 15 day

	// Set email as the subject claim
	claims := &jwt.RegisteredClaims{
		Subject:   email,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}

	// Create the token and sign it with the secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWT parses and validates a JWT token and returns the email if valid
func ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims.Subject, nil // Subject now holds the user's email
	}

	return "", nil
}
