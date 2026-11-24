package bootstrap_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"testing"
)

func TestPublicCategoryNavigationAncestorsAndCanonical(t *testing.T) {
	w := listingFixture(t)
	w.Categories = append(w.Categories, catalog.Category{ID: "grandchild", Slug: "phone-tools", Name: "زیرشاخه سوم", ParentID: "mobile", Published: true}, catalog.Category{ID: "hidden", Name: "پنهان", ParentID: "digital", Published: false}, catalog.Category{ID: "hidden-child", ParentID: "hidden", Published: true})
	repo := backofficememory.New(w)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	r := cacheRequest(h, "/storefront/navigation", "")
	if r.Code != 200 {
		t.Fatal(r.Code, r.Body.String())
	}
	var rows []backoffice.CatalogNav
	if json.Unmarshal(r.Body.Bytes(), &rows) != nil {
		t.Fatal(r.Body.String())
	}
	seen := map[string]backoffice.CatalogNav{}
	for _, n := range rows {
		seen[n.ID] = n
	}
	if seen[backoffice.PublicReference("c", "grandchild")].ParentID != backoffice.PublicReference("c", "mobile") || seen[backoffice.PublicReference("c", "grandchild")].URL != "/category/phone-tools" || seen[backoffice.PublicReference("c", "digital")].ParentID != "" {
		t.Fatal(seen)
	}
	if _, ok := seen[backoffice.PublicReference("c", "hidden-child")]; ok {
		t.Fatal("hidden ancestor leaked")
	}
	if e := repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		for i := range w.Categories {
			if w.Categories[i].ID == "digital" {
				w.Categories[i].Published = false
			}
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	r = cacheRequest(h, "/storefront/navigation", r.Header().Get("ETag"))
	if r.Code != 200 {
		t.Fatal("stale navigation", r.Code)
	}
	json.Unmarshal(r.Body.Bytes(), &rows)
	for _, n := range rows {
		if n.ID == backoffice.PublicReference("c", "digital") || n.ID == backoffice.PublicReference("c", "mobile") || n.ID == backoffice.PublicReference("c", "grandchild") {
			t.Fatal("unpublished branch visible")
		}
	}
	if r = cacheRequest(h, "/storefront/navigation?unexpected=1", ""); r.Code != 400 {
		t.Fatal(r.Code)
	}
}
