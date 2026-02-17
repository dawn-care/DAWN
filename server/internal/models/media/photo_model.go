package media

import (
	"database/sql"
	"fmt"
)

type PhotoMedia struct {
	ID            int    `json:"id"`
	URI           string `json:"uri"`
	DisplayedName string `json:"displayed_name"`
	Filename      string `json:"filename"`
	Extension     string `json:"extension"`
}

// ─────────────────────────────────────────────
// CREATE OPERATIONS
// ─────────────────────────────────────────────

// Create photo (DB)
func CreatePhoto(db *sql.DB, photo *PhotoMedia) error {
	return CreatePhotoTx(db, photo)
}

// Create photo (Tx)
func CreatePhotoTx(exec interface {
	QueryRow(query string, args ...any) *sql.Row
}, photo *PhotoMedia) error {
	query := `
		INSERT INTO "PhotoMedia" (uri, displayed_name, filename, extension)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`
	return exec.QueryRow(
		query,
		photo.URI,
		photo.DisplayedName,
		photo.Filename,
		photo.Extension,
	).Scan(&photo.ID)
}

// Create multiple photos
func CreatePhotosBulk(db *sql.DB, photos []PhotoMedia) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT INTO "PhotoMedia" (uri, displayed_name, filename, extension)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range photos {
		err := stmt.QueryRow(
			photos[i].URI,
			photos[i].DisplayedName,
			photos[i].Filename,
			photos[i].Extension,
		).Scan(&photos[i].ID)
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

// Retrieve a single photo by its ID
func GetPhotoByID(db *sql.DB, id int) (*PhotoMedia, error) {
	query := `
		SELECT id, uri, displayed_name, filename, extension
		FROM "PhotoMedia"
		WHERE id = $1;
	`
	photo := &PhotoMedia{}
	err := db.QueryRow(query, id).Scan(
		&photo.ID,
		&photo.URI,
		&photo.DisplayedName,
		&photo.Filename,
		&photo.Extension,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return photo, nil
}

// Retrieve all photos for a given carepartner and caregiver
func GetPhotosByCarepartnerAndCaregiver(db *sql.DB, carepartnerID, caregiverID int) ([]PhotoMedia, error) {
	query := `
		SELECT p.id, p.uri, p.displayed_name, p.filename, p.extension
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
		if err := rows.Scan(
			&p.ID,
			&p.URI,
			&p.DisplayedName,
			&p.Filename,
			&p.Extension,
		); err != nil {
			return nil, err
		}
		photos = append(photos, p)
	}
	return photos, nil
}

// ─────────────────────────────────────────────
// UPDATE / DELETE
// ─────────────────────────────────────────────

// Update the URI, displayed name, filename, and extension of a photo
func UpdatePhoto(db *sql.DB, photo *PhotoMedia) error {
	query := `
		UPDATE "PhotoMedia"
		SET uri = $1,
		    displayed_name = $2,
		    filename = $3,
		    extension = $4
		WHERE id = $5;
	`
	res, err := db.Exec(
		query,
		photo.URI,
		photo.DisplayedName,
		photo.Filename,
		photo.Extension,
		photo.ID,
	)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no photo found with id %d", photo.ID)
	}
	return nil
}

// Delete a photo by its ID
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
