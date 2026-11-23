// catalog-import validates or atomically appends the public catalog snapshot.
// It never connects to the source SQL Server and has no HTTP write endpoint.
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
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/postgresrepo"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	file := flag.String("file", "", "catalog JSON snapshot")
	mode := flag.String("mode", "review", "target origin: review or live")
	anonymousStores := flag.Bool("anonymous-stores", true, "use fictional Cherriva names for imported test sellers")
	apply := flag.Bool("apply", false, "append to the explicitly selected target database")
	verifyMedia := flag.Bool("verify-media", false, "download/decode media; do not write a database unless -apply is also set")
	configureHome := flag.Bool("configure-home", false, "use snapshot's four category mappings; other home settings and banners stay untouched")
	limit := flag.Int("limit-products", 0, "optional dependency-complete sample; zero selects all")
	cacheDefault, _ := os.UserCacheDir()
	cache := flag.String("cache", filepath.Join(cacheDefault, "cherriva-catalog-media"), "private raster cache directory")
	workers := flag.Int("workers", 4, "parallel image downloads, 1 through 8")
	flag.Parse()
	if *file == "" || (*mode != "review" && *mode != "live") || *limit < 0 || *workers < 1 || *workers > 8 {
		return errors.New("provide -file, valid -mode, nonnegative limit and 1..8 workers")
	}
	f, e := os.Open(*file)
	if e != nil {
		return errors.New("cannot open seed file")
	}
	defer f.Close()
	raw, e := io.ReadAll(io.LimitReader(f, (128<<20)+1))
	if e != nil {
		return errors.New("cannot read seed file")
	}
	seed, e := backoffice.DecodeCatalogSeed(raw)
	if e != nil {
		return e
	}
	seed.AnonymousStoreNames = *anonymousStores
	seed = backoffice.LimitCatalogSeed(seed, *limit)
	if e = backoffice.ValidateCatalogSeed(&seed); e != nil {
		return e
	}
	sum := sha256.Sum256(raw)
	report := map[string]any{"dataset_id": seed.DatasetID, "file_sha256": hex.EncodeToString(sum[:]), "mode": *mode, "applied": false, "products": len(seed.Catalog.Products), "categories": len(seed.Catalog.Categories), "brands": len(seed.Catalog.Brands), "variants": len(seed.Catalog.Variants), "offers": len(seed.Catalog.Offers), "media": len(seed.Media), "configure_home": *configureHome, "anonymous_stores": *anonymousStores}
	printJSON := func() { b, _ := json.Marshal(report); fmt.Println(string(b)) }
	if !*apply && !*verifyMedia {
		report["validation"] = "passed"
		printJSON()
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	media, e := hydrate(ctx, seed.Media, *cache, *workers)
	if e != nil {
		return e
	}
	var size int64
	for _, m := range media {
		size += m.Bytes
	}
	report["normalized_media_bytes"] = size
	if !*apply {
		report["validation"] = "passed_with_media"
		printJSON()
		return nil
	}
	live, review := os.Getenv("DATABASE_URL"), os.Getenv("REVIEW_DATABASE_URL")
	if live != "" && review != "" {
		return errors.New("select only one target database environment variable")
	}
	dsn := review
	if *mode == "live" {
		dsn = live
	}
	if dsn == "" {
		return errors.New("review requires REVIEW_DATABASE_URL; live requires DATABASE_URL")
	}
	db, e := pgxpool.New(ctx, dsn)
	if e != nil {
		return errors.New("invalid target database configuration")
	}
	defer db.Close()
	var origin string
	if e = db.QueryRow(ctx, "SELECT data_origin FROM review_workspace WHERE singleton").Scan(&origin); e != nil {
		return errors.New("target workspace is not initialized; run migrations and start the target app once")
	}
	if origin != *mode {
		return errors.New("target data origin does not match -mode; no data changed")
	}
	var repo *postgresrepo.Repository
	if *mode == "live" {
		repo, e = postgresrepo.NewLive(ctx, db)
	} else {
		repo, e = postgresrepo.NewReview(ctx, db, backoffice.Workspace{})
	}
	if e != nil {
		return errors.New("cannot open initialized target workspace")
	}
	result, e := backoffice.ApplyCatalogSeedFiles(ctx, repo, seed, media, *configureHome, time.Now().UTC())
	if e != nil {
		if strings.HasPrefix(e.Error(), "catalog ID already exists") || strings.HasPrefix(e.Error(), "invalid seed:") {
			return e
		}
		return errors.New("catalog import failed; transaction rolled back")
	}
	if *anonymousStores {
		ids := make([]string, 0, len(seed.Catalog.Stores))
		for _, st := range seed.Catalog.Stores {
			ids = append(ids, st.ID)
		}
		changed, e := backoffice.AnonymizeImportedStores(ctx, repo, ids, time.Now().UTC())
		if e != nil {
			return errors.New("catalog imported; store-name anonymization needs retry")
		}
		report["store_names_changed"] = changed
	}
	report["anonymous_stores"] = *anonymousStores
	report["applied"] = !result.AlreadyApplied
	report["already_applied"] = result.AlreadyApplied
	report["validation"] = "passed_with_media"
	printJSON()
	return nil
}

