package caregiver

import (
	"database/sql"
	"fmt"
)

// Caregiver represents a caregiver "user"
type Caregiver struct {
	ID            int     `json:"id"`
	CarepartnerID int     `json:"carepartner_id"`
	Name          string  `json:"name"`
	Role          string  `json:"role"`
	GoogleID      *string `json:"google_id"`
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

// CreateCaregiver inserts a single caregiver
func CreateCaregiver(db *sql.DB, c *Caregiver) error {
	query := `
		INSERT INTO "Caregiver" (carepartner_id, name, role, google_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`
	return db.QueryRow(query, c.CarepartnerID, c.Name, c.Role, c.GoogleID).Scan(&c.ID)
}

// CreateCaregiversBulk inserts many caregivers
func CreateCaregiversBulk(db *sql.DB, caregivers []Caregiver) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO "Caregiver" (carepartner_id, name, role)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range caregivers {
		err := stmt.QueryRow(
			caregivers[i].CarepartnerID,
			caregivers[i].Name,
			caregivers[i].Role,
		).Scan(&caregivers[i].ID)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

// GetCaregiverByID retrieves a caregiver by ID
func GetCaregiverByID(db *sql.DB, id int) (*Caregiver, error) {
	query := `
		SELECT id, carepartner_id, name, role
		FROM "Caregiver"
		WHERE id = $1;
	`

	c := &Caregiver{}
	err := db.QueryRow(query, id).Scan(
		&c.ID,
		&c.CarepartnerID,
		&c.Name,
		&c.Role,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return c, nil
}

// GetCaregiverByCarepartner returns the caregiver sharing the same ID
func GetCaregiverByCarepartner(db *sql.DB, crID int) (*Caregiver, error) {
	query := `
		SELECT id, carepartner_id, name, role
		FROM "Caregiver"
		WHERE carepartner_id = $1;
	`

	c := &Caregiver{}
	err := db.QueryRow(query, crID).Scan(
		&c.ID,
		&c.CarepartnerID,
		&c.Name,
		&c.Role,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return c, nil
}

// GetCaregivers returns all caregivers
func GetCaregivers(db *sql.DB) ([]Caregiver, error) {
	query := `
		SELECT id, carepartner_id, name, role
		FROM "Caregiver";
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Caregiver

	for rows.Next() {
		var c Caregiver
		if err := rows.Scan(&c.ID, &c.CarepartnerID, &c.Name, &c.Role); err != nil {
			return nil, err
		}
		list = append(list, c)
	}

	return list, nil
}

// SearchCaregiversByName allows partial name search
func SearchCaregiversByName(db *sql.DB, keyword string) ([]Caregiver, error) {
	query := `
		SELECT id, carepartner_id, name, role
		FROM "Caregiver"
		WHERE name ILIKE $1;
	`

	rows, err := db.Query(query, fmt.Sprintf("%%%s%%", keyword))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caregivers []Caregiver
	for rows.Next() {
		var c Caregiver
		if err := rows.Scan(&c.ID, &c.CarepartnerID, &c.Name, &c.Role); err != nil {
			return nil, err
		}
		caregivers = append(caregivers, c)
	}

	return caregivers, nil
}

// ─────────────────────────────────────────────
// UPDATE
// ─────────────────────────────────────────────

func UpdateCaregiver(db *sql.DB, c *Caregiver) error {
	query := `
		UPDATE "Caregiver"
		SET carepartner_id = $1, name = $2, role = $3
		WHERE id = $4;
	`

	res, err := db.Exec(query, c.CarepartnerID, c.Name, c.Role, c.ID)
	if err != nil {
		return err
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no caregiver found with id %d", c.ID)
	}
	return nil
}

// ─────────────────────────────────────────────
// DELETE
// ─────────────────────────────────────────────

func DeleteCaregiver(db *sql.DB, id int) error {
	query := `DELETE FROM "Caregiver" WHERE id = $1;`

	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no caregiver found with id %d", id)
	}

	return nil
}

// Get Caregiver by Google ID (OAuth Implementation)
func GetCaregiverByGoogleID(db *sql.DB, googleID string) (*Caregiver, error) {
	query := `
		SELECT id, carepartner_id, name, role, google_id
		FROM "Caregiver"
		WHERE google_id = $1;
	`

	c := &Caregiver{}
	err := db.QueryRow(query, googleID).Scan(
		&c.ID,
		&c.CarepartnerID,
		&c.Name,
		&c.Role,
		&c.GoogleID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}
