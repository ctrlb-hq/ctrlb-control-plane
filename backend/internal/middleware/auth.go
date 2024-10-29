package middleware

import (
	"net/http"

	sessionManager "github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/auth/session-manager"
)

// AuthMiddleware checks if the user is authenticated
func AuthMiddleware(sessionManager *sessionManager.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil || cookie.Value == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			_, err = sessionManager.GetSession(cookie.Value)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
