package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/postgresrepo"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestCatalogSnapshotPostgresRoundTrip(t *testing.T) {
	if os.Getenv("MARKETPLACE_TEST_POSTGRES_URL") == "" || os.Getenv("CATALOG_MEDIA_CACHE") == "" || os.Getenv("CATALOG_SEED_FILE") == "" {
		t.Skip("optional real-snapshot integration requires local raster cache and isolated PostgreSQL")
	}
	if e := runCatalogSnapshotRoundTrip(); e != nil {
		t.Fatal(e)
	}
}
func runCatalogSnapshotRoundTrip() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	conn, e := pgx.Connect(ctx, os.Getenv("MARKETPLACE_TEST_POSTGRES_URL"))
	if e != nil {
		return fmt.Errorf("test PG connect failed")
	}
	defer conn.Close(context.Background())
	schema := "catalog_snapshot_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	safe := pgx.Identifier{schema}.Sanitize()
	if _, e = conn.Exec(ctx, "CREATE SCHEMA "+safe); e != nil {
		return e
	}
	defer func() { _, _ = conn.Exec(context.Background(), "DROP SCHEMA "+safe+" CASCADE") }()
	opts, e := pgxpool.ParseConfig(os.Getenv("MARKETPLACE_TEST_POSTGRES_URL"))
	if e != nil {
		return e
	}
	opts.ConnConfig.RuntimeParams["search_path"] = schema
	db, e := pgxpool.NewWithConfig(ctx, opts)
	if e != nil {
		return e
	}
	defer db.Close()
	if e = postgresrepo.Migrate(ctx, db, os.DirFS("../../../../db/migrations")); e != nil {
		return e
	}
	repo, e := postgresrepo.NewReview(ctx, db, backofficememory.Fixture())
	if e != nil {
		return e
	}
	before, _ := repo.Snapshot(ctx)
	b, e := os.ReadFile(os.Getenv("CATALOG_SEED_FILE"))
	if e != nil {
		return e
	}
	seed, e := backoffice.DecodeCatalogSeed(b)
	if e != nil {
		return e
	}
	seed = backoffice.LimitCatalogSeed(seed, 20)
	if e = backoffice.ValidateCatalogSeed(&seed); e != nil {
		return e
	}
	files := []backoffice.CatalogSeedFile{}
	cache := os.Getenv("CATALOG_MEDIA_CACHE")
	normaldir, e := os.MkdirTemp(cache, "sample-normalized-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(normaldir)
	for _, m := range seed.Media {
		h := sha256.Sum256([]byte(m.URL))
		raw, e := os.ReadFile(filepath.Join(cache, hex.EncodeToString(h[:])+".source"))
		if e != nil {
			return fmt.Errorf("cached source missing %s", m.ID)
		}
		sum := sha256.Sum256(raw)
		if hex.EncodeToString(sum[:]) != m.SHA256 {
			return fmt.Errorf("raw media checksum mismatch")
		}
		data, e := backoffice.NormalizedSeedImage(raw)
		if e != nil {
			return fmt.Errorf("invalid media %s", m.ID)
		}
		sum = sha256.Sum256(data)
		path := filepath.Join(normaldir, m.ID+".png")
		if e = os.WriteFile(path, data, 0600); e != nil {
			return e
		}
		files = append(files, backoffice.CatalogSeedFile{ID: m.ID, Path: path, SHA256: hex.EncodeToString(sum[:]), Bytes: int64(len(data))})
	}
	result, e := backoffice.ApplyCatalogSeedFiles(ctx, repo, seed, files, true, time.Now())
	if e != nil {
		return e
	}
	if result.AlreadyApplied {
		return fmt.Errorf("unexpected retry")
	}
	again, e := backoffice.ApplyCatalogSeedFiles(ctx, repo, seed, files, true, time.Now())
	if e != nil || !again.AlreadyApplied {
		return fmt.Errorf("retry failed: %v", e)
	}
	db.Close()
	db, e = pgxpool.NewWithConfig(ctx, opts)
	if e != nil {
		return e
	}
	defer db.Close()
	repo, e = postgresrepo.NewReview(ctx, db, backoffice.Workspace{})
	if e != nil {
		return e
	}
	after, e := repo.Snapshot(ctx)
	if e != nil {
		return e
	}
	if !reflect.DeepEqual(before.Users, after.Users) || !reflect.DeepEqual(before.Access, after.Access) || !reflect.DeepEqual(before.Commerce, after.Commerce) {
		return fmt.Errorf("unrelated data changed")
	}
	for _, m := range after.IconMedia {
		if len(m.Data) > 0 || !m.Stored {
			return fmt.Errorf("inline blob found")
		}
	}
	handler := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	urls := []string{"/storefront/home", "/storefront/catalog?category=" + seed.Catalog.Products[0].CategoryID}
	for _, p := range seed.Catalog.Products {
		urls = append(urls, "/storefront/products/"+p.ID)
	}
	for _, u := range urls {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest("GET", u, nil))
		if rr.Code != 200 {
			return fmt.Errorf("HTTP %d %s", rr.Code, u)
		}
		if len(rr.Body.Bytes()) == 0 {
			return fmt.Errorf("empty response")
		}
	}
	p := seed.Catalog.Products[0]
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/catalog/icons/"+p.ImageID, nil))
	source, e := repo.ReadCatalogMedia(ctx, p.ImageID)
	if e != nil || rr.Code != 200 || !bytes.Equal(source, rr.Body.Bytes()) {
		return fmt.Errorf("image roundtrip failed")
	}
	var inlineCount int
	if e = db.QueryRow(ctx, "SELECT count(*) FROM review_workspace w, jsonb_array_elements(w.payload->'icon_media') m WHERE m->>'data' IS NOT NULL").Scan(&inlineCount); e != nil || inlineCount != 0 {
		return fmt.Errorf("workspace contains binary data")
	}
	b, _ = json.Marshal(map[string]any{"products": result.Products, "variants": result.Variants, "offers": result.Offers, "media": result.Media, "http_checks": len(urls) + 1, "retry": true, "restart": true, "accounts_and_finance_unchanged": true, "workspace_inline_media": inlineCount, "schema_removed_on_exit": true})
	fmt.Println(string(b))
	return nil
}
