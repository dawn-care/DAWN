package screen_widgets

import (
	"database/sql"
	"encoding/json"

	"server/internal/db"
	"server/internal/models/media"
)

/*
ScreenWidget is a widget placed in one spot of a layout preset.
Spot - where it sits: 0-based, The widgets follow reading order
Config - settings for its type, validated by the screen widget service
PhotoMediaID - the photo shown by a photo widget, or an optional photo on a message widget
PhotoURL - derived: where the tablet can fetch that photo
Hidden - derived: the current preset has no such spot, so the tablet doesn't show it
*/
type ScreenWidget struct {
	ID            int             `json:"id"`
	CarepartnerID int             `json:"carepartner_id"`
	Type          string          `json:"type"`
	Spot          int             `json:"spot"`
	Config        json.RawMessage `json:"config" swaggertype:"object"`
	PhotoMediaID  *int            `json:"photo_media_id"`
	PhotoURL      *string         `json:"photo_url"`
	Hidden        bool            `json:"hidden"`
}

const selectWidget = `
	SELECT id, carepartner_id, type, spot, config, photo_media_id
	FROM "ScreenWidget"
`

func scanWidget(row db.Scanner) (*ScreenWidget, error) {
	w := &ScreenWidget{}
	var config []byte
	var photo sql.NullInt64
	if err := row.Scan(&w.ID, &w.CarepartnerID, &w.Type, &w.Spot, &config, &photo); err != nil {
		return nil, err
	}
	w.Config = json.RawMessage(config)
	if photo.Valid {
		id := int(photo.Int64)
		w.PhotoMediaID = &id
	}
	w.setPhotoURL()
	return w, nil
}

func (w *ScreenWidget) setPhotoURL() {
	w.PhotoURL = nil
	if w.PhotoMediaID != nil {
		url := media.FileURL(*w.PhotoMediaID)
		w.PhotoURL = &url
	}
}

func nullableRow(w *ScreenWidget, err error) (*ScreenWidget, error) {
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return w, err
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

// Create a widget; fills in ID and PhotoURL
func CreateScreenWidget(exec db.QueryRower, w *ScreenWidget) error {
	query := `
		INSERT INTO "ScreenWidget" (carepartner_id, type, spot, config, photo_media_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`
	if err := exec.QueryRow(query, w.CarepartnerID, w.Type, w.Spot, []byte(w.Config), w.PhotoMediaID).Scan(&w.ID); err != nil {
		return err
	}
	w.setPhotoURL()
	return nil
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

// Retrieve one widget, or nil if it doesn't exist
func GetScreenWidgetByID(exec db.QueryRower, id int) (*ScreenWidget, error) {
	return nullableRow(scanWidget(exec.QueryRow(selectWidget+` WHERE id = $1;`, id)))
}

// Retrieve and lock one widget for the rest of the transaction, or nil if it doesn't exist
func GetScreenWidgetForUpdate(tx *sql.Tx, id int) (*ScreenWidget, error) {
	return nullableRow(scanWidget(tx.QueryRow(selectWidget+` WHERE id = $1 FOR UPDATE;`, id)))
}

// Retrieve and lock the widget in a spot, or nil if the spot is empty
func GetScreenWidgetAt(tx *sql.Tx, carepartnerID int, spot int) (*ScreenWidget, error) {
	return nullableRow(scanWidget(tx.QueryRow(
		selectWidget+` WHERE carepartner_id = $1 AND spot = $2 FOR UPDATE;`,
		carepartnerID, spot,
	)))
}

// List a carepartner's widgets ordered by spot
func ListScreenWidgets(db *sql.DB, carepartnerID int) ([]ScreenWidget, error) {
	rows, err := db.Query(selectWidget+` WHERE carepartner_id = $1 ORDER BY spot, id;`, carepartnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	widgets := []ScreenWidget{}
	for rows.Next() {
		w, err := scanWidget(rows)
		if err != nil {
			return nil, err
		}
		widgets = append(widgets, *w)
	}
	return widgets, rows.Err()
}

// Retrieve a media item for checking a widget's photo (its kind and owner), or nil
func GetPhoto(exec db.QueryRower, mediaID int) (*media.Media, error) {
	return media.GetMediaByID(exec, mediaID)
}

// ─────────────────────────────────────────────
// UPDATE / DELETE
// ─────────────────────────────────────────────

// Replace a widget's settings and photo
func UpdateScreenWidgetConfig(exec db.Execer, w *ScreenWidget) error {
	_, err := exec.Exec(
		`UPDATE "ScreenWidget" SET config = $1, photo_media_id = $2 WHERE id = $3;`,
		[]byte(w.Config), w.PhotoMediaID, w.ID,
	)
	if err == nil {
		w.setPhotoURL()
	}
	return err
}

// Move a widget to another spot. The spot uniqueness check is deferred to commit,
// so two widgets can be swapped inside one transaction.
func MoveScreenWidget(exec db.Execer, id int, spot int) error {
	_, err := exec.Exec(`UPDATE "ScreenWidget" SET spot = $1 WHERE id = $2;`, spot, id)
	return err
}

// Delete a widget; returns false if no such widget
func DeleteScreenWidget(db *sql.DB, id int) (bool, error) {
	res, err := db.Exec(`DELETE FROM "ScreenWidget" WHERE id = $1;`, id)
	if err != nil {
		return false, err
	}
	count, err := res.RowsAffected()
	return count > 0, err
}
