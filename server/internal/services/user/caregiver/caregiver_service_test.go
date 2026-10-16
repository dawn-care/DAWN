package caregiver

import (
	"database/sql"
	"regexp"
	"testing"

	model "server/internal/models/user/caregiver"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

/* -------------------------------------------------------
   SETUP
------------------------------------------------------- */

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	return db, mock
}

/* -------------------------------------------------------
   CREATE
------------------------------------------------------- */

func TestCreateCaregiver_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)

	c := &model.Caregiver{
		Name:          "john",
		Role:          "helper",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "Caregiver" (carepartner_id, name, role)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)).
		WithArgs(c.Name, c.Role).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))

	err := svc.CreateCaregiver(c)
	assert.NoError(t, err)
	assert.Equal(t, 5, c.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// func TestCreateCaregiver_ValidationError_EmptyName(t *testing.T) {
// 	db, _ := setupMockDB(t)
// 	defer db.Close()

// 	svc := NewCaregiverService(db)
// 	c := &model.Caregiver{CarepartnerID: 10}

// 	err := svc.CreateCaregiver(c)
// 	assert.Error(t, err)
// 	assert.Equal(t, "name cannot be empty", err.Error())
// }

func TestCreateCaregiver_ValidationError_InvalidCarereceiver(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)
	c := &model.Caregiver{Name: "john"}

	err := svc.CreateCaregiver(c)
	assert.Error(t, err)
	assert.Equal(t, "carepartner_id must match caregiver ID or be set", err.Error())
}

/* -------------------------------------------------------
   CREATE BULK
------------------------------------------------------- */

func TestCreateCaregiversBulk_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	list := []model.Caregiver{
		{Name: "a", Role: "helper"},
		{Name: "b", Role: "helper"},
	}

	svc := NewCaregiverService(db)

	mock.ExpectBegin()

	mock.ExpectPrepare(regexp.QuoteMeta(`
		INSERT INTO "Caregiver" (name, role)
		VALUES ($1, $2)
		RETURNING id;
	`))

	mock.ExpectQuery(`INSERT INTO "Caregiver"`).
		WithArgs("a", "helper").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	mock.ExpectQuery(`INSERT INTO "Caregiver"`).
		WithArgs("b", "helper").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))

	mock.ExpectCommit()

	err := svc.CreateCaregiversBulk(list)
	assert.NoError(t, err)
	assert.Equal(t, 10, list[0].ID)
	assert.Equal(t, 11, list[1].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCaregiversBulk_EmptyListError(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)
	err := svc.CreateCaregiversBulk([]model.Caregiver{})

	assert.Error(t, err)
	assert.Equal(t, "no caregivers provided", err.Error())
}

func TestCreateCaregiversBulk_ValidationError(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)

	list := []model.Caregiver{
		{Name: "", Role: "helper"},
	}

	err := svc.CreateCaregiversBulk(list)
	assert.Error(t, err)
	assert.Equal(t, "name cannot be empty in caregiver list", err.Error())
}

/* -------------------------------------------------------
   READ
------------------------------------------------------- */

func TestGetCaregiverByID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)

	rows := sqlmock.NewRows([]string{
		"id", "name", "role",
	}).AddRow(1, "john", "helper")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, role
		FROM "Caregiver"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	cg, err := svc.GetCaregiverByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, cg)
	assert.Equal(t, "john", cg.Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCaregiverByID_InvalidID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)

	cg, err := svc.GetCaregiverByID(0)
	assert.Nil(t, cg)
	assert.Equal(t, "invalid ID", err.Error())
}

// func TestGetCaregiverByCarepartnerID_Success(t *testing.T) {
// 	db, mock := setupMockDB(t)
// 	defer db.Close()

// 	svc := NewCaregiverService(db)

// 	rows := sqlmock.NewRows([]string{
// 		"id", "name", "role",
// 	}).AddRow(2, "mary", "primary")

// 	mock.ExpectQuery(regexp.QuoteMeta(`
// 		SELECT id, name, role
// 		FROM "Caregiver"
// 		WHERE carepartner_id = $1;
// 	`)).
// 		WithArgs(20).
// 		WillReturnRows(rows)

