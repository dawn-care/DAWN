package screen_widgets

import (
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"server/internal/utils"
)

var (
	qSettings  = regexp.QuoteMeta(`FROM "ScreenSettings"`)
	qMedia     = regexp.QuoteMeta(`FROM "Media" m`)
	qSlot      = regexp.QuoteMeta(`WHERE carepartner_id = $1 AND region = $2 AND slot = $3 FOR UPDATE;`)
	qByID      = regexp.QuoteMeta(`WHERE id = $1 FOR UPDATE;`)
	qInsert    = regexp.QuoteMeta(`INSERT INTO "ScreenWidget"`)
	qMove      = regexp.QuoteMeta(`UPDATE "ScreenWidget" SET region = $1, slot = $2 WHERE id = $3;`)
	qUpdate    = regexp.QuoteMeta(`UPDATE "ScreenWidget" SET config = $1, photo_media_id = $2 WHERE id = $3;`)
	qList      = regexp.QuoteMeta(`WHERE carepartner_id = $1 ORDER BY region, slot, id;`)
	qDelete    = regexp.QuoteMeta(`DELETE FROM "ScreenWidget" WHERE id = $1;`)
	widgetCols = []string{"id", "carepartner_id", "type", "region", "slot", "config", "photo_media_id"}
	mediaCols  = []string{"id", "carepartner_id", "kind", "display_name", "object_key", "mime_type",
		"size_bytes", "duration_seconds", "created_at", "on_screen", "audio_background", "scheduled"}
)

func intp(i int) *int { return &i }

func expectBalanced(mock sqlmock.Sqlmock, carepartnerID int) {
	mock.ExpectQuery(qSettings).WithArgs(carepartnerID).
		WillReturnRows(sqlmock.NewRows([]string{"carepartner_id", "layout_preset", "wake_time", "bed_time"}).
			AddRow(carepartnerID, "balanced", "07:00", "21:00"))
}

func expectMedia(mock sqlmock.Sqlmock, id, carepartnerID int, kind string) {
	mock.ExpectQuery(qMedia).WithArgs(id).
		WillReturnRows(sqlmock.NewRows(mediaCols).
			AddRow(id, carepartnerID, kind, "Sarah", "k", "image/png", 10, nil, time.Now(), false, false, false))
}

func expectEmptySlot(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(qSlot).WillReturnRows(sqlmock.NewRows(widgetCols))
}

func newService(t *testing.T) (*ScreenWidgetService, sqlmock.Sqlmock) {
	db, mock := utils.SetupMockDB(t)
	t.Cleanup(func() { db.Close() })
	return NewScreenWidgetService(db), mock
}

// ─────────────────────────────────────────────
// Create
// ─────────────────────────────────────────────

