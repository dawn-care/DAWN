package caregiver

import (
	"database/sql"
	"errors"
	"fmt"

	"server/internal/db"

	"github.com/lib/pq"
)

// emailIndex is the unique index on lower(email) (migration 10)
const emailIndex = "Caregiver_email_lower_key"

// IsEmailTaken reports whether err is another caregiver already having this email
func IsEmailTaken(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == emailIndex
}

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
	count, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
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
	if IsEmailTaken(err) {
		// Another sign-in saved the same address at the same moment
		return false, nil
	}
	if err != nil {
		return false, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// MergeEmailCaregiverInto folds an email-only caregiver (one with no Google ID) that holds
// email into the caregiver targetID, which has no email yet. The email-only caregiver's
// carepartners move to targetID, it is deleted (its sessions end with it), and targetID gets
// the email. It reports the merged caregiver's ID, or 0 when there was nothing to merge: no
// email-only caregiver holds the address, or targetID already has an email.
//
// The caller must already know that the person signing in to targetID owns email.
func MergeEmailCaregiverInto(conn *sql.DB, targetID int, email string) (int, error) {
	tx, err := conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var targetFound bool
	err = tx.QueryRow(`
		SELECT true FROM "Caregiver" WHERE id = $1 AND email IS NULL FOR UPDATE;
	`, targetID).Scan(&targetFound)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("locking caregiver %d: %w", targetID, err)
	}

	var sourceID int
	err = tx.QueryRow(`
		SELECT id FROM "Caregiver"
		WHERE lower(email) = lower($1) AND google_id IS NULL AND id <> $2
		FOR UPDATE;
	`, email, targetID).Scan(&sourceID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("finding the email caregiver to merge: %w", err)
	}

	if _, err := tx.Exec(`UPDATE "Carepartner" SET caregiver_id = $1 WHERE caregiver_id = $2;`, targetID, sourceID); err != nil {
		return 0, fmt.Errorf("moving carepartners from caregiver %d to %d: %w", sourceID, targetID, err)
	}
	if _, err := tx.Exec(`DELETE FROM "Caregiver" WHERE id = $1;`, sourceID); err != nil {
		return 0, fmt.Errorf("deleting merged caregiver %d: %w", sourceID, err)
	}
	if _, err := tx.Exec(`UPDATE "Caregiver" SET email = $1 WHERE id = $2;`, email, targetID); err != nil {
		return 0, fmt.Errorf("saving the email on caregiver %d: %w", targetID, err)
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return sourceID, nil
}
