package media

import (
	"database/sql"
	"errors"
	"fmt"
	
	"server/internal/models/modules/media"
)

type VideoService struct {
	DB *sql.DB
}

func NewVideoService(db *sql.DB) *VideoService {
	return &VideoService{DB: db}
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

// Delete a video
func (s *VideoService) DeleteVideo(id int) error {
	if id <= 0 {
		return errors.New("invalid video ID")
	}
	return media.DeleteVideo(s.DB, id)
}

// Get total video count
func (s *VideoService) GetVideoStats() (int, error) {
	return media.GetVideoStats(s.DB)
}
