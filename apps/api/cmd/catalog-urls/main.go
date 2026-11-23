// catalog-urls reports public URL mappings without modifying the source snapshot.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/postgresrepo"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"time"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, "catalog URL operation failed:", e)
		os.Exit(1)
	}
}
func run() error {
	mode := flag.String("mode", "review", "review or live")
	apply := flag.Bool("apply", false, "apply metadata mappings atomically")
	backup := flag.String("backup", "", "private JSON workspace backup path (required with apply)")
	flag.Parse()
	dsn := os.Getenv("REVIEW_DATABASE_URL")
	if *mode == "live" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if (*mode != "review" && *mode != "live") || dsn == "" || os.Getenv("DATABASE_URL") != "" && os.Getenv("REVIEW_DATABASE_URL") != "" {
		return fmt.Errorf("select exactly one private database environment matching mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	db, e := pgxpool.New(ctx, dsn)
	if e != nil {
		return fmt.Errorf("invalid database configuration")
	}
	defer db.Close()
	var raw []byte
	var origin string
	if e = db.QueryRow(ctx, "SELECT payload,data_origin FROM review_workspace WHERE singleton").Scan(&raw, &origin); e != nil {
		return fmt.Errorf("cannot read target workspace")
	}
	if origin != *mode {
		return fmt.Errorf("database origin mismatch")
	}
	var w backoffice.Workspace
	if e = json.Unmarshal(raw, &w); e != nil {
		return e
	}
	registry := backoffice.CatalogURLRegistry{Products: map[string]int64{}, Categories: map[string]string{}}
	var ready bool
	if e = db.QueryRow(ctx, "SELECT to_regclass('catalog_product_codes') IS NOT NULL").Scan(&ready); e != nil {
		return e
	}
	if ready {
		rows, e := db.Query(ctx, "SELECT product_id,number FROM catalog_product_codes")
		if e != nil {
			return e
		}
		for rows.Next() {
			var id string
			var n int64
			if e = rows.Scan(&id, &n); e != nil {
				rows.Close()
				return e
			}
			registry.Products[id] = n
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return e
		}
		rows, e = db.Query(ctx, "SELECT token,category_id FROM catalog_category_urls")
		if e != nil {
			return e
		}
		for rows.Next() {
			var t, id string
			if e = rows.Scan(&t, &id); e != nil {
				rows.Close()
				return e
			}
			registry.Categories[t] = id
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return e
		}
		if e = db.QueryRow(ctx, "SELECT last_value FROM catalog_product_code_seq").Scan(&registry.LastNumber); e != nil {
			return e
		}
	}
	plan, e := backoffice.PlanCatalogURLs(w, registry)
	if e != nil {
		return e
	}
	already := ready
	productFields := map[string]string{}
	categoryFields := map[string]string{}
	for _, p := range w.Products {
		productFields[p.ID] = p.PublicCode
	}
	for _, c := range w.Categories {
		categoryFields[c.ID] = c.Slug
	}
	for _, p := range plan.Products {
		if productFields[p.ID] != p.Code || registry.Products[p.ID] != p.Number {
			already = false
		}
	}
	for _, c := range plan.Categories {
		if categoryFields[c.ID] != c.Slug || registry.Categories[c.ID] != c.ID || registry.Categories[c.Slug] != c.ID {
			already = false
		}
	}
	if *apply && !already {
		if !ready || *backup == "" {
			return fmt.Errorf("apply requires migration 7 and a private -backup path")
		}
		f, e := os.OpenFile(*backup, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			return fmt.Errorf("backup must be a new writable private file")
		}
		_, e = f.Write(raw)
		ce := f.Close()
		if e != nil {
			return e
		}
		if ce != nil {
			return ce
		}
		if *mode == "review" {
			_, e = postgresrepo.NewReview(ctx, db, backoffice.EmptyWorkspace())
		} else {
			_, e = postgresrepo.NewLive(ctx, db)
		}
		if e != nil {
			return e
		}
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Mode           string                    `json:"mode"`
		AlreadyApplied bool                      `json:"already_applied"`
		Applied        bool                      `json:"applied"`
		Plan           backoffice.CatalogURLPlan `json:"plan"`
	}{*mode, already, *apply, plan})
}
