package migrations

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUnregisteredDatabaseIsNotModified(t *testing.T) {
	for _, tracked := range []bool{false, true} {
		t.Run("tracking_table="+strconv.FormatBool(tracked), func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"tracked", "tables"}).AddRow(tracked, true))
			if tracked {
				mock.ExpectQuery("SELECT COALESCE").WillReturnRows(sqlmock.NewRows([]string{"applied"}).AddRow(false))
			}
			if err := Up(context.Background(), db); !errors.Is(err, ErrBaselineRequired) {
				t.Fatalf("got %v, want baseline required", err)
			}
			// No DDL, seed inserts or version-table writes are expected.
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAllowsFreshAndRegisteredDatabases(t *testing.T) {
	for _, registered := range []bool{false, true} {
		t.Run("registered="+strconv.FormatBool(registered), func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"tracked", "tables"}).AddRow(registered, registered))
			if registered {
				mock.ExpectQuery("SELECT COALESCE").WillReturnRows(sqlmock.NewRows([]string{"applied"}).AddRow(true))
			}
			legacy, err := requiresBaseline(context.Background(), db)
			if err != nil || legacy {
				t.Fatalf("got baseline required=%v error=%v", legacy, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
