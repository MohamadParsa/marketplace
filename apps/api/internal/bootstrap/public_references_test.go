package bootstrap_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/content"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestPublicReferencesCartLegacyRetryMediaAndPrivateState(t *testing.T) {
	repo := backofficememory.New(listingFixture(t))
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	before, _ := repo.Snapshot(context.Background())
	d, r := pdpRequest(t, h, "/storefront/products/mobile?variant=mobile-blue&offer=blue-offer", 200)
	if d.SelectedVariantID != backoffice.PublicReference("v", "mobile-blue") || d.SelectedOfferID != backoffice.PublicReference("o", "blue-offer") {
		t.Fatal(r.Body.String())
	}
	// Alias input and legacy input choose the exact same offer, amount and media.
	modern, mr := pdpRequest(t, h, "/storefront/products/"+d.ID+"?variant="+d.SelectedVariantID+"&offer="+d.SelectedOfferID, 200)
	if !reflect.DeepEqual(d, modern) || r.Header().Get("ETag") != mr.Header().Get("ETag") {
		t.Fatal("alias selection or cache identity changed")
	}
	productCall(t, h, "POST", "/storefront/products/"+d.ID+"/selection", "", backoffice.ProductSelectionInput{VariantID: d.SelectedVariantID, OfferID: d.SelectedOfferID, Quantity: 1}, 200)
	// A syntactically valid alias from another product must still fail membership checks.
	pdpRequest(t, h, "/storefront/products/mobile?variant="+backoffice.PublicReference("v", "charger"), 400)
	pdpRequest(t, h, "/storefront/products/mobile?variant=chr-v-!", 400)
	for _, src := range modern.Gallery {
		if strings.Contains(src, "gsp-") {
			t.Fatal("source media ID exposed", src)
		}
		image := cacheRequest(h, src, "")
		if image.Code != 200 {
			t.Fatal("aliased media missing", image.Code)
		}
	}
	var guest map[string]any
	json.Unmarshal(commerceHTTP(t, h, "", "", "POST", "/storefront/cart/guest", nil, "", 200), &guest)
	token := guest["cart_token"].(string)
	var cart backoffice.CartView
	json.Unmarshal(commerceHTTP(t, h, "", token, "GET", "/storefront/cart", nil, "", 200), &cart)
	input := backoffice.CartMutation{Action: "add", OfferID: d.SelectedOfferID, Quantity: 1, Version: cart.Version}
	first := commerceHTTP(t, h, "", token, "POST", "/storefront/cart", input, "same-cart-reference", 200)
	input.OfferID = "blue-offer"
	retry := commerceHTTP(t, h, "", token, "POST", "/storefront/cart", input, "same-cart-reference", 200)
	if string(first) != string(retry) {
		t.Fatal("legacy/alias retry changed idempotency")
	}
	json.Unmarshal(retry, &cart)
	if len(cart.Items) != 1 || cart.Items[0].Quantity != 1 || cart.Items[0].OfferID != d.SelectedOfferID {
		t.Fatal(string(retry))
	}
	after, _ := repo.Snapshot(context.Background())
	if !reflect.DeepEqual(before.Products, after.Products) || !reflect.DeepEqual(before.Variants, after.Variants) || !reflect.DeepEqual(before.Offers, after.Offers) || !reflect.DeepEqual(before.Orders, after.Orders) || !reflect.DeepEqual(before.Users, after.Users) {
		t.Fatal("catalog or protected state rewritten")
	}
	if after.Commerce.Carts[len(after.Commerce.Carts)-1].Lines[0].OfferID != "blue-offer" {
		t.Fatal("alias was stored in private cart")
	}
}
func TestPublicReferencePostgresRestartAndETag(t *testing.T) {
	repo, db := cachePostgres(t)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	d, r := pdpRequest(t, h, "/storefront/products/mobile", 200)
	etag := r.Header().Get("ETag")
	if cacheRequest(h, "/storefront/products/"+d.ID, etag).Code != 304 {
		t.Fatal("public reference ETag not reusable")
	}
	// New service/router after restart derives the same tokens without a mapping migration.
	again, err := reviewpostgres.New(context.Background(), db, backoffice.Workspace{})
	if err != nil {
		t.Fatal(err)
	}
	reopened := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, again)
	other, _ := pdpRequest(t, reopened, "/storefront/products/"+d.ID+"?variant="+d.SelectedVariantID+"&offer="+d.SelectedOfferID, 200)
	if !reflect.DeepEqual(d, other) {
		t.Fatal("tokens or values changed after reopen")
	}
}

