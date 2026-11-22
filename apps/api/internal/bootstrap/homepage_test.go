package bootstrap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func homeFixture(t *testing.T) backoffice.Workspace {
	w := backofficememory.Fixture()
	w.Categories = []catalog.Category{}
	w.Products = []catalog.Product{}
	w.Variants = []catalog.Variant{}
	w.Offers = []catalog.Offer{}
	w.IconMedia = []catalog.IconMedia{{ID: "home-photo", Data: iconPNG(t, 8)}}
	w.HomeSettings = backoffice.DefaultHomeSettings()
	w.Stores = append(w.Stores, []store.Store{{ID: "slow", Name: "آزمون کند", Active: true, LeadDays: 3}, {ID: "fast", Name: "آزمون سریع", Active: true, LeadDays: 1}}...)
	for i, r := range w.HomeSettings.Rails {
		id := r.Key
		w.Categories = append(w.Categories, catalog.Category{ID: id, Name: id, Published: true, Version: 1})
		w.HomeSettings.Rails[i].CategoryID = id
		w.Products = append(w.Products, catalog.Product{ID: id, Name: "کالای " + id, CategoryID: id, Status: "published", ImageID: "home-photo", Version: 1})
		w.Variants = append(w.Variants, catalog.Variant{ID: id, ProductID: id, Version: 1})
		w.Offers = append(w.Offers, catalog.Offer{ID: id, ProductID: id, VariantID: id, StoreID: "slow", Stock: 2, Price: 1000, Version: 1})
	}
	return w
}
func homeHTTP(t *testing.T, h http.Handler, tag string) (backoffice.HomePage, *httptest.ResponseRecorder) {
	t.Helper()
	req := httptest.NewRequest("GET", "/storefront/home", nil)
	if tag != "" {
		req.Header.Set("If-None-Match", tag)
	}
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	var body backoffice.HomePage
	if res.Code == 200 {
		if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
	}
	return body, res
}
func TestHomePublicationBuyboxAndProjection(t *testing.T) {
	w := homeFixture(t)
	w.Offers = append(w.Offers, catalog.Offer{ID: "z-fast", ProductID: "mobile", VariantID: "mobile", StoreID: "fast", Stock: 1, Price: 1000}, catalog.Offer{ID: "a-fast", ProductID: "mobile", VariantID: "mobile", StoreID: "fast", Stock: 1, Price: 1000})
	data := backoffice.BuildHome(w)
	if len(data.Rails) != 4 || data.Rails[0].Products[0].OfferID != "a-fast" {
		t.Fatalf("tie selection %+v", data)
	}
	w.Offers[len(w.Offers)-1].Price = 2000
	w.Offers[len(w.Offers)-2].Price = 2000
	if got := backoffice.BuildHome(w).Rails[0].Products[0].OfferID; got != "mobile" {
		t.Fatal("price precedence", got)
	}
	tests := map[string]func(*backoffice.Workspace){"zero stock": func(w *backoffice.Workspace) { w.Offers[0].Stock = 0 }, "hidden ancestor": func(w *backoffice.Workspace) {
		w.Categories[0].ParentID = "hidden-parent"
		w.Categories = append(w.Categories, catalog.Category{ID: "hidden-parent", Published: false})
	}, "draft": func(w *backoffice.Workspace) { w.Products[0].Status = "draft" }, "unpublished category": func(w *backoffice.Workspace) { w.Categories[0].Published = false }, "archived": func(w *backoffice.Workspace) { w.Categories[0].Archived = true }, "no image": func(w *backoffice.Workspace) { w.Products[0].ImageID = "missing" }, "reserved": func(w *backoffice.Workspace) { w.Offers[0].Reserved = 2 }, "inactive seller": func(w *backoffice.Workspace) {
		for i := range w.Stores {
			if w.Stores[i].ID == "slow" {
				w.Stores[i].Active = false
			}
		}
	}, "wrong variant": func(w *backoffice.Workspace) { w.Variants[0].ProductID = "earbuds" }}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			state := homeFixture(t)
			change(&state)
			data := backoffice.BuildHome(state)
			for _, cat := range data.Categories {
				if cat.ID == state.Products[0].CategoryID {
					t.Fatal("unbuyable category shortcut leaked", name)
				}
			}
			if len(data.Rails) != 4 || len(data.Rails[0].Products) != 0 {
				t.Fatalf("invalid product leaked %+v", data.Rails[0])
			}
		})
	}
}
func TestHomeHTTPRevalidationAndPrivateData(t *testing.T) {
	repo := backofficememory.New(homeFixture(t))
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	data, res := homeHTTP(t, h, "")
	if res.Code != 200 || len(data.Rails) != 4 || data.Rails[0].Products[0].URL != "/product/mobile" {
		t.Fatal(res.Code, res.Body.String())
	}
	for _, key := range []string{"users", "phone", "ledger", "fingerprint", "password", "Stock", "base64"} {
		if strings.Contains(res.Body.String(), key) {
			t.Fatal("private field leaked", key)
		}
	}
	tag := res.Header().Get("ETag")
	_, res = homeHTTP(t, h, tag)
	if res.Code != 304 {
		t.Fatal("unchanged", res.Code)
	}
	if err := repo.Update(context.Background(), func(w *backoffice.Workspace) error { w.Offers[0].Price = 1100; return nil }); err != nil {
		t.Fatal(err)
	}
	data, res = homeHTTP(t, h, tag)
	if res.Code != 200 || res.Header().Get("ETag") == tag || data.Rails[0].Products[0].Price != 1100 {
		t.Fatal("stale price")
	}
}
func TestHomeCatalogWritePermissionsValidationAndMedia(t *testing.T) {
	repo := backofficememory.New(homeFixture(t))
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	cmd := backoffice.Command{Action: "catalog.product.update", Target: "mobile", Version: 1, IdempotencyKey: "home-edit", Catalog: &backoffice.CatalogInput{Name: "نام جدید", CategoryID: "mobile", Published: true}}
	call := func(token string, c backoffice.Command, want int) {
		t.Helper()
		b, _ := json.Marshal(c)
		productCall(t, h, "POST", "/backoffice/commands", token, bString(b), want)
	}
	call(viewer, cmd, 403)
	call(admin, cmd, 204)
	call(admin, cmd, 204)
	cmd.IdempotencyKey = "stale"
	call(admin, cmd, 409)
	data, _ := homeHTTP(t, h, "")
	if data.Rails[0].Products[0].Name != "نام جدید" {
		t.Fatal("mutation not public")
	}
	state, _ := repo.Snapshot(context.Background())
	p := state.Products[0]
	if p.Version != 2 {
		t.Fatal("retry duplicated", p.Version)
	}
	cmd.Version = 2
	cmd.IdempotencyKey = "bad"
	cmd.Catalog.CategoryID = "bad"
	call(admin, cmd, 400)
	response := iconUpload(t, h, admin, "product", "mobile", 2, "home-image", iconPNG(t, 16))
	if response.Code != 204 {
		t.Fatal(response.Code, response.Body.String())
	}
	state, _ = repo.Snapshot(context.Background())
	image := state.Products[0].ImageID
	if image == "home-photo" {
		t.Fatal("image not replaced")
	}
	response = iconUpload(t, h, admin, "product", "mobile", 3, "bad-image", []byte("fake"))
	if response.Code != 400 {
		t.Fatal("fake accepted", response.Code)
	}
	state, _ = repo.Snapshot(context.Background())
	if state.Products[0].ImageID != image {
		t.Fatal("invalid image overwrote original")
	}
	settings := state.HomeSettings
	settings.Rails[1].CategoryID = "mobile"
	call(admin, backoffice.Command{Action: "homepage.save", Home: &settings, Version: settings.Version, IdempotencyKey: "duplicate-category"}, 400)
}
func bString(b []byte) string { return string(b) }
func TestLegacyProductPublishEnforcesHomepageMinimum(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*backoffice.Workspace)
		status int
	}{
		{"complete", func(*backoffice.Workspace) {}, 204},
		{"missing name", func(w *backoffice.Workspace) { w.Products[0].Name = " " }, 400},
		{"missing photo", func(w *backoffice.Workspace) { w.Products[0].ImageID = "" }, 400},
		{"unpublished category", func(w *backoffice.Workspace) { w.Categories[0].Published = false }, 400},
		{"inactive brand", func(w *backoffice.Workspace) { w.Products[0].BrandID = "invalid" }, 400},
		{"no available offer", func(w *backoffice.Workspace) { w.Offers[0].Stock = 0 }, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := homeFixture(t)
			w.Products[0].Status = "draft"
			tc.change(&w)
			repo := backofficememory.New(w)
			h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
			admin := login(t, h, "09000000001")
			productCall(t, h, "POST", "/backoffice/commands", admin, backoffice.Command{Action: "product.publish", Target: "mobile", Version: 1, IdempotencyKey: "legacy-publish"}, tc.status)
			state, _ := repo.Snapshot(context.Background())
			if tc.status == 400 && (state.Products[0].Status != "draft" || state.Products[0].Version != 1 || len(state.Audit) != 0) {
				t.Fatal("invalid publication changed state")
			}
		})
	}
}
func TestProductPhotoResizeAndSeparateIconLimits(t *testing.T) {
	repo := backofficememory.New(homeFixture(t))
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	var input bytes.Buffer
	if err := png.Encode(&input, image.NewRGBA(image.Rect(0, 0, 1600, 800))); err != nil {
		t.Fatal(err)
	}
	if r := iconUpload(t, h, admin, "product", "mobile", 1, "resize-photo", input.Bytes()); r.Code != 204 {
		t.Fatal(r.Code, r.Body.String())
	}
	w, _ := repo.Snapshot(context.Background())
	res := httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest("GET", "/catalog/icons/"+w.Products[0].ImageID, nil))
	cfg, format, err := image.DecodeConfig(res.Body)
	if res.Code != 200 || err != nil || format != "png" || cfg.Width != 1024 || cfg.Height != 512 {
		t.Fatalf("resize did not preserve ratio: %d %+v %s %v", res.Code, cfg, format, err)
	}
	if r := iconUpload(t, h, admin, "category", "mobile", 1, "large-icon", input.Bytes()); r.Code != 400 {
		t.Fatal("product policy must not loosen icon dimensions", r.Code)
	}
	if r := iconUpload(t, h, admin, "product", "mobile", 2, "too-many-bytes", make([]byte, backoffice.MaxProductImageBytes+1)); r.Code != 413 {
		t.Fatal("product byte limit", r.Code)
	}
}
func TestHomePostgresRestartRetainsCatalogAndMedia(t *testing.T) {
	dsn := os.Getenv("MARKETPLACE_TEST_POSTGRES_URL")
	if dsn == "" {
		t.Skip("set disposable PostgreSQL URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	control, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer control.Close(context.Background())
	schema := "home_test_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	safe := pgx.Identifier{schema}.Sanitize()
	if _, err = control.Exec(ctx, "CREATE SCHEMA "+safe); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = control.Exec(context.Background(), "DROP SCHEMA "+safe+" CASCADE") }()
	opts, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	opts.ConnConfig.RuntimeParams["search_path"] = schema
	db, err := pgxpool.NewWithConfig(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}
	if err = reviewpostgres.Migrate(ctx, db, os.DirFS(filepath.Join("..", "..", "..", "..", "db", "migrations"))); err != nil {
		t.Fatal(err)
	}
	repo, err := reviewpostgres.New(ctx, db, homeFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Products[0].Name = "پایدار پس از راه‌اندازی"
		w.Offers[0].Price = 1234
		w.HomeSettings.Rails[0].Limit = 2
		w.Categories[0].SEO = catalog.SEO{MetaTitle: "عنوان دسته پایدار", MetaDescription: "متادیتای پایدار"}
		w.Categories[0].Bindings = []catalog.AttributeBinding{{AttributeID: "persisted-memory", Card: true, Filter: true}}
		w.Attributes = append(w.Attributes, catalog.AttributeDefinition{ID: "persisted-memory", Code: "persisted-memory", Name: "حافظه", Kind: "enum", Options: []string{"128"}})
		w.Products[0].Attributes = []catalog.AttributeValue{{AttributeID: "persisted-memory", Value: "128"}}
		w.Products[0].CreatedAt = time.Date(2026, 11, 21, 1, 2, 3, 0, time.UTC)
		grams := 120.5
		w.Products[0].Content = &catalog.ProductContent{ShortDescription: "محتوای پایدار PDP", FullDescription: "<p><strong>متن پایدار</strong></p>", Shipping: catalog.ShippingDimensions{Weight: &grams, WeightUnit: "g", DimensionUnit: "mm"}, SEO: catalog.SEO{MetaTitle: "عنوان محصول پایدار"}, ReviewsEnabled: true}
		w.Products[0].ContentRevision = &catalog.ProductContentRevision{ID: "persisted-review", State: "pending", Name: "عنوان پیشنهادی خصوصی", SubmittedBy: "seller-test", BaseFingerprint: "review-fingerprint", Content: &catalog.ProductContent{ShortDescription: "شرح پیشنهادی خصوصی"}}
		w.Products[0].GalleryIDs = []string{"home-photo"}
		w.Variants[0].ImageID = "home-photo"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	db.Close()
	db, err = pgxpool.NewWithConfig(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo, err = reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if err != nil {
		t.Fatal(err)
	}
	saved, readErr := repo.Snapshot(ctx)
	if readErr != nil || saved.Products[0].ContentRevision == nil || saved.Products[0].ContentRevision.Name != "عنوان پیشنهادی خصوصی" || saved.Products[0].ContentRevision.State != "pending" {
		t.Fatal("pending revision lost after PostgreSQL restart", readErr)
	}
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	data, res := homeHTTP(t, h, "")
	if strings.Contains(res.Body.String(), "عنوان پیشنهادی خصوصی") || strings.Contains(res.Body.String(), "شرح پیشنهادی خصوصی") {
		t.Fatal("pending review leaked to home")
	}
	if res.Code != 200 || data.Rails[0].Products[0].Name != "پایدار پس از راه‌اندازی" || data.Rails[0].Products[0].Price != 1234 {
		t.Fatal("restart lost home", res.Body.String())
	}
	listing, _ := listingRequest(t, h, "?category=mobile&attr.persisted-memory=128", 200)
	if listing.SEO.MetaTitle != "عنوان دسته پایدار" || len(listing.Products) != 1 || len(listing.Products[0].Features) != 1 || listing.Products[0].Features[0].Value != "128" {
		t.Fatal("listing metadata/attributes lost after PostgreSQL restart")
	}
	pdp, _ := pdpRequest(t, h, "/storefront/products/mobile", 200)
	if pdp.Content.ShortDescription != "محتوای پایدار PDP" || pdp.Content.SEO.MetaTitle != "عنوان محصول پایدار" || len(pdp.Gallery) != 1 || pdp.Variants[0].ImageURL == "" || pdp.Content.Shipping.Weight == nil || *pdp.Content.Shipping.Weight != 120.5 {
		t.Fatal("PDP content/media lost after PostgreSQL restart")
	}
	req := httptest.NewRequest("GET", data.Rails[0].Products[0].ImageURL, nil)
	out := httptest.NewRecorder()
	h.ServeHTTP(out, req)
	if out.Code != 200 || len(out.Body.Bytes()) == 0 {
		t.Fatal("image lost after restart")
	}
}

func TestLaunchLoginEmailPasswordAndMobileOTPBothSurfaces(t *testing.T) {
	repo := backofficememory.New(backofficememory.Fixture())
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	productCall(t, h, "PATCH", "/backoffice/account/profile", admin, `{"name":"مدیر آزمون","email":"admin@example.test","password":"test-safe-password-27"}`, 200)
	for _, surface := range []string{"backoffice", "storefront"} {
		h = bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
		for _, identifier := range []string{"09000000001", "ADMIN@example.test"} {
			body := map[string]string{"phone": identifier, "password": "test-safe-password-27"}
			res := productCall(t, h, "POST", "/"+surface+"/auth/password-login", "", body, 200)
			var session map[string]any
			_ = json.Unmarshal(res.Body.Bytes(), &session)
			if session["user"].(map[string]any)["ID"] != "admin" {
				t.Fatal("split identity")
			}
		}
		productCall(t, h, "POST", "/"+surface+"/auth/request-otp", "", map[string]string{"phone": "admin@example.test"}, 401)
		if session := loginAt(t, h, surface, "09000000001"); session["user"].(map[string]any)["ID"] != "admin" {
			t.Fatal("mobile OTP identity")
		}
	}
	productCall(t, h, "POST", "/backoffice/auth/password-login", "", map[string]string{"phone": "admin@example.test", "password": "wrong-password"}, 401)
}
func TestHomeSpecialOffersAndAllAvailableCategories(t *testing.T) {
	w := homeFixture(t)
	w.Offers[0].RegularPrice = 1250
	w.Products[1].Status = "draft"
	w.Categories = append(w.Categories, catalog.Category{ID: "other", Name: "دسته دیگر", Published: true})
	w.Products = append(w.Products, catalog.Product{ID: "other", Name: "کالای دیگر", CategoryID: "other", Status: "published", ImageID: "home-photo"})
	w.Variants = append(w.Variants, catalog.Variant{ID: "other", ProductID: "other"})
	w.Offers = append(w.Offers, catalog.Offer{ID: "other", ProductID: "other", VariantID: "other", StoreID: "fast", Price: 900, Stock: 1})
	got := backoffice.BuildHome(w)
	if len(got.SpecialOffers) != 1 || got.SpecialOffers[0].DiscountPercent != 20 || got.SpecialOffers[0].Price != 1000 || got.SpecialOffers[0].RegularPrice != 1250 {
		t.Fatalf("bad discount %+v", got.SpecialOffers)
	}
	found := false
	for _, c := range got.Categories {
		if c.ID == "earbuds" {
			t.Fatal("draft category shortcut")
		}
		if c.ID == "other" {
			found = true
		}
	}
	if !found {
		t.Fatal("categories limited to four rails")
	}
	w.Offers = append(w.Offers, catalog.Offer{ID: "cheaper-not-discounted", ProductID: "mobile", VariantID: "mobile", StoreID: "fast", Price: 950, Stock: 1})
	if len(backoffice.BuildHome(w).SpecialOffers) != 0 {
		t.Fatal("non-winning offer advertised as discount")
	}
}
