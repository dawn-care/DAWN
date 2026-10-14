package carepartner

import (
	"database/sql"
	"fmt"

	"server/internal/db"
)

/*
Carepartner represents a carepartner record.
CaregiverID - 0 until a caregiver pairs with the carepartner's device (NULL in the database)
Timezone - IANA zone the carepartner's dates and times are stored in, e.g. "America/New_York"
*/
type Carepartner struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	CaregiverID int    `json:"caregiver_id"`
	Timezone    string `json:"timezone"`
}

// nullableID stores an unset (0) caregiver ID as NULL so it passes the foreign key
func nullableID(id int) any {
	if id <= 0 {
		return nil
	}
	return id
}

func scanCarepartner(row db.Scanner) (*Carepartner, error) {
	c := &Carepartner{}
	var caregiverID sql.NullInt64
	if err := row.Scan(&c.ID, &c.Name, &caregiverID, &c.Timezone); err != nil {
		return nil, err
	}
	c.CaregiverID = int(caregiverID.Int64)
	return c, nil
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

// Create a carepartner; works with a *sql.DB or a *sql.Tx. Fills in ID and Timezone.
func CreateCarepartner(exec db.QueryRower, c *Carepartner) error {
	query := `
		INSERT INTO "Carepartner" (name, caregiver_id)
		VALUES ($1, $2)
		RETURNING id, timezone;
	`
	return exec.QueryRow(query, c.Name, nullableID(c.CaregiverID)).Scan(&c.ID, &c.Timezone)
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

func GetCarepartnerByID(db *sql.DB, id int) (*Carepartner, error) {
	query := `
		SELECT id, name, caregiver_id, timezone
		FROM "Carepartner"
		WHERE id = $1;
	`

	c, err := scanCarepartner(db.QueryRow(query, id))
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
		SELECT id, name, caregiver_id, timezone
		FROM "Carepartner";
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Carepartner
	for rows.Next() {
		c, err := scanCarepartner(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *c)
	}

	return list, nil
}

// ─────────────────────────────────────────────
// UPDATE
// ─────────────────────────────────────────────

func UpdateCarepartner(db *sql.DB, c *Carepartner) error {
	query := `
		UPDATE "Carepartner"
		SET name = $1, caregiver_id = $2
		WHERE id = $3;
	`

	res, err := db.Exec(query, c.Name, nullableID(c.CaregiverID), c.ID)
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
