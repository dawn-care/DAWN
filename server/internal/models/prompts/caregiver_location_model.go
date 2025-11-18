package prompts

import (
	"database/sql"
	"fmt"
)

/*
CaregiverLocationPrompt entries contain data the caregiver has been prompted to provide regarding their location
ID - the ID of the CaregiverLocationPrompt in the database
Location - the response provided by the caregiver regarding their location
Message - any additional message content supplied by the caregiver regarding their location, ex. "I am 15 minutes away and will be visiting at 3 PM today"
URI - the URI an image file to accompany this information, if desired
*/
type CaregiverLocationPrompt struct {
	ID       int    `json:"id"`
	Location string `json:"location"`
	Message  string `json:"message"`
	URI      string `json:"image_uri"`
}

// CREATE OPERATION

/*
Create a caregiver location entry
db - instance of the database
location - location information provided by the caregiver
*/
func CreateCaregiverLocation(db *sql.DB, location *CaregiverLocationPrompt) error {
	query := `
		INSERT INTO "CaregiverLocationPrompt" (location, message, image_uri)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	return db.QueryRow(query, location.Location, location.Message, location.URI).Scan(&location.ID)
}

// RETRIEVE OPERATION

/*
Retrieve stored caregiver location information by its ID
db - instance of the database
id - the ID of the entry in the database
*/
func GetCaregiverLocationByID(db *sql.DB, id int) (*CaregiverLocationPrompt, error) {
	query := `
		SELECT id, location, message, image_uri
		FROM "CaregiverLocationPrompt"
		WHERE id = $1;
	`
	location := &CaregiverLocationPrompt{}
	err := db.QueryRow(query, id).Scan(&location.ID, &location.Location, &location.Message, &location.URI)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return location, nil
}

/*
Update the location, message, and image of a caregiver location entry
db - instance of the database
location - location information submitted by a caregiver
*/
func UpdateCaregiverLocation(db *sql.DB, location *CaregiverLocationPrompt) error {
	query := `
		UPDATE "CaregiverLocationPrompt"
		SET location = $1, message = $2, image_uri = $3
		WHERE id = $4;
	`
	res, err := db.Exec(query, location.Location, location.Message, location.URI, location.ID)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no caregiver location found with id %d", location.ID)
	}
	return nil
}

// DELETE OPERATION

/*
Delete a caregiver location prompt by its ID
db - instance of the database
id - ID of the caregiver location entry in the database
*/
func DeleteCaregiverLocation(db *sql.DB, id int) error {
	query := `
		DELETE FROM "CaregiverLocationPrompt"
		WHERE id = $1;
	`
	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no caregiver location found with id %d", id)
	}
	return nil
}
