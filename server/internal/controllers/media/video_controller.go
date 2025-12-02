package media

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

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
	services VideoServiceInterface
}

func NewVideoController(services VideoServiceInterface) *VideoController {
	return &VideoController{services: services}
}

// CreateVideoHandler godoc
// @Summary Create a new video record
// @Description Creates a new video entry in the database
// @Tags video
// @Accept json
// @Produce json
// @Param video body media.VideoMedia true "Video data"
// @Success 200 {object} media.VideoMedia
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /video [post]
func (c *VideoController) CreateVideoHandler(w http.ResponseWriter, r *http.Request) {
	var video media.VideoMedia
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

// CreateVideoFileHandler godoc
// @Summary Upload a video file
// @Description Uploads a new video file for a user
// @Tags video
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Video file"
// @Param id formData string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "invalid file or request"
// @Failure 500 {string} string "internal server error"
// @Router /video/file [post]
func (c *VideoController) CreateVideoFileHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "unable to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

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

	file.Seek(0, io.SeekStart)

	userID := r.FormValue("id")
	uri := "users/" + userID + "/" + handler.Filename

	if err := c.services.CreateVideoFile(file, handler, uri); err != nil {
		http.Error(w, "failed to save video: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetVideoByIDHandler godoc
// @Summary Get a video record by ID
// @Description Retrieves metadata for a video
// @Tags video
// @Accept json
// @Produce json
// @Param id query int true "Video ID"
// @Success 200 {object} media.VideoMedia
// @Failure 400 {string} string "invalid video ID"
// @Failure 500 {string} string "internal server error"
// @Router /video [get]
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

// GetVideoFileByIDHandler godoc
// @Summary Download a video file by ID
// @Description Returns the raw video file for a given video ID
// @Tags video
// @Accept json
// @Produce application/octet-stream
// @Param id query int true "Video ID"
// @Success 200 {file} file "Video file"
// @Failure 400 {string} string "invalid video ID"
// @Failure 500 {string} string "internal server error"
// @Router /video/file [get]
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

	videos, err := c.services.GetVideosByCarepartnerAndCaregiver(carepartnerID, caregiverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(videos)
}

// UpdateVideoHandler godoc
// @Summary Update a video record
// @Description Updates metadata for an existing video
// @Tags video
// @Accept json
// @Produce json
// @Param video body media.VideoMedia true "Video data"
// @Success 200 {object} media.VideoMedia
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /video [put]
func (c *VideoController) UpdateVideoHandler(w http.ResponseWriter, r *http.Request) {
	var video media.VideoMedia
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

// UpdateVideoFileHandler godoc
// @Summary Update a video file
// @Description Updates the uploaded video file for a user
// @Tags video
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Video file"
// @Param id formData string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "invalid file or request"
// @Failure 500 {string} string "internal server error"
// @Router /video/file [put]
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

// DeleteVideoHandler godoc
// @Summary Delete a video record
// @Description Deletes a video metadata record by ID
// @Tags video
// @Accept json
// @Produce json
// @Param id query int true "Video ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "invalid video ID"
// @Failure 500 {string} string "internal server error"
// @Router /video [delete]
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

// DeleteVideoFileHandler godoc
// @Summary Delete a video file
// @Description Deletes the uploaded video file for a given video ID
// @Tags video
// @Accept json
// @Produce json
// @Param id query int true "Video ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "invalid video ID"
// @Failure 500 {string} string "internal server error"
// @Router /video/file [delete]
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

// GetVideoStatsHandler godoc
// @Summary Get video statistics
// @Description Returns the total number of uploaded videos
// @Tags video
// @Accept json
// @Produce json
// @Success 200 {object} map[string]int "total_videos"
// @Failure 500 {string} string "internal server error"
// @Router /video/stats [get]
func (c *VideoController) GetVideoStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := c.services.GetVideoStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}
