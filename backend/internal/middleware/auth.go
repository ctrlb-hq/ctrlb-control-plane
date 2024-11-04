package middleware

import (
	"context"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
)

// Define a custom type for the context key
type contextKey string

const emailContextKey contextKey = "email"

// AuthMiddleware verifies JWT token validity on protected routes
func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				http.Error(w, "Missing token", http.StatusUnauthorized)
				return
			}

			email, err := utils.ValidateJWT(tokenString)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Set the email in the request context
			ctx := context.WithValue(r.Context(), emailContextKey, email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
