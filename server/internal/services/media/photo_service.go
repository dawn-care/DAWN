package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mime/multipart"

	"server/internal/models/media"

	"github.com/minio/minio-go/v7"
)

type PhotoService struct {
	DB     *sql.DB
	BLOB   *minio.Client
	Bucket string
}

func NewPhotoService(db *sql.DB, blob *minio.Client, bucket string) *PhotoService {
	return &PhotoService{DB: db, BLOB: blob, Bucket: bucket}
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

// Create a single photo file
func (s *PhotoService) CreatePhotoFile(file multipart.File, handler *multipart.FileHeader, uri string) error {
	if file == nil {
		return errors.New("file data cannot be nil")
	}
	if handler == nil {
		return errors.New("file header cannot be nil")
	}
	if uri == "" {
		return errors.New("uri cannot be empty")
	}
	contentType := handler.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := s.BLOB.PutObject(context.Background(), s.Bucket, uri, file, handler.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("unable to upload media to blob storage: %w", err)
	}
	return nil
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

// Retrieve a photo file by uri
func (s *PhotoService) GetPhotoFile(uri string) (multipart.File, error) {
	if uri == "" {
		return nil, errors.New("invalid filename")
	}
	file, err := s.BLOB.GetObject(context.Background(), s.Bucket, uri, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch photo: %w", err)
	}

	return file, nil
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

// Update an existing photo file
func (s *PhotoService) UpdatePhotoFile(file multipart.File, handler *multipart.FileHeader, uri string) error {
	if file == nil {
		return errors.New("file data cannot be nil")
	}
	if handler == nil {
		return errors.New("file header cannot be nil")
	}
	if uri == "" {
		return errors.New("uri cannot be empty")
	}
	contentType := handler.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	// PutObject will override existing file, thus "Update"
	_, err := s.BLOB.PutObject(context.Background(), s.Bucket, uri, file, handler.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("unable to update photo: %w", err)
	}
	return nil
}

// Delete a photo
func (s *PhotoService) DeletePhoto(id int) error {
	if id <= 0 {
		return errors.New("invalid photo ID")
	}
	return media.DeletePhoto(s.DB, id)
}

// Delete a photo file
func (s *PhotoService) DeletePhotoFile(uri string) error {
	if uri == "" {
		return errors.New("invalid filename")
	}
	removeErr := s.BLOB.RemoveObject(context.Background(), s.Bucket, uri, minio.RemoveObjectOptions{})
	if removeErr != nil {
		return fmt.Errorf("unable to delete photo: %w", removeErr)
	}
	return nil
}

// Get total photo count
func (s *PhotoService) GetPhotoStats() (int, error) {
	return media.GetPhotoStats(s.DB)
}
