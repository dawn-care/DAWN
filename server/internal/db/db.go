package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Open connects using the shared DB_* environment settings and verifies that
// PostgreSQL is reachable. The caller owns the returned pool and must close it.
// It does not run migrations or modify the schema.
func Open(ctx context.Context) (*sql.DB, error) {
	return open(ctx, connectionString(), sql.Open)
}

func connectionString() string {
	settings := []struct {
		key, env, fallback string
	}{
		{"host", "DB_HOST", "localhost"},
		{"port", "DB_PORT", "5432"},
		{"user", "DB_USER", "myuser"},
		{"password", "DB_PASSWORD", "mypassword"},
		{"dbname", "DB_NAME", "mydb"},
	}
	parts := make([]string, 0, len(settings)+1)
	for _, setting := range settings {
		value := os.Getenv(setting.env)
		if value == "" {
			value = setting.fallback
		}
		// lib/pq keyword values escape backslashes and single quotes.
		value = strings.NewReplacer(`\`, `\\`, "'", `\'`).Replace(value)
		parts = append(parts, setting.key+"='"+value+"'")
	}
	return strings.Join(parts, " ") + " sslmode=disable"
}

func open(ctx context.Context, dsn string, openDB func(string, string) (*sql.DB, error)) (*sql.DB, error) {
	database, err := openDB("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return database, nil
}
