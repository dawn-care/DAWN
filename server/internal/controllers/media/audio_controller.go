package media

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	models "server/internal/models/media"
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
	CreateAudio(audio *models.AudioMedia) error
	CreateAudioFile(file multipart.File, handler *multipart.FileHeader, uri string) error
	GetAudioByID(id int) (*models.AudioMedia, error)
	GetAudioFile(uri string) (multipart.File, error)
	GetAudiosByCarepartnerAndCaregiver(carepartnerID, caregiverID int) ([]models.AudioMedia, error)
	UpdateAudio(audio *models.AudioMedia) error
	UpdateAudioFile(file multipart.File, handler *multipart.FileHeader, uri string) error
	DeleteAudio(id int) error
	DeleteAudioFile(uri string) error
	SearchAudios(keyword string) ([]models.AudioMedia, error)
	GetMediaStats() (*models.MediaStats, error)
}

// ---- Controller ----
type AudioController struct {
	services AudioServiceInterface
}

func NewAudioController(services AudioServiceInterface) *AudioController {
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

// Create a new audio file
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

// Get audio file by ID
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

// Update an audio file
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

// Delete an audio file
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
