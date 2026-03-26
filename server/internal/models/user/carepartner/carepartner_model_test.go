package carepartner

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

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

/* -----------------------------------------------------
   CREATE
----------------------------------------------------- */

func TestCreateCarepartner_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	c := &Carepartner{
		Name:      "test",
		AutoStart: time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "Carepartner" (name, auto_start)
		VALUES ($1, $2)
		RETURNING id;
	`)).
		WithArgs(c.Name, c.AutoStart).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := CreateCarepartner(db, c)
	assert.NoError(t, err)
	assert.Equal(t, 1, c.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCarepartner_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	c := &Carepartner{
		Name:      "bad",
		AutoStart: time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "Carepartner" (name, auto_start)
		VALUES ($1, $2)
		RETURNING id;
	`)).
		WithArgs(c.Name, c.AutoStart).
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

	autoStart := time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"id", "name", "auto_start"}).
		AddRow(1, "john", autoStart)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, auto_start
		FROM "Carepartner"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	c, err := GetCarepartnerByID(db, 1)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, 1, c.ID)
	assert.Equal(t, "john", c.Name)
	assert.Equal(t, autoStart, c.AutoStart)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCarepartnerByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, auto_start
		FROM "Carepartner"
		WHERE id = $1;
	`)).
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

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, auto_start
		FROM "Carepartner"
		WHERE id = $1;
	`)).
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

	t1 := time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)
	t2 := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"id", "name", "auto_start"}).
		AddRow(1, "a", t1).
		AddRow(2, "b", t2)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, auto_start
		FROM "Carepartner";
	`)).
		WillReturnRows(rows)

	list, err := GetAllCarepartners(db)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, t1, list[0].AutoStart)
	assert.Equal(t, t2, list[1].AutoStart)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllCarepartners_QueryError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, auto_start
		FROM "Carepartner";
	`)).
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

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, auto_start
		FROM "Carepartner";
	`)).
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

	autoStart := time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)

	c := &Carepartner{
		ID:        1,
		Name:      "updated",
		AutoStart: autoStart,
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "Carepartner"
		SET name = $1, auto_start = $2
		WHERE id = $3;
	`)).
		WithArgs(c.Name, c.AutoStart, c.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := UpdateCarepartner(db, c)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCarepartner_NoRows(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	c := &Carepartner{ID: 999}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "Carepartner"
		SET name = $1, auto_start = $2
		WHERE id = $3;
	`)).
		WithArgs(c.Name, c.AutoStart, c.ID).
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

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "Carepartner"
		SET name = $1, auto_start = $2
		WHERE id = $3;
	`)).
		WithArgs(c.Name, c.AutoStart, c.ID).
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
