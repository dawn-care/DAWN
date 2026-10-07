package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"server/internal/db"
	"server/internal/migrations"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Migration command failed:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: migrate")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	database, err := db.Open(ctx)
	if err != nil {
		return err
	}
	defer database.Close()
	return migrations.Up(ctx, database)
}
