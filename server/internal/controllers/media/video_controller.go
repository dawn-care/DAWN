package media

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	models "server/internal/models/media"
)

var VALID_VIDEO_TYPES = map[string]bool{
	"video/mp4":  true,
	"video/webm": true,
	"video/ogg":  true,
}

const MAX_VIDEO_FILE_SIZE_MB = 100

// ---- Interface ----
type VideoServiceInterface interface {
	CreateVideo(video *models.VideoMedia) error
	CreateVideoFile(file multipart.File, handler *multipart.FileHeader, uri string) error
	GetVideoByID(id int) (*models.VideoMedia, error)
	GetVideoFile(uri string) (multipart.File, error)
	GetVideosByCarepartnerAndCaregiver(carepartnerID, caregiverID int) ([]models.VideoMedia, error)
	UpdateVideo(video *models.VideoMedia) error
	UpdateVideoFile(file multipart.File, handler *multipart.FileHeader, uri string) error
	DeleteVideo(id int) error
	DeleteVideoFile(uri string) error
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

// Create a new video file
func (c *VideoController) CreateVideoFileHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(MAX_VIDEO_FILE_SIZE_MB << 20); err != nil {
		http.Error(w, "failed to parse request", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		http.Error(w, "unable to read file content", http.StatusBadRequest)
		return
	}

	detectedType := http.DetectContentType(buf)
	if !VALID_IMAGE_TYPES[detectedType] {
		http.Error(w, "file is not a valid video file", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("id")

	uri := "users/" + userID + "/" + handler.Filename

	if err := c.services.CreateVideoFile(file, handler, uri); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"filename": handler.Filename,
	}

	json.NewEncoder(w).Encode(response)
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

// Get video file by ID
func (c *VideoController) GetVideoFileByIDHandler(w http.ResponseWriter, r *http.Request) {
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

	file, err := c.services.GetVideoFile(video.URI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+video.Filename+"\"")

	io.Copy(w, file)
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

// Updaate an existing video file
func (c *VideoController) UpdateVideoFileHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(MAX_VIDEO_FILE_SIZE_MB << 20); err != nil {
		http.Error(w, "failed to parse request", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		http.Error(w, "unable to read file content", http.StatusBadRequest)
		return
	}

	detectedType := http.DetectContentType(buf)
	if !VALID_IMAGE_TYPES[detectedType] {
		http.Error(w, "file is not a valid video file", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("id")

	uri := "users/" + userID + "/" + handler.Filename

	if err := c.services.UpdateVideoFile(file, handler, uri); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"filename": handler.Filename,
	}

	json.NewEncoder(w).Encode(response)
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

// Delete a video file
func (c *VideoController) DeleteVideoFileHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid video ID", http.StatusBadRequest)
		return
	}

	video, err := c.services.GetVideoByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := c.services.DeleteVideoFile(video.URI); err != nil {
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
