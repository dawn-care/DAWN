package prompts

import (
	"database/sql"
	"fmt"
)

/*
CarepartnerLocationPrompt provides data entered by the caregiver regarding the carepartner's location, ex. "You are safe in a home in Rochester, NY"
ID - the ID of the database entry
Location - the provided location of the carepartner
Message - any additional information regarding the location of the carepartner
URI - the URI of any photo that should accompany this information, if desired
*/
type CarepartnerLocationPrompt struct {
	ID       int    `json:"id"`
	Location string `json:"location"`
	Message  string `json:"message"`
	URI      string `json:"image_uri"`
}

// CREATE OPERATION

/*
Create a carepartner location entry
db - instance of the database
location - the location information for the carepartner
*/
func CreateCarepartnerLocation(db *sql.DB, location *CarepartnerLocationPrompt) error {
	query := `
		INSERT INTO "CarepartnerLocationPrompt" (location, message, image_uri)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	return db.QueryRow(query, location.Location, location.Message, location.URI).Scan(&location.ID)
}

// RETRIEVE OPERATION

/*
Read a carepartner location entry by its ID
db - instance of the database
id - ID of the carepartner location entry
*/
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

// UPDATE OPERATION

/*
Update the location, message, and image of a carepartner location entry
db - instance of the database
lcoation - the location information for the carepartner
*/
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

// DELETE OPERATION

/*
Delete a carepartner location prompt by its ID
db - instance of the database
id - the ID of the carepartner location entry in the database
*/
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
