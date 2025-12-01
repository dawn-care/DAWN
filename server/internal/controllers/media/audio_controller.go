package media

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	models "server/internal/models/media"
)

// ---- Interface ----
type AudioServiceInterface interface {
	CreateAudio(a *models.AudioMedia) error
	CreateAudioFile(f multipart.File, h *multipart.FileHeader, uri string) error
	GetAudioByID(id int) (*models.AudioMedia, error)
	GetAudioFile(uri string) (multipart.File, error)
	UpdateAudio(a *models.AudioMedia) error
	UpdateAudioFile(f multipart.File, h *multipart.FileHeader, uri string) error
	DeleteAudio(id int) error
	DeleteAudioFile(uri string) error
	SearchAudios(keyword string) ([]models.AudioMedia, error)
	GetMediaStats() (*models.MediaStats, error)
}

// ---- Controller ----
type AudioController struct {
	services AudioServiceInterface
}

func NewAudioController(s AudioServiceInterface) *AudioController {
	return &AudioController{services: s}
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

	audio.URI = r.FormValue("uri")
	audio.Filename = r.FormValue("filename")
	audio.Extension = r.FormValue("extension")

	err := c.services.CreateAudio(&audio)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create audio: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Audio created with ID %d", audio.ID)
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
	r.ParseMultipartForm(32 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := handler.Filename
	extension := ""

	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			extension = filename[i:]
			break
		}
	}

	uri := fmt.Sprintf("audio/%s", filename)

	audio := models.AudioMedia{
		URI:       uri,
		Filename:  filename,
		Extension: extension,
	}

	// Save metadata
	err = c.services.CreateAudio(&audio)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to save audio metadata: %v", err), http.StatusInternalServerError)
		return
	}

	// Upload file
	err = c.services.CreateAudioFile(file, handler, uri)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to upload audio file: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Audio uploaded successfully with ID %d", audio.ID)
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

	audio, err := c.services.GetAudioByID(id)
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

	obj, err := c.services.GetAudioFile(uri)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch file: %v", err), http.StatusInternalServerError)
		return
	}

	io.Copy(w, obj)
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
	audio.Filename = r.FormValue("filename")
	audio.Extension = r.FormValue("extension")

	err := c.services.UpdateAudio(&audio)
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

	err = c.services.UpdateAudioFile(file, handler, uri)
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

	err := c.services.DeleteAudio(id)
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

	err := c.services.DeleteAudioFile(uri)
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

	audios, err := c.services.SearchAudios(keyword)
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
	stats, err := c.services.GetMediaStats()
	if err != nil {
		http.Error(w, "failed to get stats", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%+v", stats)
}
