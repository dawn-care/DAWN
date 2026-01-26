package media

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"server/internal/models/media"
)

type VideoService struct {
	DB     *sql.DB
	BLOB   *minio.Client
	Bucket string
}

func NewVideoService(db *sql.DB, blob *minio.Client, bucket string) *VideoService {
	return &VideoService{
		DB:     db,
		BLOB:   blob,
		Bucket: bucket,
	}
}

// Create a single video
func (s *VideoService) CreateVideo(video *media.VideoMedia) error {
	if video == nil {
		return errors.New("video cannot be nil")
	}
	if video.URI == "" || video.Filename == "" || video.Extension == "" {
		return errors.New("uri, filename, and extension are required")
	}
	return media.CreateVideo(s.DB, video)
}

// Create a single video file
func (s *VideoService) CreateVideoFile(file multipart.File, handler *multipart.FileHeader, uri string) error {
	if file == nil {
		return errors.New("file data cannot be nil")
	}
	if handler == nil {
		return errors.New("file header cannot be nil")
	}
	if uri == "" {
		return errors.New("uri cannot be empty")
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	contentType := handler.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload to blob storage
	_, err = s.BLOB.PutObject(
		context.Background(),
		s.Bucket,
		uri,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return fmt.Errorf("unable to upload video to blob storage: %w", err)
	}

	// Save metadata in DB
	video := &media.VideoMedia{
		URI:       uri,
		Filename:  handler.Filename,
		Extension: filepath.Ext(handler.Filename),
	}
	if err := s.CreateVideo(video); err != nil {
		return fmt.Errorf("failed to save video metadata to DB: %w", err)
	}

	return nil
}

// Create multiple videos in bulk
func (s *VideoService) CreateVideosBulk(videos []media.VideoMedia) error {
	if len(videos) == 0 {
		return errors.New("no videos to insert")
	}
	return media.CreateVideosBulk(s.DB, videos)
}

// Retrieve a single video by ID
func (s *VideoService) GetVideoByID(id int) (*media.VideoMedia, error) {
	if id <= 0 {
		return nil, errors.New("invalid video ID")
	}
	video, err := media.GetVideoByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get video: %w", err)
	}
	if video == nil {
		return nil, fmt.Errorf("video with id %d not found", id)
	}
	return video, nil
}

// Retrieve a video file by uri
func (s *VideoService) GetVideoFile(uri string) (multipart.File, error) {
	if uri == "" {
		return nil, errors.New("invalid filename")
	}
	file, err := s.BLOB.GetObject(context.Background(), s.Bucket, uri, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch video: %w", err)
	}
	return file, nil
}

// Retrieve all videos for a given carepartner and caregiver
func (s *VideoService) GetVideosByCarepartnerAndCaregiver(carepartnerID, caregiverID int) ([]media.VideoMedia, error) {
	if carepartnerID <= 0 || caregiverID <= 0 {
		return nil, errors.New("invalid carepartner or caregiver ID")
	}
	return media.GetVideosByCarepartnerAndCaregiver(s.DB, carepartnerID, caregiverID)
}

// Update an existing video
func (s *VideoService) UpdateVideo(video *media.VideoMedia) error {
	if video == nil {
		return errors.New("video cannot be nil")
	}
	if video.ID <= 0 {
		return errors.New("invalid video ID")
	}
	return media.UpdateVideo(s.DB, video)
}

// Update an existing video file
func (s *VideoService) UpdateVideoFile(file multipart.File, handler *multipart.FileHeader, uri string) error {
	if file == nil {
		return errors.New("file data cannot be nil")
	}
	if handler == nil {
		return errors.New("file header cannot be nil")
	}
	if uri == "" {
		return errors.New("uri cannot be empty")
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	contentType := handler.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = s.BLOB.PutObject(
		context.Background(),
		s.Bucket,
		uri,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return fmt.Errorf("unable to update video: %w", err)
	}

	return nil
}

// Delete a video
func (s *VideoService) DeleteVideo(id int) error {
	if id <= 0 {
		return errors.New("invalid video ID")
	}
	return media.DeleteVideo(s.DB, id)
}

// Delete a video file
func (s *VideoService) DeleteVideoFile(uri string) error {
	if uri == "" {
		return errors.New("invalid filename")
	}
	if err := s.BLOB.RemoveObject(context.Background(), s.Bucket, uri, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("unable to delete video: %w", err)
	}
	return nil
}

// Get total video count
func (s *VideoService) GetVideoStats() (int, error) {
	return media.GetVideoStats(s.DB)
}
