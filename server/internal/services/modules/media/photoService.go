package media

import (
	"database/sql"
	"errors"
	"fmt"
	
	"server/internal/models/modules/media"
)

type PhotoService struct {
	DB *sql.DB
}

func NewPhotoService(db *sql.DB) *PhotoService {
	return &PhotoService{DB: db}
}

// Create a single photo
func (s *PhotoService) CreatePhoto(photo *media.PhotoMedia) error {
	if photo == nil {
		return errors.New("photo cannot be nil")
	}
	if photo.URI == "" || photo.Filename == "" || photo.Extension == "" {
		return errors.New("uri, filename, and extension are required")
	}
	return media.CreatePhoto(s.DB, photo)
}

// Create multiple photos in bulk
func (s *PhotoService) CreatePhotosBulk(photos []media.PhotoMedia) error {
	if len(photos) == 0 {
		return errors.New("no photos to insert")
	}
	return media.CreatePhotosBulk(s.DB, photos)
}

// Retrieve a single photo by ID
func (s *PhotoService) GetPhotoByID(id int) (*media.PhotoMedia, error) {
	if id <= 0 {
		return nil, errors.New("invalid photo ID")
	}
	photo, err := media.GetPhotoByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get photo: %w", err)
	}
	if photo == nil {
		return nil, fmt.Errorf("photo with id %d not found", id)
	}
	return photo, nil
}

// Retrieve all photos for a given carepartner and caregiver
func (s *PhotoService) GetPhotosByCarepartnerAndCaregiver(carepartnerID, caregiverID int) ([]media.PhotoMedia, error) {
	if carepartnerID <= 0 || caregiverID <= 0 {
		return nil, errors.New("invalid carepartner or caregiver ID")
	}
	return media.GetPhotosByCarepartnerAndCaregiver(s.DB, carepartnerID, caregiverID)
}

// Update an existing photo
func (s *PhotoService) UpdatePhoto(photo *media.PhotoMedia) error {
	if photo == nil {
		return errors.New("photo cannot be nil")
	}
	if photo.ID <= 0 {
		return errors.New("invalid photo ID")
	}
	return media.UpdatePhoto(s.DB, photo)
}

// Delete a photo
func (s *PhotoService) DeletePhoto(id int) error {
	if id <= 0 {
		return errors.New("invalid photo ID")
	}
	return media.DeletePhoto(s.DB, id)
}

// Get total photo count
func (s *PhotoService) GetPhotoStats() (int, error) {
	return media.GetPhotoStats(s.DB)
}
