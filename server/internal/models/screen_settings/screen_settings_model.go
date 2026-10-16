package screen_settings

import (
	"database/sql"

	"server/internal/db"
)

/*
ScreenSettings holds a carepartner's screen layout and awake hours.
LayoutPreset - which layout the tablet uses, e.g. "balanced"
WakeTime, BedTime - awake hours as "HH:MM" in the carepartner's local time
*/
type ScreenSettings struct {
	CarepartnerID int    `json:"carepartner_id"`
	LayoutPreset  string `json:"layout_preset"`
	WakeTime      string `json:"wake_time"`
	BedTime       string `json:"bed_time"`
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

// Create a carepartner's settings row with the database defaults (Balanced, 07:00 to 21:00)
func CreateDefaultScreenSettings(exec db.Execer, carepartnerID int) error {
	_, err := exec.Exec(`INSERT INTO "ScreenSettings" (carepartner_id) VALUES ($1);`, carepartnerID)
	return err
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

// Retrieve a carepartner's settings, or nil if they have none
func GetScreenSettings(exec db.QueryRower, carepartnerID int) (*ScreenSettings, error) {
	query := `
		SELECT carepartner_id, layout_preset, to_char(wake_time, 'HH24:MI'), to_char(bed_time, 'HH24:MI')
		FROM "ScreenSettings"
		WHERE carepartner_id = $1;
	`
	s := &ScreenSettings{}
	err := exec.QueryRow(query, carepartnerID).Scan(&s.CarepartnerID, &s.LayoutPreset, &s.WakeTime, &s.BedTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Lock a carepartner's settings row for the rest of the transaction, so changes to
// their screen happen one at a time; returns false if they have no settings row
func LockScreenSettings(tx *sql.Tx, carepartnerID int) (bool, error) {
	var id int
	err := tx.QueryRow(`SELECT carepartner_id FROM "ScreenSettings" WHERE carepartner_id = $1 FOR UPDATE;`, carepartnerID).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// ─────────────────────────────────────────────
// UPDATE
// ─────────────────────────────────────────────

// Replace a carepartner's settings; returns false if they have no settings row
func UpdateScreenSettings(exec db.QueryRower, s *ScreenSettings) (bool, error) {
	query := `
		UPDATE "ScreenSettings"
		SET layout_preset = $1, wake_time = $2, bed_time = $3
		WHERE carepartner_id = $4
		RETURNING layout_preset, to_char(wake_time, 'HH24:MI'), to_char(bed_time, 'HH24:MI');
	`
	err := exec.QueryRow(query, s.LayoutPreset, s.WakeTime, s.BedTime, s.CarepartnerID).
		Scan(&s.LayoutPreset, &s.WakeTime, &s.BedTime)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