func hydrate(ctx context.Context, items []backoffice.SeedMedia, cache string, workers int) ([]backoffice.CatalogSeedFile, error) {
	if e := os.MkdirAll(cache, 0700); e != nil {
		return nil, errors.New("cannot create media cache")
	}
	client := &http.Client{Timeout: 25 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) > 3 || req.URL.Scheme != "https" || req.URL.Host != "cdn.gooshishop.ir" || req.URL.User != nil {
			return errors.New("media redirect rejected")
		}
		return nil
	}}
	jobs := make(chan backoffice.SeedMedia)
	out := make([]backoffice.CatalogSeedFile, 0, len(items))
	var mu sync.Mutex
	var first error
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range jobs {
				data, err := loadMedia(ctx, client, m, cache)
				mu.Lock()
				if err != nil {
					if first == nil {
						first = fmt.Errorf("media %s: %w; no database write performed", m.ID, err)
					}
				} else {
					sum := sha256.Sum256(data)
					path := filepath.Join(cache, hex.EncodeToString(sum[:])+".normalized.png")
					if e := os.WriteFile(path, data, 0600); e != nil {
						if first == nil {
							first = errors.New("cannot write normalized image cache")
						}
					} else {
						out = append(out, backoffice.CatalogSeedFile{ID: m.ID, Path: path, SHA256: hex.EncodeToString(sum[:]), Bytes: int64(len(data))})
					}
				}
				if (len(out))%500 == 0 && len(out) > 0 {
					fmt.Fprintf(os.Stderr, "Prepared %d/%d media files\n", len(out), len(items))
				}
				mu.Unlock()
			}
		}()
	}
	for _, m := range items {
		select {
		case jobs <- m:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return nil, errors.New("media preparation timed out")
		}
	}
	close(jobs)
	wg.Wait()
	if first != nil {
		return nil, first
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func loadMedia(ctx context.Context, client *http.Client, m backoffice.SeedMedia, cache string) ([]byte, error) {
	h := sha256.Sum256([]byte(m.URL))
	path := filepath.Join(cache, hex.EncodeToString(h[:])+".source")
	raw, e := os.ReadFile(path)
	if e != nil {
		req, e := http.NewRequestWithContext(ctx, http.MethodGet, m.URL, nil)
		if e != nil {
			return nil, errors.New("invalid image URL")
		}
		req.Header.Set("User-Agent", "Cherriva-Catalog-Import/1.0")
		response, e := client.Do(req)
		if e != nil {
			return nil, errors.New("image request failed")
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d", response.StatusCode)
		}
		raw, e = io.ReadAll(io.LimitReader(response.Body, backoffice.MaxProductImageBytes+1))
		if e != nil {
			return nil, errors.New("cannot read image")
		}
	}
	if len(raw) == 0 || len(raw) > backoffice.MaxProductImageBytes {
		return nil, errors.New("image exceeds size limit")
	}
	sum := sha256.Sum256(raw)
	if m.SHA256 != "" && hex.EncodeToString(sum[:]) != m.SHA256 {
		return nil, errors.New("source image checksum changed; refresh snapshot rather than silently replacing it")
	}
	normalized, e := backoffice.NormalizedSeedImage(raw)
	if e != nil {
		return nil, errors.New("unsupported or invalid raster")
	}
	if e = os.WriteFile(path, raw, 0600); e != nil {
		return nil, errors.New("cannot save private image cache")
	}
	return normalized, nil
}
