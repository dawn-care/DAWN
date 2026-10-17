package screen_widgets

import (
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"

	"server/internal/db"
	"server/internal/models/media"
)

/*
ScreenWidget is a widget placed on the tablet's grid.
TopLeft, BottomRight - the cells at its corners, both inclusive, so a 1x1 widget has the
same cell for both
Config - settings for its type, validated by the screen widget service
PhotoMediaID - the photo shown by a photo widget, or an optional photo on a message widget
PhotoURL - derived: where the tablet can fetch that photo
*/
type ScreenWidget struct {
	ID            int             `json:"id"`
	CarepartnerID int             `json:"carepartner_id"`
	Type          string          `json:"type"`
	TopLeft       Cell            `json:"top_left"`
	BottomRight   Cell            `json:"bottom_right"`
	Config        json.RawMessage `json:"config" swaggertype:"object"`
	PhotoMediaID  *int            `json:"photo_media_id"`
	PhotoURL      *string         `json:"photo_url"`
}

// Area is where the widget sits, as its top-left cell and size
func (w ScreenWidget) Area() GridArea {
	return AreaFrom(w.TopLeft, w.BottomRight)
}

const selectWidget = `
	SELECT id, carepartner_id, type, col_start, row_start, col_end, row_end, config, photo_media_id
	FROM "ScreenWidget"
`

func scanWidget(row db.Scanner) (*ScreenWidget, error) {
	w := &ScreenWidget{}
	var config []byte
	var photo sql.NullInt64
	if err := row.Scan(&w.ID, &w.CarepartnerID, &w.Type,
		&w.TopLeft.Col, &w.TopLeft.Row, &w.BottomRight.Col, &w.BottomRight.Row,
		&config, &photo); err != nil {
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

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

// Create a widget; fills in ID and PhotoURL
func CreateScreenWidget(exec db.QueryRower, w *ScreenWidget) error {
	query := `
		INSERT INTO "ScreenWidget" (carepartner_id, type, col_start, row_start, col_end, row_end, config, photo_media_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id;
	`
	err := exec.QueryRow(query, w.CarepartnerID, w.Type,
		w.TopLeft.Col, w.TopLeft.Row, w.BottomRight.Col, w.BottomRight.Row,
		[]byte(w.Config), w.PhotoMediaID,
	).Scan(&w.ID)
	if err != nil {
		return err
	}
	w.setPhotoURL()
	return nil
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

func listWidgets(exec db.Querier, query string, carepartnerID int) ([]ScreenWidget, error) {
	rows, err := exec.Query(query, carepartnerID)
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

// List a carepartner's widgets in reading order: by top row, then left column
func ListScreenWidgets(exec db.Querier, carepartnerID int) ([]ScreenWidget, error) {
	return listWidgets(exec, selectWidget+` WHERE carepartner_id = $1 ORDER BY row_start, col_start, id;`, carepartnerID)
}

// List and lock a carepartner's widgets for the rest of the transaction
func ListScreenWidgetsForUpdate(tx *sql.Tx, carepartnerID int) ([]ScreenWidget, error) {
	return listWidgets(tx, selectWidget+` WHERE carepartner_id = $1 ORDER BY row_start, col_start, id FOR UPDATE;`, carepartnerID)
}

// Retrieve a media item for checking a widget's photo (its kind and owner), or nil
func GetPhoto(exec db.QueryRower, mediaID int) (*media.Media, error) {
	return media.GetMediaByID(exec, mediaID)
}

// ─────────────────────────────────────────────
// UPDATE / DELETE
// ─────────────────────────────────────────────

// Replace a widget's corners, settings and photo; its type doesn't change
func UpdateScreenWidget(exec db.Execer, w *ScreenWidget) error {
	_, err := exec.Exec(`
		UPDATE "ScreenWidget"
		SET col_start = $1, row_start = $2, col_end = $3, row_end = $4, config = $5, photo_media_id = $6
		WHERE id = $7;`,
		w.TopLeft.Col, w.TopLeft.Row, w.BottomRight.Col, w.BottomRight.Row,
		[]byte(w.Config), w.PhotoMediaID, w.ID,
	)
	if err == nil {
		w.setPhotoURL()
	}
	return err
}

// Delete every widget of the carepartner whose ID isn't in keep; an empty keep deletes them all
func DeleteScreenWidgetsExcept(exec db.Execer, carepartnerID int, keep []int) error {
	// Never nil: a nil array is sent as NULL, and id = ANY(NULL) would keep every widget
	ids := make(pq.Int64Array, 0, len(keep))
	for _, id := range keep {
		ids = append(ids, int64(id))
	}
	_, err := exec.Exec(
		`DELETE FROM "ScreenWidget" WHERE carepartner_id = $1 AND NOT (id = ANY($2));`,
		carepartnerID, ids,
	)
	return err
}
