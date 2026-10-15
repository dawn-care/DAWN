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
	qSpot      = regexp.QuoteMeta(`WHERE carepartner_id = $1 AND spot = $2 FOR UPDATE;`)
	qByID      = regexp.QuoteMeta(`WHERE id = $1 FOR UPDATE;`)
	qInsert    = regexp.QuoteMeta(`INSERT INTO "ScreenWidget"`)
	qMove      = regexp.QuoteMeta(`UPDATE "ScreenWidget" SET spot = $1 WHERE id = $2;`)
	qUpdate    = regexp.QuoteMeta(`UPDATE "ScreenWidget" SET config = $1, photo_media_id = $2 WHERE id = $3;`)
	qList      = regexp.QuoteMeta(`WHERE carepartner_id = $1 ORDER BY spot, id;`)
	qDelete    = regexp.QuoteMeta(`DELETE FROM "ScreenWidget" WHERE id = $1;`)
	widgetCols = []string{"id", "carepartner_id", "type", "spot", "config", "photo_media_id"}
	mediaCols  = []string{"id", "carepartner_id", "kind", "display_name", "object_key", "mime_type",
		"size_bytes", "duration_seconds", "created_at", "on_screen", "audio_background", "scheduled"}
)

func intp(i int) *int { return &i }

func expectPreset(mock sqlmock.Sqlmock, carepartnerID int, preset string) {
	mock.ExpectQuery(qSettings).WithArgs(carepartnerID).
		WillReturnRows(sqlmock.NewRows([]string{"carepartner_id", "layout_preset", "wake_time", "bed_time"}).
			AddRow(carepartnerID, preset, "07:00", "21:00"))
}

// Balanced has 4 spots, 0 to 3
func expectBalanced(mock sqlmock.Sqlmock, carepartnerID int) {
	expectPreset(mock, carepartnerID, "balanced")
}

func expectMedia(mock sqlmock.Sqlmock, id, carepartnerID int, kind string) {
	mock.ExpectQuery(qMedia).WithArgs(id).
		WillReturnRows(sqlmock.NewRows(mediaCols).
			AddRow(id, carepartnerID, kind, "Sarah", "k", "image/png", 10, nil, time.Now(), false, false, false))
}

