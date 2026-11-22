package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"html"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pdpRequest(t *testing.T, h http.Handler, path string, status int) (backoffice.ProductDetail, *httptest.ResponseRecorder) {
	t.Helper()
	r := httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest("GET", path, nil))
	if r.Code != status {
		t.Fatalf("%s: %d %s", path, r.Code, r.Body.String())
	}
	var d backoffice.ProductDetail
	if status == 200 && strings.HasPrefix(path, "/storefront/") {
		if e := json.Unmarshal(r.Body.Bytes(), &d); e != nil {
			t.Fatal(e)
		}
	}
	return d, r
}
func TestProductDetailSelectionsVisibilityAndPrivacy(t *testing.T) {
	w := listingFixture(t)
	w.Products[0].BrandID = "pdp-brand"
	w.Brands = append(w.Brands, catalog.Brand{ID: "pdp-brand", Name: "برند آزمون", Active: true})
	w.Products[0].Content = &catalog.ProductContent{ShortDescription: "شرح کوتاه", FullDescription: "<p>متن <strong>پررنگ</strong></p>", Advantages: []string{"مزیت"}, ReviewsEnabled: true}
	repo := backofficememory.New(w)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	d, r := pdpRequest(t, h, "/storefront/products/mobile", 200)
	if len(d.Breadcrumbs) != 2 || d.Brand == nil || d.Brand.Name != "برند آزمون" || d.SelectedOfferID != "mobile" || d.Content.ShortDescription != "شرح کوتاه" || d.Interactions.Reviews || !d.CheckoutEnabled {
		t.Fatal("public DTO", r.Body.String())
	}
	for _, private := range []string{`"stock"`, `"reserved"`, `"owner"`, `"users"`, `"icon_media"`, `"password"`, `"sku"`} {
		if strings.Contains(r.Body.String(), private) {
			t.Fatal("private field", private)
		}
	}
	blue, _ := pdpRequest(t, h, "/storefront/products/mobile?variant=mobile-blue&offer=blue-offer", 200)
	if blue.SelectedVariantID != "mobile-blue" || blue.SelectedOfferID != "blue-offer" || *blue.Offers[0].Price != 1500 {
		t.Fatal("selection mismatch")
	}
	for _, q := range []string{"?variant=charger", "?offer=blue-offer", "?variant=mobile-blue&offer=fast", "?q=x", "?variant=mobile-blue&variant=mobile-blue", "?variant=%ZZ"} {
		pdpRequest(t, h, "/storefront/products/mobile"+q, 400)
	}
	for _, id := range []string{"draft", "missing"} {
		pdpRequest(t, h, "/storefront/products/"+id, 404)
	}
	out, _ := pdpRequest(t, h, "/storefront/products/unavailable", 200)
	if out.Content.SEO.MetaTitle == "" || len(out.Offers) != 0 || out.SelectedOfferID != "" {
		t.Fatal("unavailable content")
	}
	repo.Update(context.Background(), func(w *backoffice.Workspace) error { w.Categories[0].Published = false; return nil })
	pdpRequest(t, h, "/storefront/products/mobile", 404)
}
func TestProductSelectionIsReadOnlyAndServerAuthoritative(t *testing.T) {
	w := listingFixture(t)
	repo := backofficememory.New(w)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	input := backoffice.ProductSelectionInput{VariantID: "mobile", OfferID: "mobile", Quantity: 1}
	res := productCall(t, h, "POST", "/storefront/products/mobile/selection", "", input, 200)
	var s backoffice.ProductSelection
	json.Unmarshal(res.Body.Bytes(), &s)
	if s.ProductID != "mobile" || s.OfferID != "mobile" || s.UnitPrice != 1000 {
		t.Fatal(res.Body.String())
	}
	for i := 0; i < 2; i++ {
		productCall(t, h, "POST", "/storefront/products/mobile/selection", "", input, 200)
	}
	after, _ := repo.Snapshot(context.Background())
	if len(after.Orders) != len(w.Orders) || after.Offers[0].Reserved != w.Offers[0].Reserved || len(after.Audit) != len(w.Audit) {
		t.Fatal("selection wrote state")
	}
	input.Quantity = 99
	productCall(t, h, "POST", "/storefront/products/mobile/selection", "", input, 409)
	input.Quantity = 0
	productCall(t, h, "POST", "/storefront/products/mobile/selection", "", input, 400)
	input.Quantity = 1
	input.VariantID = "charger"
	productCall(t, h, "POST", "/storefront/products/mobile/selection", "", input, 400)
	productCall(t, h, "POST", "/storefront/products/draft/selection", "", input, 404)
	productCall(t, h, "POST", "/storefront/products/mobile/selection", "", map[string]any{"variant_id": "mobile", "offer_id": "mobile", "quantity": 1, "unit_price": 1}, 400)
}
func TestProductContentEditorSanitizationPermissionsAndVersion(t *testing.T) {
	w := listingFixture(t)
	repo := backofficememory.New(w)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	p := w.Products[0]
	weight := 123.456
	content := catalog.ProductContent{ShortDescription: "شرح کوتاه", FullDescription: `<h2>عنوان</h2><p><strong>درست</strong><img src=x onerror=alert(1)><script>alert(2)</script><a href="javascript:alert(3)">بد</a><a href="https://example.invalid/info" onclick="alert(4)">لینک سالم</a></p>`, Advantages: []string{"مزیت"}, Disadvantages: []string{"عیب"}, ReviewsEnabled: true, QAEnabled: false, Shipping: catalog.ShippingDimensions{Weight: &weight, WeightUnit: "g", DimensionUnit: "mm"}, SEO: catalog.SEO{MetaTitle: `عنوان </title><script>x</script>`, MetaDescription: "متای محصول"}}
	cmd := backoffice.Command{Action: "catalog.product.update", Target: p.ID, Version: p.Version, IdempotencyKey: "pdp-content", Catalog: &backoffice.CatalogInput{Name: p.Name, EnglishName: "Phone", Description: p.Description, CategoryID: p.CategoryID, BrandID: p.BrandID, Published: true, Content: &content}}
	seller := sellerLogin(t, h, "09120000012")
	productCall(t, h, "POST", "/seller/commands", seller, cmd, 403)
	productCall(t, h, "POST", "/backoffice/commands", viewer, cmd, 403)
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	cmd.IdempotencyKey = "pdp-stale"
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 409)
	after, _ := repo.Snapshot(context.Background())
	saved := after.Products[0].Content
	if saved == nil || !saved.ReviewsEnabled || saved.QAEnabled || !strings.Contains(saved.FullDescription, "<strong>درست</strong>") || !strings.Contains(saved.FullDescription, `href="https://example.invalid/info"`) {
		t.Fatal("formatted content lost")
	}
	for _, unsafe := range []string{"<img", "<script", "javascript:", "onclick", "onerror"} {
		if strings.Contains(saved.FullDescription, unsafe) {
			t.Fatal("unsafe stored HTML", unsafe)
		}
	}
	if after.Offers[0] != w.Offers[0] || after.Offers[1] != w.Offers[1] || after.Products[0].ImageID != p.ImageID {
		t.Fatal("content changed offer/media")
	}
	cmd.Version = after.Products[0].Version
	cmd.IdempotencyKey = "bad-content"
	cmd.Catalog.Content.FullDescription = strings.Repeat("x", 20001)
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 400)
	cmd.Catalog.Content.FullDescription = "<p>محفوظ</p>"
	bad := -1.0
	cmd.Catalog.Content.Shipping.Weight = &bad
	cmd.IdempotencyKey = "bad-weight"
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 400)
	cmd.Catalog.Content.Shipping.Weight = &weight
	cmd.Catalog.Content.Shipping.DimensionUnit = "cm"
	cmd.IdempotencyKey = "bad-unit"
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 400)
	cmd.Catalog.Content = nil
	cmd.IdempotencyKey = "legacy-client"
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	final, _ := repo.Snapshot(context.Background())
	if final.Products[0].Content.FullDescription != saved.FullDescription {
		t.Fatal("legacy client erased content")
	}
}
func TestProductGalleryOrderingVariantImageAndIsolation(t *testing.T) {
	w := listingFixture(t)
	repo := backofficememory.New(w)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	p := w.Products[0]
	image := iconPNG(t, 8)
	if r := iconUpload(t, h, viewer, "product-gallery", p.ID, p.Version, "viewer-gallery", image); r.Code != 403 {
		t.Fatal(r.Code)
	}
	for i, key := range []string{"gallery-1", "gallery-2"} {
		r := iconUpload(t, h, admin, "product-gallery", p.ID, p.Version+i, key, image)
		if r.Code != 204 {
			t.Fatal(r.Code, r.Body.String())
		}
		r = iconUpload(t, h, admin, "product-gallery", p.ID, p.Version+i, key, image)
		if r.Code != 204 {
			t.Fatal("retry", r.Code)
		}
	}
	state, _ := repo.Snapshot(context.Background())
	ids := state.Products[0].GalleryIDs
	if len(ids) != 2 {
		t.Fatal("gallery duplicated")
	}
	cmd := backoffice.Command{Action: "product.media.update", Target: p.ID, Version: state.Products[0].Version, IdempotencyKey: "order-gallery", ProductMedia: &backoffice.ProductMediaInput{GalleryIDs: []string{ids[1], ids[0]}, VariantImages: map[string]string{"mobile-blue": ids[1]}}}
	productCall(t, h, "POST", "/backoffice/commands", viewer, cmd, 403)
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	d, _ := pdpRequest(t, h, "/storefront/products/mobile?variant=mobile-blue", 200)
	if !strings.HasSuffix(d.Gallery[0], ids[1]) {
		t.Fatal("variant image not first")
	}
	state, _ = repo.Snapshot(context.Background())
	cmd.Version = state.Products[0].Version
	cmd.IdempotencyKey = "foreign-variant"
	cmd.ProductMedia.VariantImages = map[string]string{"charger": ids[0]}
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 400)
	cmd.IdempotencyKey = "remove-gallery"
	cmd.ProductMedia.GalleryIDs = []string{}
	cmd.ProductMedia.VariantImages = map[string]string{}
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	r := httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest("GET", "/catalog/icons/"+ids[1], nil))
	if r.Code != 200 {
		t.Fatal("referenced variant file deleted")
	}
	r = httptest.NewRecorder()
	h.ServeHTTP(r, httptest.NewRequest("GET", "/catalog/icons/"+ids[0], nil))
	if r.Code != 404 {
		t.Fatal("orphan gallery file public")
	}
}
func TestProductInitialHTMLRobotsIsolationAndCache(t *testing.T) {
	for _, flags := range []struct {
		index, follow bool
		want          string
	}{{false, false, "index, follow"}, {true, false, "noindex, follow"}, {false, true, "index, nofollow"}, {true, true, "noindex, nofollow"}} {
		t.Run(flags.want, func(t *testing.T) {
			w := listingFixture(t)
			w.Products[0].Content = &catalog.ProductContent{FullDescription: `<img onerror=x><p>متن</p>`, NoIndex: flags.index, NoFollow: flags.follow, SEO: catalog.SEO{MetaTitle: `محصول </title><script>x</script>`, MetaDescription: `" onload="x`, MetaKeywords: "موبایل"}}
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><head><meta name="robots" content="noindex,nofollow"><!--cherriva-page-meta:start--><title>default</title><!--cherriva-page-meta:end--></head></html>`), 0600)
			repo := backofficememory.New(w)
			h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "production", WebRoot: dir, WebSurface: "storefront"}, repo)
			_, r := pdpRequest(t, h, "/product/mobile", 200)
			body := r.Body.String()
			if !strings.Contains(body, html.EscapeString(w.Products[0].Content.SEO.MetaTitle)) || strings.Contains(body, "<script>") || strings.Count(body, `name="robots"`) != 1 || !strings.Contains(body, `content="`+flags.want+`"`) {
				t.Fatal("unsafe/incorrect metadata", body)
			}
			_, other := pdpRequest(t, h, "/product/charger", 200)
			if strings.Contains(other.Body.String(), "onload") || strings.Contains(other.Body.String(), "&lt;script&gt;") {
				t.Fatal("SEO leaked to other product")
			}
			_, before := pdpRequest(t, h, "/storefront/products/mobile", 200)
			req := httptest.NewRequest("GET", "/storefront/products/mobile", nil)
			req.Header.Set("If-None-Match", before.Header().Get("ETag"))
			cached := httptest.NewRecorder()
			h.ServeHTTP(cached, req)
			if cached.Code != 304 {
				t.Fatal("ETag")
			}
			repo.Update(context.Background(), func(w *backoffice.Workspace) error { w.Offers[0].Price++; return nil })
			_, after := pdpRequest(t, h, "/storefront/products/mobile", 200)
			if after.Header().Get("ETag") == before.Header().Get("ETag") {
				t.Fatal("offer cache stale")
			}
			review := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test", WebRoot: dir, WebSurface: "storefront"}, repo)
			_, rr := pdpRequest(t, review, "/product/mobile", 200)
			if !strings.Contains(rr.Body.String(), `content="noindex, nofollow"`) {
				t.Fatal("review indexable")
			}
			pdpRequest(t, h, "/product/draft", 404)
		})
	}
}

func TestProductGalleryBoundsAndForeignMedia(t *testing.T) {
	w := listingFixture(t)
	w.IconMedia = append(w.IconMedia, catalog.IconMedia{ID: "foreign-photo", Data: iconPNG(t, 4)})
	w.Products[1].ImageID = "foreign-photo"
	repo := backofficememory.New(w)
	h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test"}, repo)
	admin := login(t, h, "09000000001")
	productCall(t, h, "POST", "/backoffice/commands", admin, backoffice.Command{Action: "product.media.update", Target: "mobile", Version: 1, IdempotencyKey: "foreign-image", ProductMedia: &backoffice.ProductMediaInput{GalleryIDs: []string{"foreign-photo"}}}, 400)
	if r := iconUpload(t, h, admin, "product-gallery", "mobile", 1, "large-gallery", make([]byte, backoffice.MaxProductImageBytes+1)); r.Code != 413 {
		t.Fatal("gallery size bound", r.Code)
	}
	repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		for i := 0; i < backoffice.MaxGalleryImages; i++ {
			id := fmt.Sprintf("gallery-%d", i)
			w.Products[0].GalleryIDs = append(w.Products[0].GalleryIDs, id)
			w.IconMedia = append(w.IconMedia, catalog.IconMedia{ID: id, Data: iconPNG(t, 4)})
		}
		return nil
	})
	if r := iconUpload(t, h, admin, "product-gallery", "mobile", 1, "gallery-overflow", iconPNG(t, 4)); r.Code != 400 {
		t.Fatal("gallery count bound", r.Code)
	}
	state, _ := repo.Snapshot(context.Background())
	if len(state.Products[0].GalleryIDs) != backoffice.MaxGalleryImages || state.Products[0].Version != 1 {
		t.Fatal("invalid upload mutated product")
	}
}
