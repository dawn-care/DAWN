package db

import "database/sql"

// QueryRower runs a query that returns one row; *sql.DB and *sql.Tx both satisfy it
type QueryRower interface {
	QueryRow(query string, args ...any) *sql.Row
}

// Execer runs a statement that returns no rows; *sql.DB and *sql.Tx both satisfy it
type Execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

// Scanner reads the current row; *sql.Row and *sql.Rows both satisfy it
type Scanner interface {
	Scan(dest ...any) error
}
