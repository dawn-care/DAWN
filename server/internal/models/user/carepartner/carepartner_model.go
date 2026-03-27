package carepartner

import (
	"database/sql"
	"fmt"
	"time"
)

// Carepartner represents a carepartner record
type Carepartner struct {
	ID       	int       `json:"id"`
	Name      	string    `json:"name"`
	CaregiverID int	  	  `json:"caregiver_id"`
	AutoStart 	time.Time `json:"auto_start"`
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

func CreateCarepartner(db *sql.DB, c *Carepartner) error {
	query := `
		INSERT INTO "Carepartner" (name, auto_start, caregiver_id)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	return db.QueryRow(query, c.Name, c.AutoStart, c.CaregiverID).Scan(&c.ID)
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

func GetCarepartnerByID(db *sql.DB, id int) (*Carepartner, error) {
	query := `
		SELECT id, name, auto_start, caregiver_id
		FROM "Carepartner"
		WHERE id = $1;
	`

	c := &Carepartner{}
	err := db.QueryRow(query, id).Scan(&c.ID, &c.Name, &c.AutoStart, &c.CaregiverID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return c, nil
}

func GetAllCarepartners(db *sql.DB) ([]Carepartner, error) {
	query := `
		SELECT id, name, auto_start, caregiver_id
		FROM "Carepartner";
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Carepartner
	for rows.Next() {
		var c Carepartner
		if err := rows.Scan(&c.ID, &c.Name, &c.AutoStart, &c.CaregiverID); err != nil {
			return nil, err
		}
		list = append(list, c)
	}

	return list, nil
}

// ─────────────────────────────────────────────
// UPDATE
// ─────────────────────────────────────────────

func UpdateCarepartner(db *sql.DB, c *Carepartner) error {
	query := `
		UPDATE "Carepartner"
		SET name = $1, auto_start = $2, caregiver_id = $3
		WHERE id = $4;
	`

	res, err := db.Exec(query, c.Name, c.AutoStart, c.CaregiverID, c.ID)
	if err != nil {
		return err
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no carepartner found with id %d", c.ID)
	}

	return nil
}

// ─────────────────────────────────────────────
// DELETE
// ─────────────────────────────────────────────

func DeleteCarepartner(db *sql.DB, id int) error {
	query := `DELETE FROM "Carepartner" WHERE id = $1;`

	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no carepartner found with id %d", id)
	}

	return nil
}
