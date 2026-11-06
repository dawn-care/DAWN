package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	mw "server/internal/middleware"
	"server/internal/services/auth"
)

// Error codes returned in the "code" field of error responses from the sign-in code endpoints
const (
	CodeInvalidRequest  = "INVALID_REQUEST"
	CodeInvalidEmail    = "INVALID_EMAIL"
	CodeRateLimited     = "RATE_LIMITED"
	CodeEmailSendFailed = "EMAIL_SEND_FAILED"
	CodeCodeRejected    = "CODE_REJECTED"
	CodeTooManyAttempts = "TOO_MANY_ATTEMPTS"
	CodeTooManyGuesses  = "TOO_MANY_GUESSES"
	CodeInternal        = "INTERNAL_ERROR"
)

// maxEmailLoginBody is far more than an address and a 6-digit code need
const maxEmailLoginBody = 4 << 10

// AuthError is the body of every error response from the sign-in code endpoints. The error
// text is written to be shown to the caregiver.
type AuthError struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

func writeAuthJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeAuthError(w http.ResponseWriter, status int, code, message string) {
	writeAuthJSON(w, status, AuthError{Code: code, Error: message})
}

func decodeEmailLoginBody(w http.ResponseWriter, r *http.Request, body any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxEmailLoginBody)
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		writeAuthError(w, http.StatusBadRequest, CodeInvalidRequest, "The request could not be read.")
		return false
	}
	return true
}

// clientIP is the address the connection came from. Forwarding headers are ignored on purpose:
// the server is reached directly, so a client could write anything in them and so get around
// the per-IP limit.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// writeLimitError answers 429 with a Retry-After header and the wait in the message
func writeLimitError(w http.ResponseWriter, err error, code, message string) {
	var limit *auth.LimitError
	if errors.As(err, &limit) && limit.RetryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(int(limit.RetryAfter.Seconds())))
		message += " Try again " + waitText(limit.RetryAfter) + "."
	} else {
		message += " Try again later."
	}
	writeAuthError(w, http.StatusTooManyRequests, code, message)
}

// waitText says how long the longest wait is, for the caregiver
func waitText(d time.Duration) string {
	switch {
	case d >= 24*time.Hour:
		return "tomorrow"
	case d >= time.Hour:
		return "in an hour"
	default:
		return "in " + strconv.Itoa(int(d.Round(time.Minute).Minutes())) + " minutes"
	}
}

// RequestEmailCode godoc
// @Summary Email a sign-in code
// @Description Emails a 6-digit sign-in code to the address. It works the same whether or not the address has an account, so it cannot be used to find out who has one. The code expires after 10 minutes, and a new code replaces any earlier one. Limited to 3 codes per address per 15 minutes, 10 per address per day, and 10 per IP address per hour. A code that could not be emailed is not stored and does not count toward the limits.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body object{email=string} true "Email address"
// @Success 200 {object} map[string]string "status: sent"
// @Failure 400 {object} AuthError "INVALID_REQUEST or INVALID_EMAIL"
// @Failure 403 {object} AuthError "CROSS_ORIGIN_BLOCKED: the request came from another site"
// @Failure 429 {object} AuthError "RATE_LIMITED, with a Retry-After header in seconds"
// @Failure 500 {object} AuthError "INTERNAL_ERROR"
// @Failure 502 {object} AuthError "EMAIL_SEND_FAILED: the mail server did not accept the email; an earlier code still works"
// @Router /auth/email/code [post]
func (c *AuthController) RequestEmailCode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if !decodeEmailLoginBody(w, r, &body) {
		return
	}

	err := c.AuthService.RequestEmailCode(r.Context(), body.Email, clientIP(r))
	switch {
	case errors.Is(err, auth.ErrInvalidEmail):
		writeAuthError(w, http.StatusBadRequest, CodeInvalidEmail, "Enter a valid email address.")
	case errors.Is(err, auth.ErrTooManyRequests):
		writeLimitError(w, err, CodeRateLimited, "Too many codes requested.")
	case errors.Is(err, auth.ErrSendFailed):
		log.Printf("emailing a sign-in code failed: %v", err)
		writeAuthError(w, http.StatusBadGateway, CodeEmailSendFailed, "We couldn't send the email. Check the address, or try again in a few minutes.")
	case err != nil:
		log.Printf("requesting a sign-in code failed: %v", err)
		writeAuthError(w, http.StatusInternalServerError, CodeInternal, "Something went wrong. Please try again.")
	default:
		writeAuthJSON(w, http.StatusOK, map[string]string{"status": "sent"})
	}
}

// VerifyEmailCode godoc
// @Summary Sign in with an emailed code
// @Description Checks the code emailed to the address and starts a session, creating the caregiver if this is their first sign-in. The session cookie is the same as for Google sign-in. A code works once, for 10 minutes, and is burned after 5 wrong guesses. One IP address may make at most 20 wrong guesses per hour, across every address. If the code is right but signing in fails (500), the code still works.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body object{email=string,code=string} true "Email address and the 6-digit code"
// @Success 200 {object} map[string]string "status: authenticated"
// @Header 200 {string} Set-Cookie "session=<token>; Max-Age=7776000; HttpOnly; SameSite=Strict"
// @Failure 400 {object} AuthError "INVALID_REQUEST"
// @Failure 401 {object} AuthError "CODE_REJECTED: the code is wrong, expired or already used"
// @Failure 403 {object} AuthError "CROSS_ORIGIN_BLOCKED: the request came from another site"
// @Failure 429 {object} AuthError "TOO_MANY_ATTEMPTS: the code was guessed wrong too many times, request a new one; or TOO_MANY_GUESSES: too many wrong codes from this IP, with a Retry-After header in seconds"
// @Failure 500 {object} AuthError "INTERNAL_ERROR"
// @Router /auth/email/verify [post]
func (c *AuthController) VerifyEmailCode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if !decodeEmailLoginBody(w, r, &body) {
		return
	}

	token, err := c.AuthService.LoginWithEmailCode(r.Context(), body.Email, body.Code, clientIP(r))
	switch {
	case errors.Is(err, auth.ErrCodeRejected):
		writeAuthError(w, http.StatusUnauthorized, CodeCodeRejected, "That code is wrong or has expired.")
	case errors.Is(err, auth.ErrTooManyAttempts):
		writeAuthError(w, http.StatusTooManyRequests, CodeTooManyAttempts, "Too many wrong codes. Request a new code.")
	case errors.Is(err, auth.ErrTooManyGuesses):
		writeLimitError(w, err, CodeTooManyGuesses, "Too many wrong codes from this network.")
	case err != nil:
		// A database or session error, while checking the code or after it was accepted. The
		// code is not used up in either case.
		log.Printf("email code login failed: %v", err)
		writeAuthError(w, http.StatusInternalServerError, CodeInternal, "Something went wrong. Please try again.")
	default:
		// The same cookie as Google sign-in: a new session lasts the full idle window
		mw.SetSessionCookie(w, token, auth.SessionIdleTTL)
		writeAuthJSON(w, http.StatusOK, map[string]string{"status": "authenticated"})
	}
}
