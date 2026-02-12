package media

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
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

// Create a single audio
func (s *AudioService) CreateAudio(
	audio *media.AudioMedia,
) error {
	if audio == nil {
		return errors.New("audio cannot be nil")
	}
	if audio.URI == "" || audio.Filename == "" || audio.Extension == "" {
		return errors.New("uri, filename, and extension are required")
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create audio metadata
	if err := media.CreateAudioTx(tx, audio); err != nil {
		return err
	}

	// Commit
	return tx.Commit()
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

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file for upload: %w", err)
	}
	reader := bytes.NewReader(data)

	contentType := handler.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = s.BLOB.PutObject(
		context.Background(),
		s.Bucket,
		uri,
		reader,
		int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return fmt.Errorf("unable to upload media to blob storage: %w", err)
	}
	return nil
}

func (s *AudioService) CreateAudiosBulk(audios []media.AudioMedia) error {
	if len(audios) == 0 {
		return errors.New("no audio entries provided")
	}

	for _, a := range audios {
		if a.URI == "" || a.DisplayedName == "" || a.Filename == "" || a.Extension == "" {
			return errors.New("all audio entries must include uri, displayed_name, filename, and extension")
		}
	}

	return media.CreateAudiosBulk(s.DB, audios)
}

func (s *AudioService) CreateAudioFilesBulk(
	files []multipart.File,
	handlers []*multipart.FileHeader,
	uris []string,
) error {
	if len(files) != len(handlers) || len(files) != len(uris) {
		return errors.New("files, handlers, and uris length mismatch")
	}

	uploaded := make([]string, 0, len(files))

	for i := range files {
		file := files[i]
		handler := handlers[i]
		uri := uris[i]

		data, err := io.ReadAll(file)
		if err != nil {
			for _, name := range uploaded {
				_ = s.BLOB.RemoveObject(context.Background(), s.Bucket, name, minio.RemoveObjectOptions{})
			}
			return fmt.Errorf("failed to read file for upload: %w", err)
		}
		reader := bytes.NewReader(data)

		contentType := handler.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		_, err = s.BLOB.PutObject(
			context.Background(),
			s.Bucket,
			uri,
			reader,
			int64(len(data)),
			minio.PutObjectOptions{ContentType: contentType},
		)
		if err != nil {
			for _, name := range uploaded {
				_ = s.BLOB.RemoveObject(context.Background(), s.Bucket, name, minio.RemoveObjectOptions{})
			}
			return fmt.Errorf("failed to upload file: %w", err)
		}

		uploaded = append(uploaded, uri)
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
		return nil, errors.New("invalid uri")
	}
	file, err := s.BLOB.GetObject(context.Background(), s.Bucket, uri, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch audio: %w", err)
	}
	return file, nil
}

func (s *AudioService) GetAudiosByDayCarepartnerandCaregiver(
	date string,
	carepartnerID,
	caregiverID int,
) ([]int, error) {
	if date == "" {
		return nil, errors.New("date is required")
	}
	if carepartnerID <= 0 || caregiverID <= 0 {
		return nil, errors.New("invalid user IDs")
	}
	return media.GetAudiosByDayCarepartnerandCaregiver(s.DB, date, carepartnerID, caregiverID)
}

func (s *AudioService) GetMediaByTypeCarepartnerAndCaregiver(
	mediaType string,
	carepartnerID,
	caregiverID int,
) ([]int, error) {
	if mediaType == "" {
		return nil, errors.New("media type is required")
	}
	if carepartnerID <= 0 || caregiverID <= 0 {
		return nil, errors.New("invalid user IDs")
	}
	return media.GetMediaByTypeCarepartnerAndCaregiver(s.DB, mediaType, carepartnerID, caregiverID)
}

func (s *AudioService) GetAudiosByCarepartnerAndCaregiver(
	carepartnerID,
	caregiverID int,
) ([]media.AudioMedia, error) {
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
	if audio.DisplayedName == "" {
		return errors.New("displayed_name cannot be empty")
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

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	reader := bytes.NewReader(data)

	contentType := handler.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = s.BLOB.PutObject(
		context.Background(),
		s.Bucket,
		uri,
		reader,
		int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return fmt.Errorf("unable to update audio: %w", err)
	}
	return nil
}

func (s *AudioService) DeleteAudio(id int) error {
	if id <= 0 {
		return errors.New("invalid audio ID")
	}
	return media.DeleteAudio(s.DB, id)
}

func (s *AudioService) DeleteAudioFile(uri string) error {
	if uri == "" {
		return errors.New("invalid uri")
	}
	if err := s.BLOB.RemoveObject(context.Background(), s.Bucket, uri, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("unable to delete audio: %w", err)
	}
	return nil
}
