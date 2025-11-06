package prompts

import (
	"database/sql"
	"fmt"
	"time"
)

type EventPrompt struct {
	List_ID   int        `json:"list_id"`
	ID        int        `json:"id"`
	Time      time.Time  `json:"time"`
	Text      string     `json:"text"`
}

// Create a list of events
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

// get a list of events based on list_id
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

// Update an event
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

// Delete an event by it's id
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

