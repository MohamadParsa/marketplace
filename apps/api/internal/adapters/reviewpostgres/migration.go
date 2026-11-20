package reviewpostgres

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"slices"

	"github.com/jackc/pgx/v5"
)

const migrationLock int64 = 634003706638276104

// Migrate applies the immutable up files in filename order under one transaction.
func Migrate(ctx context.Context, db interface {
	Begin(context.Context) (pgx.Tx, error)
}, migrations fs.FS) error {
	names, err := fs.Glob(migrations, "*.up.sql")
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return fmt.Errorf("no .up.sql migrations found")
	}
	slices.Sort(names)
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", migrationLock); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "CREATE TABLE IF NOT EXISTS review_schema_migrations (name text PRIMARY KEY, checksum text NOT NULL)"); err != nil {
		return err
	}
	for _, name := range names {
		sql, err := fs.ReadFile(migrations, name)
		if err != nil {
			return err
		}
		checksum := fmt.Sprintf("%x", sha256.Sum256(sql))
		var applied string
		switch err := tx.QueryRow(ctx, "SELECT checksum FROM review_schema_migrations WHERE name = $1", name).Scan(&applied); err {
		case nil:
			if applied != checksum {
				return fmt.Errorf("migration %s was modified after application", name)
			}
			continue
		case pgx.ErrNoRows:
		default:
			return err
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO review_schema_migrations VALUES ($1, $2)", name, checksum); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
