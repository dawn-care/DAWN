package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // or your database driver
)

var database *sql.DB

// Initialize connects to the database and stores the handle
func Initialize(connStr string) {
	var err error
	database, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	if err := database.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
}

// GetDB returns the singleton *sql.DB
func GetDB() *sql.DB {
	if database == nil {
		log.Fatal("database not initialized. Call db.Initialize() first")
	}
	return database
}

// Close closes the database connection
func Close() {
	if database != nil {
		database.Close()
	}
}
