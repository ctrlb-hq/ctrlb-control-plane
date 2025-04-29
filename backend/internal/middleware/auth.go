package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

// Define a custom type for the context key
type contextKey string

const EmailContextKey contextKey = "email"

// AuthMiddleware verifies JWT token validity on protected routes
func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				utils.SendJSONError(w, http.StatusUnauthorized, "Missing token")
				return
			}

			tokenString = strings.TrimPrefix(tokenString, "Bearer ")

			// Validate the access token
			email, err := utils.ValidateJWTFunc(tokenString, "access")
			if err != nil {
				// Check if the error is due to token expiration
				if errors.Is(err, jwt.ErrTokenExpired) {
					utils.SendJSONError(w, http.StatusUnauthorized, "Token expired, please refresh")
					return
				}
				utils.SendJSONError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// Set the email in the request context
			ctx := context.WithValue(r.Context(), EmailContextKey, email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
