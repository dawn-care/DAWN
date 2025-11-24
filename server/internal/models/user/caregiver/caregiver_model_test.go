package caregiver

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

/* -----------------------------------------------------
   CREATE (single)
----------------------------------------------------- */

func TestCreateCaregiver(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	c := &Caregiver{
		CarereceiverID: 101,
		Username:       "john",
		Role:           "admin",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "Caregiver" (carereceiver_id, username, role)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WithArgs(c.CarereceiverID, c.Username, c.Role).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	err := CreateCaregiver(db, c)
	assert.NoError(t, err)
	assert.Equal(t, 10, c.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCaregiver_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	c := &Caregiver{
		CarereceiverID: 999,
		Username:       "bad",
		Role:           "role",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "Caregiver" (carereceiver_id, username, role)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WithArgs(c.CarereceiverID, c.Username, c.Role).
		WillReturnError(errors.New("insert error"))

	err := CreateCaregiver(db, c)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

/* -----------------------------------------------------
   CREATE BULK
----------------------------------------------------- */

func TestCreateCaregiversBulk_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	cgs := []Caregiver{
		{CarereceiverID: 1, Username: "a", Role: "x"},
		{CarereceiverID: 2, Username: "b", Role: "y"},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare(regexp.QuoteMeta(`
		INSERT INTO "Caregiver" (carereceiver_id, username, role)
		VALUES ($1, $2, $3)
		RETURNING id;
	`))

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "Caregiver" (carereceiver_id, username, role)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WithArgs(cgs[0].CarereceiverID, cgs[0].Username, cgs[0].Role).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(101))

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "Caregiver" (carereceiver_id, username, role)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WithArgs(cgs[1].CarereceiverID, cgs[1].Username, cgs[1].Role).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(102))

	mock.ExpectCommit()

	err := CreateCaregiversBulk(db, cgs)
	assert.NoError(t, err)
	assert.Equal(t, 101, cgs[0].ID)
	assert.Equal(t, 102, cgs[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCaregiversBulk_StatementError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	cgs := []Caregiver{{CarereceiverID: 1, Username: "a", Role: "x"}}

	mock.ExpectBegin()

	mock.ExpectPrepare(regexp.QuoteMeta(`
		INSERT INTO "Caregiver" (carereceiver_id, username, role)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WillReturnError(errors.New("prepare error"))

	err := CreateCaregiversBulk(db, cgs)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCaregiversBulk_QueryError_Rollback(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	cgs := []Caregiver{{CarereceiverID: 1, Username: "a", Role: "x"}}

	mock.ExpectBegin()

	mock.ExpectPrepare(regexp.QuoteMeta(`
		INSERT INTO "Caregiver" (carereceiver_id, username, role)
		VALUES ($1, $2, $3)
		RETURNING id;
	`))

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "Caregiver" (carereceiver_id, username, role)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WithArgs(1, "a", "x").
		WillReturnError(errors.New("insert error"))

	mock.ExpectRollback()

	err := CreateCaregiversBulk(db, cgs)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

/* -----------------------------------------------------
   GET BY ID
----------------------------------------------------- */

func TestGetCaregiverByID_Found(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "carereceiver_id", "username", "role"}).
		AddRow(10, 100, "john", "admin")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, carereceiver_id, username, role
		FROM "Caregiver"
		WHERE id = $1;
	`)).
		WithArgs(10).
		WillReturnRows(rows)

	c, err := GetCaregiverByID(db, 10)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, 10, c.ID)
	assert.Equal(t, "john", c.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCaregiverByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, carereceiver_id, username, role
		FROM "Caregiver"
		WHERE id = $1;
	`)).
		WithArgs(5).
		WillReturnError(sql.ErrNoRows)

	c, err := GetCaregiverByID(db, 5)
	assert.NoError(t, err)
	assert.Nil(t, c)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCaregiverByID_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, carereceiver_id, username, role
		FROM "Caregiver"
		WHERE id = $1;
	`)).
		WithArgs(5).
		WillReturnError(errors.New("db error"))

	c, err := GetCaregiverByID(db, 5)
	assert.Error(t, err)
	assert.Nil(t, c)
	assert.NoError(t, mock.ExpectationsWereMet())
}

/* -----------------------------------------------------
   GET BY CARERECEIVER ID
----------------------------------------------------- */

func TestGetCaregiverByCarereceiver_Found(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "carereceiver_id", "username", "role"}).
		AddRow(20, 999, "care", "basic")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, carereceiver_id, username, role
		FROM "Caregiver"
		WHERE carereceiver_id = $1;
	`)).
		WithArgs(999).
		WillReturnRows(rows)

	c, err := GetCaregiverByCarereceiver(db, 999)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, 20, c.ID)
	assert.Equal(t, "care", c.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCaregiverByCarereceiver_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, carereceiver_id, username, role
		FROM "Caregiver"
		WHERE carereceiver_id = $1;
	`)).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	c, err := GetCaregiverByCarereceiver(db, 999)
	assert.NoError(t, err)
	assert.Nil(t, c)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCaregiverByCarereceiver_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, carereceiver_id, username, role
		FROM "Caregiver"
		WHERE carereceiver_id = $1;
	`)).
		WithArgs(999).
		WillReturnError(errors.New("db fail"))

	c, err := GetCaregiverByCarereceiver(db, 999)
	assert.Error(t, err)
	assert.Nil(t, c)
	assert.NoError(t, mock.ExpectationsWereMet())
}

/* -----------------------------------------------------
   GET ALL
----------------------------------------------------- */

func TestGetCaregivers_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "carereceiver_id", "username", "role"}).
		AddRow(1, 10, "one", "x").
		AddRow(2, 20, "two", "y")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, carereceiver_id, username, role
		FROM "Caregiver";
	`)).
		WillReturnRows(rows)

	result, err := GetCaregivers(db)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCaregivers_QueryError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, carereceiver_id, username, role
		FROM "Caregiver";
	`)).
		WillReturnError(errors.New("query failed"))

	result, err := GetCaregivers(db)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCaregivers_ScanError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	// Wrong number/type of columns forces Scan error
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, carereceiver_id, username, role
		FROM "Caregiver";
	`)).
		WillReturnRows(rows)

	result, err := GetCaregivers(db)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

