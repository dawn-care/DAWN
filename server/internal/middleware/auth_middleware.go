package middleware

import (
	"context"
	"net/http"
	"server/internal/services/auth"
)

type contextKey string

const UserIDKey contextKey = "userID"

// Basically implement this wherever we want to secure a page/endpoint/etc.
func RequireAuth(sessionSvc auth.SessionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie("session")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := sessionSvc.Validate(r.Context(), cookie.Value)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
