package bootstrap_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"net/url"
	"testing"
)

func TestListingPriceRangeAllOffersIgnoresPriceAndPagination(t *testing.T) {
	w := listingFixture(t)
	w.Offers = append(w.Offers, catalog.Offer{ID: "range-high", ProductID: "mobile", VariantID: "mobile", StoreID: "slow", Stock: 2, Price: 777777}, catalog.Offer{ID: "range-no-stock", ProductID: "mobile", VariantID: "mobile", StoreID: "slow", Stock: 0, Price: 999999})
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, backofficememory.New(w))
	read := func(path string) backoffice.ListingResult {
		t.Helper()
		r := cacheRequest(h, path, "")
		if r.Code != 200 {
			t.Fatal(r.Code, r.Body.String())
		}
		var d backoffice.ListingResult
		if e := json.Unmarshal(r.Body.Bytes(), &d); e != nil {
			t.Fatal(e)
		}
		return d
	}
	a := read("/storefront/catalog?category=mobile&page_size=1")
	if a.PriceRange.Min == nil || *a.PriceRange.Min != 1000 || *a.PriceRange.Max != 777777 {
		t.Fatal(a.PriceRange)
	}
	b := read("/storefront/catalog?category=mobile&min_price=500000&page_size=1")
	if *b.PriceRange.Min != 1000 || *b.PriceRange.Max != 777777 || b.Total == 0 {
		t.Fatal("price filter collapsed bounds", b.PriceRange, b.Total)
	}
	c := read("/storefront/catalog?category=mobile&min_price=900000")
	if c.Total != 0 || *c.PriceRange.Max != 777777 {
		t.Fatal("empty result cannot be widened", c.Total, c.PriceRange)
	}
}
func TestListingPriceRangeSameVariantAndEmpty(t *testing.T) {
	w := listingFixture(t)
	repo := backofficememory.New(w)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	r := cacheRequest(h, "/storefront/catalog?category=mobile&attr.color="+url.QueryEscape("آبی"), "")
	if r.Code != 200 {
		t.Fatal(r.Code, r.Body.String())
	}
	var d backoffice.ListingResult
	json.Unmarshal(r.Body.Bytes(), &d)
	if d.PriceRange.Min == nil || *d.PriceRange.Min != 1500 || *d.PriceRange.Max != 1500 {
		t.Fatal("wrong variant range", d.PriceRange)
	}
	if e := repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		for i := range w.Offers {
			w.Offers[i].Stock = 0
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	r = cacheRequest(h, "/storefront/catalog?category=mobile", "")
	json.Unmarshal(r.Body.Bytes(), &d)
	if d.PriceRange.Min != nil || d.PriceRange.Max != nil {
		t.Fatal("unavailable offers gave slider bounds", d.PriceRange)
	}
}
