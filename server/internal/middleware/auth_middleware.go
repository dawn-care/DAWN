package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"server/internal/services/auth"
	"time"
)

type contextKey string

const UserIDKey contextKey = "userID"

// Basically implement this wherever we want to secure a page/endpoint/etc.
//
// A session that is unknown or expired gets a 401 and its cookie is deleted. Any other failure
// (the database being down) is a 500 and leaves the cookie alone, so a brief outage does not
// sign anyone out. A session that renewed gets its cookie sent again with the new expiry.
func RequireAuth(sessionSvc auth.SessionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie(SessionCookieName)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			session, err := sessionSvc.Validate(r.Context(), cookie.Value)
			if errors.Is(err, auth.ErrSessionNotFound) || errors.Is(err, auth.ErrSessionExpired) {
				ClearSessionCookie(w)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			if err != nil {
				log.Printf("checking session failed: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if session.Refreshed {
				SetSessionCookie(w, cookie.Value, time.Until(session.ExpiresAt))
			}

			ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
