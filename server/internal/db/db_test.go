package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestConnectionString(t *testing.T) {
	for _, key := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"} {
		t.Setenv(key, "")
	}
	want := "host='localhost' port='5432' user='myuser' password='mypassword' dbname='mydb' sslmode=disable"
	if got := connectionString(); got != want {
		t.Fatal("empty settings did not use the existing defaults")
	}

	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "test user")
	t.Setenv("DB_PASSWORD", `p'ass\word@?#`)
	t.Setenv("DB_NAME", "test database")
	want = `host='db' port='5433' user='test user' password='p\'ass\\word@?#' dbname='test database' sslmode=disable`
	if got := connectionString(); got != want {
		t.Fatal("environment overrides or credential escaping are incorrect")
	}
}

func TestOpen(t *testing.T) {
	pingErr := errors.New("database unavailable")
	for _, test := range []struct {
		name string
		err  error
	}{
		{"reachable", nil},
		{"unreachable", pingErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			database, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			if err != nil {
				t.Fatal(err)
			}
			defer database.Close()
			mock.ExpectPing().WillReturnError(test.err)
			mock.ExpectClose()
			got, err := open(context.Background(), "test-dsn", func(driver, dsn string) (*sql.DB, error) {
				if driver != "postgres" || dsn != "test-dsn" {
					t.Fatal("incorrect driver or connection settings")
				}
				return database, nil
			})
			if !errors.Is(err, test.err) {
				t.Fatalf("got error %v, want %v", err, test.err)
			}
			if test.err == nil {
				if got != database {
					t.Fatal("did not return the connected pool")
				}
				if err := got.Close(); err != nil {
					t.Fatal(err)
				}
			} else if got != nil {
				t.Fatal("returned a pool after a failed ping")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestOpenFailure(t *testing.T) {
	want := errors.New("cannot open pool")
	got, err := open(context.Background(), "", func(string, string) (*sql.DB, error) {
		return nil, want
	})
	if got != nil || !errors.Is(err, want) {
		t.Fatalf("got (%v, %v), want (nil, wrapped open error)", got, err)
	}
}

func TestOpenCanceledContextClosesPool(t *testing.T) {
	database, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	mock.ExpectClose()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err := open(ctx, "", func(string, string) (*sql.DB, error) {
		return database, nil
	})
	if got != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("got (%v, %v), want (nil, context canceled)", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
