package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	mw "server/internal/middleware"
	"server/internal/services/auth"
	"strconv"
)

type AuthController struct {
	AuthService *auth.AuthService
}

// GoogleLogin godoc
// @Summary Log in with Google
// @Description Verifies a Google ID token and starts a session. The session cookie is HttpOnly, SameSite=Strict, and Secure outside dev. It lasts 90 days from when it last renewed (a request to /me renews it, at most once a day) and always ends 1 year after login.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body object{idToken=string} true "Google ID token"
// @Success 200 {object} map[string]string "status: authenticated"
// @Header 200 {string} Set-Cookie "session=<token>; Max-Age=7776000; HttpOnly; SameSite=Strict"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /auth/google [post]
func (c *AuthController) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDToken string `json:"idToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	token, err := c.AuthService.LoginWithGoogle(r.Context(), body.IDToken)
	if errors.Is(err, auth.ErrGoogleTokenRejected) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		// The caregiver is who they say they are, but the server could not sign them in
		log.Printf("google login failed: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// A new session lasts the full idle window; it renews from there as it is used
	mw.SetSessionCookie(w, token, auth.SessionIdleTTL)

	//w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "authenticated",
	})
}

// MeHandler godoc
// @Summary Get the signed-in caregiver
// @Description Returns the caregiver the session cookie belongs to, with their carepartners. A session used more than a day after it last renewed is renewed here, and the cookie is sent again with its new expiry. An unknown or expired session gets 401 and its cookie is deleted. A failure loading the session, the caregiver or their carepartners gets 500, and the cookie is left alone.
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{} "id, name, role and carepartners"
// @Header 200 {string} Set-Cookie "session=<token>; Max-Age=<seconds left>; HttpOnly; SameSite=Strict (only when the session renewed)"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /me [get]
func (c *AuthController) MeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(mw.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := c.AuthService.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
		return
	}

	carepartners, err := c.AuthService.Caregivers.GetCarepartnersByCaregiverID(userID)
	if err != nil {
		http.Error(w, "Failed to fetch carepartners", http.StatusInternalServerError)
		return
	}

	type cpEntry struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	cpList := make([]cpEntry, len(carepartners))
	for i, cp := range carepartners {
		cpList[i] = cpEntry{ID: cp.ID, Name: cp.Name}
	}

	resp := map[string]interface{}{
		"id":           user.ID,
		"name":         user.Name,
		"role":         user.Role,
		"carepartners": cpList,
	}
	json.NewEncoder(w).Encode(resp)
}

func (c *AuthController) UpdateMeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user exists
	userID, ok := r.Context().Value(mw.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Check that carepartner ID is included
	pairingCode := r.FormValue("pairingCode")
	if pairingCode == "" {
		http.Error(w, "Pairing code required", http.StatusBadRequest)
		return
	}

	// Check that carepartner ID is a valid string
	carepartnerID, err := strconv.Atoi(pairingCode)
	if err != nil {
		http.Error(w, "Invalid pairing code", http.StatusBadRequest)
		return
	}

	// Check that the consent form is included
	file, _, err := r.FormFile("consentForm")
	if err != nil {
		http.Error(w, "Consent form required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// TODO: Store file with minio

	// Update the users care partner ID
	err = c.AuthService.UpdateCarepartnerID(r.Context(), userID, carepartnerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Logout godoc
// @Summary Log out
// @Description Ends every session of the caregiver, on every device, then clears this browser's session cookie. With no cookie, or one that is unknown or expired, it still succeeds and clears the cookie, but other devices stay signed in, since the cookie no longer says whose sessions to end. If the sessions cannot be ended it answers 500 and keeps the cookie, so the caregiver can try again.
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string "status: logged out"
// @Header 200 {string} Set-Cookie "session=; Max-Age=0 (deletes the cookie)"
// @Failure 500 {string} string "Failed to log out"
// @Router /auth/logout [post]
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(mw.SessionCookieName); err == nil {
		if err := c.AuthService.LogoutEverywhere(r.Context(), cookie.Value); err != nil {
			// The sessions may still be valid, so the cookie stays and the caregiver can try again
			log.Printf("logout failed: %v", err)
			http.Error(w, "Failed to log out", http.StatusInternalServerError)
			return
		}
	}

	mw.ClearSessionCookie(w)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"logged out"}`))
}
