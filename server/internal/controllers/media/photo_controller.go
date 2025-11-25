package media

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"server/internal/models/media"
)

var VALID_IMAGE_TYPES = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

const MAX_PHOTO_FILE_SIZE_MB = 100

// ---- Interface ----
type PhotoServiceInterface interface {
	CreatePhoto(photo *media.PhotoMedia) error
	CreatePhotoFile(file multipart.File, handler *multipart.FileHeader, uri string) error
	GetPhotoByID(id int) (*media.PhotoMedia, error)
	GetPhotoFile(uri string) (multipart.File, error)
	GetPhotosByCarepartnerAndCaregiver(carepartnerID, caregiverID int) ([]media.PhotoMedia, error)
	UpdatePhoto(photo *media.PhotoMedia) error
	UpdatePhotoFile(file multipart.File, handler *multipart.FileHeader, uri string) error
	DeletePhoto(id int) error
	DeletePhotoFile(uri string) error
	GetPhotoStats() (int, error)
}

// ---- Controller ----
type PhotoController struct {
	services PhotoServiceInterface
}

func NewPhotoController(services PhotoServiceInterface) *PhotoController {
	return &PhotoController{services: services}
}

// CreatePhotoHandler godoc
// @Summary Create a new photo record
// @Description Creates a new photo entry in the database
// @Tags photo
// @Accept json
// @Produce json
// @Param photo body media.PhotoMedia true "Photo data"
// @Success 200 {object} media.PhotoMedia
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /photo [post]
func (c *PhotoController) CreatePhotoHandler(w http.ResponseWriter, r *http.Request) {
	var photo media.PhotoMedia
	if err := json.NewDecoder(r.Body).Decode(&photo); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.services.CreatePhoto(&photo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(photo)
}

// CreatePhotoFileHandler godoc
// @Summary Upload a photo file
// @Description Uploads a new photo file for a user
// @Tags photo
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Photo file"
// @Param id formData string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "invalid file or request"
// @Failure 500 {string} string "internal server error"
// @Router /photo/file [post]
func (c *PhotoController) CreatePhotoFileHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(MAX_PHOTO_FILE_SIZE_MB << 20); err != nil {
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
		http.Error(w, "file is not a valid photo file", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("id")

	uri := "users/" + userID + "/" + handler.Filename

	if err := c.services.CreatePhotoFile(file, handler, uri); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"filename": handler.Filename,
	}

	json.NewEncoder(w).Encode(response)
}

// GetPhotoByIDHandler godoc
// @Summary Get a photo record by ID
// @Description Retrieves metadata for a photo
// @Tags photo
// @Accept json
// @Produce json
// @Param id query int true "Photo ID"
// @Success 200 {object} media.PhotoMedia
// @Failure 400 {string} string "invalid photo ID"
// @Failure 500 {string} string "internal server error"
// @Router /photo [get]
func (c *PhotoController) GetPhotoByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid photo ID", http.StatusBadRequest)
		return
	}

	photo, err := c.services.GetPhotoByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(photo)
}

// GetPhotoFileByIDHandler godoc
// @Summary Download a photo file by ID
// @Description Returns the raw photo file for a given photo ID
// @Tags photo
// @Accept json
// @Produce application/octet-stream
// @Param id query int true "Photo ID"
// @Success 200 {file} file "Photo file"
// @Failure 400 {string} string "invalid photo ID"
// @Failure 500 {string} string "internal server error"
// @Router /photo/file [get]
func (c *PhotoController) GetPhotoFileByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid photo ID", http.StatusBadRequest)
		return
	}

	photo, err := c.services.GetPhotoByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, err := c.services.GetPhotoFile(photo.URI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+photo.Filename+"\"")

	io.Copy(w, file)
}

// GetPhotosByCarepartnerAndCaregiverHandler godoc
// @Summary Get all photos for a carepartner and caregiver
// @Description Returns all photo records associated with a carepartner and caregiver
// @Tags photo
// @Accept json
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param caregiver_id query int true "Caregiver ID"
// @Success 200 {array} media.PhotoMedia
// @Failure 500 {string} string "internal server error"
// @Router /photo/list [get]
func (c *PhotoController) GetPhotosByCarepartnerAndCaregiverHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, _ := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	caregiverID, _ := strconv.Atoi(r.URL.Query().Get("caregiver_id"))

	photos, err := c.services.GetPhotosByCarepartnerAndCaregiver(carepartnerID, caregiverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(photos)
}

// UpdatePhotoHandler godoc
// @Summary Update a photo record
// @Description Updates metadata for an existing photo
// @Tags photo
// @Accept json
// @Produce json
// @Param photo body media.PhotoMedia true "Photo data"
// @Success 200 {object} media.PhotoMedia
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /photo [put]
func (c *PhotoController) UpdatePhotoHandler(w http.ResponseWriter, r *http.Request) {
	var photo media.PhotoMedia
	if err := json.NewDecoder(r.Body).Decode(&photo); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.services.UpdatePhoto(&photo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(photo)
}

// UpdatePhotoFileHandler godoc
// @Summary Update a photo file
// @Description Updates the uploaded photo file for a user
// @Tags photo
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Photo file"
// @Param id formData string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "invalid file or request"
// @Failure 500 {string} string "internal server error"
// @Router /photo/file [put]
func (c *PhotoController) UpdatePhotoFileHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(MAX_PHOTO_FILE_SIZE_MB << 20); err != nil {
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
		http.Error(w, "file is not a valid photo file", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("id")

	uri := "users/" + userID + "/" + handler.Filename

	if err := c.services.UpdatePhotoFile(file, handler, uri); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"filename": handler.Filename,
	}

	json.NewEncoder(w).Encode(response)
}

// DeletePhotoHandler godoc
// @Summary Delete a photo record
// @Description Deletes a photo metadata record by ID
// @Tags photo
// @Accept json
// @Produce json
// @Param id query int true "Photo ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "invalid photo ID"
// @Failure 500 {string} string "internal server error"
// @Router /photo [delete]
func (c *PhotoController) DeletePhotoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid photo ID", http.StatusBadRequest)
		return
	}

	if err := c.services.DeletePhoto(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeletePhotoFileHandler godoc
// @Summary Delete a photo file
// @Description Deletes the uploaded photo file for a given photo ID
// @Tags photo
// @Accept json
// @Produce json
// @Param id query int true "Photo ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "invalid photo ID"
// @Failure 500 {string} string "internal server error"
// @Router /photo/file [delete]
func (c *PhotoController) DeletePhotoFileHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid photo ID", http.StatusBadRequest)
		return
	}

	photo, err := c.services.GetPhotoByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := c.services.DeletePhotoFile(photo.URI); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetPhotoStatsHandler godoc
// @Summary Get photo statistics
// @Description Returns the total number of uploaded photos
// @Tags photo
// @Accept json
// @Produce json
// @Success 200 {object} map[string]int "total_photos"
// @Failure 500 {string} string "internal server error"
// @Router /photo/stats [get]
func (c *PhotoController) GetPhotoStatsHandler(w http.ResponseWriter, r *http.Request) {
	count, err := c.services.GetPhotoStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"total_photos": count})
}
