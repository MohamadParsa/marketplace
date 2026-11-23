package bootstrap_test

import (
	"context"
	"encoding/json"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
)

func listingFixture(t *testing.T) backoffice.Workspace {
	w := homeFixture(t)
	w.Categories = append(w.Categories, catalog.Category{ID: "digital", Name: "کالای دیجیتال", Published: true, Version: 1}, catalog.Category{ID: "charger", ParentID: "digital", Name: "شارژر", Published: true, Version: 1})
	w.Categories[0].ParentID = "digital"
	w.Categories[0].Name = "موبایل"
	w.Categories[0].Description = "توضیح دسته موبایل"
	w.Categories[0].Bindings = []catalog.AttributeBinding{{AttributeID: "memory", Card: true, Filter: true, Position: 1}, {AttributeID: "ram", Card: true, Filter: true, Position: 2}, {AttributeID: "camera", Card: true, Position: 3}, {AttributeID: "color", Axis: true, Filter: true, Position: 4}}
	w.Attributes = []catalog.AttributeDefinition{{ID: "memory", Code: "memory", Name: "حافظه", Kind: "enum", Options: []string{"64", "128"}, Version: 1}, {ID: "ram", Code: "ram", Name: "رم", Kind: "number", Unit: "GB", Version: 1}, {ID: "camera", Code: "camera", Name: "دوربین", Kind: "text", Version: 1}, {ID: "color", Code: "color", Name: "رنگ", Kind: "enum", Options: []string{"قرمز", "آبی"}, Version: 1}}
	w.Products[0].Name = "گوشی یک"
	w.Products[0].EnglishName = "Phone One"
	w.Products[0].Attributes = []catalog.AttributeValue{{AttributeID: "memory", Value: "64"}, {AttributeID: "ram", Value: "8"}, {AttributeID: "camera", Value: "48MP"}}
	w.Products[0].CreatedAt = time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	w.Variants[0].Values = []catalog.AttributeValue{{AttributeID: "color", Value: "قرمز"}}
	w.Variants = append(w.Variants, catalog.Variant{ID: "mobile-blue", ProductID: "mobile", Values: []catalog.AttributeValue{{AttributeID: "color", Value: "آبی"}}})
	w.Offers = append(w.Offers, catalog.Offer{ID: "blue-offer", ProductID: "mobile", VariantID: "mobile-blue", StoreID: "fast", Price: 1500, Stock: 2})
	w.Products = append(w.Products, catalog.Product{ID: "unavailable", Name: "گوشی تازه ناموجود", CategoryID: "mobile", Status: "published", ImageID: "home-photo", CreatedAt: time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)}, catalog.Product{ID: "draft", Name: "راز پیش‌نویس", CategoryID: "mobile", Status: "draft", ImageID: "home-photo"}, catalog.Product{ID: "charger", Name: "شارژر آزمون", CategoryID: "charger", Status: "published", ImageID: "home-photo", Attributes: []catalog.AttributeValue{{AttributeID: "memory", Value: "64"}}})
	w.Variants = append(w.Variants, catalog.Variant{ID: "unavailable", ProductID: "unavailable"}, catalog.Variant{ID: "charger", ProductID: "charger"})
	w.Offers = append(w.Offers, catalog.Offer{ID: "charger", VariantID: "charger", ProductID: "charger", StoreID: "fast", Price: 20, Stock: 1})
	return w
}
func listingRequest(t *testing.T, h http.Handler, query string, status int) (backoffice.ListingResult, *httptest.ResponseRecorder) {
	t.Helper()
	r := httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest("GET", "/storefront/catalog"+query, nil))
	if r.Code != status {
		t.Fatalf("listing %s: %d %s", query, r.Code, r.Body.String())
	}
	var data backoffice.ListingResult
	if status == 200 {
		if e := json.Unmarshal(r.Body.Bytes(), &data); e != nil {
			t.Fatal(e)
		}
	}
	return data, r
}
func TestListingHierarchyVisibilityAndNoInheritance(t *testing.T) {
	w := listingFixture(t)
	repo := backofficememory.New(w)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	parent, _ := listingRequest(t, h, "?category=digital", 200)
	if parent.Total != 0 || len(parent.Children) != 2 {
		t.Fatal("parent must show links, not descendants", parent)
	}
	leaf, _ := listingRequest(t, h, "?category=charger", 200)
	if len(leaf.Children) != 0 || len(leaf.Filters) != 0 || len(leaf.Products[0].Features) != 0 {
		t.Fatal("leaf inherited mobile attribute", leaf)
	}
	mobile, _ := listingRequest(t, h, "?category=mobile", 200)
	if len(mobile.Breadcrumbs) != 2 || mobile.Breadcrumbs[0].ID != "digital" || mobile.Breadcrumbs[1].ID != "mobile" {
		t.Fatal("breadcrumb", mobile.Breadcrumbs)
	}
	admin := login(t, h, "09000000001")
	productCall(t, h, "POST", "/backoffice/commands", admin, backoffice.Command{Action: "catalog.category.update", Target: "digital", Version: 1, IdempotencyKey: "hide-parent", Catalog: &backoffice.CatalogInput{Name: "کالای دیجیتال", Published: false}}, 204)
	listingRequest(t, h, "?category=mobile", 404)
	listingRequest(t, h, "?category=charger", 404)
	listingRequest(t, h, "?category=digital", 404)
	data, res := listingRequest(t, h, "", 200)
	if strings.Contains(res.Body.String(), "گوشی") || strings.Contains(res.Body.String(), "راز") || strings.Contains(res.Body.String(), "شارژر") {
		t.Fatal("hidden branch leaked", data)
	}
	state, _ := repo.Snapshot(context.Background())
	if !state.Categories[0].Published || !state.Categories[len(state.Categories)-1].Published {
		t.Fatal("parent mutation rewrote child flags")
	}
}
func TestListingFiltersVariantsCashSortAndUnavailable(t *testing.T) {
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, backofficememory.New(listingFixture(t)))
	data, res := listingRequest(t, h, "?category=mobile", 200)
	if data.Total != 2 || !data.Products[0].Available || data.Products[1].Available || data.Products[1].Price != nil || data.Products[1].OfferID != "" {
		t.Fatal("stock or draft", data.Products)
	}
	if len(data.Products[0].Features) != 3 {
		t.Fatal("card limit")
	}
	for _, secret := range []string{"password", "stock\"", "reserved", "audit", "ledger", "phone", "base64", "راز پیش"} {
		if strings.Contains(strings.ToLower(res.Body.String()), secret) {
			t.Fatal("public DTO leak", secret)
		}
	}
	cases := []struct {
		query string
		count int
		price int64
	}{{"&in_stock=1", 1, 1000}, {"&attr.color=" + url.QueryEscape("آبی"), 1, 1500}, {"&attr.memory=128", 0, 0}, {"&attr.memory=64&attr.ram=8", 1, 1000}, {"&min_price=1200&max_price=1600", 1, 1500}, {"&attr.color=" + url.QueryEscape("قرمز") + "&min_price=1200", 0, 0}, {"&min_price=2000", 0, 0}}
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			d, _ := listingRequest(t, h, "?category=mobile"+tc.query, 200)
			if d.Total != tc.count {
				t.Fatal("count", d.Total, tc.count)
			}
			if tc.count > 0 && (d.Products[0].Price == nil || *d.Products[0].Price != tc.price) {
				t.Fatal("wrong representative variant", d.Products)
			}
		})
	}
	desc, _ := listingRequest(t, h, "?sort=price-desc&page_size=2", 200)
	asc, _ := listingRequest(t, h, "?sort=price-asc&page_size=2", 200)
	if *desc.Products[0].Price < *desc.Products[1].Price || asc.Products[0].ID != "charger" {
		t.Fatal("sort")
	}
	second, _ := listingRequest(t, h, "?sort=price-asc&page_size=2&page=2", 200)
	if second.Total != asc.Total || second.Page != 2 || second.Products[0].ID == asc.Products[0].ID {
		t.Fatal("pagination")
	}
}
func TestListingQueryValidationSearchAndSuggestions(t *testing.T) {
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, backofficememory.New(listingFixture(t)))
	for _, q := range []string{"?sort=unsupported", "?sort=popular", "?page=0", "?page_size=41", "?page=1&page=2", "?min_price=-1", "?min_price=20&max_price=10", "?attr.memory=64", "?category=charger&attr.memory=64", "?category=mobile&attr.secret=1", "?category=mobile&attr.memory=evil", "?category=mobile&attr.ram=NaN", "?in_stock=yes", "?unknown=1", "?q=%zz"} {
		listingRequest(t, h, q, 400)
	}
	for _, q := range []string{"گوشي یک", "PHONE one"} {
		d, _ := listingRequest(t, h, "?q="+url.QueryEscape(q), 200)
		if d.Total != 1 || d.Products[0].ID != "mobile" {
			t.Fatal("normalization", q, d.Products)
		}
	}
	d, _ := listingRequest(t, h, "?q="+url.QueryEscape("%' OR 1=1 --"), 200)
	if d.Total != 0 {
		t.Fatal("search interpreted executable syntax")
	}
	r := httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest("GET", "/storefront/suggestions?q="+url.QueryEscape("گوشی"), nil))
	if r.Code != 200 || !strings.Contains(r.Body.String(), "گوشی یک") || strings.Contains(r.Body.String(), "راز") {
		t.Fatal(r.Code, r.Body.String())
	}
}
func TestListingCategorySEOCardEditorPermissionsAndCache(t *testing.T) {
	repo := backofficememory.New(listingFixture(t))
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	_, before := listingRequest(t, h, "?category=mobile", 200)
	seo := catalog.SEO{MetaTitle: "عنوان اختصاصی موبایل", MetaKeywords: "گوشی,موبایل", MetaDescription: "متن اختصاصی فهرست"}
	cmd := backoffice.Command{Action: "catalog.category.update", Target: "mobile", Version: 1, IdempotencyKey: "category-seo", Catalog: &backoffice.CatalogInput{Name: "موبایل", EnglishName: "Mobile", Description: "توضیح دسته", ParentID: "digital", Published: true, SEO: &seo}}
	productCall(t, h, "POST", "/backoffice/commands", viewer, cmd, 403)
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	cmd.IdempotencyKey = "stale"
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 409)
	after, res := listingRequest(t, h, "?category=mobile", 200)
	if after.SEO != seo || res.Header().Get("ETag") == before.Header().Get("ETag") {
		t.Fatal("SEO/cache")
	}
	state, _ := repo.Snapshot(context.Background())
	if state.Categories[0].ParentID != "digital" || len(state.Categories[0].Bindings) != 4 || state.Offers[0].Price != 1000 {
		t.Fatal("content update corrupted links/offers")
	}
	cmd.Version = 2
	cmd.IdempotencyKey = "invalid-seo"
	cmd.Catalog.SEO = &catalog.SEO{MetaTitle: strings.Repeat("x", 161)}
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 400)
	productCall(t, h, "POST", "/backoffice/commands", viewer, backoffice.Command{Action: "settings.card-limit", Target: "display", Version: state.DisplaySettings.Version, Amount: 1, IdempotencyKey: "viewer-limit"}, 403)
	productCall(t, h, "POST", "/backoffice/commands", admin, backoffice.Command{Action: "settings.card-limit", Target: "display", Version: state.DisplaySettings.Version, Amount: 1, IdempotencyKey: "card-limit"}, 204)
	after, res = listingRequest(t, h, "?category=mobile", 200)
	if len(after.Products[0].Features) != 1 {
		t.Fatal("display setting not consumed")
	}
	req := httptest.NewRequest("GET", "/storefront/catalog?category=mobile", nil)
	req.Header.Set("If-None-Match", res.Header().Get("ETag"))
	out := httptest.NewRecorder()
	h.ServeHTTP(out, req)
	if out.Code != 304 {
		t.Fatal("conditional GET", out.Code)
	}
}
func TestListingInitialHTMLMetadataIsolationAndEscaping(t *testing.T) {
	w := listingFixture(t)
	w.Categories[0].SEO = catalog.SEO{MetaTitle: `موبایل </title><script>alert(1)</script>`, MetaKeywords: `"><img src=x onerror=alert(2)>`, MetaDescription: "توضیح موبایل"}
	dir := t.TempDir()
	template := `<html><head><!--cherriva-page-meta:start--><title>default</title><meta name="description" content="generic"><!--cherriva-page-meta:end--></head><body><div id="root"></div></body></html>`
	if e := os.WriteFile(filepath.Join(dir, "index.html"), []byte(template), 0600); e != nil {
		t.Fatal(e)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test", WebSurface: "storefront", WebRoot: dir}, backofficememory.New(w))
	a := httptest.NewRecorder()
	h.ServeHTTP(a, httptest.NewRequest("GET", "/category/mobile", nil))
	body := a.Body.String()
	if a.Code != 200 || !strings.Contains(body, html.EscapeString(w.Categories[0].SEO.MetaTitle)) || strings.Contains(body, "<script>alert") || strings.Contains(body, "<img src=x") || strings.Count(body, "<title>") != 1 || strings.Count(body, `name="description"`) != 1 {
		t.Fatal("unsafe or duplicate initial metadata", a.Code, body)
	}
	b := httptest.NewRecorder()
	h.ServeHTTP(b, httptest.NewRequest("GET", "/category/charger", nil))
	if b.Code != 200 || strings.Contains(b.Body.String(), "alert(1)") || !strings.Contains(b.Body.String(), "شارژر | Cherriva") {
		t.Fatal("category metadata crossed boundary", b.Body.String())
	}
	missing := httptest.NewRecorder()
	h.ServeHTTP(missing, httptest.NewRequest("GET", "/category/missing", nil))
	if missing.Code != 404 {
		t.Fatal("missing category HTML", missing.Code)
	}
}

func TestListingDoesNotCombineDifferentVariantsAndValidatesAttributeEdits(t *testing.T) {
	w := listingFixture(t)
	w.Attributes = append(w.Attributes, catalog.AttributeDefinition{ID: "warranty", Code: "warranty", Name: "گارانتی", Kind: "enum", Options: []string{"1", "2"}, Version: 1})
	w.Categories[0].Bindings = append(w.Categories[0].Bindings, catalog.AttributeBinding{AttributeID: "warranty", Axis: true, Filter: true})
	for i := range w.Variants {
		if w.Variants[i].ProductID == "mobile" {
			value := "1"
			if w.Variants[i].ID == "mobile-blue" {
				value = "2"
			}
			w.Variants[i].Values = append(w.Variants[i].Values, catalog.AttributeValue{AttributeID: "warranty", Value: value})
		}
	}
	repo := backofficememory.New(w)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	d, _ := listingRequest(t, h, "?category=mobile&attr.color="+url.QueryEscape("قرمز")+"&attr.warranty=2", 200)
	if d.Total != 0 {
		t.Fatal("cross-variant false positive", d.Products)
	}
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	values := []catalog.AttributeValue{{AttributeID: "memory", Value: "128"}, {AttributeID: "ram", Value: "12"}, {AttributeID: "camera", Value: "64MP"}}
	cmd := backoffice.Command{Action: "catalog.product.update", Target: "mobile", Version: 1, IdempotencyKey: "edit-product-attributes", Catalog: &backoffice.CatalogInput{Name: "گوشی یک", CategoryID: "mobile", Published: true, Attributes: &values}}
	productCall(t, h, "POST", "/backoffice/commands", viewer, cmd, 403)
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	d, _ = listingRequest(t, h, "?category=mobile&attr.memory=128", 200)
	if d.Total != 1 || d.Products[0].Features[0].Value != "128" {
		t.Fatal("attribute edit not consumed", d.Products)
	}
	invalid := []catalog.AttributeValue{{AttributeID: "secret", Value: "bad"}}
	cmd.Version = 2
	cmd.IdempotencyKey = "invalid-attributes"
	cmd.Catalog.Attributes = &invalid
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 400)
	state, _ := repo.Snapshot(context.Background())
	if state.Products[0].Attributes[0].Value != "128" || state.Offers[0].Price != 1000 {
		t.Fatal("failed edit changed stored attributes or price")
	}
}

func TestListingPublishedUnavailableContentCanBeEditedWithoutRepublishing(t *testing.T) {
	w := listingFixture(t)
	repo := backofficememory.New(w)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	values := []catalog.AttributeValue{{AttributeID: "memory", Value: "64"}}
	productCall(t, h, "POST", "/backoffice/commands", admin, backoffice.Command{Action: "catalog.product.update", Target: "unavailable", Version: 0, IdempotencyKey: "edit-unavailable-content", Catalog: &backoffice.CatalogInput{Name: "گوشی ناموجود ویرایش‌شده", CategoryID: "mobile", Published: true, Attributes: &values}}, 204)
	d, _ := listingRequest(t, h, "?category=mobile", 200)
	found := false
	for _, p := range d.Products {
		if p.ID == "unavailable" {
			found = true
			if p.Available || p.Price != nil || len(p.Features) != 1 {
				t.Fatal("content edit changed purchasability", p)
			}
		}
	}
	if !found {
		t.Fatal("unavailable publication lost")
	}
}
