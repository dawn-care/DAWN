package media

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"server/internal/models/media"
)

var VALID_AUDIO_TYPES = map[string]bool{
	"audio/mpeg": true,
	"audio/wav":  true,
	"audio/ogg":  true,
	"audio/flac": true,
	"audio/aac":  true,
}

const MAX_AUDIO_FILE_SIZE_MB = 100

// ---- Interface ----
type AudioServiceInterface interface {
	CreateAudio(audio *media.AudioMedia) error
	CreateAudioFile(file multipart.File, handler *multipart.FileHeader, uri string) error
	GetAudioByID(id int) (*media.AudioMedia, error)
	GetAudioFile(uri string) (multipart.File, error)
	GetAudiosByCarepartnerAndCaregiver(carepartnerID, caregiverID int) ([]media.AudioMedia, error)
	UpdateAudio(audio *media.AudioMedia) error
	UpdateAudioFile(file multipart.File, handler *multipart.FileHeader, uri string) error
	DeleteAudio(id int) error
	DeleteAudioFile(uri string) error
	SearchAudios(keyword string) ([]media.AudioMedia, error)
	GetMediaStats() (*media.MediaStats, error)
}

// ---- Controller ----
type AudioController struct {
	services AudioServiceInterface
}

func NewAudioController(services AudioServiceInterface) *AudioController {
	return &AudioController{services: services}
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
	var audio media.AudioMedia
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
	if err := r.ParseMultipartForm(MAX_AUDIO_FILE_SIZE_MB << 20); err != nil {
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
	if !VALID_AUDIO_TYPES[detectedType] {
		http.Error(w, "file is not a valid audio file", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("id")

	uri := "users/" + userID + "/" + handler.Filename

	if err := c.services.CreateAudioFile(file, handler, uri); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"filename": handler.Filename,
	}

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

	file, err := c.services.GetAudioFile(audio.URI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+audio.Filename+"\"")

	io.Copy(w, file)
}

// GetAudiosByCarepartnerAndCaregiverHandler godoc
// @Summary Get all audios for a carepartner and caregiver
// @Description Retrieves all audio entries for given carepartner and caregiver
// @Tags audio
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param caregiver_id query int true "Caregiver ID"
// @Success 200 {array} media.AudioMedia
// @Failure 500 {string} string "internal server error"
// @Router /audio/list [get]
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
	var audio media.AudioMedia
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
	if err := r.ParseMultipartForm(MAX_AUDIO_FILE_SIZE_MB << 20); err != nil {
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
	if !VALID_AUDIO_TYPES[detectedType] {
		http.Error(w, "file is not a valid audio file", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("id")

	uri := "users/" + userID + "/" + handler.Filename

	if err := c.services.UpdateAudioFile(file, handler, uri); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"filename": handler.Filename,
	}

	json.NewEncoder(w).Encode(response)
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
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid audio ID", http.StatusBadRequest)
		return
	}

	audio, err := c.services.GetAudioByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := c.services.DeleteAudioFile(audio.URI); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

// GetMediaStatsHandler godoc
// @Summary Get media statistics
// @Description Returns statistics for media entries
// @Tags audio
// @Produce json
// @Success 200 {object} media.MediaStats
// @Failure 500 {string} string "internal server error"
// @Router /audio/stats [get]
func (c *AudioController) GetMediaStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := c.services.GetMediaStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}
