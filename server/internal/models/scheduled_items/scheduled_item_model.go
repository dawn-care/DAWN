package scheduled_items

import (
	"database/sql"

	"github.com/lib/pq"

	"server/internal/db"
	"server/internal/models/media"
)

// Repeat values, matching the CHECK constraint on "ScheduledItem".repeat
const (
	RepeatOnce    = "once"
	RepeatDaily   = "daily"
	RepeatWeekly  = "weekly"
	RepeatMonthly = "monthly"
)

// Shapes of a scheduled item, derived from its media kind and background photo
const (
	ShapeVideo               = "video"
	ShapeAudio               = "audio"
	ShapeAudioWithBackground = "audio_with_background"
	ShapePhoto               = "photo"
)

// MediaRef is the media a scheduled item plays, as the calendar and tablet need it
type MediaRef struct {
	ID              int    `json:"id"`
	Kind            string `json:"kind"`
	DisplayName     string `json:"display_name"`
	DurationSeconds *int   `json:"duration_seconds"`
	URL             string `json:"url"`
}

// PhotoRef is a background photo
type PhotoRef struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	URL         string `json:"url"`
}

/*
ScheduledItem plays a video, audio file or photo at a set time, once or on repeat.
Dates are "YYYY-MM-DD" and StartTime is "HH:MM", in the carepartner's local time.
Shape, Media, BackgroundPhoto and SkipDates are filled in when read. OutsideAwakeHours
is set by the service against the carepartner's current awake hours.
*/
type ScheduledItem struct {
	ID                int       `json:"id"`
	CarepartnerID     int       `json:"carepartner_id"`
	MediaID           int       `json:"media_id"`
	BackgroundPhotoID *int      `json:"background_photo_id"`
	DisplaySeconds    *int      `json:"display_seconds"`
	StartDate         string    `json:"start_date"`
	StartTime         string    `json:"start_time"`
	Repeat            string    `json:"repeat"`
	EndDate           *string   `json:"end_date"`
	Shape             string    `json:"shape"`
	Media             MediaRef  `json:"media"`
	BackgroundPhoto   *PhotoRef `json:"background_photo"`
	SkipDates         []string  `json:"skip_dates"`
	OutsideAwakeHours bool      `json:"outside_awake_hours"`
}

// Occurrence is one dated showing of a scheduled item, as the calendar lists it
type Occurrence struct {
	ScheduledItemID   int       `json:"scheduled_item_id"`
	Date              string    `json:"date"`
	Time              string    `json:"time"`
	Repeat            string    `json:"repeat"`
	Shape             string    `json:"shape"`
	Media             MediaRef  `json:"media"`
	BackgroundPhoto   *PhotoRef `json:"background_photo"`
	DisplaySeconds    *int      `json:"display_seconds"`
	DurationSeconds   int       `json:"duration_seconds"` // how long it stays on screen
	OutsideAwakeHours bool      `json:"outside_awake_hours"`
}

// DurationSeconds is how long the item stays on screen: the media's length, or a photo's display time
func (s *ScheduledItem) DurationSeconds() int {
	if s.DisplaySeconds != nil {
		return *s.DisplaySeconds
	}
	if s.Media.DurationSeconds != nil {
		return *s.Media.DurationSeconds
	}
	return 0
}

const selectItem = `
	SELECT s.id, s.carepartner_id, s.media_id, s.background_photo_id, s.display_seconds,
	       to_char(s.start_date, 'YYYY-MM-DD'), to_char(s.start_time, 'HH24:MI'), s.repeat,
	       to_char(s.end_date, 'YYYY-MM-DD'),
	       m.kind, m.display_name, m.duration_seconds, b.display_name
	FROM "ScheduledItem" s
	JOIN "Media" m ON m.id = s.media_id
	LEFT JOIN "Media" b ON b.id = s.background_photo_id
`

func scanItem(row db.Scanner) (*ScheduledItem, error) {
	s := &ScheduledItem{SkipDates: []string{}}
	var background, display, duration sql.NullInt64
	var endDate, backgroundName sql.NullString
	if err := row.Scan(
		&s.ID, &s.CarepartnerID, &s.MediaID, &background, &display,
		&s.StartDate, &s.StartTime, &s.Repeat, &endDate,
		&s.Media.Kind, &s.Media.DisplayName, &duration, &backgroundName,
	); err != nil {
		return nil, err
	}
	s.Media.ID = s.MediaID
	s.Media.URL = media.FileURL(s.MediaID)
	if duration.Valid {
		d := int(duration.Int64)
		s.Media.DurationSeconds = &d
	}
	if display.Valid {
		d := int(display.Int64)
		s.DisplaySeconds = &d
	}
	if endDate.Valid {
		s.EndDate = &endDate.String
	}
	if background.Valid {
		id := int(background.Int64)
		s.BackgroundPhotoID = &id
		s.BackgroundPhoto = &PhotoRef{ID: id, DisplayName: backgroundName.String, URL: media.FileURL(id)}
	}
	s.Shape = shapeOf(s.Media.Kind, s.BackgroundPhotoID != nil)
	return s, nil
}

