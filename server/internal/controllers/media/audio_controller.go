package media

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	models "server/internal/models/media"
)

// ---- Interface ----
type AudioServiceInterface interface {
	CreateAudio(a *models.AudioMedia) error
	CreateAudioFile(f multipart.File, h *multipart.FileHeader, uri string) error
	GetAudioByID(id int) (*models.AudioMedia, error)
	GetAudioFile(uri string) (multipart.File, error)
	GetAudiosByCarepartnerAndCaregiver(carepartnerID, caregiverID int) ([]models.AudioMedia, error)
	UpdateAudio(a *models.AudioMedia) error
	UpdateAudioFile(f multipart.File, h *multipart.FileHeader, uri string) error
	DeleteAudio(id int) error
	DeleteAudioFile(uri string) error
	SearchAudios(keyword string) ([]models.AudioMedia, error)
	GetMediaStats() (*models.MediaStats, error)
}

// ---- Controller ----
type AudioController struct {
	Services AudioServiceInterface
}

func NewAudioController(s AudioServiceInterface) *AudioController {
	return &AudioController{Services: s}
}

var VALID_AUDIO_TYPES = map[string]bool{
	"audio/mpeg":      true,
	"audio/mp3":       true,
	"audio/wav":       true,
	"audio/x-wav":     true,
	"audio/aac":       true,
	"audio/ogg":       true,
	"audio/webm":      true,
	"audio/mp4":       true,
}

// CreateAudioHandler godoc
// @Summary Create a new audio record
// @Description Creates a new audio entry in the database
// @Tags audio
// @Accept json
// @Produce json
// @Param audio body media.AudioMedia true "Audio data"
// @Success 200 {object} media.AudioMedia
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /audio [post]
func (c *AudioController) CreateAudioHandler(w http.ResponseWriter, r *http.Request) {
	var audio models.AudioMedia
	if err := json.NewDecoder(r.Body).Decode(&audio); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.Services.CreateAudio(&audio); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(audio)
}

// CreateAudioFileHandler godoc
// @Summary Upload a new audio file
// @Description Uploads an audio file to blob storage
// @Tags audio
// @Accept mpfd
// @Produce json
// @Param file formData file true "Audio file"
// @Param id formData string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /audio/file [post]
func (c *AudioController) CreateAudioFileHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read first 512 bytes for MIME detection
	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil {
		http.Error(w, "unable to read file content", http.StatusBadRequest)
		return
	}

	detectedType := http.DetectContentType(buf)
	if !VALID_AUDIO_TYPES[detectedType] {
		http.Error(w, "file is not a valid audio file", http.StatusBadRequest)
		return
	}

	// Reset file pointer
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "failed to reset file pointer", http.StatusInternalServerError)
		return
	}

	// Required form values
	userID := r.FormValue("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	caregiverIDStr := r.FormValue("caregiver_id")
	carepartnerIDStr := r.FormValue("carepartner_id")

	caregiverID, err := strconv.Atoi(caregiverIDStr)
	if err != nil || caregiverID <= 0 {
		http.Error(w, "invalid caregiver_id", http.StatusBadRequest)
		return
	}

	carepartnerID, err := strconv.Atoi(carepartnerIDStr)
	if err != nil || carepartnerID <= 0 {
		http.Error(w, "invalid carepartner_id", http.StatusBadRequest)
		return
	}

	// Build audio metadata
	uri := "users/" + userID + "/" + handler.Filename

	extension := ""
	if idx := strings.LastIndex(handler.Filename, "."); idx != -1 {
		extension = handler.Filename[idx:]
	}

	displayedName := r.FormValue("displayed_name")
	if displayedName == "" {
		displayedName = handler.Filename
	}

	// Upload file to MinIO
	if err := c.Services.CreateAudioFile(file, handler, uri); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	audio := &models.AudioMedia{
		URI:           uri,
		Filename:      handler.Filename,
		DisplayedName: displayedName,
		Extension:     extension,
	}

	if err := c.Services.CreateAudio(audio); err != nil {
		http.Error(w, "failed to save audio metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":             audio.ID,
		"filename":       audio.Filename,
		"displayed_name": audio.DisplayedName,
		"uri":            audio.URI,
		"extension":      audio.Extension,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetAudioByIDHandler godoc
// @Summary Get audio metadata by ID
// @Description Retrieves audio metadata from the database
// @Tags audio
// @Produce json
// @Param id query int true "Audio ID"
// @Success 200 {object} media.AudioMedia
// @Failure 400 {string} string "invalid audio ID"
// @Failure 500 {string} string "internal server error"
// @Router /audio [get]
func (c *AudioController) GetAudioByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	audio, err := c.Services.GetAudioByID(id)
	if err != nil {
		http.Error(w, "Audio not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%+v", audio)
}

// GetAudioFileByIDHandler godoc
// @Summary Download audio file by ID
// @Description Retrieves the audio file from blob storage
// @Tags audio
// @Produce application/octet-stream
// @Param id query int true "Audio ID"
// @Success 200 {file} file "Audio file"
// @Failure 400 {string} string "invalid audio ID"
// @Failure 500 {string} string "internal server error"
// @Router /audio/file [get]
func (c *AudioController) GetAudioFileByIDHandler(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "uri required", http.StatusBadRequest)
		return
	}

	obj, err := c.Services.GetAudioFile(uri)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch file: %v", err), http.StatusInternalServerError)
		return
	}

	io.Copy(w, obj)
}

