package screen_widgets

import (
	"database/sql"
	"encoding/json"

	"server/internal/models/media"
)

/*
ScreenWidget is a widget placed in one slot of a layout preset's region.
Region, Slot - where it sits, e.g. "tiles", 1 for the second tile
Config - settings for its type, validated by the screen widget service
PhotoMediaID - the photo shown by a photo widget, or an optional photo on a message widget
PhotoURL - derived: where the tablet can fetch that photo
Hidden - derived: the current preset has no such region or slot, so the tablet doesn't show it
*/
type ScreenWidget struct {
	ID            int             `json:"id"`
	CarepartnerID int             `json:"carepartner_id"`
	Type          string          `json:"type"`
	Region        string          `json:"region"`
	Slot          int             `json:"slot"`
	Config        json.RawMessage `json:"config" swaggertype:"object"`
	PhotoMediaID  *int            `json:"photo_media_id"`
	PhotoURL      *string         `json:"photo_url"`
	Hidden        bool            `json:"hidden"`
}

type queryRower interface {
	QueryRow(query string, args ...any) *sql.Row
}

type execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

const selectWidget = `
	SELECT id, carepartner_id, type, region, slot, config, photo_media_id
	FROM "ScreenWidget"
`

type scanner interface {
	Scan(dest ...any) error
}

func scanWidget(row scanner) (*ScreenWidget, error) {
	w := &ScreenWidget{}
	var config []byte
	var photo sql.NullInt64
	if err := row.Scan(&w.ID, &w.CarepartnerID, &w.Type, &w.Region, &w.Slot, &config, &photo); err != nil {
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
func CreateScreenWidget(exec queryRower, w *ScreenWidget) error {
	query := `
		INSERT INTO "ScreenWidget" (carepartner_id, type, region, slot, config, photo_media_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`
	if err := exec.QueryRow(query, w.CarepartnerID, w.Type, w.Region, w.Slot, []byte(w.Config), w.PhotoMediaID).Scan(&w.ID); err != nil {
		return err
	}
	w.setPhotoURL()
	return nil
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

// Retrieve one widget, or nil if it doesn't exist
func GetScreenWidgetByID(exec queryRower, id int) (*ScreenWidget, error) {
	return nullableRow(scanWidget(exec.QueryRow(selectWidget+` WHERE id = $1;`, id)))
}

// Retrieve and lock one widget for the rest of the transaction, or nil if it doesn't exist
func GetScreenWidgetForUpdate(tx *sql.Tx, id int) (*ScreenWidget, error) {
	return nullableRow(scanWidget(tx.QueryRow(selectWidget+` WHERE id = $1 FOR UPDATE;`, id)))
}

// Retrieve and lock the widget in a slot, or nil if the slot is empty
func GetScreenWidgetAt(tx *sql.Tx, carepartnerID int, region string, slot int) (*ScreenWidget, error) {
	return nullableRow(scanWidget(tx.QueryRow(
		selectWidget+` WHERE carepartner_id = $1 AND region = $2 AND slot = $3 FOR UPDATE;`,
		carepartnerID, region, slot,
	)))
}

// List a carepartner's widgets ordered by region and slot
func ListScreenWidgets(db *sql.DB, carepartnerID int) ([]ScreenWidget, error) {
	rows, err := db.Query(selectWidget+` WHERE carepartner_id = $1 ORDER BY region, slot, id;`, carepartnerID)
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
func GetPhoto(exec queryRower, mediaID int) (*media.Media, error) {
	return media.GetMediaByID(exec, mediaID)
}

// ─────────────────────────────────────────────
// UPDATE / DELETE
// ─────────────────────────────────────────────

// Replace a widget's settings and photo
func UpdateScreenWidgetConfig(exec execer, w *ScreenWidget) error {
	_, err := exec.Exec(
		`UPDATE "ScreenWidget" SET config = $1, photo_media_id = $2 WHERE id = $3;`,
		[]byte(w.Config), w.PhotoMediaID, w.ID,
	)
	if err == nil {
		w.setPhotoURL()
	}
	return err
}

// Move a widget to another region and slot. The slot uniqueness check is deferred
// to commit, so two widgets can be swapped inside one transaction.
func MoveScreenWidget(exec execer, id int, region string, slot int) error {
	_, err := exec.Exec(`UPDATE "ScreenWidget" SET region = $1, slot = $2 WHERE id = $3;`, region, slot, id)
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
