package media

import (
	"database/sql"
	"fmt"
)

// AudioMedia represents one audio media record
type AudioMedia struct {
	ID        int    `json:"id"`
	URI       string `json:"uri"`
	Filename  string `json:"filename"`
	Extension string `json:"extension"` // replace with string later if needed
}

// ─────────────────────────────────────────────
// CREATE OPERATIONS
// ─────────────────────────────────────────────

// Add one audio file
func CreateAudio(db *sql.DB, audio *AudioMedia) error {
	query := `
		INSERT INTO "AudioMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	return db.QueryRow(query, audio.URI, audio.Filename, audio.Extension).Scan(&audio.ID)
}

// Add multiple audio files at once
func CreateAudiosBulk(db *sql.DB, audios []AudioMedia) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT INTO "AudioMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range audios {
		err := stmt.QueryRow(audios[i].URI, audios[i].Filename, audios[i].Extension).Scan(&audios[i].ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// ─────────────────────────────────────────────
// READ OPERATIONS
// ─────────────────────────────────────────────

// Get one audio record by ID
func GetAudioByID(db *sql.DB, id int) (*AudioMedia, error) {
	query := `
		SELECT id, uri, filename, extension
		FROM "AudioMedia"
		WHERE id = $1;
	`
	audio := &AudioMedia{}
	err := db.QueryRow(query, id).Scan(&audio.ID, &audio.URI, &audio.Filename, &audio.Extension)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return audio, nil
}

// Get multiple audio IDs by date and users (assuming filename/URI encodes date)
func GetAudiosByDayAndUsers(db *sql.DB, date string, carereceiverID, caregiverID int) ([]int, error) {
	query := `
		SELECT a.id
		FROM "AudioMedia" a
		JOIN "Module" m ON m.metadata = a.id
		WHERE m.carereceiver_id = $1 AND m.caregiver_id = $2
		AND a.filename LIKE $3;
	`
	rows, err := db.Query(query, carereceiverID, caregiverID, fmt.Sprintf("%%%s%%", date))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Get all media (video, audio, photo) by type and user (returns IDs)
func GetMediaByTypeAndUser(db *sql.DB, mediaType string, carereceiverID, caregiverID int) ([]int, error) {
	validTables := map[string]string{
		"audio": "AudioMedia",
		"video": "VideoMedia",
		"photo": "PhotoMedia",
	}
	table, ok := validTables[mediaType]
	if !ok {
		return nil, fmt.Errorf("invalid media type: %s", mediaType)
	}

	query := fmt.Sprintf(`
		SELECT m.metadata
		FROM "%s" t
		JOIN "Module" m ON m.metadata = t.id
		WHERE m.carereceiver_id = $1 AND m.caregiver_id = $2;
	`, table)

	rows, err := db.Query(query, carereceiverID, caregiverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Get all audio files by user (full records)
func GetAudiosByUsers(db *sql.DB, carereceiverID, caregiverID int) ([]AudioMedia, error) {
	query := `
		SELECT a.id, a.uri, a.filename, a.extension
		FROM "AudioMedia" a
		JOIN "Module" m ON m.metadata = a.id
		WHERE m.carereceiver_id = $1 AND m.caregiver_id = $2;
	`
	rows, err := db.Query(query, carereceiverID, caregiverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var audios []AudioMedia
	for rows.Next() {
		var a AudioMedia
		if err := rows.Scan(&a.ID, &a.URI, &a.Filename, &a.Extension); err != nil {
			return nil, err
		}
		audios = append(audios, a)
	}
	return audios, nil
}

// ─────────────────────────────────────────────
// UPDATE / DELETE
// ─────────────────────────────────────────────

// Update an existing audio entry
func UpdateAudio(db *sql.DB, audio *AudioMedia) error {
	query := `
		UPDATE "AudioMedia"
		SET uri = $1, filename = $2, extension = $3
		WHERE id = $4;
	`
	res, err := db.Exec(query, audio.URI, audio.Filename, audio.Extension, audio.ID)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no audio found with id %d", audio.ID)
	}
	return nil
}

// Delete an audio record by ID
func DeleteAudio(db *sql.DB, id int) error {
	query := `DELETE FROM "AudioMedia" WHERE id = $1;`
	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no audio found with id %d", id)
	}
	return nil
}

// ─────────────────────────────────────────────
// ANALYTICS / STATS
// ─────────────────────────────────────────────

// Get media stats: count of videos, audio, and photos
type MediaStats struct {
	AudioCount int `json:"audio_count"`
	VideoCount int `json:"video_count"`
	PhotoCount int `json:"photo_count"`
	Total      int `json:"total"`
}

func GetMediaStats(db *sql.DB) (*MediaStats, error) {
	stats := &MediaStats{}
	query := `
		SELECT
			(SELECT COUNT(*) FROM "AudioMedia") AS audio_count,
			(SELECT COUNT(*) FROM "VideoMedia") AS video_count,
			(SELECT COUNT(*) FROM "PhotoMedia") AS photo_count;
	`
	err := db.QueryRow(query).Scan(&stats.AudioCount, &stats.VideoCount, &stats.PhotoCount)
	if err != nil {
		return nil, err
	}
	stats.Total = stats.AudioCount + stats.VideoCount + stats.PhotoCount
	return stats, nil
}

// ─────────────────────────────────────────────
// ADDITIONAL USEFUL FUNCTIONALITY
// ─────────────────────────────────────────────

// Search audios by filename keyword (partial match)
func SearchAudios(db *sql.DB, keyword string) ([]AudioMedia, error) {
	query := `
		SELECT id, uri, filename, extension
		FROM "AudioMedia"
		WHERE filename ILIKE $1;
	`
	rows, err := db.Query(query, fmt.Sprintf("%%%s%%", keyword))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AudioMedia
	for rows.Next() {
		var a AudioMedia
		if err := rows.Scan(&a.ID, &a.URI, &a.Filename, &a.Extension); err != nil {
			return nil, err
		}
		results = append(results, a)
	}
	return results, nil
}