func TestCreateWidget_Photo(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	expectBalanced(mock, 1)
	expectMedia(mock, 4, 1, "photo")
	expectEmptySlot(mock)
	mock.ExpectQuery(qInsert).
		WithArgs(1, "photo", "photo_frame", 0, []byte(`{"caption":"Sarah"}`), 4).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))
	mock.ExpectCommit()

	w, err := s.CreateWidget(CreateInput{
		CarepartnerID: 1, Type: "photo", Region: "photo_frame", Slot: intp(0),
		Config: json.RawMessage(`{"caption":" Sarah "}`), PhotoMediaID: intp(4),
	})

	require.NoError(t, err)
	assert.Equal(t, 9, w.ID)
	assert.Equal(t, "/media/4/file", *w.PhotoURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWidget_WeatherDefaultZip(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	expectBalanced(mock, 1)
	expectEmptySlot(mock)
	mock.ExpectQuery(qInsert).
		WithArgs(1, "weather", "tiles", 2, []byte(`{"zip":"14617"}`), nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectCommit()

	w, err := s.CreateWidget(CreateInput{CarepartnerID: 1, Type: "weather", Region: "tiles", Slot: intp(2)})

	require.NoError(t, err)
	assert.JSONEq(t, `{"zip":"14617"}`, string(w.Config))
	assert.Nil(t, w.PhotoURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWidget_Placement(t *testing.T) {
	cases := []struct {
		name       string
		widgetType string
		region     string
		slot       int
		want       error
	}{
		{"region not in preset", "weather", "sidebar", 0, ErrUnknownSlot},
		{"slot past the end", "weather", "tiles", 3, ErrUnknownSlot},
		{"negative slot", "weather", "tiles", -1, ErrUnknownSlot},
		{"weather in the photo frame", "weather", "photo_frame", 0, ErrTypeNotAllowed},
		{"message in a tile", "message", "tiles", 0, ErrTypeNotAllowed},
		{"news on the plate", "news", "plate", 0, ErrTypeNotAllowed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock := newService(t)
			mock.ExpectBegin()
			expectBalanced(mock, 1)
			mock.ExpectRollback()

			config := json.RawMessage(`{}`)
			if tc.widgetType == "message" {
				config = json.RawMessage(`{"text":"hi"}`)
			}
			_, err := s.CreateWidget(CreateInput{CarepartnerID: 1, Type: tc.widgetType, Region: tc.region, Slot: intp(tc.slot), Config: config})

			assert.ErrorIs(t, err, tc.want)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateWidget_SlotTaken(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	expectBalanced(mock, 1)
	mock.ExpectQuery(qSlot).WithArgs(1, "tiles", 0).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(5, 1, "news", "tiles", 0, []byte(`{}`), nil))
	mock.ExpectRollback()

	_, err := s.CreateWidget(CreateInput{CarepartnerID: 1, Type: "weather", Region: "tiles", Slot: intp(0)})

	assert.ErrorIs(t, err, ErrSlotTaken)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWidget_LostRaceAtCommit(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	expectBalanced(mock, 1)
	expectEmptySlot(mock)
	mock.ExpectQuery(qInsert).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	// The unique constraint is deferred, so a concurrent insert fails at commit
	mock.ExpectCommit().WillReturnError(&pq.Error{Code: "23505"})

	_, err := s.CreateWidget(CreateInput{CarepartnerID: 1, Type: "weather", Region: "tiles", Slot: intp(0)})

	assert.ErrorIs(t, err, ErrSlotTaken)
}

func TestCreateWidget_Photos(t *testing.T) {
	cases := []struct {
		name       string
		widgetType string
		region     string
		config     string
		photo      *int
		mediaOwner int
		mediaKind  string
	}{
		{"photo widget without a photo", "photo", "photo_frame", `{}`, nil, 0, ""},
		{"photo of another carepartner", "photo", "photo_frame", `{}`, intp(4), 2, "photo"},
		{"video as a photo", "photo", "photo_frame", `{}`, intp(4), 1, "video"},
		{"message with a video", "message", "plate", `{"text":"hi"}`, intp(4), 1, "video"},
		{"weather with a photo", "weather", "tiles", `{}`, intp(4), 0, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock := newService(t)
			mock.ExpectBegin()
			expectBalanced(mock, 1)
			if tc.mediaKind != "" {
				expectMedia(mock, 4, tc.mediaOwner, tc.mediaKind)
			}
			mock.ExpectRollback()

			_, err := s.CreateWidget(CreateInput{CarepartnerID: 1, Type: tc.widgetType, Region: tc.region,
				Slot: intp(0), Config: json.RawMessage(tc.config), PhotoMediaID: tc.photo})

			assert.ErrorIs(t, err, ErrInvalidPhoto)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateWidget_InvalidInput(t *testing.T) {
	s, mock := newService(t)

	_, err := s.CreateWidget(CreateInput{CarepartnerID: 0, Type: "news", Region: "tiles", Slot: intp(0)})
	assert.ErrorIs(t, err, ErrInvalidCarepartner)
	_, err = s.CreateWidget(CreateInput{CarepartnerID: 1, Type: "clock", Region: "tiles", Slot: intp(0)})
	assert.ErrorIs(t, err, ErrInvalidType)
	_, err = s.CreateWidget(CreateInput{CarepartnerID: 1, Type: "news", Region: "tiles"})
	assert.ErrorIs(t, err, ErrUnknownSlot, "slot is required")
	_, err = s.CreateWidget(CreateInput{CarepartnerID: 1, Type: "weather", Region: "tiles", Slot: intp(0), Config: json.RawMessage(`{"zip":"abc"}`)})
	assert.ErrorIs(t, err, ErrInvalidConfig)
	assert.NoError(t, mock.ExpectationsWereMet(), "invalid input never reaches the database")
}

func TestCreateWidget_CarepartnerMissing(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectQuery(qSettings).WillReturnRows(sqlmock.NewRows([]string{"carepartner_id", "layout_preset", "wake_time", "bed_time"}))
	mock.ExpectRollback()

	_, err := s.CreateWidget(CreateInput{CarepartnerID: 99, Type: "news", Region: "tiles", Slot: intp(0)})
	assert.ErrorIs(t, err, ErrCarepartnerMissing)
}

// ─────────────────────────────────────────────
// List / Get
// ─────────────────────────────────────────────

func TestListWidgets_OrderAndHidden(t *testing.T) {
	s, mock := newService(t)
	expectBalanced(mock, 1)
	// The database orders by region name; the service orders by the preset's regions
	mock.ExpectQuery(qList).WithArgs(1).
		WillReturnRows(sqlmock.NewRows(widgetCols).
			AddRow(1, 1, "photo", "photo_frame", 0, []byte(`{}`), 4).
			AddRow(2, 1, "message", "plate", 0, []byte(`{"text":"Ana"}`), nil).
			AddRow(3, 1, "news", "sidebar", 0, []byte(`{}`), nil).
			AddRow(4, 1, "weather", "tiles", 1, []byte(`{"zip":"14617"}`), nil).
			AddRow(5, 1, "sports", "tiles", 0, []byte(`{"league":"mlb","team":"NYY"}`), nil).
			AddRow(6, 1, "reminder", "tiles", 5, []byte(`{"items":[]}`), nil))

	widgets, err := s.ListWidgets(1)

	require.NoError(t, err)
	var ids []int
	var hidden []bool
	for _, w := range widgets {
		ids = append(ids, w.ID)
		hidden = append(hidden, w.Hidden)
	}
	assert.Equal(t, []int{1, 5, 4, 2, 3, 6}, ids, "photo frame, tiles 0 and 1, plate, then hidden")
	assert.Equal(t, []bool{false, false, false, false, true, true}, hidden)
	assert.Equal(t, "/media/4/file", *widgets[0].PhotoURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListWidgets_Errors(t *testing.T) {
	s, mock := newService(t)
	_, err := s.ListWidgets(0)
	assert.ErrorIs(t, err, ErrInvalidCarepartner)

	mock.ExpectQuery(qSettings).WillReturnRows(sqlmock.NewRows([]string{"carepartner_id", "layout_preset", "wake_time", "bed_time"}))
	_, err = s.ListWidgets(99)
	assert.ErrorIs(t, err, ErrCarepartnerMissing)
}

func TestGetWidget(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectQuery(regexp.QuoteMeta(`WHERE id = $1;`)).WithArgs(3).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(3, 1, "news", "sidebar", 0, []byte(`{}`), nil))
	expectBalanced(mock, 1)

	w, err := s.GetWidget(3)
	require.NoError(t, err)
	assert.True(t, w.Hidden)

	mock.ExpectQuery(regexp.QuoteMeta(`WHERE id = $1;`)).WithArgs(8).WillReturnRows(sqlmock.NewRows(widgetCols))
	_, err = s.GetWidget(8)
	assert.ErrorIs(t, err, ErrNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ─────────────────────────────────────────────
// Update
// ─────────────────────────────────────────────

func TestUpdateWidget(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectQuery(qByID).WithArgs(4).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(4, 1, "weather", "tiles", 1, []byte(`{"zip":"14617"}`), nil))
	mock.ExpectExec(qUpdate).WithArgs([]byte(`{"zip":"10001"}`), nil, 4).WillReturnResult(sqlmock.NewResult(0, 1))
	expectBalanced(mock, 1)
	mock.ExpectCommit()

	w, err := s.UpdateWidget(4, UpdateInput{Config: json.RawMessage(`{"zip":"10001"}`)})

	require.NoError(t, err)
	assert.JSONEq(t, `{"zip":"10001"}`, string(w.Config))
	assert.Equal(t, "tiles", w.Region, "slot unchanged")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateWidget_Errors(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectQuery(qByID).WithArgs(4).WillReturnRows(sqlmock.NewRows(widgetCols))
	mock.ExpectRollback()
	_, err := s.UpdateWidget(4, UpdateInput{})
	assert.ErrorIs(t, err, ErrNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(qByID).WithArgs(4).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(4, 1, "message", "plate", 0, []byte(`{"text":"Ana"}`), nil))
	mock.ExpectRollback()
	_, err = s.UpdateWidget(4, UpdateInput{Config: json.RawMessage(`{}`)})
	assert.ErrorIs(t, err, ErrInvalidConfig, "message text is still required")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ─────────────────────────────────────────────
// Move
// ─────────────────────────────────────────────

func TestMoveWidget_ToEmptySlot(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectQuery(qByID).WithArgs(4).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(4, 1, "weather", "tiles", 0, []byte(`{}`), nil))
	expectBalanced(mock, 1)
	mock.ExpectQuery(qSlot).WithArgs(1, "tiles", 2).WillReturnRows(sqlmock.NewRows(widgetCols))
	mock.ExpectExec(qMove).WithArgs("tiles", 2, 4).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	w, err := s.MoveWidget(4, MoveInput{Region: "tiles", Slot: intp(2)})

	require.NoError(t, err)
	assert.Equal(t, 2, w.Slot)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMoveWidget_Swap(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectQuery(qByID).WithArgs(4).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(4, 1, "weather", "tiles", 0, []byte(`{}`), nil))
	expectBalanced(mock, 1)
	mock.ExpectQuery(qSlot).WithArgs(1, "tiles", 1).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(5, 1, "sports", "tiles", 1, []byte(`{}`), nil))
	mock.ExpectExec(qMove).WithArgs("tiles", 0, 5).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(qMove).WithArgs("tiles", 1, 4).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	w, err := s.MoveWidget(4, MoveInput{Region: "tiles", Slot: intp(1)})

	require.NoError(t, err)
	assert.Equal(t, 1, w.Slot)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMoveWidget_SwapNotPossible(t *testing.T) {
	// A hidden widget can't swap into a visible slot: the other widget would become hidden
	s, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectQuery(qByID).WithArgs(4).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(4, 1, "weather", "sidebar", 0, []byte(`{}`), nil))
	expectBalanced(mock, 1)
	mock.ExpectQuery(qSlot).WithArgs(1, "tiles", 1).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(5, 1, "sports", "tiles", 1, []byte(`{}`), nil))
	mock.ExpectRollback()

	_, err := s.MoveWidget(4, MoveInput{Region: "tiles", Slot: intp(1)})

	assert.ErrorIs(t, err, ErrSlotTaken)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMoveWidget_SameSlotAndBadTarget(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectQuery(qByID).WithArgs(4).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(4, 1, "weather", "tiles", 0, []byte(`{}`), nil))
	expectBalanced(mock, 1)
	mock.ExpectRollback()
	w, err := s.MoveWidget(4, MoveInput{Region: "tiles", Slot: intp(0)})
	require.NoError(t, err)
	assert.Equal(t, 0, w.Slot, "moving to its own slot is a no-op")

	mock.ExpectBegin()
	mock.ExpectQuery(qByID).WithArgs(4).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(4, 1, "weather", "tiles", 0, []byte(`{}`), nil))
	expectBalanced(mock, 1)
	mock.ExpectRollback()
	_, err = s.MoveWidget(4, MoveInput{Region: "plate", Slot: intp(0)})
	assert.ErrorIs(t, err, ErrTypeNotAllowed)

	_, err = s.MoveWidget(4, MoveInput{Region: "tiles"})
	assert.ErrorIs(t, err, ErrUnknownSlot)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ─────────────────────────────────────────────
// Delete
// ─────────────────────────────────────────────

func TestDeleteWidget(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectExec(qDelete).WithArgs(4).WillReturnResult(sqlmock.NewResult(0, 1))
	assert.NoError(t, s.DeleteWidget(4))

	mock.ExpectExec(qDelete).WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 0))
	assert.ErrorIs(t, s.DeleteWidget(5), ErrNotFound)

	mock.ExpectExec(qDelete).WithArgs(6).WillReturnError(errors.New("db down"))
	err := s.DeleteWidget(6)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}
