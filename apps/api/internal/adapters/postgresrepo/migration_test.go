package postgresrepo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestMigrationLineEndingsRetainContentGuard(t *testing.T) {
	lf := []byte("-- schema\nCREATE TABLE unchanged (id integer);\n")
	crlf := []byte(strings.ReplaceAll(string(lf), "\n", "\r\n"))
	for _, original := range [][]byte{lf, crlf} {
		for _, read := range [][]byte{lf, crlf} {
			if !matchesMigrationChecksum(migrationChecksum(original), read) {
				t.Fatal("line ending variation rejected")
			}
		}
	}
	for _, change := range []string{"-- schema\nCREATE TABLE changed (id integer);\n", "-- changed comment\nCREATE TABLE unchanged (id integer);\n", string(lf) + "\n"} {
		if matchesMigrationChecksum(migrationChecksum(crlf), []byte(change)) {
			t.Fatal("content change accepted")
		}
	}
}
func TestMigrationLegacyCRLFRegistryRemainsUnchanged(t *testing.T) {
	dsn := os.Getenv("MARKETPLACE_TEST_POSTGRES_URL")
	if dsn == "" {
		t.Skip("set isolated PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	control, e := pgx.Connect(ctx, dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer control.Close(context.Background())
	schema := "migration_eol_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	safe := pgx.Identifier{schema}.Sanitize()
	if _, e = control.Exec(ctx, "CREATE SCHEMA "+safe); e != nil {
		t.Fatal(e)
	}
	defer func() { _, _ = control.Exec(context.Background(), "DROP SCHEMA "+safe+" CASCADE") }()
	config, e := pgxpool.ParseConfig(dsn)
	if e != nil {
		t.Fatal(e)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	db, e := pgxpool.NewWithConfig(ctx, config)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	lf := []byte("-- original\nCREATE TABLE eol_check(id integer PRIMARY KEY);\n")
	legacy := migrationChecksum([]byte(strings.ReplaceAll(string(lf), "\n", "\r\n")))
	if _, e = db.Exec(ctx, "CREATE TABLE review_schema_migrations(name text PRIMARY KEY,checksum text NOT NULL); CREATE TABLE eol_check(id integer PRIMARY KEY)"); e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(ctx, "INSERT INTO review_schema_migrations VALUES($1,$2)", "000001.up.sql", legacy); e != nil {
		t.Fatal(e)
	}
	fs := fstest.MapFS{"000001.up.sql": &fstest.MapFile{Data: lf}, "000002.up.sql": &fstest.MapFile{Data: []byte("ALTER TABLE eol_check ADD COLUMN name text;\n")}}
	if e = Migrate(ctx, db, fs); e != nil {
		t.Fatal(e)
	}
	var stored string
	_ = db.QueryRow(ctx, "SELECT checksum FROM review_schema_migrations WHERE name='000001.up.sql'").Scan(&stored)
	if stored != legacy {
		t.Fatal("legacy registry rewritten")
	}
	fs["000001.up.sql"] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(string(lf), "id integer", "id bigint"))}
	if e = Migrate(ctx, db, fs); e == nil {
		t.Fatal("modified SQL accepted")
	}
	var n int
	_ = db.QueryRow(ctx, "SELECT count(*) FROM review_schema_migrations").Scan(&n)
	if n != 2 {
		t.Fatal("registry changed after failed migration")
	}
}
