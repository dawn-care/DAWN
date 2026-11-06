package caregiver

import (
	"database/sql"
	"fmt"

	"server/internal/db"
)

// Caregiver represents a caregiver "user"
type Caregiver struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Role     string  `json:"role"`
	GoogleID *string `json:"google_id"`
	// Email is stored lowercase. It is never sent in a response: several caregiver routes
	// are not behind login yet, so the address stays on the server.
	Email *string `json:"-"`
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

// CreateCaregiver inserts a single caregiver
func CreateCaregiver(db *sql.DB, c *Caregiver) error {
	query := `
		INSERT INTO "Caregiver" (name, role, google_id, email)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`
	return db.QueryRow(query, c.Name, c.Role, c.GoogleID, c.Email).Scan(&c.ID)
}

// CreateCaregiversBulk inserts many caregivers
func CreateCaregiversBulk(db *sql.DB, caregivers []Caregiver) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO "Caregiver" (name, role)
		VALUES ($1, $2)
		RETURNING id;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range caregivers {
		err := stmt.QueryRow(
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
		SELECT id, name, role, google_id, email
		FROM "Caregiver"
		WHERE id = $1;
	`

	c := &Caregiver{}
	err := db.QueryRow(query, id).Scan(
		&c.ID,
		&c.Name,
		&c.Role,
		&c.GoogleID,
		&c.Email,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return c, nil
}

// GetCarepartnersByCaregiverID retrieves all carepartners for a given caregiver
func GetCarepartnersByCaregiverID(db *sql.DB, caregiverID int) ([]Caregiver, error) {
	query := `
		SELECT id, name
		FROM "Carepartner"
		WHERE caregiver_id = $1;
	`

	rows, err := db.Query(query, caregiverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var carepartners []Caregiver

	for rows.Next() {
		var cp Caregiver
		if err := rows.Scan(&cp.ID, &cp.Name); err != nil {
			return nil, err
		}
		carepartners = append(carepartners, cp)
	}

	return carepartners, nil
}

// GetCaregiverByCarepartner returns the caregiver sharing the same ID
// func GetCaregiverByCarepartner(db *sql.DB, crID int) (*Caregiver, error) {
// 	query := `
// 		SELECT id, name, role
// 		FROM "Caregiver"
// 		WHERE carepartner_id = $1;
// 	`

// 	c := &Caregiver{}
// 	err := db.QueryRow(query, crID).Scan(
// 		&c.ID,
// 		&c.Name,
// 		&c.Role,
// 	)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}

// 	return c, nil
// }

// GetCaregivers returns all caregivers
func GetCaregivers(db *sql.DB) ([]Caregiver, error) {
	query := `
		SELECT id, name, role
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
		if err := rows.Scan(&c.ID, &c.Name, &c.Role); err != nil {
			return nil, err
		}
		list = append(list, c)
	}

	return list, nil
}

// SearchCaregiversByName allows partial name search
func SearchCaregiversByName(db *sql.DB, keyword string) ([]Caregiver, error) {
	query := `
		SELECT id, name, role
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
		if err := rows.Scan(&c.ID, &c.Name, &c.Role); err != nil {
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
		SET name = $1, role = $2
		WHERE id = $3;
	`

	res, err := db.Exec(query, c.Name, c.Role, c.ID)
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

// Delete a caregiver; works with a *sql.DB or a *sql.Tx
func DeleteCaregiver(exec db.Execer, id int) error {
	query := `DELETE FROM "Caregiver" WHERE id = $1;`

	res, err := exec.Exec(query, id)
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
		SELECT id, name, role, google_id, email
		FROM "Caregiver"
		WHERE google_id = $1;
	`

	c := &Caregiver{}
	err := db.QueryRow(query, googleID).Scan(
		&c.ID,
		&c.Name,
		&c.Role,
		&c.GoogleID,
		&c.Email,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

// GetCaregiverByEmail finds a caregiver by email, ignoring case (email sign-in)
func GetCaregiverByEmail(db *sql.DB, email string) (*Caregiver, error) {
	query := `
		SELECT id, name, role, google_id, email
		FROM "Caregiver"
		WHERE lower(email) = lower($1);
	`

	c := &Caregiver{}
	err := db.QueryRow(query, email).Scan(
		&c.ID,
		&c.Name,
		&c.Role,
		&c.GoogleID,
		&c.Email,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

// SetGoogleID links a Google account to a caregiver that has none yet.
// It reports whether the caregiver was changed; it never replaces a Google ID.
func SetGoogleID(db *sql.DB, id int, googleID string) (bool, error) {
	res, err := db.Exec(`
		UPDATE "Caregiver"
		SET google_id = $1
		WHERE id = $2 AND google_id IS NULL;
	`, googleID, id)
	if err != nil {
		return false, err
	}
	count, _ := res.RowsAffected()
	return count > 0, nil
}

// SetEmail saves the email of a caregiver that has none yet.
// It reports whether the caregiver was changed. It does nothing when the caregiver already
// has an email or when another caregiver already uses this one.
func SetEmail(db *sql.DB, id int, email string) (bool, error) {
	res, err := db.Exec(`
		UPDATE "Caregiver"
		SET email = $1
		WHERE id = $2 AND email IS NULL
		  AND NOT EXISTS (SELECT 1 FROM "Caregiver" WHERE lower(email) = lower($1));
	`, email, id)
	if err != nil {
		return false, err
	}
	count, _ := res.RowsAffected()
	return count > 0, nil
}
