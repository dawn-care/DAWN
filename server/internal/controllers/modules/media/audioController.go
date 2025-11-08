package media

import (
	"encoding/json"
	"net/http"
	"strconv"

	models "server/internal/models/modules/media"
	services "server/internal/services/modules/media"
)

type AudioController struct {
	services *services.AudioService
}

func NewAudioController(services *services.AudioService) *AudioController {
	return &AudioController{services: services}
}

// Create a new audio
func (c *AudioController) CreateAudioHandler(w http.ResponseWriter, r *http.Request) {
	var audio models.AudioMedia
	if err := json.NewDecoder(r.Body).Decode(&audio); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.services.CreateAudio(&audio); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(audio)
}

// Get audio by ID
func (c *AudioController) GetAudioByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid audio ID", http.StatusBadRequest)
		return
	}

	audio, err := c.services.GetAudioByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(audio)
}

// Get all audios for a carepartner and caregiver
func (c *AudioController) GetAudiosByCarepartnerAndCaregiverHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, _ := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	caregiverID, _ := strconv.Atoi(r.URL.Query().Get("caregiver_id"))

	audios, err := c.services.GetAudiosByCarepartnerAndCaregiver(carepartnerID, caregiverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(audios)
}

// Update an audio
func (c *AudioController) UpdateAudioHandler(w http.ResponseWriter, r *http.Request) {
	var audio models.AudioMedia
	if err := json.NewDecoder(r.Body).Decode(&audio); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.services.UpdateAudio(&audio); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(audio)
}

// Delete an audio
func (c *AudioController) DeleteAudioHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid audio ID", http.StatusBadRequest)
		return
	}

	if err := c.services.DeleteAudio(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Search audios by keyword
func (c *AudioController) SearchAudiosHandler(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		http.Error(w, "keyword is required", http.StatusBadRequest)
		return
	}

	results, err := c.services.SearchAudios(keyword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}

// Get media stats
func (c *AudioController) GetMediaStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := c.services.GetMediaStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}