func shapeOf(kind string, hasBackground bool) string {
	switch {
	case kind == media.KindAudio && hasBackground:
		return ShapeAudioWithBackground
	case kind == media.KindAudio:
		return ShapeAudio
	case kind == media.KindVideo:
		return ShapeVideo
	default:
		return ShapePhoto
	}
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

// Create a scheduled item; fills in ID
func CreateScheduledItem(exec db.QueryRower, s *ScheduledItem) error {
	query := `
		INSERT INTO "ScheduledItem"
			(carepartner_id, media_id, background_photo_id, display_seconds, start_date, start_time, repeat, end_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id;
	`
	return exec.QueryRow(query,
		s.CarepartnerID, s.MediaID, s.BackgroundPhotoID, s.DisplaySeconds,
		s.StartDate, s.StartTime, s.Repeat, s.EndDate,
	).Scan(&s.ID)
}

// Skip one date of a repeating item ("delete just this day"); skipping twice is harmless
func CreateSkip(exec db.Execer, itemID int, date string) error {
	_, err := exec.Exec(
		`INSERT INTO "ScheduledItemSkip" (scheduled_item_id, skip_date) VALUES ($1, $2) ON CONFLICT DO NOTHING;`,
		itemID, date,
	)
	return err
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

// Retrieve one item with its skip dates, or nil if it doesn't exist
func GetScheduledItemByID(db *sql.DB, id int) (*ScheduledItem, error) {
	s, err := scanItem(db.QueryRow(selectItem+` WHERE s.id = $1;`, id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	skips, err := ListSkips(db, []int{id})
	if err != nil {
		return nil, err
	}
	if dates, ok := skips[id]; ok {
		s.SkipDates = dates
	}
	return s, nil
}

/*
List a carepartner's items with their skip dates. With from and to ("YYYY-MM-DD"),
only items that could occur in that range are returned.
*/
func ListScheduledItems(db *sql.DB, carepartnerID int, from, to string) ([]ScheduledItem, error) {
	query := selectItem + ` WHERE s.carepartner_id = $1`
	args := []any{carepartnerID}
	if from != "" && to != "" {
		query += ` AND s.start_date <= $2 AND (s.end_date IS NULL OR s.end_date >= $3)`
		args = append(args, to, from)
	}
	rows, err := db.Query(query+` ORDER BY s.start_time, s.id;`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []ScheduledItem{}
	var ids []int
	for rows.Next() {
		s, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *s)
		ids = append(ids, s.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	skips, err := ListSkips(db, ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if dates, ok := skips[items[i].ID]; ok {
			items[i].SkipDates = dates
		}
	}
	return items, nil
}

// ListSkips returns the skip dates of each item, in date order
func ListSkips(db *sql.DB, itemIDs []int) (map[int][]string, error) {
	skips := map[int][]string{}
	if len(itemIDs) == 0 {
		return skips, nil
	}
	rows, err := db.Query(
		`SELECT scheduled_item_id, to_char(skip_date, 'YYYY-MM-DD') FROM "ScheduledItemSkip"
		 WHERE scheduled_item_id = ANY($1) ORDER BY skip_date;`,
		pq.Array(itemIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var date string
		if err := rows.Scan(&id, &date); err != nil {
			return nil, err
		}
		skips[id] = append(skips[id], date)
	}
	return skips, rows.Err()
}

// ─────────────────────────────────────────────
// UPDATE / DELETE
// ─────────────────────────────────────────────

// Replace an item's media, timing and repeat; returns false if no such item
func UpdateScheduledItem(exec db.Execer, s *ScheduledItem) (bool, error) {
	res, err := exec.Exec(`
		UPDATE "ScheduledItem"
		SET media_id = $1, background_photo_id = $2, display_seconds = $3,
		    start_date = $4, start_time = $5, repeat = $6, end_date = $7
		WHERE id = $8;
	`, s.MediaID, s.BackgroundPhotoID, s.DisplaySeconds, s.StartDate, s.StartTime, s.Repeat, s.EndDate, s.ID)
	if err != nil {
		return false, err
	}
	count, err := res.RowsAffected()
	return count > 0, err
}

// Delete an item and its skip dates; returns false if no such item
func DeleteScheduledItem(exec db.Execer, id int) (bool, error) {
	res, err := exec.Exec(`DELETE FROM "ScheduledItem" WHERE id = $1;`, id)
	if err != nil {
		return false, err
	}
	count, err := res.RowsAffected()
	return count > 0, err
}
