package prompts

import (
	"database/sql"
	"fmt"
)

/*
MessagePrompt defines a general message accompanied by an image if desired
ID - the ID of the entry in the database
Message - the message content submitted by the caregiver
URI - the URI of the accompanying image
*/
type MessagePrompt struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
	URI     string `json:"image_uri"`
}

// CREATE OPERATION

/*
Create a message entry
db - instance of database
message - the message content
*/
func CreateMessage(db *sql.DB, message *MessagePrompt) error {
	query := `
		INSERT INTO "MessagePrompt" (message, image_uri)
		VALUES ($1, $2)
		RETURNING id;
	`
	return db.QueryRow(query, message.Message, message.URI).Scan(&message.ID)
}

// RETRIEVE OPERATION

/*
Retrieve message entry by ID
db - instance of the database
id - the ID of the message in the database
*/
func GetMessageByID(db *sql.DB, id int) (*MessagePrompt, error) {
	query := `
		SELECT id, message, image_uri
		FROM "MessagePrompt"
		WHERE id = $1;
	`
	message := &MessagePrompt{}
	err := db.QueryRow(query, id).Scan(&message.ID, &message.Message, &message.URI)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return message, nil
}

/*
Update the message and image of a message prompt
db - instance of the database
message - message prompt data to be updated
*/
func UpdateMessagePrompt(db *sql.DB, message *MessagePrompt) error {
	query := `
		UPDATE "MessagePrompt"
		SET message = $1, image_uri = $2
		WHERE id = $3;
	`
	res, err := db.Exec(query, message.Message, message.URI, message.ID)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no message found with id %d", message.ID)
	}
	return nil
}

// DELETE OPERATION

/*
Delete a carepartner location prompt by its ID
db - instance of the database
id - the ID of the message to be deleted
*/
func DeleteMessage(db *sql.DB, id int) error {
	query := `
		DELETE FROM "MessagePrompt"
		WHERE id = $1;
	`
	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no message found with id %d", id)
	}
	return nil
}
