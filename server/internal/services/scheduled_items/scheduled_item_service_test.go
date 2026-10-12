package scheduled_items

import (
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
	qMedia    = regexp.QuoteMeta(`FROM "Media" m`)
	qSettings = regexp.QuoteMeta(`FROM "ScreenSettings"`)
	qInsert   = regexp.QuoteMeta(`INSERT INTO "ScheduledItem"`)
	qGetItem  = regexp.QuoteMeta(`WHERE s.id = $1;`)
	qList     = regexp.QuoteMeta(`WHERE s.carepartner_id = $1`)
	qSkips    = regexp.QuoteMeta(`FROM "ScheduledItemSkip"`)
	qAddSkip  = regexp.QuoteMeta(`INSERT INTO "ScheduledItemSkip"`)
	qUpdate   = regexp.QuoteMeta(`UPDATE "ScheduledItem"`)
	qDelete   = regexp.QuoteMeta(`DELETE FROM "ScheduledItem" WHERE id = $1;`)

	mediaCols = []string{"id", "carepartner_id", "kind", "display_name", "object_key", "mime_type",
		"size_bytes", "duration_seconds", "created_at", "on_screen", "audio_background", "scheduled"}
	itemCols = []string{"id", "carepartner_id", "media_id", "background_photo_id", "display_seconds",
		"start_date", "start_time", "repeat", "end_date", "kind", "display_name", "duration_seconds", "bg_name"}
)

func intp(i int) *int       { return &i }
func strp(s string) *string { return &s }

func newService(t *testing.T) (*ScheduledItemService, sqlmock.Sqlmock) {
	db, mock := utils.SetupMockDB(t)
	t.Cleanup(func() { db.Close() })
	return NewScheduledItemService(db), mock
}

// expectMedia: a media item; duration is nil for photos
func expectMedia(mock sqlmock.Sqlmock, id, carepartnerID int, kind string, duration any) {
	mock.ExpectQuery(qMedia).WithArgs(id).
		WillReturnRows(sqlmock.NewRows(mediaCols).
			AddRow(id, carepartnerID, kind, "name", "k", "x/x", 1, duration, time.Now(), false, false, false))
}

func expectSettings(mock sqlmock.Sqlmock, wake, bed string) {
	mock.ExpectQuery(qSettings).WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"carepartner_id", "layout_preset", "wake_time", "bed_time"}).
			AddRow(1, "balanced", wake, bed))
}

func itemRow(id, mediaID int, bg, display any, start, at, repeat string, end any, kind string, duration any) *sqlmock.Rows {
	return sqlmock.NewRows(itemCols).AddRow(id, 1, mediaID, bg, display, start, at, repeat, end, kind, "name", duration, "Sarah")
}

func expectGetItem(mock sqlmock.Sqlmock, rows *sqlmock.Rows, skips ...string) {
	mock.ExpectQuery(qGetItem).WillReturnRows(rows)
	skipRows := sqlmock.NewRows([]string{"scheduled_item_id", "skip_date"})
	for _, s := range skips {
		skipRows.AddRow(5, s)
	}
	mock.ExpectQuery(qSkips).WithArgs(sqlmock.AnyArg()).WillReturnRows(skipRows)
}

func videoInput() ItemInput {
	return ItemInput{CarepartnerID: 1, MediaID: 2, StartDate: "2026-10-12", StartTime: "10:00", Repeat: "daily"}
}

// ─────────────────────────────────────────────
// Create
// ─────────────────────────────────────────────

