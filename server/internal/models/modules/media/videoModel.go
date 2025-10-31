package media

import (
	"database/sql"
	"fmt"
)

type VideoMedia struct {
	ID        int    `json:"id"`
	URI       string `json:"uri"`
	Filename  string `json:"filename"`
	Extension int    `json:"extension"`
}

// Create a single video
func CreateVideo(db *sql.DB, video *VideoMedia) error {
	query := `
		INSERT INTO "VideoMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	return db.QueryRow(query, video.URI, video.Filename, video.Extension).Scan(&video.ID)
}

// Create multiple videos
func CreateVideosBulk(db *sql.DB, videos []VideoMedia) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT INTO "VideoMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range videos {
		err := stmt.QueryRow(videos[i].URI, videos[i].Filename, videos[i].Extension).Scan(&videos[i].ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// Retrieve a single video by its ID
func GetVideoByID(db *sql.DB, id int) (*VideoMedia, error) {
	query := `
		SELECT id, uri, filename, extension
		FROM "VideoMedia"
		WHERE id = $1;
	`
	video := &VideoMedia{}
	err := db.QueryRow(query, id).Scan(&video.ID, &video.URI, &video.Filename, &video.Extension)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return video, nil
}

// Update the URI, filename, and extension of a video
func UpdateVideo(db *sql.DB, video *VideoMedia) error {
	query := `
		UPDATE "VideoMedia"
		SET uri = $1, filename = $2, extension = $3
		WHERE id = $4;
	`
	res, err := db.Exec(query, video.URI, video.Filename, video.Extension, video.ID)
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

// Get the total number of videos
func GetVideoStats(db *sql.DB) (int, error) {
	query := `SELECT COUNT(*) FROM "VideoMedia";`
	var count int
	err := db.QueryRow(query).Scan(&count)
	return count, err
}
