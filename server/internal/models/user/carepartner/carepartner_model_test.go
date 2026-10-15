package carepartner

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

/* -----------------------------------------------------
   SETUP
----------------------------------------------------- */

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	return db, mock
}

var carepartnerColumns = []string{"id", "name", "caregiver_id", "timezone"}

const (
	insertQuery = `
		INSERT INTO "Carepartner" (name, caregiver_id)
		VALUES ($1, $2)
		RETURNING id, timezone;
	`
	selectByIDQuery = `
		SELECT id, name, caregiver_id, timezone
		FROM "Carepartner"
		WHERE id = $1;
	`
	selectAllQuery = `
		SELECT id, name, caregiver_id, timezone
		FROM "Carepartner";
	`
	updateQuery = `
		UPDATE "Carepartner"
		SET name = $1, caregiver_id = $2
		WHERE id = $3;
	`
)

/* -----------------------------------------------------
   CREATE
----------------------------------------------------- */

func TestCreateCarepartner_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	c := &Carepartner{Name: "test"}

	// No caregiver yet: caregiver_id is stored as NULL, not 0
	mock.ExpectQuery(regexp.QuoteMeta(insertQuery)).
		WithArgs(c.Name, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id", "timezone"}).AddRow(1, "America/New_York"))

	err := CreateCarepartner(db, c)
	assert.NoError(t, err)
	assert.Equal(t, 1, c.ID)
	assert.Equal(t, "America/New_York", c.Timezone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCarepartner_WithCaregiver(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	c := &Carepartner{Name: "test", CaregiverID: 4}

	mock.ExpectQuery(regexp.QuoteMeta(insertQuery)).
		WithArgs(c.Name, 4).
		WillReturnRows(sqlmock.NewRows([]string{"id", "timezone"}).AddRow(2, "America/New_York"))

	assert.NoError(t, CreateCarepartner(db, c))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCarepartner_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	c := &Carepartner{Name: "bad"}

	mock.ExpectQuery(regexp.QuoteMeta(insertQuery)).
		WithArgs(c.Name, nil).
		WillReturnError(errors.New("insert error"))

	err := CreateCarepartner(db, c)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

/* -----------------------------------------------------
   GET BY ID
----------------------------------------------------- */

func TestGetCarepartnerByID_Found(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows(carepartnerColumns).AddRow(1, "john", 3, "America/New_York")

	mock.ExpectQuery(regexp.QuoteMeta(selectByIDQuery)).
		WithArgs(1).
		WillReturnRows(rows)

	c, err := GetCarepartnerByID(db, 1)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, 1, c.ID)
	assert.Equal(t, "john", c.Name)
	assert.Equal(t, 3, c.CaregiverID)
	assert.Equal(t, "America/New_York", c.Timezone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCarepartnerByID_NoCaregiver(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows(carepartnerColumns).AddRow(1, "john", nil, "America/New_York")

	mock.ExpectQuery(regexp.QuoteMeta(selectByIDQuery)).
		WithArgs(1).
		WillReturnRows(rows)

	c, err := GetCarepartnerByID(db, 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, c.CaregiverID, "a NULL caregiver reads as 0")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCarepartnerByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(selectByIDQuery)).
		WithArgs(5).
		WillReturnError(sql.ErrNoRows)

	c, err := GetCarepartnerByID(db, 5)
	assert.NoError(t, err)
	assert.Nil(t, c)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCarepartnerByID_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(selectByIDQuery)).
		WithArgs(5).
		WillReturnError(errors.New("db error"))

	c, err := GetCarepartnerByID(db, 5)
	assert.Error(t, err)
	assert.Nil(t, c)
	assert.NoError(t, mock.ExpectationsWereMet())
}

/* -----------------------------------------------------
   GET ALL
----------------------------------------------------- */

func TestGetAllCarepartners_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows(carepartnerColumns).
		AddRow(1, "a", 1, "America/New_York").
		AddRow(2, "b", nil, "America/Chicago")

	mock.ExpectQuery(regexp.QuoteMeta(selectAllQuery)).
		WillReturnRows(rows)

	list, err := GetAllCarepartners(db)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, 1, list[0].CaregiverID)
	assert.Equal(t, 0, list[1].CaregiverID)
	assert.Equal(t, "America/Chicago", list[1].Timezone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllCarepartners_QueryError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(selectAllQuery)).
		WillReturnError(errors.New("query fail"))

	list, err := GetAllCarepartners(db)
	assert.Error(t, err)
	assert.Nil(t, list)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllCarepartners_ScanError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(regexp.QuoteMeta(selectAllQuery)).
		WillReturnRows(rows)

	list, err := GetAllCarepartners(db)
	assert.Error(t, err)
	assert.Nil(t, list)
	assert.NoError(t, mock.ExpectationsWereMet())
}

/* -----------------------------------------------------
   UPDATE
----------------------------------------------------- */

func TestUpdateCarepartner_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	c := &Carepartner{ID: 1, Name: "updated", CaregiverID: 2}

	mock.ExpectExec(regexp.QuoteMeta(updateQuery)).
		WithArgs(c.Name, 2, c.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := UpdateCarepartner(db, c)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCarepartner_NoRows(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	c := &Carepartner{ID: 999}

	mock.ExpectExec(regexp.QuoteMeta(updateQuery)).
		WithArgs(c.Name, nil, c.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := UpdateCarepartner(db, c)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("no carepartner found with id %d", c.ID), err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCarepartner_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	c := &Carepartner{ID: 1}

	mock.ExpectExec(regexp.QuoteMeta(updateQuery)).
		WithArgs(c.Name, nil, c.ID).
		WillReturnError(errors.New("update fail"))

	err := UpdateCarepartner(db, c)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

/* -----------------------------------------------------
   DELETE
----------------------------------------------------- */

func TestDeleteCarepartner_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "Carepartner" WHERE id = $1;`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := DeleteCarepartner(db, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCarepartner_NoRows(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "Carepartner" WHERE id = $1;`)).
		WithArgs(123).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := DeleteCarepartner(db, 123)
	assert.Error(t, err)
	assert.Equal(t, "no carepartner found with id 123", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCarepartner_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "Carepartner" WHERE id = $1;`)).
		WithArgs(1).
		WillReturnError(errors.New("delete fail"))

	err := DeleteCarepartner(db, 1)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
