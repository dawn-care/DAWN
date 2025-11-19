package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/minio/minio-go/v7"

	"server/internal/models/media"
)

type AudioService struct {
	DB     *sql.DB
	BLOB   *minio.Client
	Bucket string
}

func NewAudioService(db *sql.DB, blob *minio.Client, bucket string) *AudioService {
	return &AudioService{DB: db, BLOB: blob, Bucket: bucket}
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

func (s *AudioService) CreateAudioFile(file multipart.File, handler *multipart.FileHeader, uri string) error {
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

func (s *AudioService) CreateAudiosBulk(audios []media.AudioMedia) error {
	if len(audios) == 0 {
		return errors.New("no audio entries provided")
	}

	return media.CreateAudiosBulk(s.DB, audios)
}

func (s *AudioService) CreateAudioFilesBulk(files []multipart.File, handlers []*multipart.FileHeader) error {
	uploaded := make([]string, 0, len(files))

	for i := range files {
		file := files[i]
		handler := handlers[i]

		contentType := handler.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		_, err := s.BLOB.PutObject(
			context.Background(),
			s.Bucket,
			handler.Filename,
			file,
			handler.Size,
			minio.PutObjectOptions{ContentType: contentType},
		)
		if err != nil {
			ctx := context.Background()

			for _, name := range uploaded {
				err := s.BLOB.RemoveObject(ctx, s.Bucket, name, minio.RemoveObjectOptions{})
				if err != nil {
					return fmt.Errorf("database insert failed: %w", err)
				}
			}
		}

		uploaded = append(uploaded, handler.Filename)
	}

	return nil
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

func (s *AudioService) GetAudioFile(uri string) (multipart.File, error) {
	if uri == "" {
		return nil, errors.New("invalid filename")
	}
	file, err := s.BLOB.GetObject(context.Background(), s.Bucket, uri, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch audio: %w", err)
	}

	return file, nil
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

func (s *AudioService) UpdateAudioFile(file multipart.File, handler *multipart.FileHeader, uri string) error {
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
		return fmt.Errorf("unable to update audio: %w", err)
	}
	return nil
}

func (s *AudioService) DeleteAudio(id int) error {
	if id <= 0 {
		return errors.New("invalid audio ID")
	}
	_, err := media.GetAudioByID(s.DB, id)
	if err != nil {
		return fmt.Errorf("failed to get audio with that id: %w", err)
	}
	return media.DeleteAudio(s.DB, id)
}

func (s *AudioService) DeleteAudioFile(uri string) error {
	if uri == "" {
		return errors.New("invalid filename")
	}
	removeErr := s.BLOB.RemoveObject(context.Background(), s.Bucket, uri, minio.RemoveObjectOptions{})
	if removeErr != nil {
		return fmt.Errorf("unable to delete audio: %w", removeErr)
	}
	return nil
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
