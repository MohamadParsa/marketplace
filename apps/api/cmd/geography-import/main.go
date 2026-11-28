package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/postgresrepo"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(os.Args[1:], os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func targetDSN(mode string, getenv func(string) string) (string, error) {
	live, review := getenv("DATABASE_URL"), getenv("REVIEW_DATABASE_URL")
	if live != "" && review != "" {
		return "", errors.New("select only one target database environment variable")
	}
	if mode == "review" && review != "" {
		return review, nil
	}
	if mode == "live" && live != "" {
		return live, nil
	}
	return "", errors.New("review requires REVIEW_DATABASE_URL; live requires DATABASE_URL")
}

func run(args []string, getenv func(string) string, output io.Writer) error {
	flags := flag.NewFlagSet("geography-import", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	file := flags.String("file", "", "pinned Iran geography JSON file")
	mode := flags.String("mode", "review", "target origin: review or live")
	validateOnly := flags.Bool("validate-only", false, "validate the file without connecting to a database")
	apply := flags.Bool("apply", false, "merge the directory into the selected database")
	backup := flags.String("backup", "", "new JSON backup path; required with -apply")
	if err := flags.Parse(args); err != nil {
		return errors.New("invalid geography-import flags")
	}
	if *file == "" || (*mode != "review" && *mode != "live") || (*apply && (*validateOnly || *backup == "")) || flags.NArg() != 0 {
		return errors.New("provide -file, valid -mode and a new -backup path with -apply; -validate-only cannot apply")
	}
	f, err := os.Open(*file)
	if err != nil {
		return errors.New("cannot open geography seed")
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, backoffice.MaxGeographySeedBytes+1))
	if err != nil {
		return errors.New("cannot read geography seed")
	}
	seed, err := backoffice.DecodeGeographySeed(raw)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(raw)
	report := map[string]any{"dataset_id": seed.DatasetID, "file_sha256": hex.EncodeToString(sum[:]), "mode": *mode, "applied": false, "source_provinces": seed.Source.ProvinceCount, "source_cities": seed.Source.CityCount}
	printReport := func() error { return json.NewEncoder(output).Encode(report) }
	if *validateOnly {
		report["validation"] = "passed"
		return printReport()
	}
	dsn, err := targetDSN(*mode, getenv)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return errors.New("invalid target database configuration")
	}
	defer db.Close()
	var origin string
	var geography []byte
	var version int
	err = db.QueryRow(ctx, `SELECT data_origin, payload->'commerce'->'settings'->'geography', (payload->'commerce'->'settings'->>'version')::int FROM review_workspace WHERE singleton`).Scan(&origin, &geography, &version)
	if err != nil {
		return errors.New("cannot read initialized target geography; start the target app first")
	}
	if origin != *mode {
		return errors.New("target data origin does not match -mode; no data changed")
	}
	var previous []commerce.Province
	if err := json.Unmarshal(geography, &previous); err != nil {
		return errors.New("invalid existing geography")
	}
	_, plan, err := backoffice.MergeGeographySeed(previous, seed)
	if err != nil {
		return err
	}
	plan.SettingsVersion = version
	report["result"] = plan
	if !*apply || plan.AlreadyApplied {
		return printReport()
	}
	previousSnapshot := struct {
		Version   int                 `json:"version"`
		Geography []commerce.Province `json:"geography"`
	}{version, previous}
	encoded, err := json.MarshalIndent(previousSnapshot, "", "  ")
	if err != nil {
		return errors.New("cannot encode previous geography")
	}
	backupFile, err := os.OpenFile(*backup, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return errors.New("cannot create new geography backup; existing files are never overwritten")
	}
	_, writeErr := backupFile.Write(encoded)
	closeErr := backupFile.Close()
	if writeErr != nil || closeErr != nil {
		return errors.New("cannot finish geography backup; no import performed")
	}
	var repo *postgresrepo.Repository
	if *mode == "live" {
		repo, err = postgresrepo.NewLive(ctx, db)
	} else {
		repo, err = postgresrepo.NewReview(ctx, db, backoffice.Workspace{})
	}
	if err != nil {
		return errors.New("cannot open initialized target workspace")
	}
	result, err := backoffice.ApplyGeographySeed(ctx, repo, seed, version, time.Now().UTC())
	if err != nil {
		return err
	}
	report["applied"] = !result.AlreadyApplied
	report["result"] = result
	return printReport()
}
