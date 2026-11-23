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
	"strconv"
	"testing"
)

func priorityNumber(n int) *int { return &n }
func TestCategoryPriorityIsSiblingScopedAndStable(t *testing.T) {
	w := listingFixture(t)
	w.Categories = append(w.Categories,
		catalog.Category{ID: "priority-zero", Name: "Z", ParentID: "digital", Published: true, DisplayPriority: priorityNumber(0)},
		catalog.Category{ID: "priority-one-z", Name: "B", ParentID: "digital", Published: true, DisplayPriority: priorityNumber(1)},
		catalog.Category{ID: "priority-one-b", Name: "A", ParentID: "digital", Published: true, DisplayPriority: priorityNumber(1)},
		catalog.Category{ID: "priority-one-a", Name: "A", ParentID: "digital", Published: true, DisplayPriority: priorityNumber(1)},
		catalog.Category{ID: "priority-hidden", Name: "hidden", ParentID: "digital", Published: false, DisplayPriority: priorityNumber(0)})
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, backofficememory.New(w))
	r := cacheRequest(h, "/storefront/navigation", "")
	var rows []backoffice.CatalogNav
	if r.Code != 200 || json.Unmarshal(r.Body.Bytes(), &rows) != nil {
		t.Fatal(r.Code, r.Body.String())
	}
	var siblings, roots []string
	for _, c := range rows {
		if c.ParentID == "digital" {
			siblings = append(siblings, c.ID)
		}
		if c.ParentID == "" {
			roots = append(roots, c.ID)
		}
	}
	want := []string{"priority-zero", "priority-one-a", "priority-one-b", "priority-one-z", "charger", "mobile"}
	if !reflect.DeepEqual(siblings, want) {
		t.Fatal(siblings)
	}
	if roots[0] != "earbuds" {
		t.Fatal("child priority reordered roots", roots)
	}
}
func TestCategoryPriorityPostgresPermissionsLegacyClearCacheRestart(t *testing.T) {
	repo, db := cachePostgres(t)
	ctx := context.Background()
	cache := backoffice.NewPublicCache(&publicMemoryCache{values: map[string][]byte{}})
	h := bootstrap.NewHTTPHandlerWithPublicCache(config.Config{Environment: "test"}, repo, cache)
	before := cacheRequest(h, "/storefront/navigation", "")
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	cmd := backoffice.Command{Action: "catalog.category.update", Target: "smartwatch", Version: 1, IdempotencyKey: "priority-zero", Catalog: &backoffice.CatalogInput{Name: "smartwatch", Published: true, DisplayPriority: priorityNumber(0)}}
	productCall(t, h, "POST", "/backoffice/commands", viewer, cmd, 403)
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	cmd.IdempotencyKey = "priority-stale"
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 409)
	after := cacheRequest(h, "/storefront/navigation", before.Header().Get("ETag"))
	var rows []backoffice.CatalogNav
	json.Unmarshal(after.Body.Bytes(), &rows)
	if after.Code != 200 || len(rows) == 0 || rows[0].ID != "smartwatch" || after.Header().Get("ETag") == before.Header().Get("ETag") {
		t.Fatal("stale navigation", after.Code, after.Body.String())
	}
	// Old clients omit the optional field; updating other category content must retain it.
	cmd.Version = 2
	cmd.IdempotencyKey = "priority-legacy"
	cmd.Catalog.DisplayPriority = nil
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	state, _ := repo.Snapshot(ctx)
	for _, c := range state.Categories {
		if c.ID == "smartwatch" && (c.DisplayPriority == nil || *c.DisplayPriority != 0) {
			t.Fatal("legacy update erased priority")
		}
	}
	cmd.Version = 3
	for _, n := range []int{-1, 1000001} {
		cmd.IdempotencyKey = "priority-invalid-" + strconv.Itoa(n)
		cmd.Catalog.DisplayPriority = priorityNumber(n)
		productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 400)
	}
	cmd.IdempotencyKey = "priority-contradictory"
	cmd.Catalog.DisplayPriority = priorityNumber(2)
	cmd.Catalog.ClearDisplayPriority = true
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 400)
	cmd.IdempotencyKey = "priority-clear"
	cmd.Catalog.DisplayPriority = nil
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	state, _ = repo.Snapshot(ctx)
	for _, c := range state.Categories {
		if c.ID == "smartwatch" && c.DisplayPriority != nil {
			t.Fatal("explicit clear failed")
		}
	}
	cmd.Version = 4
	cmd.IdempotencyKey = "priority-seven"
	cmd.Catalog.ClearDisplayPriority = false
	cmd.Catalog.DisplayPriority = priorityNumber(7)
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	reopened, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	state, _ = reopened.Snapshot(ctx)
	for _, c := range state.Categories {
		if c.ID == "smartwatch" && (c.DisplayPriority == nil || *c.DisplayPriority != 7 || c.Version != 5) {
			t.Fatal("priority lost on restart", c)
		}
	}
	create := backoffice.Command{Action: "catalog.category.create", IdempotencyKey: "new-priority-category", Catalog: &backoffice.CatalogInput{Name: "دسته آزمون اولویت", EnglishName: "Priority Test", Published: false, DisplayPriority: priorityNumber(0)}}
	productCall(t, h, "POST", "/backoffice/commands", admin, create, 204)
	state, _ = repo.Snapshot(ctx)
	found := false
	for _, c := range state.Categories {
		if c.Name == create.Catalog.Name {
			found = true
			if c.DisplayPriority == nil || *c.DisplayPriority != 0 {
				t.Fatal(c)
			}
		}
	}
	if !found {
		t.Fatal("create failed")
	}
}
