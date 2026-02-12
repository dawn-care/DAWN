package media

import (
	"database/sql"
	"fmt"
)

type VideoMedia struct {
	ID            int    `json:"id"`
	URI           string `json:"uri"`
	DisplayedName string `json:"displayed_name"`
	Filename      string `json:"filename"`
	Extension     string `json:"extension"`
}

// ─────────────────────────────────────────────
// CREATE OPERATIONS
// ─────────────────────────────────────────────

// Create video (DB)
func CreateVideo(db *sql.DB, video *VideoMedia) error {
	return CreateVideoTx(db, video)
}

// Create video (Tx)
func CreateVideoTx(exec interface {
	QueryRow(query string, args ...any) *sql.Row
}, video *VideoMedia) error {
	query := `
		INSERT INTO "VideoMedia" (uri, displayed_name, filename, extension)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`
	return exec.QueryRow(
		query,
		video.URI,
		video.DisplayedName,
		video.Filename,
		video.Extension,
	).Scan(&video.ID)
}

// Create multiple videos
func CreateVideosBulk(db *sql.DB, videos []VideoMedia) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT INTO "VideoMedia" (uri, displayed_name, filename, extension)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range videos {
		err := stmt.QueryRow(
			videos[i].URI,
			videos[i].DisplayedName,
			videos[i].Filename,
			videos[i].Extension,
		).Scan(&videos[i].ID)
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

// Retrieve a single video by its ID
func GetVideoByID(db *sql.DB, id int) (*VideoMedia, error) {
	query := `
		SELECT id, uri, displayed_name, filename, extension
		FROM "VideoMedia"
		WHERE id = $1;
	`
	video := &VideoMedia{}
	err := db.QueryRow(query, id).Scan(
		&video.ID,
		&video.URI,
		&video.DisplayedName,
		&video.Filename,
		&video.Extension,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return video, nil
}

// Retrieve all videos for a given carepartner and caregiver
func GetVideosByCarepartnerAndCaregiver(db *sql.DB, carepartnerID, caregiverID int) ([]VideoMedia, error) {
	query := `
		SELECT v.id, v.uri, v.displayed_name, v.filename, v.extension
		FROM "VideoMedia" v
		JOIN "Module" m ON m.metadata_id = v.id
		WHERE m.carepartner_id = $1 AND m.caregiver_id = $2;
	`
	rows, err := db.Query(query, carepartnerID, caregiverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []VideoMedia
	for rows.Next() {
		var v VideoMedia
		if err := rows.Scan(
			&v.ID,
			&v.URI,
			&v.DisplayedName,
			&v.Filename,
			&v.Extension,
		); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, nil
}

// ─────────────────────────────────────────────
// UPDATE / DELETE
// ─────────────────────────────────────────────

// Update the URI, displayed name, filename, and extension of a video
func UpdateVideo(db *sql.DB, video *VideoMedia) error {
	query := `
		UPDATE "VideoMedia"
		SET uri = $1,
		    displayed_name = $2,
		    filename = $3,
		    extension = $4
		WHERE id = $5;
	`
	res, err := db.Exec(
		query,
		video.URI,
		video.DisplayedName,
		video.Filename,
		video.Extension,
		video.ID,
	)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no video found with id %d", video.ID)
	}
	return nil
}

// Delete a video by its ID
func DeleteVideo(db *sql.DB, id int) error {
	query := `
		DELETE FROM "VideoMedia"
		WHERE id = $1;
	`
	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no video found with id %d", id)
	}
	return nil
}
