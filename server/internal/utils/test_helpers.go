package utils

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// setupMockDB initializes a mock database
func SetupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	return db, mock
}
