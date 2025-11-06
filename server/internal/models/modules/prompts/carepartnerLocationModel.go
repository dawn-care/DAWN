package prompts

import (
	"database/sql"
	"fmt"
)

type CarepartnerLocationPrompt struct {
	ID       int    `json:"id"`
	Location string `json:"location"`
	Message  string `json:"message"`
	URI      string `json:"image_uri"`
}

// Create a carepartner location
func CreateCarepartnerLocation(db *sql.DB, location *CarepartnerLocationPrompt) error {
	query := `
		INSERT INTO "CarepartnerLocationPrompt" (location, message, image_uri)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	return db.QueryRow(query, location.Location, location.Message, location.URI).Scan(&location.ID)
}

// Read a carepartner location by its ID
func GetCarepartnerLocationByID(db *sql.DB, id int) (*CarepartnerLocationPrompt, error) {
	query := `
		SELECT id, location, message, image_uri
		FROM "CarepartnerLocationPrompt"
		WHERE id = $1;
	`
	location := &CarepartnerLocationPrompt{}
	err := db.QueryRow(query, id).Scan(&location.ID, &location.Location, &location.Message, &location.URI)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return location, nil
}

// Update the location, message, and image of a carepartner location
func UpdateCarepartnerLocation(db *sql.DB, location *CarepartnerLocationPrompt) error {
	query := `
		UPDATE "CarepartnerLocationPrompt"
		SET location = $1, message = $2, image_uri = $3
		WHERE id = $4;
	`
	res, err := db.Exec(query, location.Location, location.Message, location.URI, location.ID)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no carepartner location found with id %d", location.ID)
	}
	return nil
}

// Delete a carepartner location prompt by its ID
func DeleteCarepartnerLocation(db *sql.DB, id int) error {
	query := `
		DELETE FROM "CarepartnerLocationPrompt"
		WHERE id = $1;
	`
	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no carepartner location found with id %d", id)
	}
	return nil
}
