package media

import (
	"encoding/json"
	"net/http"
	"strconv"

	models "server/internal/models/modules/media"
)

// ---- Interface ----
type VideoServiceInterface interface {
	CreateVideo(video *models.VideoMedia) error
	GetVideoByID(id int) (*models.VideoMedia, error)
	GetVideosByCarepartnerAndCaregiver(carepartnerID, caregiverID int) ([]models.VideoMedia, error)
	UpdateVideo(video *models.VideoMedia) error
	DeleteVideo(id int) error
	GetVideoStats() (int, error)
}

// ---- Controller ----
type VideoController struct {
	services VideoServiceInterface
}

func NewVideoController(services VideoServiceInterface) *VideoController {
	return &VideoController{services: services}
}

// Create a new video
func (c *VideoController) CreateVideoHandler(w http.ResponseWriter, r *http.Request) {
	var video models.VideoMedia
	if err := json.NewDecoder(r.Body).Decode(&video); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.services.CreateVideo(&video); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(video)
}

// Get video by ID
func (c *VideoController) GetVideoByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid video ID", http.StatusBadRequest)
		return
	}

	video, err := c.services.GetVideoByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(video)
}

// Get all videos for a carepartner and caregiver
func (c *VideoController) GetVideosByCarepartnerAndCaregiverHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, _ := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	caregiverID, _ := strconv.Atoi(r.URL.Query().Get("caregiver_id"))

	videos, err := c.services.GetVideosByCarepartnerAndCaregiver(carepartnerID, caregiverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(videos)
}

// Update an existing video
func (c *VideoController) UpdateVideoHandler(w http.ResponseWriter, r *http.Request) {
	var video models.VideoMedia
	if err := json.NewDecoder(r.Body).Decode(&video); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.services.UpdateVideo(&video); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(video)
}

// Delete a video
func (c *VideoController) DeleteVideoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid video ID", http.StatusBadRequest)
		return
	}

	if err := c.services.DeleteVideo(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Get video stats
func (c *VideoController) GetVideoStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := c.services.GetVideoStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}
