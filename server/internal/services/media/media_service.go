package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	"server/internal/models/media"
)

// MaxDurationSeconds is the longest video or audio file that can be uploaded (10 minutes)
const MaxDurationSeconds = 600

var (
	ErrInvalidKind        = errors.New("kind must be photo, video or audio")
	ErrInvalidName        = errors.New("display_name must be 1 to 200 characters")
	ErrTooLong            = fmt.Errorf("video and audio files can be at most %d minutes long", MaxDurationSeconds/60)
	ErrUnreadableMedia    = errors.New("could not read the length of this file")
	ErrNotFound           = errors.New("media not found")
	ErrInvalidCarepartner = errors.New("invalid carepartner_id")
	ErrInvalidSituation   = errors.New("situations must come from the list of situations")
	ErrInvalidFamily      = errors.New("family_member_ids must be this carepartner's family members")
)

// ObjectStore is the part of blob storage the media service needs
type ObjectStore interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadSeekCloser, error)
	Remove(ctx context.Context, key string) error
}

// MinioStore stores objects in one MinIO bucket
type MinioStore struct {
	Client *minio.Client
	Bucket string
}

func (s *MinioStore) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.Client.PutObject(ctx, s.Bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (s *MinioStore) Get(ctx context.Context, key string) (io.ReadSeekCloser, error) {
	return s.Client.GetObject(ctx, s.Bucket, key, minio.GetObjectOptions{})
}

func (s *MinioStore) Remove(ctx context.Context, key string) error {
	return s.Client.RemoveObject(ctx, s.Bucket, key, minio.RemoveObjectOptions{})
}

// ProbeDuration returns the length in seconds of the media file at path
type ProbeDuration func(ctx context.Context, path string) (float64, error)

// FFProbeDuration reads a file's length with ffprobe, which must be on PATH
func FFProbeDuration(ctx context.Context, path string) (float64, error) {
	out, err := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	).Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}
	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, fmt.Errorf("ffprobe returned no duration: %w", err)
	}
	return seconds, nil
}

type MediaService struct {
	DB    *sql.DB
	Store ObjectStore
	Probe ProbeDuration
}

func NewMediaService(db *sql.DB, blob *minio.Client, bucket string) *MediaService {
	return &MediaService{
		DB:    db,
		Store: &MinioStore{Client: blob, Bucket: bucket},
		Probe: FFProbeDuration,
	}
}

// UploadInput is a file the caller has already checked is the right type for Kind
type UploadInput struct {
	CarepartnerID int
	Kind          string
	DisplayName   string
	Extension     string // including the dot, e.g. ".mp4"; may be empty
	MimeType      string
	File          io.Reader
	Tags          Tags
}

// Tags say how staff find a media item. No situations means every situation.
type Tags struct {
	Situations      []string
	FamilyMemberIDs []int
}

/*
cleanTags checks tags and fills in the default: no situations means every one.
Repeats are dropped; situations end up in the order of media.Situations and family
members by ID, as they are read back.
*/
func cleanTags(t Tags) (Tags, error) {
	for _, s := range t.Situations {
		if !media.ValidSituation(s) {
			return Tags{}, ErrInvalidSituation
		}
	}
	clean := Tags{Situations: []string{}, FamilyMemberIDs: []int{}}
	for _, s := range media.Situations {
		if len(t.Situations) == 0 || slices.Contains(t.Situations, s) {
			clean.Situations = append(clean.Situations, s)
		}
	}
	for _, id := range t.FamilyMemberIDs {
		if id <= 0 {
			return Tags{}, ErrInvalidFamily
		}
		if !slices.Contains(clean.FamilyMemberIDs, id) {
			clean.FamilyMemberIDs = append(clean.FamilyMemberIDs, id)
		}
	}
	slices.Sort(clean.FamilyMemberIDs)
	return clean, nil
}

func validKind(kind string) bool {
	return kind == media.KindPhoto || kind == media.KindVideo || kind == media.KindAudio
}

func cleanName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 200 {
		return "", ErrInvalidName
	}
	return name, nil
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

/*
Upload stores the file in blob storage and creates its Media row.
Video and audio are measured with Probe first and rejected if longer than 10 minutes.
*/
func (s *MediaService) Upload(ctx context.Context, in UploadInput) (*media.Media, error) {
	if in.CarepartnerID <= 0 {
		return nil, ErrInvalidCarepartner
	}
	if !validKind(in.Kind) {
		return nil, ErrInvalidKind
	}
	name, err := cleanName(in.DisplayName)
	if err != nil {
		return nil, err
	}
	tags, err := cleanTags(in.Tags)
	if err != nil {
		return nil, err
	}

	// Spool to disk: ffprobe needs a path, and the store needs the size up front
	tmp, err := os.CreateTemp("", "media-upload-*"+in.Extension)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	size, err := io.Copy(tmp, in.File)
	if err != nil {
		return nil, fmt.Errorf("failed to read upload: %w", err)
	}

	var duration *int
	if in.Kind != media.KindPhoto {
		seconds, err := s.Probe(ctx, tmp.Name())
		if err != nil || seconds < 0 || math.IsNaN(seconds) {
			return nil, ErrUnreadableMedia
		}
		rounded := int(math.Round(seconds))
		if rounded > MaxDurationSeconds {
			return nil, ErrTooLong
		}
		duration = &rounded
	}

	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to rewind upload: %w", err)
	}

	m := &media.Media{
		CarepartnerID:   in.CarepartnerID,
		Kind:            in.Kind,
		DisplayName:     name,
		ObjectKey:       fmt.Sprintf("carepartners/%d/%s/%s%s", in.CarepartnerID, in.Kind, uuid.NewString(), strings.ToLower(in.Extension)),
		MimeType:        in.MimeType,
		SizeBytes:       size,
		DurationSeconds: duration,
	}

	if err := s.Store.Put(ctx, m.ObjectKey, tmp, size, m.MimeType); err != nil {
		return nil, fmt.Errorf("unable to upload media to blob storage: %w", err)
	}

	if err := s.createWithTags(m, tags); err != nil {
		// Don't leave an orphaned file behind
		if rmErr := s.Store.Remove(context.Background(), m.ObjectKey); rmErr != nil {
			log.Printf("media: failed to remove orphaned object %s: %v", m.ObjectKey, rmErr)
		}
		if errors.Is(err, ErrInvalidFamily) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to save media: %w", err)
	}

	return m, nil
}

