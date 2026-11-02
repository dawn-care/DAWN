package middleware

import (
	"net/http"
	"os"
	"time"
)

const SessionCookieName = "session"

// SetSessionCookie sends the session token to the browser. HttpOnly keeps scripts from reading
// it, Secure keeps it off plain HTTP (except in dev), SameSite=Strict keeps other sites from
// using it, and the Max-Age keeps it across browser restarts until the session ends.
func SetSessionCookie(w http.ResponseWriter, token string, maxAge time.Duration) {
	seconds := int(maxAge.Seconds())
	if seconds <= 0 {
		// A Max-Age of 0 would make this a cookie that dies with the browser, not one that is already over
		ClearSessionCookie(w)
		return
	}
	http.SetCookie(w, sessionCookie(token, seconds))
}

// ClearSessionCookie tells the browser to delete the session cookie right away
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, sessionCookie("", -1))
}

func sessionCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   os.Getenv("RUN_MODE") != "dev",
		SameSite: http.SameSiteStrictMode,
	}
}