/* -----------------------------------------------------
   SEARCH
----------------------------------------------------- */

func TestSearchCaregiversByUsername_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "carereceiver_id", "username", "role"}).
		AddRow(10, 100, "alice", "admin").
		AddRow(11, 101, "alicex", "user")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, carereceiver_id, username, role
		FROM "Caregiver"
		WHERE username ILIKE $1;
	`)).
		WithArgs("%ali%").
		WillReturnRows(rows)

	list, err := SearchCaregiversByUsername(db, "ali")
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchCaregiversByUsername_QueryError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, carereceiver_id, username, role
		FROM "Caregiver"
		WHERE username ILIKE $1;
	`)).
		WithArgs("%fail%").
		WillReturnError(errors.New("query error"))

	list, err := SearchCaregiversByUsername(db, "fail")
	assert.Error(t, err)
	assert.Nil(t, list)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchCaregiversByUsername_ScanError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	// Wrong columns → scan fails
	rows := sqlmock.NewRows([]string{"id"}).AddRow(5)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, carereceiver_id, username, role
		FROM "Caregiver"
		WHERE username ILIKE $1;
	`)).
		WithArgs("%bad%").
		WillReturnRows(rows)

	list, err := SearchCaregiversByUsername(db, "bad")
	assert.Error(t, err)
	assert.Nil(t, list)
	assert.NoError(t, mock.ExpectationsWereMet())
}

/* -----------------------------------------------------
   UPDATE
----------------------------------------------------- */

func TestUpdateCaregiver_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	c := &Caregiver{ID: 1, CarereceiverID: 200, Username: "updated", Role: "boss"}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "Caregiver"
		SET carereceiver_id = $1, username = $2, role = $3
		WHERE id = $4;
	`)).
		WithArgs(c.CarereceiverID, c.Username, c.Role, c.ID).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

	err := UpdateCaregiver(db, c)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCaregiver_NoRows(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	c := &Caregiver{ID: 999, CarereceiverID: 0, Username: "none", Role: "none"}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "Caregiver"
		SET carereceiver_id = $1, username = $2, role = $3
		WHERE id = $4;
	`)).
		WithArgs(c.CarereceiverID, c.Username, c.Role, c.ID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	err := UpdateCaregiver(db, c)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("no caregiver found with id %d", c.ID), err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCaregiver_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	c := &Caregiver{ID: 5}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "Caregiver"
		SET carereceiver_id = $1, username = $2, role = $3
		WHERE id = $4;
	`)).
		WithArgs(c.CarereceiverID, c.Username, c.Role, c.ID).
		WillReturnError(errors.New("update fail"))

	err := UpdateCaregiver(db, c)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

/* -----------------------------------------------------
   DELETE
----------------------------------------------------- */

func TestDeleteCaregiver_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "Caregiver" WHERE id = $1;`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := DeleteCaregiver(db, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCaregiver_NoRows(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "Caregiver" WHERE id = $1;`)).
		WithArgs(123).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := DeleteCaregiver(db, 123)
	assert.Error(t, err)
	assert.Equal(t, "no caregiver found with id 123", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCaregiver_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "Caregiver" WHERE id = $1;`)).
		WithArgs(1).
		WillReturnError(errors.New("delete fail"))

	err := DeleteCaregiver(db, 1)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
