package screen_widgets

import (
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	model "server/internal/models/screen_widgets"
	"server/internal/utils"
)

var (
	qSettings   = regexp.QuoteMeta(`FROM "ScreenSettings"`)
	qLock       = regexp.QuoteMeta(`SELECT carepartner_id FROM "ScreenSettings" WHERE carepartner_id = $1 FOR UPDATE;`)
	qList       = regexp.QuoteMeta(`WHERE carepartner_id = $1 ORDER BY row_start, col_start, id;`)
	qListLocked = regexp.QuoteMeta(`WHERE carepartner_id = $1 ORDER BY row_start, col_start, id FOR UPDATE;`)
	qMedia      = regexp.QuoteMeta(`FROM "Media" m`)
	qDelete     = regexp.QuoteMeta(`DELETE FROM "ScreenWidget" WHERE carepartner_id = $1 AND NOT (id = ANY($2));`)
	qUpdate     = regexp.QuoteMeta(`UPDATE "ScreenWidget"`)
	qInsert     = regexp.QuoteMeta(`INSERT INTO "ScreenWidget"`)
	widgetCols  = []string{"id", "carepartner_id", "type", "col_start", "row_start", "col_end", "row_end", "config", "photo_media_id"}
	mediaCols   = []string{"id", "carepartner_id", "kind", "display_name", "object_key", "mime_type",
		"size_bytes", "duration_seconds", "created_at", "on_screen", "audio_background", "scheduled"}
)

func intp(i int) *int { return &i }

func newService(t *testing.T) (*ScreenWidgetService, sqlmock.Sqlmock) {
	db, mock := utils.SetupMockDB(t)
	t.Cleanup(func() { db.Close() })
	return NewScreenWidgetService(db), mock
}

func expectMedia(mock sqlmock.Sqlmock, id, carepartnerID int, kind string) {
	mock.ExpectQuery(qMedia).WithArgs(id).
		WillReturnRows(sqlmock.NewRows(mediaCols).
			AddRow(id, carepartnerID, kind, "Sarah", "k", "image/png", 10, nil, time.Now(), false, false, false))
}

// The carepartner's stored widgets: a 2x2 photo, a weather tile and a news tile
func storedRows() *sqlmock.Rows {
	return sqlmock.NewRows(widgetCols).
		AddRow(1, 1, "photo", 0, 0, 1, 1, []byte(`{"caption":"Sarah"}`), 4).
		AddRow(2, 1, "weather", 2, 0, 2, 0, []byte(`{"zip":"14617"}`), nil).
		AddRow(3, 1, "news", 3, 0, 3, 0, []byte(`{}`), nil)
}

// expectSaveStart expects the transaction, the settings lock and the locked widget list
func expectSaveStart(mock sqlmock.Sqlmock, rows *sqlmock.Rows) {
	mock.ExpectBegin()
	mock.ExpectQuery(qLock).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"carepartner_id"}).AddRow(1))
	mock.ExpectQuery(qListLocked).WithArgs(1).WillReturnRows(rows)
}

func widget(id *int, widgetType string, top, left, bottom, right int, config string) model.WidgetInput {
	return model.WidgetInput{
		ID: id, Type: widgetType,
		TopLeft:     &model.Cell{Col: left, Row: top},
		BottomRight: &model.Cell{Col: right, Row: bottom},
		Config:      json.RawMessage(config),
	}
}

func layout(widgets ...model.WidgetInput) model.LayoutInput {
	if widgets == nil {
		widgets = []model.WidgetInput{}
	}
	return model.LayoutInput{Widgets: &widgets}
}

// ─────────────────────────────────────────────
// GetLayout
// ─────────────────────────────────────────────