// createWithTags saves a new media row and its tags together
func (s *MediaService) createWithTags(m *media.Media, tags Tags) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := media.CreateMediaTx(tx, m); err != nil {
		return err
	}
	linked, err := media.SetTagsTx(tx, m.ID, m.CarepartnerID, tags.Situations, tags.FamilyMemberIDs)
	if err != nil {
		return err
	}
	if !linked {
		return ErrInvalidFamily
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	m.Situations = &tags.Situations
	m.FamilyMemberIDs = &tags.FamilyMemberIDs
	return nil
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

// GetMedia returns one media item with its tags
func (s *MediaService) GetMedia(id int) (*media.Media, error) {
	m, err := s.findMedia(id)
	if err != nil {
		return nil, err
	}
	items := []media.Media{*m}
	if err := media.LoadTags(s.DB, items); err != nil {
		return nil, fmt.Errorf("failed to get media tags: %w", err)
	}
	return &items[0], nil
}

// findMedia returns one media item without its tags
func (s *MediaService) findMedia(id int) (*media.Media, error) {
	if id <= 0 {
		return nil, ErrNotFound
	}
	m, err := media.GetMediaByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get media: %w", err)
	}
	if m == nil {
		return nil, ErrNotFound
	}
	return m, nil
}

// List a carepartner's media that matches filter, with tags; zero values don't filter
func (s *MediaService) ListMedia(carepartnerID int, filter media.ListFilter) ([]media.Media, error) {
	if carepartnerID <= 0 {
		return nil, ErrInvalidCarepartner
	}
	if filter.Kind != "" && !validKind(filter.Kind) {
		return nil, ErrInvalidKind
	}
	if filter.Situation != "" && !media.ValidSituation(filter.Situation) {
		return nil, ErrInvalidSituation
	}
	if filter.FamilyMemberID < 0 {
		return nil, ErrInvalidFamily
	}
	items, err := media.ListMedia(s.DB, carepartnerID, filter)
	if err != nil {
		return nil, err
	}
	if err := media.LoadTags(s.DB, items); err != nil {
		return nil, fmt.Errorf("failed to get media tags: %w", err)
	}
	return items, nil
}

// OpenFile returns a media item and a reader for its file; the caller closes the reader.
// It skips the tags, since serving a file doesn't need them.
func (s *MediaService) OpenFile(ctx context.Context, id int) (*media.Media, io.ReadSeekCloser, error) {
	m, err := s.findMedia(id)
	if err != nil {
		return nil, nil, err
	}
	file, err := s.Store.Get(ctx, m.ObjectKey)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to fetch media file: %w", err)
	}
	return m, file, nil
}

// ─────────────────────────────────────────────
// UPDATE / DELETE
// ─────────────────────────────────────────────

// Rename changes only the display name; the stored file is untouched
func (s *MediaService) Rename(id int, displayName string) (*media.Media, error) {
	if id <= 0 {
		return nil, ErrNotFound
	}
	name, err := cleanName(displayName)
	if err != nil {
		return nil, err
	}
	found, err := media.RenameMedia(s.DB, id, name)
	if err != nil {
		return nil, fmt.Errorf("failed to rename media: %w", err)
	}
	if !found {
		return nil, ErrNotFound
	}
	return s.GetMedia(id)
}

/*
SetTags replaces how staff find a media item. No situations means every situation;
family members must be the item's carepartner's.
*/
func (s *MediaService) SetTags(id int, tags Tags) (*media.Media, error) {
	if id <= 0 {
		return nil, ErrNotFound
	}
	clean, err := cleanTags(tags)
	if err != nil {
		return nil, err
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	carepartnerID, found, err := media.CarepartnerOfMediaTx(tx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to set media tags: %w", err)
	}
	if !found {
		return nil, ErrNotFound
	}
	linked, err := media.SetTagsTx(tx, id, carepartnerID, clean.Situations, clean.FamilyMemberIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to set media tags: %w", err)
	}
	if !linked {
		return nil, ErrInvalidFamily
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to set media tags: %w", err)
	}
	return s.GetMedia(id)
}

/*
Delete removes a media item and everything that depends on it (see DeleteMediaTx),
then its file. The file is removed only after the database commit; if that fails the
orphaned object is logged rather than reported, since the item is already gone.
*/
func (s *MediaService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrNotFound
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	objectKey, err := media.DeleteMediaTx(tx, id)
	if err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
	}
	if objectKey == "" {
		return ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
	}

	if err := s.Store.Remove(ctx, objectKey); err != nil {
		log.Printf("media: deleted media %d but failed to remove object %s: %v", id, objectKey, err)
	}
	return nil
}
