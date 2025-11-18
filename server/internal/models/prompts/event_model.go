package prompts

import (
	"database/sql"
	"fmt"
	"time"
)

/*
EventPrompt holds information provided regarding an event that will occur, ex. "There will be a band today at 3 PM"
List_ID - the ID of the event in the listing
ID - the ID of the event in the database table
Time - the time the event will occur
Text - information regarding the event
*/
type EventPrompt struct {
	List_ID int       `json:"list_id"`
	ID      int       `json:"id"`
	Time    time.Time `json:"time"`
	Text    string    `json:"text"`
}

// CREATE OPERATION

/*
Create a list of events
db - instance of the database
events - array of event detail entries
*/
func CreateEventList(db *sql.DB, events []*EventPrompt) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT INTO "EventPrompt" (list_id, time, text)
		VALUES ($1, $2, $3)
		RETURNING id
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range events {
		err := stmt.QueryRow(events[i].List_ID, events[i].Time, events[i].Text).Scan(&events[i].ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// RETRIEVE OPERATIONS

/*
Get a list of events based on list_id
db - instance of the database
list_id - the list ID value to retrieve events associated with
*/
func GetEventListByListID(db *sql.DB, list_id *int) ([]EventPrompt, error) {
	query := `
		SELECT *
		FROM "EventPrompt"
		WHERE list_id = $1;
	`
	rows, err := db.Query(query, list_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eventlist []EventPrompt
	for rows.Next() {
		var event EventPrompt
		err := rows.Scan(&event.List_ID, &event.ID, &event.Time, &event.Text)
		if err != nil {
			return nil, err
		}
		eventlist = append(eventlist, event)
	}
	return eventlist, nil
}

// UPDATE OPERATION

/*
Update the content of an event
db - instance of the database
event - an event entry with updated information
*/
func UpdateEvent(db *sql.DB, event *EventPrompt) error {
	query := `
		UPDATE "EventPrompt"
		SET list_id = $1, time = $2, text = $3
		WHERE id = $4;
	`
	res, err := db.Exec(query, event.List_ID, event.Time, event.Text, event.ID)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no event found with id %d", event.ID)
	}
	return nil
}

// DELETE OPERATION

/*
Delete an event by the ID
db - instance of the database
id - the ID of the event in the database
*/
func DeleteEvent(db *sql.DB, id *int) error {
	query := `
		DELETE FROM "EventPrompt"
		WHERE id = $1;
	`
	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no event found with id %d", id)
	}
	return nil
}

// DELETE OPERATION

/*
Delete an event entry from a list
db - instance of the database
list_id - list ID that identifies what entry to delete
*/
func DeleteEventList(db *sql.DB, list_id *int) error {
	query := `
		DELETE FROM "EventPrompt"
		WHERE list_id = $1;
	`
	res, err := db.Exec(query, list_id)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no event list found with id %d", list_id)
	}
	return nil
}
