package media

import (
	"database/sql"
	"fmt"
)

/*
PhotoMedia represents one photo record
ID - the ID of the photo in the database
URI - the URI of the photo in the database
Filename - the name of the file
Extension - the file extension, i.e. what type of file it is
*/
type PhotoMedia struct {
	ID        int    `json:"id"`
	URI       string `json:"uri"`
	Filename  string `json:"filename"`
	Extension string `json:"extension"`
}

// CREATE OPERATIONS

/*
Create a single photo record
db - instance of the database
photo - the photo record to be added to the database
*/
func CreatePhoto(db *sql.DB, photo *PhotoMedia) error {
	query := `
		INSERT INTO "PhotoMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	return db.QueryRow(query, photo.URI, photo.Filename, photo.Extension).Scan(&photo.ID)
}

/*
Create multiple photo records at once
db - instance of the database
photos - array of photo records to be added to the database
*/
func CreatePhotosBulk(db *sql.DB, photos []PhotoMedia) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT INTO "PhotoMedia" (uri, filename, extension)
		VALUES ($1, $2, $3)
		RETURNING id
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range photos {
		err := stmt.QueryRow(photos[i].URI, photos[i].Filename, photos[i].Extension).Scan(&photos[i].ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// RETRIEVE OPERATIONS

/*
Retrieve a single photo by the ID
db - instance of the database
id - ID of the photo in the database
*/
func GetPhotoByID(db *sql.DB, id int) (*PhotoMedia, error) {
	query := `
		SELECT id, uri, filename, extension
		FROM "PhotoMedia"
		WHERE id = $1;
	`
	photo := &PhotoMedia{}
	err := db.QueryRow(query, id).Scan(&photo.ID, &photo.URI, &photo.Filename, &photo.Extension)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return photo, nil
}

/*
Retrieve all photos exchanged between a given carepartner and caregiver
db - instance of the database
carepartnerID - the ID of the carepartner who receives the media
caregiverID - the ID of the caregiver who uploaded the media
*/
func GetPhotosByCarepartnerAndCaregiver(db *sql.DB, carepartnerID, caregiverID int) ([]PhotoMedia, error) {
	query := `
		SELECT p.id, p.uri, p.filename, p.extension
		FROM "PhotoMedia" p
		JOIN "Module" m ON m.metadata_id = p.id
		WHERE m.carepartner_id = $1 AND m.caregiver_id = $2;
	`
	rows, err := db.Query(query, carepartnerID, caregiverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []PhotoMedia
	for rows.Next() {
		var p PhotoMedia
		if err := rows.Scan(&p.ID, &p.URI, &p.Filename, &p.Extension); err != nil {
			return nil, err
		}
		photos = append(photos, p)
	}
	return photos, nil
}

// UPDATE OPERATION

/*
Update the URI, filename, and extension of a photo
db - instance of the database
photo - the photo record to update in the database
*/
func UpdatePhoto(db *sql.DB, photo *PhotoMedia) error {
	query := `
		UPDATE "PhotoMedia"
		SET uri = $1, filename = $2, extension = $3
		WHERE id = $4;
	`
	res, err := db.Exec(query, photo.URI, photo.Filename, photo.Extension, photo.ID)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no photo found with id %d", photo.ID)
	}
	return nil
}

// DELETE OPERATION

/*
Delete a photo by its ID
db - instance of the database
id - ID of the photo in the database
*/
func DeletePhoto(db *sql.DB, id int) error {
	query := `
		DELETE FROM "PhotoMedia"
		WHERE id = $1;
	`
	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no photo found with id %d", id)
	}
	return nil
}

// ANALYTICS & STATS

/*
Get a count of how many photo uploads are available for all users cumulative
db - instance of the database
*/
func GetPhotoStats(db *sql.DB) (int, error) {
	query := `SELECT COUNT(*) FROM "PhotoMedia";`
	var count int
	err := db.QueryRow(query).Scan(&count)
	return count, err
}
