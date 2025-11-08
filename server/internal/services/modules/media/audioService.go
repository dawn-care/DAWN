package media

import (
	"database/sql"
	"errors"
	"fmt"
	
	"server/internal/models/modules/media"
)

type AudioService struct {
	DB *sql.DB
}

func NewAudioService(db *sql.DB) *AudioService {
	return &AudioService{DB: db}
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

func (s *AudioService) CreateAudio(audio *media.AudioMedia) error {
	if audio == nil {
		return errors.New("audio cannot be nil")
	}
	if audio.URI == "" || audio.Filename == "" || audio.Extension == "" {
		return errors.New("uri, filename, and extension are required")
	}
	return media.CreateAudio(s.DB, audio)
}

func (s *AudioService) CreateAudiosBulk(audios []media.AudioMedia) error {
	if len(audios) == 0 {
		return errors.New("no audio entries provided")
	}
	return media.CreateAudiosBulk(s.DB, audios)
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

func (s *AudioService) GetAudioByID(id int) (*media.AudioMedia, error) {
	if id <= 0 {
		return nil, errors.New("invalid audio ID")
	}
	audio, err := media.GetAudioByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio: %w", err)
	}
	if audio == nil {
		return nil, fmt.Errorf("audio with id %d not found", id)
	}
	return audio, nil
}

func (s *AudioService) GetAudiosByDayCarepartnerandCaregiver(date string, carepartnerID, caregiverID int) ([]int, error) {
	if date == "" {
		return nil, errors.New("date is required")
	}
	if carepartnerID <= 0 || caregiverID <= 0 {
		return nil, errors.New("invalid user IDs")
	}
	return media.GetAudiosByDayCarepartnerandCaregiver(s.DB, date, carepartnerID, caregiverID)
}

func (s *AudioService) GetMediaByTypeCarepartnerAndCaregiver(mediaType string, carepartnerID, caregiverID int) ([]int, error) {
	if mediaType == "" {
		return nil, errors.New("media type is required")
	}
	if carepartnerID <= 0 || caregiverID <= 0 {
		return nil, errors.New("invalid user IDs")
	}
	return media.GetMediaByTypeCarepartnerAndCaregiver(s.DB, mediaType, carepartnerID, caregiverID)
}

func (s *AudioService) GetAudiosByCarepartnerAndCaregiver(carepartnerID, caregiverID int) ([]media.AudioMedia, error) {
	if carepartnerID <= 0 || caregiverID <= 0 {
		return nil, errors.New("invalid user IDs")
	}
	return media.GetAudiosByCarepartnerAndCaregiver(s.DB, carepartnerID, caregiverID)
}

// ─────────────────────────────────────────────
// UPDATE / DELETE
// ─────────────────────────────────────────────

func (s *AudioService) UpdateAudio(audio *media.AudioMedia) error {
	if audio == nil {
		return errors.New("audio cannot be nil")
	}
	if audio.ID <= 0 {
		return errors.New("invalid audio ID")
	}
	return media.UpdateAudio(s.DB, audio)
}

func (s *AudioService) DeleteAudio(id int) error {
	if id <= 0 {
		return errors.New("invalid audio ID")
	}
	return media.DeleteAudio(s.DB, id)
}

// ─────────────────────────────────────────────
// ANALYTICS / STATS
// ─────────────────────────────────────────────

func (s *AudioService) GetMediaStats() (*media.MediaStats, error) {
	stats, err := media.GetMediaStats(s.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to get media stats: %w", err)
	}
	return stats, nil
}

// ─────────────────────────────────────────────
// SEARCH
// ─────────────────────────────────────────────

func (s *AudioService) SearchAudios(keyword string) ([]media.AudioMedia, error) {
	if keyword == "" {
		return nil, errors.New("keyword cannot be empty")
	}
	results, err := media.SearchAudios(s.DB, keyword)
	if err != nil {
		return nil, fmt.Errorf("failed to search audios: %w", err)
	}
	return results, nil
}
