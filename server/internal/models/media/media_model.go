package media

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	KindPhoto = "photo"
	KindVideo = "video"
	KindAudio = "audio"
)

// Usage values reported on a Media item. An empty list means the item is only saved.
const (
	UsageOnScreen        = "on_screen"        // shown by a photo or message widget
	UsageAudioBackground = "audio_background" // background photo of a scheduled audio item
	UsageScheduled       = "scheduled"        // the media of a scheduled item
)

/*
Media is a photo, video or audio file in a carepartner's library.
ObjectKey - where the file is stored in MinIO; never sent to clients, who use URL instead
DurationSeconds - length of a video or audio file, nil for photos
Usage - derived when read: where the item is currently used
URL - path the file can be fetched from, relative to the API host
*/
type Media struct {
	ID              int       `json:"id"`
	CarepartnerID   int       `json:"carepartner_id"`
	Kind            string    `json:"kind"`
	DisplayName     string    `json:"display_name"`
	ObjectKey       string    `json:"-"`
	MimeType        string    `json:"mime_type"`
	SizeBytes       int64     `json:"size_bytes"`
	DurationSeconds *int      `json:"duration_seconds"`
	CreatedAt       time.Time `json:"created_at"`
	Usage           []string  `json:"usage"`
	URL             string    `json:"url"`
}

type queryRower interface {
	QueryRow(query string, args ...any) *sql.Row
}

// selectMedia reads every column plus the derived usage flags
const selectMedia = `
	SELECT m.id, m.carepartner_id, m.kind, m.display_name, m.object_key, m.mime_type,
	       m.size_bytes, m.duration_seconds, m.created_at,
	       EXISTS (SELECT 1 FROM "ScreenWidget" w WHERE w.photo_media_id = m.id) AS on_screen,
	       EXISTS (SELECT 1 FROM "ScheduledItem" s WHERE s.background_photo_id = m.id) AS audio_background,
	       EXISTS (SELECT 1 FROM "ScheduledItem" s WHERE s.media_id = m.id) AS scheduled
	FROM "Media" m
`

type scanner interface {
	Scan(dest ...any) error
}

func scanMedia(row scanner) (*Media, error) {
	m := &Media{}
	var duration sql.NullInt64
	var onScreen, audioBackground, scheduled bool
	if err := row.Scan(
		&m.ID, &m.CarepartnerID, &m.Kind, &m.DisplayName, &m.ObjectKey, &m.MimeType,
		&m.SizeBytes, &duration, &m.CreatedAt,
		&onScreen, &audioBackground, &scheduled,
	); err != nil {
		return nil, err
	}
	if duration.Valid {
		d := int(duration.Int64)
		m.DurationSeconds = &d
	}
	m.Usage = []string{}
	if onScreen {
		m.Usage = append(m.Usage, UsageOnScreen)
	}
	if audioBackground {
		m.Usage = append(m.Usage, UsageAudioBackground)
	}
	if scheduled {
		m.Usage = append(m.Usage, UsageScheduled)
	}
	m.URL = FileURL(m.ID)
	return m, nil
}

// FileURL is the API path that serves a media item's file
func FileURL(id int) string {
	return fmt.Sprintf("/api/media/%d/file", id)
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

// Create a media row; fills in ID and CreatedAt
func CreateMediaTx(exec queryRower, m *Media) error {
	query := `
		INSERT INTO "Media" (carepartner_id, kind, display_name, object_key, mime_type, size_bytes, duration_seconds)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at;
	`
	if err := exec.QueryRow(
		query,
		m.CarepartnerID,
		m.Kind,
		m.DisplayName,
		m.ObjectKey,
		m.MimeType,
		m.SizeBytes,
		m.DurationSeconds,
	).Scan(&m.ID, &m.CreatedAt); err != nil {
		return err
	}
	m.Usage = []string{}
	m.URL = FileURL(m.ID)
	return nil
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

// Retrieve one media item, or nil if it doesn't exist
func GetMediaByID(exec queryRower, id int) (*Media, error) {
	m, err := scanMedia(exec.QueryRow(selectMedia+` WHERE m.id = $1;`, id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return m, err
}

// List a carepartner's media, newest first. An empty kind lists every kind.
func ListMedia(db *sql.DB, carepartnerID int, kind string) ([]Media, error) {
	rows, err := db.Query(
		selectMedia+` WHERE m.carepartner_id = $1 AND ($2 = '' OR m.kind = $2) ORDER BY m.created_at DESC, m.id DESC;`,
		carepartnerID, kind,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Media{}
	for rows.Next() {
		m, err := scanMedia(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *m)
	}
	return items, rows.Err()
}

// ─────────────────────────────────────────────
// UPDATE / DELETE
// ─────────────────────────────────────────────

// Change the name the caregiver sees; returns false if no such item
func RenameMedia(db *sql.DB, id int, displayName string) (bool, error) {
	res, err := db.Exec(`UPDATE "Media" SET display_name = $1 WHERE id = $2;`, displayName, id)
	if err != nil {
		return false, err
	}
	count, err := res.RowsAffected()
	return count > 0, err
}

/*
Delete a media item inside tx and return its object key, or "" if no such item.
Photo widgets showing it are deleted first; the foreign keys then clear message
widget photos and audio backgrounds and delete scheduled items that play it.
*/
func DeleteMediaTx(tx *sql.Tx, id int) (string, error) {
	var objectKey string
	err := tx.QueryRow(`SELECT object_key FROM "Media" WHERE id = $1 FOR UPDATE;`, id).Scan(&objectKey)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if _, err := tx.Exec(`DELETE FROM "ScreenWidget" WHERE type = 'photo' AND photo_media_id = $1;`, id); err != nil {
		return "", err
	}
	if _, err := tx.Exec(`DELETE FROM "Media" WHERE id = $1;`, id); err != nil {
		return "", err
	}
	return objectKey, nil
}