// GetAudiosByCarepartnerAndCaregiverHandler godoc
// @Summary Get audios by carepartner and caregiver
// @Description Retrieves a list of audios filtered by carepartner ID and caregiver ID
// @Tags audio
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param caregiver_id query int true "Caregiver ID"
// @Success 200 {array} media.AudioMedia
// @Failure 400 {string} string "invalid or missing parameters"
// @Failure 500 {string} string "internal server error"
// @Router /audio/list [get]
func (c *AudioController) GetAudiosByCarepartnerAndCaregiverHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, _ := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	caregiverID, _ := strconv.Atoi(r.URL.Query().Get("caregiver_id"))

	audios, err := c.Services.GetAudiosByCarepartnerAndCaregiver(carepartnerID, caregiverID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get audios: %v", err), http.StatusInternalServerError)
		return
	}
	
	if audios == nil {
		audios = []models.AudioMedia{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(audios)
}

// UpdateAudioHandler godoc
// @Summary Update an audio metadata record
// @Description Updates audio metadata in the database
// @Tags audio
// @Accept json
// @Produce json
// @Param audio body media.AudioMedia true "Audio data"
// @Success 200 {object} media.AudioMedia
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /audio [put]
func (c *AudioController) UpdateAudioHandler(w http.ResponseWriter, r *http.Request) {
	var audio models.AudioMedia

	id, _ := strconv.Atoi(r.FormValue("id"))
	audio.ID = id
	audio.URI = r.FormValue("uri")
	audio.DisplayedName = r.FormValue("displayed_name")
	audio.Filename = r.FormValue("filename")
	audio.Extension = r.FormValue("extension")

	err := c.Services.UpdateAudio(&audio)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update audio: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Audio updated")
}

// UpdateAudioFileHandler godoc
// @Summary Update an audio file
// @Description Uploads and replaces an existing audio file
// @Tags audio
// @Accept mpfd
// @Produce json
// @Param file formData file true "Audio file"
// @Param id formData string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /audio/file [put]
func (c *AudioController) UpdateAudioFileHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)

	uri := r.FormValue("uri")
	if uri == "" {
		http.Error(w, "uri required", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed reading file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	err = c.Services.UpdateAudioFile(file, handler, uri)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update file: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Audio file updated")
}

// DeleteAudioHandler godoc
// @Summary Delete an audio metadata record
// @Description Deletes audio metadata from the database
// @Tags audio
// @Param id query int true "Audio ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "invalid audio ID"
// @Failure 500 {string} string "internal server error"
// @Router /audio [delete]
func (c *AudioController) DeleteAudioHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))

	err := c.Services.DeleteAudio(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Audio deleted")
}

// DeleteAudioFileHandler godoc
// @Summary Delete an audio file
// @Description Deletes audio file from blob storage
// @Tags audio
// @Param id query int true "Audio ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "invalid audio ID"
// @Failure 500 {string} string "internal server error"
// @Router /audio/file [delete]
func (c *AudioController) DeleteAudioFileHandler(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Query().Get("uri")

	err := c.Services.DeleteAudioFile(uri)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete file: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Audio file deleted")
}

// SearchAudiosHandler godoc
// @Summary Search audio entries by keyword
// @Description Performs keyword search on audio metadata
// @Tags audio
// @Produce json
// @Param keyword query string true "Search keyword"
// @Success 200 {array} media.AudioMedia
// @Failure 400 {string} string "keyword is required"
// @Failure 500 {string} string "internal server error"
// @Router /audio/search [get]
func (c *AudioController) SearchAudiosHandler(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")

	audios, err := c.Services.SearchAudios(keyword)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%+v", audios)
}

// GetMediaStatsHandler godoc
// @Summary Get media statistics
// @Description Returns statistics for media entries
// @Tags audio
// @Produce json
// @Success 200 {object} media.MediaStats
// @Failure 500 {string} string "internal server error"
// @Router /audio/stats [get]
func (c *AudioController) GetMediaStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := c.Services.GetMediaStats()
	if err != nil {
		http.Error(w, "failed to get stats", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%+v", stats)
}
