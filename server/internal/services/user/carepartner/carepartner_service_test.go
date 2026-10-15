package carepartner

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"server/internal/models/user/carepartner"
	"server/internal/utils"
)

var (
	insertCarepartner    = regexp.QuoteMeta(`INSERT INTO "Carepartner"`)
	insertScreenSettings = regexp.QuoteMeta(`INSERT INTO "ScreenSettings" (carepartner_id) VALUES ($1);`)
)

func TestCreateCarepartner_CreatesDefaultScreenSettings(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()
	service := NewCarepartnerService(db)

	mock.ExpectBegin()
	mock.ExpectQuery(insertCarepartner).
		WithArgs("Ray", nil).
		WillReturnRows(sqlmock.NewRows([]string{"id", "timezone"}).AddRow(5, "America/New_York"))
	mock.ExpectExec(insertScreenSettings).
		WithArgs(5).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	c := &carepartner.Carepartner{Name: "Ray"}
	err := service.CreateCarepartner(c)

	assert.NoError(t, err)
	assert.Equal(t, 5, c.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCarepartner_SettingsFailureRollsBack(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()
	service := NewCarepartnerService(db)

	mock.ExpectBegin()
	mock.ExpectQuery(insertCarepartner).
		WillReturnRows(sqlmock.NewRows([]string{"id", "timezone"}).AddRow(5, "America/New_York"))
	mock.ExpectExec(insertScreenSettings).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	err := service.CreateCarepartner(&carepartner.Carepartner{Name: "Ray"})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet(), "no carepartner is left without settings")
}

func TestCreateCarepartner_InsertFailureRollsBack(t *testing.T) {
	db, mock := utils.SetupMockDB(t)
	defer db.Close()
	service := NewCarepartnerService(db)

	mock.ExpectBegin()
	mock.ExpectQuery(insertCarepartner).WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	err := service.CreateCarepartner(&carepartner.Carepartner{Name: "Ray"})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCarepartner_Nil(t *testing.T) {
	db, _ := utils.SetupMockDB(t)
	defer db.Close()

	assert.EqualError(t, NewCarepartnerService(db).CreateCarepartner(nil), "carepartner is nil")
}
