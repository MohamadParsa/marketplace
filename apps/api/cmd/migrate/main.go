package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/postgresrepo"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"time"
)

func main() {
	directory := flag.String("dir", "../../db/migrations", "SQL migration directory")
	flag.Parse()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("REVIEW_DATABASE_URL")
	}
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "database configuration required")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	db, e := pgxpool.New(ctx, dsn)
	if e != nil {
		fmt.Fprintln(os.Stderr, "database connection failed")
		os.Exit(1)
	}
	defer db.Close()
	if e = postgresrepo.Migrate(ctx, db, os.DirFS(*directory)); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	fmt.Println("Migrations applied.")
}
