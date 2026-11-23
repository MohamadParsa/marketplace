package bootstrap_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestCatalogURLsPostgresAliasesRestartAndConcurrency(t *testing.T) {
	repo, db := cachePostgres(t)
	ctx := context.Background()
	before, e := repo.Snapshot(ctx)
	if e != nil {
		t.Fatal(e)
	}
	code := before.Products[0].PublicCode
	if !strings.HasPrefix(code, "chr-") {
		t.Fatal(code)
	}
	other, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	after, _ := other.Snapshot(ctx)
	if before.Products[0].PublicCode != after.Products[0].PublicCode {
		t.Fatal("restart changed code")
	}
	if e = repo.Update(ctx, func(w *backoffice.Workspace) error { w.Categories[0].Slug = "phones"; return nil }); e != nil {
		t.Fatal(e)
	}
	if e = other.Update(ctx, func(w *backoffice.Workspace) error {
		w.Categories[0].Slug = "new-phones"
		w.Categories[0].EnglishName = "Renamed English"
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	for _, token := range []string{"mobile", "phones", "new-phones"} {
		id, e := other.ResolveCatalogToken(ctx, "category", token)
		if e != nil || id != "mobile" {
			t.Fatal(token, id, e)
		}
	}
	if e = repo.Update(ctx, func(w *backoffice.Workspace) error { w.Categories[1].Slug = "phones"; return nil }); !errors.Is(e, backoffice.ErrConflict) {
		t.Fatal("alias collision", e)
	}
	stable, _ := repo.Snapshot(ctx)
	if stable.Categories[1].Slug == "phones" {
		t.Fatal("failed write leaked")
	}
	// Two process/repository writers receive distinct durable numbers.
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, id := range []string{"new-one", "new-two"} {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			errs <- repo.Update(ctx, func(w *backoffice.Workspace) error {
				w.Products = append(w.Products, catalog.Product{ID: id, Name: id})
				return nil
			})
		}(id)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	final, _ := repo.Snapshot(ctx)
	seen := map[string]bool{}
	for _, p := range final.Products {
		if seen[p.PublicCode] || p.PublicCode == "" {
			t.Fatal("duplicate code")
		}
		seen[p.PublicCode] = true
	}
	if e = repo.Update(ctx, func(w *backoffice.Workspace) error { w.Products = w.Products[1:]; return nil }); e != nil {
		t.Fatal(e)
	}
	if _, e = repo.ResolveCatalogToken(ctx, "product", code); e != nil {
		t.Fatal("deleted number not reserved", e)
	}
}
func TestCatalogURLHTTPRedirectSEOAndHiddenAliases(t *testing.T) {
	repo, _ := cachePostgres(t)
	ctx := context.Background()
	w, _ := repo.Snapshot(ctx)
	p := w.Products[0]
	root := t.TempDir()
	if e := os.WriteFile(filepath.Join(root, "index.html"), []byte(`<html><head><!--cherriva-page-meta:start--><!--cherriva-page-meta:end--></head><body></body></html>`), 0600); e != nil {
		t.Fatal(e)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test", WebRoot: root, WebSurface: "storefront"}, repo)
	r := cacheRequest(h, "/product/mobile?variant=mobile&offer=mobile", "")
	if r.Code != 301 || r.Header().Get("Location") != "/product/"+p.PublicCode+"?offer=mobile&variant=mobile" {
		t.Fatal(r.Code, r.Header())
	}
	r = cacheRequest(h, r.Header().Get("Location"), "")
	if r.Code != 200 || !strings.Contains(r.Body.String(), `rel="canonical" href="/product/`+p.PublicCode+`"`) {
		t.Fatal(r.Code, r.Body.String())
	}
	r = cacheRequest(h, "/storefront/products/"+p.PublicCode, "")
	if r.Code != 200 {
		t.Fatal(r.Code, r.Body.String())
	}
	var d backoffice.ProductDetail
	json.Unmarshal(r.Body.Bytes(), &d)
	if d.ID != p.ID {
		t.Fatal("internal id renamed")
	}
	productCall(t, h, "POST", "/storefront/products/"+p.PublicCode+"/selection", "", backoffice.ProductSelectionInput{VariantID: "mobile", OfferID: "mobile", Quantity: 1}, 200)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { w.Categories[0].Slug = "phones"; return nil }); e != nil {
		t.Fatal(e)
	}
	r = cacheRequest(h, "/category/mobile?sort=price-asc&category=mobile", "")
	if r.Code != 301 || r.Header().Get("Location") != "/category/phones?sort=price-asc" {
		t.Fatal(r.Code, r.Header())
	}
	r = cacheRequest(h, r.Header().Get("Location"), "")
	if r.Code != 200 || !strings.Contains(r.Body.String(), `rel="canonical" href="/category/phones"`) {
		t.Fatal(r.Code)
	}
	r = cacheRequest(h, "/product/mobile?offer=foreign", "")
	if r.Code != 400 || r.Header().Get("Location") != "" {
		t.Fatal("invalid query redirected", r.Code)
	}
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { w.Categories[0].Published = false; return nil }); e != nil {
		t.Fatal(e)
	}
	for _, path := range []string{"/product/mobile", "/product/" + p.PublicCode, "/category/mobile", "/category/phones"} {
		r = cacheRequest(h, path, "")
		if r.Code != http.StatusNotFound || r.Header().Get("Location") != "" {
			t.Fatal("hidden alias leaks", path, r.Code)
		}
	}
}
func TestCatalogSlugCommandPermissionsConfirmationAndVersion(t *testing.T) {
	repo, _ := cachePostgres(t)
	ctx := context.Background()
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	w, _ := repo.Snapshot(ctx)
	c := w.Categories[0]
	slug := "renamed-mobile"
	command := backoffice.Command{Action: "catalog.category.update", Target: c.ID, Version: c.Version, IdempotencyKey: "url-rename", Catalog: &backoffice.CatalogInput{Name: c.Name, EnglishName: "Mobile", Published: true, Slug: &slug}}
	productCall(t, h, "POST", "/backoffice/commands", viewer, command, 403)
	productCall(t, h, "POST", "/backoffice/commands", admin, command, 400)
	command.Catalog.ConfirmSlugChange = true
	productCall(t, h, "POST", "/backoffice/commands", admin, command, 204)
	productCall(t, h, "POST", "/backoffice/commands", admin, command, 204)
	command.IdempotencyKey = "stale-url"
	productCall(t, h, "POST", "/backoffice/commands", admin, command, 409)
}
