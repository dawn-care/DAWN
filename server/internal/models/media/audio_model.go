package media

import (
	"database/sql"
	"fmt"
)

/*
AudioMedia represents one audio media record
ID - the ID in the database
URI - the URL of the file
Filename - the name of the file
Extension - the file extension, i.e. what file type it is
*/
type AudioMedia struct {
	ID        int    `json:"id"`
	URI       string `json:"uri"`
	Filename  string `json:"filename"`
	Extension string `json:"extension"`
}

// CREATE OPERATIONS

/*
Add one audio file
db - the instance of the database
audio - an audio media record
*/
func CreateAudio(db *sql.DB, audio *AudioMedia) error {
	query := `
		INSERT INTO "AudioMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	return db.QueryRow(query, audio.URI, audio.Filename, audio.Extension).Scan(&audio.ID)
}

/*
Add multiple audio files at once
db - the instance of the database
audios - an array of audio media files
*/
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

// READ OPERATIONS

/*
Get one audio record by ID
db - the instance of the database
id - the ID of the audio record in the database
*/
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

/*
Get multiple audio IDs by date and users (assuming filename/URI encodes date)
db - the instance of the database
date - date related to the audio record
carereceiverID - the ID of the carepartner
caregiverID - the ID of the caregiver who uploaded the file
*/
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

/*
Get all media (video, audio, photo) by type and user (returns IDs)
db - instance of the database
mediaType - type of media to get
carereceiverID - ID of the carepartner
caregiverID - ID of the caregiver who uploaded the media
*/
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

/*
Get all audio files by user (full records)
db - instance of the database
carereceiverID - ID of the carepartner
caregiverID - ID of the caregiver who uploaded the audio record
*/
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

// UPDATE OPERATION

/*
Update an existing audio entry
db - instance of the database
audio - the audio record to update
*/
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

// DELETE OPERATION

/*
Delete an audio record by ID
db - instance of the database
id - the ID of the audio record to be deleted
*/
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

// ANALYTICS & STATISTICS

/*
Stats on the current media uploaded: count of videos, audio, and photos
AudioCount - the number of audio records uploaded
VideoCount - the number of video records uploaded
PhotoCount - the number of photos uploaded
Total - total media records available
*/
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

// ADDITIONAL FUNCTIONALITY

/*
Search audios by filename keyword (partial match)
db - instance of the database
keyword - string we are matching when searching the audios
*/
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