func TestPublicReferencesWishlistAndRecentKeepPrivateIDs(t *testing.T) {
	repo := backofficememory.New(listingFixture(t))
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	buyer := loginAt(t, h, "storefront", "09000000003")["token"].(string)
	d, _ := pdpRequest(t, h, "/storefront/products/mobile", 200)
	commerceHTTP(t, h, buyer, "", "PUT", "/storefront/account/wishlist/"+d.ID, nil, "", 200)
	commerceHTTP(t, h, buyer, "", "POST", "/storefront/account/recently-viewed/"+d.ID, nil, "", 200)
	wishlist := commerceHTTP(t, h, buyer, "", "GET", "/storefront/account/wishlist", nil, "", 200)
	recent := commerceHTTP(t, h, buyer, "", "GET", "/storefront/account/recently-viewed", nil, "", 200)
	if !strings.Contains(string(wishlist), d.ID) || !strings.Contains(string(recent), d.ID) {
		t.Fatal("personal list lost alias")
	}
	state, _ := repo.Snapshot(context.Background())
	if len(state.Wishlist) != 1 || state.Wishlist[0].ProductID != "mobile" || len(state.RecentlyViewed) != 1 || state.RecentlyViewed[0].ProductID != "mobile" {
		t.Fatal("personal list stored alias")
	}
	commerceHTTP(t, h, buyer, "", "DELETE", "/storefront/account/wishlist/mobile", nil, "", 200)
	commerceHTTP(t, h, buyer, "", "DELETE", "/storefront/account/recently-viewed/mobile", nil, "", 200)
}

func TestPublicBannerReferencesMatchHomeAndKeepPreviewImages(t *testing.T) {
	w := homeFixture(t)
	w.Banners = append(w.Banners, content.Banner{ID: "reference-promo", Name: "test promo", Published: &content.PublishedBanner{Revision: 1, Config: content.BannerConfig{Enabled: true, Placement: "promo", CatalogCategoryID: "mobile", Alt: "موبایل"}}})
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, backofficememory.New(w))
	r := cacheRequest(h, "/content/banners", "")
	var banners []content.PublicBanner
	if r.Code != 200 || json.Unmarshal(r.Body.Bytes(), &banners) != nil || len(banners) != 1 {
		t.Fatal(r.Code, r.Body.String())
	}
	var home backoffice.HomePage
	r = cacheRequest(h, "/storefront/home", "")
	json.Unmarshal(r.Body.Bytes(), &home)
	if home.Rails[0].CategoryID != banners[0].Config.CatalogCategoryID {
		t.Fatal("promo and rail references diverged")
	}
	r = cacheRequest(h, "/content/banners/reference-promo/view", "")
	if r.Code != 200 || !strings.Contains(r.Body.String(), "/catalog/icons/chr-m-") {
		t.Fatal(r.Code, r.Body.String())
	}
}

func TestPublicReferenceListingFiltersMatchLegacy(t *testing.T) {
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, backofficememory.New(listingFixture(t)))
	old, _ := listingRequest(t, h, "?category=mobile&attr.color="+url.QueryEscape("آبی"), 200)
	modern, _ := listingRequest(t, h, "?category="+backoffice.PublicReference("c", "mobile")+"&attr."+backoffice.PublicReference("a", "color")+"="+url.QueryEscape("آبی"), 200)
	if !reflect.DeepEqual(old, modern) || modern.Total != 1 || *modern.Products[0].Price != 1500 {
		t.Fatal("aliased PLP filter changed results")
	}
}
