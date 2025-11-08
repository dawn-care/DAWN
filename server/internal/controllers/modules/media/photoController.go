package media

import (
	"encoding/json"
	"net/http"
	"strconv"

	models "server/internal/models/modules/media"
	services "server/internal/services/modules/media"
)

type PhotoController struct {
	Service *services.PhotoService
}

func NewPhotoController(service *services.PhotoService) *PhotoController {
	return &PhotoController{Service: service}
}

// Create a new photo
func (c *PhotoController) CreatePhotoHandler(w http.ResponseWriter, r *http.Request) {
	var photo models.PhotoMedia
	if err := json.NewDecoder(r.Body).Decode(&photo); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.Service.CreatePhoto(&photo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(photo)
}

// Get photo by ID
func (c *PhotoController) GetPhotoByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid photo ID", http.StatusBadRequest)
		return
	}

	photo, err := c.Service.GetPhotoByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(photo)
}

// Get all photos for a carepartner and caregiver
func (c *PhotoController) GetPhotosByCarepartnerAndCaregiverHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, _ := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	caregiverID, _ := strconv.Atoi(r.URL.Query().Get("caregiver_id"))

	photos, err := c.Service.GetPhotosByCarepartnerAndCaregiver(carepartnerID, caregiverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(photos)
}

// Update an existing photo
func (c *PhotoController) UpdatePhotoHandler(w http.ResponseWriter, r *http.Request) {
	var photo models.PhotoMedia
	if err := json.NewDecoder(r.Body).Decode(&photo); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.Service.UpdatePhoto(&photo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(photo)
}

// Delete a photo
func (c *PhotoController) DeletePhotoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid photo ID", http.StatusBadRequest)
		return
	}

	if err := c.Service.DeletePhoto(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Get photo stats
func (c *PhotoController) GetPhotoStatsHandler(w http.ResponseWriter, r *http.Request) {
	count, err := c.Service.GetPhotoStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"total_photos": count})
}
