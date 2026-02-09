package media

import (
	"database/sql"
	"fmt"
)

type AudioMedia struct {
	ID            int    `json:"id"`
	URI           string `json:"uri"`
	Filename      string `json:"filename"`
	Extension     string `json:"extension"`
	DisplayedName string `json:"displayed_name"`
}

// ─────────────────────────────────────────────
// CREATE OPERATIONS
// ─────────────────────────────────────────────

// Add one audio file
func CreateAudio(db *sql.DB, audio *AudioMedia) error {
	return CreateAudioTx(db, audio)
}

// Add one audio file (Tx)
func CreateAudioTx(exec interface {
	QueryRow(query string, args ...any) *sql.Row
}, audio *AudioMedia) error {
	query := `
		INSERT INTO "AudioMedia" (uri, filename, extension, displayed_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`
	return exec.QueryRow(
		query,
		audio.URI,
		audio.Filename,
		audio.Extension,
		audio.DisplayedName,
	).Scan(&audio.ID)
}

// Add multiple audio files at once (DB)
func CreateAudiosBulk(db *sql.DB, audios []AudioMedia) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO "AudioMedia" (uri, filename, extension, displayed_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for i := range audios {
		err := stmt.QueryRow(
			audios[i].URI,
			audios[i].Filename,
			audios[i].Extension,
			audios[i].DisplayedName,
		).Scan(&audios[i].ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

// Get one audio record by ID
func GetAudioByID(db *sql.DB, id int) (*AudioMedia, error) {
	query := `
		SELECT id, uri, filename, extension, displayed_name
		FROM "AudioMedia"
		WHERE id = $1;
	`
	audio := &AudioMedia{}
	err := db.QueryRow(query, id).Scan(
		&audio.ID,
		&audio.URI,
		&audio.Filename,
		&audio.Extension,
		&audio.DisplayedName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return audio, nil
}

// Get multiple audio IDs by date and users
func GetAudiosByDayCarepartnerandCaregiver(db *sql.DB, date string, carepartnerID, caregiverID int) ([]int, error) {
	query := `
		SELECT a.id
		FROM "AudioMedia" a
		JOIN "Module" m ON m.metadata = a.id
		WHERE m.carepartner_id = $1
		  AND m.caregiver_id = $2
		  AND a.filename LIKE $3;
	`
	rows, err := db.Query(query, carepartnerID, caregiverID, fmt.Sprintf("%%%s%%", date))
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

// Get all media IDs by type, carepartner, and caregiver
func GetMediaByTypeCarepartnerAndCaregiver(db *sql.DB, mediaType string, carepartnerID, caregiverID int) ([]int, error) {
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
		WHERE m.carepartner_id = $1
		  AND m.caregiver_id = $2;
	`, table)

	rows, err := db.Query(query, carepartnerID, caregiverID)
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

// Get all audio files by carepartner and caregiver (full records)
func GetAudiosByCarepartnerAndCaregiver(db *sql.DB, carepartnerID, caregiverID int) ([]AudioMedia, error) {
	query := `
		SELECT a.id, a.uri, a.filename, a.extension, a.displayed_name
		FROM "AudioMedia" a
		JOIN "Module" m ON m.metadata_id = a.id
		WHERE m.carepartner_id = $1 AND m.caregiver_id = $2;
	`
	rows, err := db.Query(query, carepartnerID, caregiverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var audios []AudioMedia
	for rows.Next() {
		var a AudioMedia
		if err := rows.Scan(
			&a.ID,
			&a.URI,
			&a.Filename,
			&a.Extension,
			&a.DisplayedName,
		); err != nil {
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
		SET uri = $1,
		    filename = $2,
		    extension = $3,
		    displayed_name = $4
		WHERE id = $5;
	`
	res, err := db.Exec(
		query,
		audio.URI,
		audio.Filename,
		audio.Extension,
		audio.DisplayedName,
		audio.ID,
	)
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
	err := db.QueryRow(query).Scan(
		&stats.AudioCount,
		&stats.VideoCount,
		&stats.PhotoCount,
	)
	if err != nil {
		return nil, err
	}

	stats.Total = stats.AudioCount + stats.VideoCount + stats.PhotoCount
	return stats, nil
}

// ─────────────────────────────────────────────
// SEARCH
// ─────────────────────────────────────────────

// Search audios by filename OR displayed name
func SearchAudios(db *sql.DB, keyword string) ([]AudioMedia, error) {
	query := `
		SELECT id, uri, filename, extension, displayed_name
		FROM "AudioMedia"
		WHERE filename ILIKE $1
		   OR displayed_name ILIKE $1;
	`
	rows, err := db.Query(query, fmt.Sprintf("%%%s%%", keyword))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AudioMedia
	for rows.Next() {
		var a AudioMedia
		if err := rows.Scan(
			&a.ID,
			&a.URI,
			&a.Filename,
			&a.Extension,
			&a.DisplayedName,
		); err != nil {
			return nil, err
		}
		results = append(results, a)
	}
	return results, nil
}
