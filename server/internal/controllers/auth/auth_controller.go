package auth

import (
	"encoding/json"
	"net/http"
	"os"
	mw "server/internal/middleware"
	"server/internal/services/auth"
	"strconv"
)

type AuthController struct {
	AuthService *auth.AuthService
}

func (c *AuthController) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDToken string `json:"idToken"`
	}

	json.NewDecoder(r.Body).Decode(&body)

	sessionID, err := c.AuthService.LoginWithGoogle(r.Context(), body.IDToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   os.Getenv("RUN_MODE") != "dev",
		SameSite: http.SameSiteLaxMode,
	})

	//w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "authenticated",
	})
}

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

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // tells browser to delete immediately
		HttpOnly: true,
		Secure:   os.Getenv("RUN_MODE") != "dev",
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"logged out"}`))
}
