package media

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	models "server/internal/models/modules/media"
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
	CreatePhoto(photo *models.PhotoMedia) error
	CreatePhotoFile(file multipart.File, handler *multipart.FileHeader, uri string) error
	GetPhotoByID(id int) (*models.PhotoMedia, error)
	GetPhotoFile(uri string) (multipart.File, error)
	GetPhotosByCarepartnerAndCaregiver(carepartnerID, caregiverID int) ([]models.PhotoMedia, error)
	UpdatePhoto(photo *models.PhotoMedia) error
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

// Create a new photo
func (c *PhotoController) CreatePhotoHandler(w http.ResponseWriter, r *http.Request) {
	var photo models.PhotoMedia
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

// Create a new photo file
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

// Get photo by ID
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

// Get photo file by ID
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
    w.Header().Set("Content-Disposition", "attachment; filename=\"" + photo.Filename + "\"")

    io.Copy(w, file)
}


// Get all photos for a carepartner and caregiver
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

// Update an existing photo
func (c *PhotoController) UpdatePhotoHandler(w http.ResponseWriter, r *http.Request) {
	var photo models.PhotoMedia
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

// Update a photo file
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

// Delete a photo
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

// Delete a photo file
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

// Get photo stats
func (c *PhotoController) GetPhotoStatsHandler(w http.ResponseWriter, r *http.Request) {
	count, err := c.services.GetPhotoStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"total_photos": count})
}
