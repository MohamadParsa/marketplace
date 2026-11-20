package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/jackc/pgx/v5"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "up" {
		fmt.Fprintln(os.Stderr, "usage: migrate up")
		os.Exit(2)
	}

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	if cfg.ReviewDatabaseURL == "" || (cfg.Environment != "local" && cfg.Environment != "test") {
		return fmt.Errorf("set REVIEW_DATABASE_URL for a disposable local/test PostgreSQL database")
	}
	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "../../db/migrations"
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			dir = "db/migrations"
		}
	}
	options, err := pgx.ParseConfig(cfg.ReviewDatabaseURL)
	if err != nil {
		return fmt.Errorf("invalid REVIEW_DATABASE_URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	db, err := pgx.ConnectConfig(ctx, options)
	if err != nil {
		return err
	}
	defer db.Close(context.Background())
	return reviewpostgres.Migrate(ctx, db, os.DirFS(dir))
}