func expectEmptySpot(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(qSpot).WillReturnRows(sqlmock.NewRows(widgetCols))
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
	expectEmptySpot(mock)
	mock.ExpectQuery(qInsert).
		WithArgs(1, "photo", 0, []byte(`{"caption":"Sarah"}`), 4).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))
	mock.ExpectCommit()

	w, err := s.CreateWidget(CreateInput{
		CarepartnerID: 1, Type: "photo", Spot: intp(0),
		Config: json.RawMessage(`{"caption":" Sarah "}`), PhotoMediaID: intp(4),
	})

	require.NoError(t, err)
	assert.Equal(t, 9, w.ID)
	assert.Equal(t, 0, w.Spot)
	assert.Equal(t, "/media/4/file", *w.PhotoURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWidget_WeatherDefaultZip(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	expectBalanced(mock, 1)
	expectEmptySpot(mock)
	mock.ExpectQuery(qInsert).
		WithArgs(1, "weather", 2, []byte(`{"zip":"14617"}`), nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectCommit()

	w, err := s.CreateWidget(CreateInput{CarepartnerID: 1, Type: "weather", Spot: intp(2)})

	require.NoError(t, err)
	assert.JSONEq(t, `{"zip":"14617"}`, string(w.Config))
	assert.Nil(t, w.PhotoURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWidget_AnyTypeInAnySpot(t *testing.T) {
	cases := []struct {
		widgetType string
		spot       int
		config     string
	}{
		{"weather", 0, `{"zip":"14617"}`},
		{"message", 1, `{"text":"hi"}`},
		{"news", 0, `{}`},
		{"reminder", 3, `{"items":[{"time":"09:00","text":"Pills"}]}`},
	}
	for _, tc := range cases {
		t.Run(tc.widgetType, func(t *testing.T) {
			s, mock := newService(t)
			mock.ExpectBegin()
			expectBalanced(mock, 1)
			expectEmptySpot(mock)
			mock.ExpectQuery(qInsert).
				WithArgs(1, tc.widgetType, tc.spot, sqlmock.AnyArg(), nil).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
			mock.ExpectCommit()

			w, err := s.CreateWidget(CreateInput{CarepartnerID: 1, Type: tc.widgetType, Spot: intp(tc.spot), Config: json.RawMessage(tc.config)})

			require.NoError(t, err)
			assert.Equal(t, tc.spot, w.Spot)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateWidget_UnknownSpot(t *testing.T) {
	cases := []struct {
		name   string
		preset string
		spot   int
	}{
		{"past the end of balanced", "balanced", 4},
		{"past the end of simple", "simple", 2},
		{"negative", "busier", -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock := newService(t)
			mock.ExpectBegin()
			expectPreset(mock, 1, tc.preset)
			mock.ExpectRollback()

			_, err := s.CreateWidget(CreateInput{CarepartnerID: 1, Type: "news", Spot: intp(tc.spot)})

			assert.ErrorIs(t, err, ErrUnknownSpot)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateWidget_SpotTaken(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	expectBalanced(mock, 1)
	mock.ExpectQuery(qSpot).WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(5, 1, "news", 1, []byte(`{}`), nil))
	mock.ExpectRollback()

	_, err := s.CreateWidget(CreateInput{CarepartnerID: 1, Type: "weather", Spot: intp(1)})

	assert.ErrorIs(t, err, ErrSpotTaken)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWidget_LostRaceAtCommit(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	expectBalanced(mock, 1)
	expectEmptySpot(mock)
	mock.ExpectQuery(qInsert).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	// The unique constraint is deferred, so a concurrent insert fails at commit
	mock.ExpectCommit().WillReturnError(&pq.Error{Code: "23505"})

	_, err := s.CreateWidget(CreateInput{CarepartnerID: 1, Type: "weather", Spot: intp(0)})

	assert.ErrorIs(t, err, ErrSpotTaken)
}

func TestCreateWidget_Photos(t *testing.T) {
	cases := []struct {
		name       string
		widgetType string
		config     string
		photo      *int
		mediaOwner int
		mediaKind  string
	}{
		{"photo widget without a photo", "photo", `{}`, nil, 0, ""},
		{"photo of another carepartner", "photo", `{}`, intp(4), 2, "photo"},
		{"video as a photo", "photo", `{}`, intp(4), 1, "video"},
		{"message with a video", "message", `{"text":"hi"}`, intp(4), 1, "video"},
		{"weather with a photo", "weather", `{}`, intp(4), 0, ""},
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

			_, err := s.CreateWidget(CreateInput{CarepartnerID: 1, Type: tc.widgetType,
				Spot: intp(0), Config: json.RawMessage(tc.config), PhotoMediaID: tc.photo})

			assert.ErrorIs(t, err, ErrInvalidPhoto)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateWidget_InvalidInput(t *testing.T) {
	s, mock := newService(t)

	_, err := s.CreateWidget(CreateInput{CarepartnerID: 0, Type: "news", Spot: intp(0)})
	assert.ErrorIs(t, err, ErrInvalidCarepartner)
	_, err = s.CreateWidget(CreateInput{CarepartnerID: 1, Type: "clock", Spot: intp(0)})
	assert.ErrorIs(t, err, ErrInvalidType)
	_, err = s.CreateWidget(CreateInput{CarepartnerID: 1, Type: "news"})
	assert.ErrorIs(t, err, ErrUnknownSpot, "spot is required")
	_, err = s.CreateWidget(CreateInput{CarepartnerID: 1, Type: "weather", Spot: intp(0), Config: json.RawMessage(`{"zip":"abc"}`)})
	assert.ErrorIs(t, err, ErrInvalidConfig)
	assert.NoError(t, mock.ExpectationsWereMet(), "invalid input never reaches the database")
}

func TestCreateWidget_CarepartnerMissing(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectQuery(qSettings).WillReturnRows(sqlmock.NewRows([]string{"carepartner_id", "layout_preset", "wake_time", "bed_time"}))
	mock.ExpectRollback()

	_, err := s.CreateWidget(CreateInput{CarepartnerID: 99, Type: "news", Spot: intp(0)})
	assert.ErrorIs(t, err, ErrCarepartnerMissing)
}

// ─────────────────────────────────────────────
// List / Get
// ─────────────────────────────────────────────

func expectListRows(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(qList).WithArgs(1).
		WillReturnRows(sqlmock.NewRows(widgetCols).
			AddRow(1, 1, "photo", 0, []byte(`{}`), 4).
			AddRow(2, 1, "weather", 1, []byte(`{"zip":"14617"}`), nil).
			AddRow(3, 1, "message", 3, []byte(`{"text":"Ana"}`), nil).
			AddRow(4, 1, "news", 5, []byte(`{}`), nil))
}

func TestListWidgets_OrderAndHidden(t *testing.T) {
	s, mock := newService(t)
	expectBalanced(mock, 1)
	expectListRows(mock)

	widgets, err := s.ListWidgets(1)

	require.NoError(t, err)
	var ids []int
	var hidden []bool
	for _, w := range widgets {
		ids = append(ids, w.ID)
		hidden = append(hidden, w.Hidden)
	}
	assert.Equal(t, []int{1, 2, 3, 4}, ids, "spot order, so hidden widgets come last")
	assert.Equal(t, []bool{false, false, false, true}, hidden, "balanced has spots 0 to 3")
	assert.Equal(t, "/media/4/file", *widgets[0].PhotoURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListWidgets_PresetChangesWhatIsHidden(t *testing.T) {
	cases := []struct {
		preset string
		hidden []bool
	}{
		{"simple", []bool{false, false, true, true}},
		{"grid", []bool{false, false, false, true}},
		{"busier", []bool{false, false, false, false}},
	}
	for _, tc := range cases {
		t.Run(tc.preset, func(t *testing.T) {
			s, mock := newService(t)
			expectPreset(mock, 1, tc.preset)
			expectListRows(mock)

			widgets, err := s.ListWidgets(1)

			require.NoError(t, err)
			var hidden []bool
			for _, w := range widgets {
				hidden = append(hidden, w.Hidden)
			}
			assert.Equal(t, tc.hidden, hidden)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
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
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(3, 1, "news", 5, []byte(`{}`), nil))
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
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(4, 1, "weather", 1, []byte(`{"zip":"14617"}`), nil))
	mock.ExpectExec(qUpdate).WithArgs([]byte(`{"zip":"10001"}`), nil, 4).WillReturnResult(sqlmock.NewResult(0, 1))
	expectBalanced(mock, 1)
	mock.ExpectCommit()

	w, err := s.UpdateWidget(4, UpdateInput{Config: json.RawMessage(`{"zip":"10001"}`)})

	require.NoError(t, err)
	assert.JSONEq(t, `{"zip":"10001"}`, string(w.Config))
	assert.Equal(t, 1, w.Spot, "spot unchanged")
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
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(4, 1, "message", 3, []byte(`{"text":"Ana"}`), nil))
	mock.ExpectRollback()
	_, err = s.UpdateWidget(4, UpdateInput{Config: json.RawMessage(`{}`)})
	assert.ErrorIs(t, err, ErrInvalidConfig, "message text is still required")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ─────────────────────────────────────────────
// Move
// ─────────────────────────────────────────────

func TestMoveWidget_ToEmptySpot(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectQuery(qByID).WithArgs(4).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(4, 1, "weather", 1, []byte(`{}`), nil))
	expectBalanced(mock, 1)
	mock.ExpectQuery(qSpot).WithArgs(1, 3).WillReturnRows(sqlmock.NewRows(widgetCols))
	mock.ExpectExec(qMove).WithArgs(3, 4).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	w, err := s.MoveWidget(4, MoveInput{Spot: intp(3)})

	require.NoError(t, err)
	assert.Equal(t, 3, w.Spot)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMoveWidget_Swap(t *testing.T) {
	// Any two widgets can swap: a weather widget can take the photo's main spot
	s, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectQuery(qByID).WithArgs(4).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(4, 1, "weather", 2, []byte(`{}`), nil))
	expectBalanced(mock, 1)
	mock.ExpectQuery(qSpot).WithArgs(1, 0).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(5, 1, "photo", 0, []byte(`{}`), 7))
	mock.ExpectExec(qMove).WithArgs(2, 5).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(qMove).WithArgs(0, 4).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	w, err := s.MoveWidget(4, MoveInput{Spot: intp(0)})

	require.NoError(t, err)
	assert.Equal(t, 0, w.Spot)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMoveWidget_HiddenWidgetSwapsIntoView(t *testing.T) {
	// The widget in the target spot takes the hidden spot and becomes hidden
	s, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectQuery(qByID).WithArgs(4).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(4, 1, "weather", 5, []byte(`{}`), nil))
	expectBalanced(mock, 1)
	mock.ExpectQuery(qSpot).WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(5, 1, "sports", 1, []byte(`{}`), nil))
	mock.ExpectExec(qMove).WithArgs(5, 5).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(qMove).WithArgs(1, 4).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	w, err := s.MoveWidget(4, MoveInput{Spot: intp(1)})

	require.NoError(t, err)
	assert.Equal(t, 1, w.Spot)
	assert.False(t, w.Hidden)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMoveWidget_SameSpotAndBadTarget(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectBegin()
	mock.ExpectQuery(qByID).WithArgs(4).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(4, 1, "weather", 1, []byte(`{}`), nil))
	expectBalanced(mock, 1)
	mock.ExpectRollback()
	w, err := s.MoveWidget(4, MoveInput{Spot: intp(1)})
	require.NoError(t, err)
	assert.Equal(t, 1, w.Spot, "moving to its own spot is a no-op")

	mock.ExpectBegin()
	mock.ExpectQuery(qByID).WithArgs(4).
		WillReturnRows(sqlmock.NewRows(widgetCols).AddRow(4, 1, "weather", 1, []byte(`{}`), nil))
	expectBalanced(mock, 1)
	mock.ExpectRollback()
	_, err = s.MoveWidget(4, MoveInput{Spot: intp(4)})
	assert.ErrorIs(t, err, ErrUnknownSpot, "balanced has no spot 4")

	_, err = s.MoveWidget(4, MoveInput{})
	assert.ErrorIs(t, err, ErrUnknownSpot, "spot is required")
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
