package bootstrap_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"reflect"
	"strings"
	"testing"
)

func brandManagementSeed(t *testing.T) backoffice.Workspace {
	w := commerceSeed()
	w.Orders = backofficememory.Fixture().Orders
	w.Orders[0].Items[0].ProductID = "product"
	w.Brands = []catalog.Brand{{ID: "brand", Name: "برند ساختگی", Active: true, Version: 1}}
	w.Products[0].BrandID = "brand"
	w.IconMedia[0].Data = iconPNG(t, 8)
	return w
}
func TestBrandManagementCommandsValidationPermissionsAndImages(t *testing.T) {
	ctx := context.Background()
	repo := backofficememory.New(brandManagementSeed(t))
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	cmd := backoffice.Command{Action: "catalog.brand.update", Target: "brand", Version: 1, IdempotencyKey: "brand-save", Catalog: &backoffice.CatalogInput{Name: " برند ویرایش‌شده ", EnglishName: " Sample ", Description: " توضیحات ساختگی ", Published: true}}
	commerceHTTP(t, h, "", "", "POST", "/backoffice/commands", cmd, "", 401)
	commerceHTTP(t, h, viewer, "", "POST", "/backoffice/commands", cmd, "", 403)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", cmd, "", 204)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", cmd, "", 204)
	w, _ := repo.Snapshot(ctx)
	b := w.Brands[0]
	if b.Name != "برند ویرایش‌شده" || b.EnglishName != "Sample" || b.Description != "توضیحات ساختگی" || b.Version != 2 {
		t.Fatalf("metadata %#v", b)
	}
	cmd.IdempotencyKey = "stale-brand"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", cmd, "", 409)
	cmd.Version = 2
	cmd.Target = "missing"
	cmd.IdempotencyKey = "missing-brand"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", cmd, "", 404)
	cmd.Target = "brand"
	cmd.Catalog.EnglishName = strings.Repeat("x", 161)
	cmd.IdempotencyKey = "long-brand"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", cmd, "", 400)
	create := backoffice.Command{Action: "catalog.brand.create", IdempotencyKey: "duplicate-brand", Catalog: &backoffice.CatalogInput{Name: "نام دیگر", EnglishName: "sample", Published: true}}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", create, "", 409)
	create.Catalog.EnglishName = "Unique"
	create.Catalog.Name = "برند ویرایش‌شده"
	create.IdempotencyKey = "duplicate-persian"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", create, "", 409)
	create.Catalog.Name = "برند جدید"
	create.IdempotencyKey = "create-brand"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", create, "", 204)
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", create, "", 204)
	w, _ = repo.Snapshot(ctx)
	if len(w.Brands) != 2 {
		t.Fatal("duplicate retry")
	}
	// Inactive brands can prepare media without publishing it.
	cmd.Catalog = &backoffice.CatalogInput{Name: b.Name, EnglishName: b.EnglishName, Description: b.Description, Published: false}
	cmd.IdempotencyKey = "deactivate-brand"
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", cmd, "", 204)
	if r := iconUpload(t, h, admin, "brand", "brand", 3, "brand-image", iconPNG(t, 8)); r.Code != 204 {
		t.Fatalf("inactive image %d %s", r.Code, r.Body.String())
	}
	if r := iconUpload(t, h, admin, "brand", "brand", 4, "bad-brand-image", []byte("not-image")); r.Code != 400 {
		t.Fatal("invalid image accepted")
	}
	w, _ = repo.Snapshot(ctx)
	if w.Brands[0].IconID == "" || w.Brands[0].Version != 4 {
		t.Fatal("image lost")
	}
	if r := cacheRequest(h, "/catalog/icons/"+w.Brands[0].IconID, ""); r.Code != 404 {
		t.Fatal("inactive image public")
	}
	remove := backoffice.Command{Action: "brand.icon.remove", Target: "brand", Version: 4, IdempotencyKey: "remove-brand-icon"}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", remove, "", 204)
	after, _ := repo.Snapshot(ctx)
	if after.Brands[0].IconID != "" {
		t.Fatal("fallback")
	}
	if !reflect.DeepEqual(w.Products, after.Products) || !reflect.DeepEqual(w.Orders, after.Orders) {
		t.Fatal("history changed")
	}
}
func TestBrandMetadataPostgresRestartVisibilityCacheAndHistory(t *testing.T) {
	ctx := context.Background()
	repo, db := cachePostgres(t)
	seed := brandManagementSeed(t)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); e != nil {
		t.Fatal(e)
	}
	before, _ := repo.Snapshot(ctx)
	cache := backoffice.NewPublicCache(&publicMemoryCache{values: map[string][]byte{}})
	h := bootstrap.NewHTTPHandlerWithPublicCache(config.Config{Environment: "test"}, repo, cache)
	admin := login(t, h, "09000000001")
	old := cacheRequest(h, "/storefront/products/product", "")
	if old.Code != 200 {
		t.Fatal("seed hidden")
	}
	c := backoffice.Command{Action: "catalog.brand.update", Target: "brand", Version: 1, IdempotencyKey: "pg-brand", Catalog: &backoffice.CatalogInput{Name: "برند تازه", EnglishName: "Fresh", Description: "توضیح محفوظ", Published: true}}
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", c, "", 204)
	updated := cacheRequest(h, "/storefront/products/product", old.Header().Get("ETag"))
	if updated.Code != 200 || updated.Header().Get("ETag") == old.Header().Get("ETag") || !strings.Contains(updated.Body.String(), "برند تازه") {
		t.Fatal("brand cache stale")
	}
	c.Version = 2
	c.IdempotencyKey = "pg-brand-hide"
	c.Catalog.Published = false
	commerceHTTP(t, h, admin, "", "POST", "/backoffice/commands", c, "", 204)
	if r := cacheRequest(h, "/storefront/products/product", ""); r.Code != 404 {
		t.Fatalf("inactive brand PDP %d", r.Code)
	}
	reopened, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	after, _ := reopened.Snapshot(ctx)
	if after.Brands[0].EnglishName != "Fresh" || after.Brands[0].Description != "توضیح محفوظ" || after.Brands[0].Active {
		t.Fatal("restart metadata")
	}
	for name, pair := range map[string][2]any{"products": {before.Products, after.Products}, "offers": {before.Offers, after.Offers}, "orders": {before.Orders, after.Orders}, "users": {before.Users, after.Users}, "roles": {before.Access, after.Access}, "ledger": {before.Ledger, after.Ledger}, "commerce": {before.Commerce, after.Commerce}} {
		if !reflect.DeepEqual(pair[0], pair[1]) {
			t.Fatal("changed " + name)
		}
	}
	// Optional fields must not rewrite old records with empty values on serialization.
	raw, _ := json.Marshal(catalog.Brand{ID: "legacy", Name: "قدیمی"})
	if strings.Contains(string(raw), "english_name") || strings.Contains(string(raw), "description") {
		t.Fatal("legacy serialization changed")
	}
}
