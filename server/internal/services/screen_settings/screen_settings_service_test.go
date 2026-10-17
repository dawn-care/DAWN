package screen_settings

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"server/internal/utils"
)

var (
	selectSettings = regexp.QuoteMeta(`FROM "ScreenSettings"`)
	updateSettings = regexp.QuoteMeta(`UPDATE "ScreenSettings"`)
)

func TestGetScreenSettings(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(selectSettings).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"carepartner_id", "wake_time", "bed_time"}).
			AddRow(1, "07:00", "21:00"))

	s, err := NewScreenSettingsService(db).GetScreenSettings(1)

	require.NoError(t, err)
	assert.Equal(t, "07:00", s.WakeTime)
	assert.Equal(t, "21:00", s.BedTime)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetScreenSettings_NotFound(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(selectSettings).WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"carepartner_id", "wake_time", "bed_time"}))

	_, err := NewScreenSettingsService(db).GetScreenSettings(9)
	assert.ErrorIs(t, err, ErrNotFound)

	_, err = NewScreenSettingsService(db).GetScreenSettings(0)
	assert.ErrorIs(t, err, ErrInvalidCarepartner)
}

func TestGetScreenSettings_DBError(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(selectSettings).WillReturnError(errors.New("db down"))

	_, err := NewScreenSettingsService(db).GetScreenSettings(1)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrNotFound)
}

func TestUpdateScreenSettings(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	// Times are zero-padded before they reach the database
	mock.ExpectQuery(updateSettings).
		WithArgs("06:30", "22:00", 1).
		WillReturnRows(sqlmock.NewRows([]string{"wake_time", "bed_time"}).
			AddRow("06:30", "22:00"))

	s, err := NewScreenSettingsService(db).UpdateScreenSettings(1, UpdateInput{WakeTime: "6:30", BedTime: "22:00"})

	require.NoError(t, err)
	assert.Equal(t, 1, s.CarepartnerID)
	assert.Equal(t, "06:30", s.WakeTime)
	assert.Equal(t, "22:00", s.BedTime)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateScreenSettings_NotFound(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(updateSettings).
		WillReturnRows(sqlmock.NewRows([]string{"wake_time", "bed_time"}))

	_, err := NewScreenSettingsService(db).UpdateScreenSettings(9, UpdateInput{WakeTime: "07:00", BedTime: "21:00"})
	assert.ErrorIs(t, err, ErrNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateScreenSettings_Validation(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()
	service := NewScreenSettingsService(db)

	valid := UpdateInput{WakeTime: "07:00", BedTime: "21:00"}
	cases := []struct {
		name string
		edit func(*UpdateInput)
		want error
	}{
		{"wake before 6:00", func(in *UpdateInput) { in.WakeTime = "05:59" }, ErrInvalidWakeTime},
		{"wake after 10:00", func(in *UpdateInput) { in.WakeTime = "10:01" }, ErrInvalidWakeTime},
		{"wake in the evening", func(in *UpdateInput) { in.WakeTime = "20:00" }, ErrInvalidWakeTime},
		{"wake not a time", func(in *UpdateInput) { in.WakeTime = "seven" }, ErrInvalidWakeTime},
		{"wake with seconds", func(in *UpdateInput) { in.WakeTime = "07:00:00" }, ErrInvalidWakeTime},
		{"missing wake", func(in *UpdateInput) { in.WakeTime = "" }, ErrInvalidWakeTime},
		{"bed before 19:00", func(in *UpdateInput) { in.BedTime = "18:59" }, ErrInvalidBedTime},
		{"bed after 22:00", func(in *UpdateInput) { in.BedTime = "22:01" }, ErrInvalidBedTime},
		{"bed as 12 hour time", func(in *UpdateInput) { in.BedTime = "9:00" }, ErrInvalidBedTime},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := valid
			tc.edit(&in)
			_, err := service.UpdateScreenSettings(1, in)
			assert.ErrorIs(t, err, tc.want)
		})
	}

	_, err := service.UpdateScreenSettings(0, valid)
	assert.ErrorIs(t, err, ErrInvalidCarepartner)
	assert.NoError(t, mock.ExpectationsWereMet(), "invalid input never reaches the database")
}

func TestUpdateScreenSettings_Bounds(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()
	service := NewScreenSettingsService(db)

	for _, in := range []UpdateInput{
		{WakeTime: "06:00", BedTime: "19:00"},
		{WakeTime: "10:00", BedTime: "22:00"},
	} {
		mock.ExpectQuery(updateSettings).
			WithArgs(in.WakeTime, in.BedTime, 1).
			WillReturnRows(sqlmock.NewRows([]string{"wake_time", "bed_time"}).
				AddRow(in.WakeTime, in.BedTime))
		_, err := service.UpdateScreenSettings(1, in)
		assert.NoError(t, err, "%+v is inside the allowed range", in)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}
