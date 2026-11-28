package bootstrap_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func catalogManagementFixture(t *testing.T) backoffice.Workspace {
	w := panelAttributeFixture(t)
	for i := range w.Categories {
		if w.Categories[i].ID == "mobile" {
			w.Categories[i].EnglishName = "Mobile"
			w.Categories[i].Description = "توضیح دسته موبایل"
		}
	}
	for i := 0; i < 45; i++ {
		w.Categories = append(w.Categories, catalog.Category{ID: fmt.Sprintf("filter-cat-%02d", i), Name: fmt.Sprintf("دسته آزمون %02d", i), EnglishName: fmt.Sprintf("Catalog Test %02d", i), ParentID: "digital", Published: i%2 == 0, Archived: i == 44, Version: 1})
	}
	found := false
	for _, m := range w.StoreMemberships {
		found = found || (m.UserID == "ops" && m.StoreID == "s1")
	}
	if !found {
		w.StoreMemberships = append(w.StoreMemberships, identity.StoreMembership{UserID: "ops", StoreID: "s1"})
	}
	return w
}
func TestCatalogAdministrationCannotUseSellerSurface(t *testing.T) {
	ctx := context.Background()
	repo := backofficememory.New(catalogManagementFixture(t))
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	seller := sellerLogin(t, h, "09000000002")
	before, _ := repo.Snapshot(ctx)
	commands := []backoffice.Command{
		{Action: "category.bindings.set", Target: "mobile", Version: 1, Bindings: before.Categories[0].Bindings},
		{Action: "attribute.update", Target: before.Attributes[0].ID, Version: before.Attributes[0].Version, Attribute: &before.Attributes[0]},
		{Action: "catalog.category.update", Target: "mobile", Version: 1, Catalog: &backoffice.CatalogInput{Name: "نام غیرمجاز", EnglishName: "Invalid"}},
		{Action: "catalog.brand.create", Catalog: &backoffice.CatalogInput{Name: "برند غیرمجاز"}},
		{Action: "homepage.save", Target: "home", Version: before.HomeSettings.Version, Home: &before.HomeSettings},
		{Action: "category.archive", Target: "mobile", Reason: "test"},
	}
	for i, c := range commands {
		c.IdempotencyKey = fmt.Sprintf("seller-catalog-denied-%d", i)
		productCall(t, h, "POST", "/seller/commands", seller, c, 403)
	}
	after, _ := repo.Snapshot(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("forbidden seller command changed state")
	}
	// The same account may still manage catalog with its authorized backoffice session.
	ops := login(t, h, "09000000002")
	c := backoffice.Command{Action: "catalog.brand.create", Catalog: &backoffice.CatalogInput{Name: "برند مجاز"}, IdempotencyKey: "backoffice-brand"}
	productCall(t, h, "POST", "/backoffice/commands", ops, c, 204)
}
func TestCategoryBindingSafeFlagsRefreshCachedListingAndPreserveScope(t *testing.T) {
	repo, _ := cachePostgres(t)
	ctx := context.Background()
	seed := catalogManagementFixture(t)
	if e := repo.Update(ctx, func(w *backoffice.Workspace) error { *w = seed; return nil }); e != nil {
		t.Fatal(e)
	}
	cache := backoffice.NewPublicCache(&publicMemoryCache{values: map[string][]byte{}})
	site := bootstrap.NewHTTPHandlerWithPublicCache(config.Config{Environment: "test"}, repo, cache)
	admin := login(t, site, "09000000001")
	viewer := login(t, site, "09000000005")
	_, initial := listingRequest(t, site, "?category=mobile", 200)
	initialTag := initial.Header().Get("ETag")
	if cacheRequest(site, "/storefront/catalog?category=mobile", initialTag).Code != 304 {
		t.Fatal("listing was not cacheable")
	}
	before, _ := repo.Snapshot(ctx)
	c := before.Categories[0]
	bindings := append([]catalog.AttributeBinding{}, c.Bindings...)
	for i := range bindings {
		if bindings[i].AttributeID == "memory" {
			bindings[i].Position = 10
			bindings[i].Filter = false
			bindings[i].Card = false
		}
	}
	command := backoffice.Command{Action: "category.bindings.set", Target: c.ID, Version: c.Version, Bindings: bindings, IdempotencyKey: "safe-display-flags"}
	productCall(t, site, "POST", "/backoffice/commands", viewer, command, 403)
	productCall(t, site, "POST", "/backoffice/commands", admin, command, 204)
	productCall(t, site, "POST", "/backoffice/commands", admin, command, 204)
	after, res := listingRequest(t, site, "?category=mobile", 200)
	if res.Header().Get("ETag") == initialTag || res.Code != 200 {
		t.Fatal("cached listing was not renewed")
	}
	for _, f := range after.Filters {
		if f.ID == backoffice.PublicReference("a", "memory") {
			t.Fatal("removed PLP flag still cached")
		}
	}
	for _, p := range after.Products {
		for _, f := range p.Features {
			if f.AttributeID == backoffice.PublicReference("a", "memory") {
				t.Fatal("removed card flag still cached")
			}
		}
	}
	command.IdempotencyKey = "stale-display-flags"
	productCall(t, site, "POST", "/backoffice/commands", admin, command, 409)
	command.Version++
	command.IdempotencyKey = "invalid-axis-change"
	command.Bindings = append(command.Bindings, catalog.AttributeBinding{AttributeID: "gsp-axis-warranty", Axis: true})
	productCall(t, site, "POST", "/backoffice/commands", admin, command, 409)
	state, _ := repo.Snapshot(ctx)
	if !reflect.DeepEqual(before.Categories[1:], state.Categories[1:]) || !reflect.DeepEqual(before.Attributes, state.Attributes) || !reflect.DeepEqual(before.Products, state.Products) || !reflect.DeepEqual(before.Variants, state.Variants) || !reflect.DeepEqual(before.Offers, state.Offers) || !reflect.DeepEqual(before.Users, state.Users) || !reflect.DeepEqual(before.Orders, state.Orders) || !reflect.DeepEqual(before.Commerce, state.Commerce) {
		t.Fatal("binding changed unrelated data")
	}
	child, _ := listingRequest(t, site, "?category=charger", 200)
	if len(child.Filters) != 0 {
		t.Fatal("child inherited parent attributes")
	}
}