func TestGetLayout(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectQuery(qSettings).WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"carepartner_id", "wake_time", "bed_time"}).
			AddRow(1, "07:00", "21:00"))
	mock.ExpectQuery(qList).WithArgs(1).WillReturnRows(storedRows())

	l, err := s.GetLayout(1)

	require.NoError(t, err)
	require.Len(t, l.Widgets, 3)
	photo := l.Widgets[0]
	assert.Equal(t, model.Cell{Col: 0, Row: 0}, photo.TopLeft)
	assert.Equal(t, model.Cell{Col: 1, Row: 1}, photo.BottomRight)
	assert.Equal(t, model.GridArea{Col: 0, Row: 0, Width: 2, Height: 2}, photo.Area())
	assert.Equal(t, "/media/4/file", *photo.PhotoURL)
	assert.NoError(t, mock.ExpectationsWereMet())

	raw, err := json.Marshal(l)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"top_left":{"col":0,"row":0},"bottom_right":{"col":1,"row":1}`)
}

func TestGetLayout_Errors(t *testing.T) {
	s, mock := newService(t)
	_, err := s.GetLayout(0)
	assert.ErrorIs(t, err, ErrInvalidCarepartner)

	mock.ExpectQuery(qSettings).WithArgs(9).WillReturnRows(sqlmock.NewRows([]string{"carepartner_id", "wake_time", "bed_time"}))
	_, err = s.GetLayout(9)
	assert.ErrorIs(t, err, ErrCarepartnerMissing)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ─────────────────────────────────────────────
// SaveLayout
// ─────────────────────────────────────────────

func TestSaveLayout_CreatesUpdatesAndDeletes(t *testing.T) {
	// Keep the photo but make it 1x2, move the weather tile, drop the news tile, add a
	// 2x1 sports widget that touches the weather tile edge to edge
	s, mock := newService(t)
	expectSaveStart(mock, storedRows())
	expectMedia(mock, 4, 1, "photo")
	mock.ExpectExec(qDelete).WithArgs(1, "{1,2}").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(qUpdate).WithArgs(0, 0, 0, 1, []byte(`{"caption":"Sarah"}`), 4, 1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(qUpdate).WithArgs(1, 0, 1, 0, []byte(`{"zip":"10001"}`), nil, 2).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(qInsert).WithArgs(1, "sports", 2, 0, 3, 0, []byte(`{"league":"mlb","team":"NYY"}`), nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))
	mock.ExpectCommit()

	photo := widget(intp(1), "photo", 0, 0, 1, 0, `{"caption":" Sarah "}`)
	photo.PhotoMediaID = intp(4)
	l, err := s.SaveLayout(1, layout(
		photo,
		widget(intp(2), "weather", 0, 1, 0, 1, `{"zip":"10001"}`),
		widget(nil, "sports", 0, 2, 0, 3, `{"league":"mlb","team":"nyy"}`),
	))

	require.NoError(t, err)
	require.Len(t, l.Widgets, 3)
	assert.Equal(t, []int{1, 2, 9}, []int{l.Widgets[0].ID, l.Widgets[1].ID, l.Widgets[2].ID}, "in the order sent, with the new widget's id")
	assert.Equal(t, model.GridArea{Col: 0, Row: 0, Width: 1, Height: 2}, l.Widgets[0].Area())
	assert.Equal(t, model.GridArea{Col: 2, Row: 0, Width: 2, Height: 1}, l.Widgets[2].Area())
	assert.Equal(t, "/media/4/file", *l.Widgets[0].PhotoURL)
	assert.JSONEq(t, `{"league":"mlb","team":"NYY"}`, string(l.Widgets[2].Config), "config is normalized")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveLayout_EmptyListClearsTheScreen(t *testing.T) {
	s, mock := newService(t)
	expectSaveStart(mock, storedRows())
	mock.ExpectExec(qDelete).WithArgs(1, "{}").WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	l, err := s.SaveLayout(1, layout())

	require.NoError(t, err)
	assert.Empty(t, l.Widgets)
	raw, _ := json.Marshal(l)
	assert.JSONEq(t, `{"widgets":[]}`, string(raw))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveLayout_WidgetsRequired(t *testing.T) {
	// A missing key or null must never be read as "clear the screen"
	s, mock := newService(t)

	_, err := s.SaveLayout(1, model.LayoutInput{})

	assert.ErrorIs(t, err, ErrInvalidRequest)
	assert.Contains(t, err.Error(), "send [] to clear the screen")
	assert.NoError(t, mock.ExpectationsWereMet(), "nothing reaches the database")
}

func TestSaveLayout_InvalidWidgets(t *testing.T) {
	missingCorners := widget(nil, "news", 0, 0, 0, 0, `{}`)
	missingCorners.BottomRight = nil
	weatherWithPhoto := widget(nil, "weather", 0, 0, 0, 0, `{}`)
	weatherWithPhoto.PhotoMediaID = intp(4)
	ok := widget(nil, "news", 2, 3, 2, 3, `{}`)

	cases := []struct {
		name    string
		widgets []model.WidgetInput
		want    error
		message string
	}{
		{"unknown type", []model.WidgetInput{ok, widget(nil, "clock", 0, 0, 0, 0, `{}`)}, ErrInvalidType, "widgets[1]"},
		{"id of another carepartner's widget", []model.WidgetInput{widget(intp(77), "news", 0, 0, 0, 0, `{}`)}, ErrNotFound, "widget 77 isn't one of this carepartner's widgets"},
		{"id listed twice", []model.WidgetInput{widget(intp(3), "news", 0, 0, 0, 0, `{}`), widget(intp(3), "news", 1, 0, 1, 0, `{}`)}, ErrInvalidRequest, "widgets[1]: invalid request: widget 3 is listed more than once"},
		{"type changed under the same id", []model.WidgetInput{widget(intp(2), "news", 0, 0, 0, 0, `{}`)}, ErrInvalidType, "widget 2 is a weather widget"},
		{"missing corner", []model.WidgetInput{missingCorners}, ErrInvalidPosition, "widgets[0]"},
		{"column 4 is off the grid", []model.WidgetInput{widget(nil, "news", 0, 3, 0, 4, `{}`)}, ErrInvalidPosition, "widgets[0]"},
		{"row 3 is off the grid", []model.WidgetInput{widget(nil, "news", 2, 0, 3, 0, `{}`)}, ErrInvalidPosition, "widgets[0]"},
		{"corners reversed", []model.WidgetInput{widget(nil, "news", 1, 1, 0, 0, `{}`)}, ErrInvalidPosition, "widgets[0]"},
		{"bad config", []model.WidgetInput{ok, widget(nil, "weather", 0, 0, 0, 0, `{"zip":"abc"}`)}, ErrInvalidConfig, "widgets[1]"},
		{"photo on a weather widget", []model.WidgetInput{weatherWithPhoto}, ErrInvalidPhoto, "widgets[0]"},
		{"a 1x1 inside a 2x2", []model.WidgetInput{widget(nil, "news", 0, 0, 1, 1, `{}`), ok, widget(nil, "weather", 1, 1, 1, 1, `{}`)}, ErrOverlap, "widgets[0] and widgets[2] cover the same cell"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock := newService(t)
			expectSaveStart(mock, storedRows())
			mock.ExpectRollback()

			_, err := s.SaveLayout(1, layout(tc.widgets...))

			assert.ErrorIs(t, err, tc.want)
			assert.Contains(t, err.Error(), tc.message)
			assert.NoError(t, mock.ExpectationsWereMet(), "nothing is written")
		})
	}
}

func TestSaveLayout_PhotoChecks(t *testing.T) {
	cases := []struct {
		name       string
		mediaOwner int
		mediaKind  string
	}{
		{"photo of another carepartner", 2, "photo"},
		{"video as a photo", 1, "video"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock := newService(t)
			expectSaveStart(mock, sqlmock.NewRows(widgetCols))
			expectMedia(mock, 4, tc.mediaOwner, tc.mediaKind)
			mock.ExpectRollback()

			photo := widget(nil, "photo", 0, 0, 1, 1, `{}`)
			photo.PhotoMediaID = intp(4)
			_, err := s.SaveLayout(1, layout(photo))

			assert.ErrorIs(t, err, ErrInvalidPhoto)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSaveLayout_Errors(t *testing.T) {
	s, mock := newService(t)
	_, err := s.SaveLayout(0, layout())
	assert.ErrorIs(t, err, ErrInvalidCarepartner)

	mock.ExpectBegin()
	mock.ExpectQuery(qLock).WithArgs(9).WillReturnRows(sqlmock.NewRows([]string{"carepartner_id"}))
	mock.ExpectRollback()
	_, err = s.SaveLayout(9, layout())
	assert.ErrorIs(t, err, ErrCarepartnerMissing)

	expectSaveStart(mock, sqlmock.NewRows(widgetCols))
	mock.ExpectExec(qDelete).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(qInsert).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	mock.ExpectCommit().WillReturnError(errors.New("connection lost"))
	_, err = s.SaveLayout(1, layout(widget(nil, "news", 0, 0, 0, 0, `{}`)))
	assert.EqualError(t, err, "connection lost")
	assert.NoError(t, mock.ExpectationsWereMet())
}