func TestCreateItem_AudioWithBackground(t *testing.T) {
	s, mock := newService(t)
	expectMedia(mock, 2, 1, "audio", 214)
	expectMedia(mock, 4, 1, "photo", nil)
	expectSettings(mock, "07:00", "21:00")
	mock.ExpectQuery(qInsert).
		WithArgs(1, 2, 4, nil, "2026-10-12", "09:15", "weekly", "2026-12-31").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	expectGetItem(mock, itemRow(5, 2, 4, nil, "2026-10-12", "09:15", "weekly", "2026-12-31", "audio", 214))
	expectSettings(mock, "07:00", "21:00")

	item, err := s.CreateItem(ItemInput{CarepartnerID: 1, MediaID: 2, BackgroundPhotoID: intp(4),
		StartDate: "2026-10-12", StartTime: "9:15", Repeat: "weekly", EndDate: strp("2026-12-31")})

	require.NoError(t, err)
	assert.Equal(t, "audio_with_background", item.Shape)
	assert.Equal(t, "/media/4/file", item.BackgroundPhoto.URL)
	assert.Equal(t, 214, *item.Media.DurationSeconds)
	assert.Equal(t, []string{}, item.SkipDates)
	assert.False(t, item.OutsideAwakeHours)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateItem_TimedPhoto(t *testing.T) {
	s, mock := newService(t)
	expectMedia(mock, 3, 1, "photo", nil)
	expectSettings(mock, "07:00", "21:00")
	mock.ExpectQuery(qInsert).
		WithArgs(1, 3, nil, 300, "2026-10-12", "14:30", "once", nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(6))
	expectGetItem(mock, itemRow(6, 3, nil, 300, "2026-10-12", "14:30", "once", nil, "photo", nil))
	expectSettings(mock, "07:00", "21:00")

	item, err := s.CreateItem(ItemInput{CarepartnerID: 1, MediaID: 3, DisplaySeconds: intp(300),
		StartDate: "2026-10-12", StartTime: "14:30", Repeat: "once"})

	require.NoError(t, err)
	assert.Equal(t, "photo", item.Shape)
	assert.Equal(t, 300, item.DurationSeconds())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateItem_RejectedBeforeAnyQuery(t *testing.T) {
	cases := []struct {
		name string
		edit func(*ItemInput)
		want error
	}{
		{"no carepartner", func(in *ItemInput) { in.CarepartnerID = 0 }, ErrInvalidCarepartner},
		{"unknown repeat", func(in *ItemInput) { in.Repeat = "yearly" }, ErrInvalidRepeat},
		{"bad start date", func(in *ItemInput) { in.StartDate = "10/12/2026" }, ErrInvalidDate},
		{"end before start", func(in *ItemInput) { in.EndDate = strp("2026-10-01") }, ErrInvalidDate},
		{"bad end date", func(in *ItemInput) { in.EndDate = strp("soon") }, ErrInvalidDate},
		{"once with an end date", func(in *ItemInput) { in.Repeat = "once"; in.EndDate = strp("2026-12-01") }, ErrInvalidDate},
		{"bad time", func(in *ItemInput) { in.StartTime = "10am" }, ErrInvalidTime},
		{"10:10 is not a 15 minute step", func(in *ItemInput) { in.StartTime = "10:10" }, ErrNotOn15MinStep},
		{"seconds", func(in *ItemInput) { in.StartTime = "10:00:30" }, ErrInvalidTime},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock := newService(t)
			in := videoInput()
			tc.edit(&in)
			_, err := s.CreateItem(in)
			assert.ErrorIs(t, err, tc.want)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateItem_Shapes(t *testing.T) {
	cases := []struct {
		name   string
		in     ItemInput
		expect func(sqlmock.Sqlmock)
		want   error
	}{
		{"video with a background", ItemInput{MediaID: 2, BackgroundPhotoID: intp(4)},
			func(m sqlmock.Sqlmock) { expectMedia(m, 2, 1, "video", 30) }, ErrInvalidShape},
		{"video with display_seconds", ItemInput{MediaID: 2, DisplaySeconds: intp(30)},
			func(m sqlmock.Sqlmock) { expectMedia(m, 2, 1, "video", 30) }, ErrInvalidShape},
		{"audio with display_seconds", ItemInput{MediaID: 2, DisplaySeconds: intp(30)},
			func(m sqlmock.Sqlmock) { expectMedia(m, 2, 1, "audio", 30) }, ErrInvalidShape},
		{"audio with a video background", ItemInput{MediaID: 2, BackgroundPhotoID: intp(4)},
			func(m sqlmock.Sqlmock) { expectMedia(m, 2, 1, "audio", 30); expectMedia(m, 4, 1, "video", 30) }, ErrInvalidShape},
		{"photo without display_seconds", ItemInput{MediaID: 3},
			func(m sqlmock.Sqlmock) { expectMedia(m, 3, 1, "photo", nil) }, ErrInvalidShape},
		{"photo shown too long", ItemInput{MediaID: 3, DisplaySeconds: intp(601)},
			func(m sqlmock.Sqlmock) { expectMedia(m, 3, 1, "photo", nil) }, ErrInvalidShape},
		{"photo with a background", ItemInput{MediaID: 3, DisplaySeconds: intp(60), BackgroundPhotoID: intp(4)},
			func(m sqlmock.Sqlmock) { expectMedia(m, 3, 1, "photo", nil) }, ErrInvalidShape},
		{"another carepartner's media", ItemInput{MediaID: 2},
			func(m sqlmock.Sqlmock) { expectMedia(m, 2, 7, "video", 30) }, ErrInvalidMedia},
		{"missing media", ItemInput{MediaID: 2},
			func(m sqlmock.Sqlmock) { m.ExpectQuery(qMedia).WillReturnRows(sqlmock.NewRows(mediaCols)) }, ErrInvalidMedia},
		{"another carepartner's background", ItemInput{MediaID: 2, BackgroundPhotoID: intp(4)},
			func(m sqlmock.Sqlmock) { expectMedia(m, 2, 1, "audio", 30); expectMedia(m, 4, 7, "photo", nil) }, ErrInvalidMedia},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock := newService(t)
			tc.expect(mock)
			in := tc.in
			in.CarepartnerID, in.StartDate, in.StartTime, in.Repeat = 1, "2026-10-12", "10:00", "daily"
			_, err := s.CreateItem(in)
			assert.ErrorIs(t, err, tc.want)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateItem_AwakeHours(t *testing.T) {
	cases := []struct {
		name, at string
		duration int
		want     error
	}{
		{"before wake", "06:45", 60, ErrOutsideAwakeHours},
		{"runs past bed", "20:45", 900 + 1, ErrOutsideAwakeHours},
		{"ends exactly at bed", "20:45", 900, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock := newService(t)
			expectMedia(mock, 2, 1, "video", tc.duration)
			expectSettings(mock, "07:00", "21:00")
			if tc.want == nil {
				mock.ExpectQuery(qInsert).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
				expectGetItem(mock, itemRow(5, 2, nil, nil, "2026-10-12", tc.at, "daily", nil, "video", tc.duration))
				expectSettings(mock, "07:00", "21:00")
			}
			in := videoInput()
			in.StartTime = tc.at
			_, err := s.CreateItem(in)
			if tc.want == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.want)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateItem_MediaDeletedMeanwhile(t *testing.T) {
	s, mock := newService(t)
	expectMedia(mock, 2, 1, "video", 30)
	expectSettings(mock, "07:00", "21:00")
	mock.ExpectQuery(qInsert).WillReturnError(&pq.Error{Code: "23503"})

	_, err := s.CreateItem(videoInput())
	assert.ErrorIs(t, err, ErrInvalidMedia)
}

// ─────────────────────────────────────────────
// Read / Calendar
// ─────────────────────────────────────────────

func TestGetItem_FlagsWithCurrentAwakeHours(t *testing.T) {
	s, mock := newService(t)
	expectGetItem(mock, itemRow(5, 2, nil, nil, "2026-10-12", "07:15", "daily", nil, "video", 60), "2026-10-14")
	expectSettings(mock, "08:00", "21:00") // wake time moved later than the item

	item, err := s.GetItem(5)

	require.NoError(t, err)
	assert.True(t, item.OutsideAwakeHours, "flagged, not rejected or edited")
	assert.Equal(t, []string{"2026-10-14"}, item.SkipDates)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCalendar(t *testing.T) {
	s, mock := newService(t)
	expectSettings(mock, "07:00", "21:00")
	mock.ExpectQuery(qList).WithArgs(1, "2026-10-18", "2026-10-12").
		WillReturnRows(sqlmock.NewRows(itemCols).
			AddRow(5, 1, 2, nil, nil, "2026-10-01", "10:00", "daily", "2026-10-14", "video", "Clip", 60, nil).
			AddRow(6, 1, 3, nil, 300, "2026-10-13", "09:00", "weekly", nil, "photo", "Beach", nil, nil).
			AddRow(7, 1, 8, 4, nil, "2026-10-12", "20:45", "once", nil, "audio", "Hymn", 1200, "Sarah"))
	mock.ExpectQuery(qSkips).WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"scheduled_item_id", "skip_date"}).AddRow(5, "2026-10-13"))

	got, err := s.Calendar(1, "2026-10-12", "2026-10-18")

	require.NoError(t, err)
	var keys []string
	for _, o := range got {
		keys = append(keys, o.Date+" "+o.Time)
	}
	assert.Equal(t, []string{
		"2026-10-12 10:00", // daily video
		"2026-10-12 20:45", // once audio
		"2026-10-13 09:00", // weekly photo; the daily video is skipped this day
		"2026-10-14 10:00", // daily video's last day
	}, keys)
	assert.Equal(t, "audio_with_background", got[1].Shape)
	assert.True(t, got[1].OutsideAwakeHours, "20 minutes from 20:45 runs past 21:00")
	assert.Equal(t, 300, got[2].DurationSeconds)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCalendar_InvalidRange(t *testing.T) {
	s, mock := newService(t)
	for _, r := range [][2]string{
		{"2026-10-18", "2026-10-12"}, // backwards
		{"2026-10-01", "2026-12-02"}, // 63 days
		{"2026-10-01", ""},
		{"yesterday", "2026-10-12"},
	} {
		_, err := s.Calendar(1, r[0], r[1])
		assert.ErrorIs(t, err, ErrInvalidRange, r)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCalendar_62DaysAllowed(t *testing.T) {
	s, mock := newService(t)
	expectSettings(mock, "07:00", "21:00")
	mock.ExpectQuery(qList).WillReturnRows(sqlmock.NewRows(itemCols))
	_, err := s.Calendar(1, "2026-10-01", "2026-12-01")
	assert.NoError(t, err)
}

// ─────────────────────────────────────────────
// Update / Delete
// ─────────────────────────────────────────────

func TestUpdateItem(t *testing.T) {
	s, mock := newService(t)
	expectGetItem(mock, itemRow(5, 2, nil, nil, "2026-10-12", "10:00", "daily", nil, "video", 60))
	expectMedia(mock, 2, 1, "video", 60)
	expectSettings(mock, "07:00", "21:00")
	mock.ExpectExec(qUpdate).WithArgs(2, nil, nil, "2026-10-12", "11:30", "daily", nil, 5).
		WillReturnResult(sqlmock.NewResult(0, 1))
	expectGetItem(mock, itemRow(5, 2, nil, nil, "2026-10-12", "11:30", "daily", nil, "video", 60))
	expectSettings(mock, "07:00", "21:00")

	in := videoInput()
	in.StartTime = "11:30"
	item, err := s.UpdateItem(5, in)

	require.NoError(t, err)
	assert.Equal(t, "11:30", item.StartTime)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateItem_NotFound(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectQuery(qGetItem).WillReturnRows(sqlmock.NewRows(itemCols))
	_, err := s.UpdateItem(5, videoInput())
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDeleteItem_Series(t *testing.T) {
	s, mock := newService(t)
	mock.ExpectExec(qDelete).WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 1))
	assert.NoError(t, s.DeleteItem(5, "series", ""))

	mock.ExpectExec(qDelete).WithArgs(6).WillReturnResult(sqlmock.NewResult(0, 0))
	assert.ErrorIs(t, s.DeleteItem(6, "series", ""), ErrNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteItem_Occurrence(t *testing.T) {
	s, mock := newService(t)
	expectGetItem(mock, itemRow(5, 2, nil, nil, "2026-10-12", "10:00", "weekly", nil, "video", 60))
	mock.ExpectExec(qAddSkip).WithArgs(5, "2026-10-19").WillReturnResult(sqlmock.NewResult(0, 1))

	assert.NoError(t, s.DeleteItem(5, "occurrence", "2026-10-19"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteItem_OccurrenceOfOnceItemDeletesIt(t *testing.T) {
	s, mock := newService(t)
	expectGetItem(mock, itemRow(5, 2, nil, nil, "2026-10-12", "10:00", "once", nil, "video", 60))
	mock.ExpectExec(qDelete).WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 1))

	assert.NoError(t, s.DeleteItem(5, "occurrence", "2026-10-12"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteItem_Errors(t *testing.T) {
	s, mock := newService(t)
	assert.ErrorIs(t, s.DeleteItem(5, "", ""), ErrInvalidScope)
	assert.ErrorIs(t, s.DeleteItem(5, "occurrence", "tuesday"), ErrInvalidDate)

	expectGetItem(mock, itemRow(5, 2, nil, nil, "2026-10-12", "10:00", "weekly", nil, "video", 60))
	assert.ErrorIs(t, s.DeleteItem(5, "occurrence", "2026-10-20"), ErrNotAnOccurrence, "a Tuesday, not a Monday")

	mock.ExpectQuery(qGetItem).WillReturnRows(sqlmock.NewRows(itemCols))
	assert.ErrorIs(t, s.DeleteItem(9, "occurrence", "2026-10-12"), ErrNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}