// 	cg, err := svc.GetCaregiverByCarepartnerID(20)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, cg)
// 	assert.Equal(t, "mary", cg.Name)

// 	assert.NoError(t, mock.ExpectationsWereMet())
// }

// func TestGetCaregiverByCarepartnerID_InvalidID(t *testing.T) {
// 	db, _ := setupMockDB(t)
// 	defer db.Close()

// 	svc := NewCaregiverService(db)

// 	cg, err := svc.GetCaregiverByCarepartnerID(0)
// 	assert.Nil(t, cg)
// 	assert.Equal(t, "invalid carereceiver ID", err.Error())
// }

func TestGetAllCaregivers(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)

	rows := sqlmock.NewRows([]string{"id", "name", "role"}).
		AddRow(1, "john", "helper").
		AddRow(2, "mary", "primary")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, role
		FROM "Caregiver";
	`)).
		WillReturnRows(rows)

	list, err := svc.GetAllCaregivers()
	assert.NoError(t, err)
	assert.Len(t, list, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchCaregivers_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)

	rows := sqlmock.NewRows([]string{"id", "name", "role"}).
		AddRow(3, "ann", "helper")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, role
		FROM "Caregiver"
		WHERE name ILIKE $1;
	`)).
		WithArgs("%an%").
		WillReturnRows(rows)

	results, err := svc.SearchCaregivers("an")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "ann", results[0].Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchCaregivers_KeywordTooShort(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)

	results, err := svc.SearchCaregivers("a")
	assert.Nil(t, results)
	assert.Equal(t, "search keyword too short", err.Error())
}

/* -------------------------------------------------------
   UPDATE
------------------------------------------------------- */

func TestUpdateCaregiver_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)

	c := &model.Caregiver{
		ID:            10,
		Name:          "bob",
		Role:          "helper",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "Caregiver"
		SET name = $1, role = $2
		WHERE id = $3;
	`)).
		WithArgs(c.Name, c.Role, c.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.UpdateCaregiver(c)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCaregiver_InvalidID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)
	c := &model.Caregiver{ID: 0, Name: "bob"}

	err := svc.UpdateCaregiver(c)
	assert.Equal(t, "invalid caregiver ID", err.Error())
}

func TestUpdateCaregiver_EmptyName(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)
	c := &model.Caregiver{ID: 5, Name: ""}

	err := svc.UpdateCaregiver(c)
	assert.Equal(t, "name cannot be empty", err.Error())
}

/* -------------------------------------------------------
   DELETE
------------------------------------------------------- */

var (
	deletePairingsSQL  = regexp.QuoteMeta(`DELETE FROM "DevicePairing"`)
	unpairSQL          = regexp.QuoteMeta(`UPDATE "Carepartner" SET caregiver_id = NULL WHERE caregiver_id = $1;`)
	deleteCaregiverSQL = regexp.QuoteMeta(`DELETE FROM "Caregiver" WHERE id = $1;`)
)

func TestDeleteCaregiver_UnpairsCarepartnersFirst(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)

	mock.ExpectBegin()
	mock.ExpectExec(deletePairingsSQL).WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(unpairSQL).WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(deleteCaregiverSQL).WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := svc.DeleteCaregiver(5)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCaregiver_InvalidID(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)
	err := svc.DeleteCaregiver(0)

	assert.Equal(t, "invalid ID", err.Error())
}

func TestDeleteCaregiver_UnpairError_DoesNotDelete(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)

	mock.ExpectBegin()
	mock.ExpectExec(deletePairingsSQL).WithArgs(10).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(unpairSQL).WithArgs(10).WillReturnError(sqlmock.ErrCancelled)
	mock.ExpectRollback()

	err := svc.DeleteCaregiver(10)
	assert.Equal(t, sqlmock.ErrCancelled, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCaregiver_DBDeleteError_RollsBackUnpair(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	svc := NewCaregiverService(db)

	mock.ExpectBegin()
	mock.ExpectExec(deletePairingsSQL).WithArgs(7).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(unpairSQL).WithArgs(7).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(deleteCaregiverSQL).WithArgs(7).WillReturnError(sqlmock.ErrCancelled)
	mock.ExpectRollback()

	err := svc.DeleteCaregiver(7)
	assert.Equal(t, sqlmock.ErrCancelled, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
