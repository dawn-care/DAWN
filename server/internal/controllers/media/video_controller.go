package media

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"server/internal/models/media"
)

var VALID_VIDEO_TYPES = map[string]bool{
	"video/mp4":  true,
	"video/webm": true,
	"video/ogg":  true,
}

const MAX_VIDEO_FILE_SIZE_MB = 100

// ---- Interface ----
type VideoServiceInterface interface {
	CreateVideo(video *media.VideoMedia) error
	CreateVideoFile(file multipart.File, handler *multipart.FileHeader, uri string) error
	GetVideoByID(id int) (*media.VideoMedia, error)
	GetVideoFile(uri string) (multipart.File, error)
	GetVideosByCarepartnerAndCaregiver(carepartnerID, caregiverID int) ([]media.VideoMedia, error)
	UpdateVideo(video *media.VideoMedia) error
	UpdateVideoFile(file multipart.File, handler *multipart.FileHeader, uri string) error
	DeleteVideo(id int) error
	DeleteVideoFile(uri string) error
	GetVideoStats() (int, error)
}

// ---- Controller ----
type VideoController struct {
	Services VideoServiceInterface
}

func NewVideoController(Services VideoServiceInterface) *VideoController {
	return &VideoController{Services: Services}
}

// CreateVideoHandler godoc
func (c *VideoController) CreateVideoHandler(w http.ResponseWriter, r *http.Request) {
	var video media.VideoMedia
	if err := json.NewDecoder(r.Body).Decode(&video); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.Services.CreateVideo(&video); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(video)
}

// CreateVideoFileHandler godoc
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

	// Validate MIME
	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		http.Error(w, "unable to read file", http.StatusBadRequest)
		return
	}

	detectedType := http.DetectContentType(buf)
	if !VALID_VIDEO_TYPES[detectedType] {
		http.Error(w, "file is not a valid video file", http.StatusBadRequest)
		return
	}

	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	userID := r.FormValue("id")
	displayedName := r.FormValue("displayed_name")

	if displayedName == "" {
		http.Error(w, "displayed_name is required", http.StatusBadRequest)
		return
	}

	uri := "users/" + userID + "/" + handler.Filename
	extension := "." + handler.Filename[strings.LastIndex(handler.Filename, ".")+1:]

	// Upload to storage
	if err := c.Services.CreateVideoFile(file, handler, uri); err != nil {
		http.Error(w, "failed to save video file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create DB record
	video := &media.VideoMedia{
		URI:           uri,
		Filename:      handler.Filename,
		Extension:     extension,
		DisplayedName: displayedName,
	}

	if err := c.Services.CreateVideo(video); err != nil {
		http.Error(w, "failed to save video metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":             video.ID,
		"filename":       video.Filename,
		"displayed_name": video.DisplayedName,
		"uri":            video.URI,
		"extension":      video.Extension,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetVideosByCarepartnerAndCaregiverHandler godoc
// @Summary Get all videos for a carepartner and caregiver
// @Description Returns all video records associated with a carepartner and caregiver
// @Tags video
// @Accept json
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param caregiver_id query int true "Caregiver ID"
// @Success 200 {array} media.VideoMedia
// @Failure 500 {string} string "internal server error"
// @Router /video/list [get]
func (c *VideoController) GetVideosByCarepartnerAndCaregiverHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, _ := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	caregiverID, _ := strconv.Atoi(r.URL.Query().Get("caregiver_id"))

	videos, err := c.Services.GetVideosByCarepartnerAndCaregiver(carepartnerID, caregiverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(videos)
}

// GetVideoByIDHandler godoc
func (c *VideoController) GetVideoByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid video ID", http.StatusBadRequest)
		return
	}

	video, err := c.Services.GetVideoByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(video)
}

// GetVideoFileByIDHandler godoc
func (c *VideoController) GetVideoFileByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid video ID", http.StatusBadRequest)
		return
	}

	video, err := c.Services.GetVideoByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, err := c.Services.GetVideoFile(video.URI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+video.Filename+"\"")

	io.Copy(w, file)
}

// UpdateVideoHandler godoc
func (c *VideoController) UpdateVideoHandler(w http.ResponseWriter, r *http.Request) {
	var video media.VideoMedia
	if err := json.NewDecoder(r.Body).Decode(&video); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.Services.UpdateVideo(&video); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(video)
}

// UpdateVideoFileHandler godoc
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
	if !VALID_VIDEO_TYPES[detectedType] {
		http.Error(w, "file is not a valid video file", http.StatusBadRequest)
		return
	}

	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	userID := r.FormValue("id")
	uri := "users/" + userID + "/" + handler.Filename

	if err := c.Services.UpdateVideoFile(file, handler, uri); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"filename": handler.Filename,
	})
}

// DeleteVideoHandler godoc
func (c *VideoController) DeleteVideoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid video ID", http.StatusBadRequest)
		return
	}

	if err := c.Services.DeleteVideo(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteVideoFileHandler godoc
func (c *VideoController) DeleteVideoFileHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid video ID", http.StatusBadRequest)
		return
	}

	video, err := c.Services.GetVideoByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := c.Services.DeleteVideoFile(video.URI); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetVideoStatsHandler godoc
func (c *VideoController) GetVideoStatsHandler(w http.ResponseWriter, r *http.Request) {
	count, err := c.Services.GetVideoStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"total_videos": count})
}
