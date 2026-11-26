package bootstrap_test

import (
	"context"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSalesGridBrowserHarness(t *testing.T) {
	if os.Getenv("CHERRIVA_SALES_GRID_BROWSER") != "1" {
		t.Skip("isolated sales/grid browser fixture")
	}
	w := commerceSeed()
	w.Users = append(w.Users, backoffice.User{ID: "inactive-admin-fixture", Name: "ادمین غیرفعال آزمون", Phone: "09120000888", Role: identity.RoleSuperAdmin})
	w.IconMedia[0].Data = iconPNG(t, 12)
	w.Categories = append(w.Categories, catalog.Category{ID: "child", Name: "فرزند دیجیتال", ParentID: "regular", Published: true, Version: 1})
	w.Products[0].CategoryID = "child"
	for i := 0; i < 150; i++ {
		name := fmt.Sprintf("کاربر نمونه %03d", i)
		if i == 100 {
			name = "=SUM(1,1)"
		}
		w.Users = append(w.Users, backoffice.User{ID: fmt.Sprintf("grid-user-%03d", i), Name: name, Phone: fmt.Sprintf("09%09d", 10000+i), Active: true, Role: identity.RoleViewer})
	}
	if os.Getenv("CHERRIVA_PANEL_ATTRIBUTE_BROWSER") == "1" {
		w = panelAttributeFixture(t)
	}
	if os.Getenv("CHERRIVA_CATALOG_MANAGEMENT_BROWSER") == "1" {
		w = catalogManagementFixture(t)
	}
	if os.Getenv("CHERRIVA_AXES_BROWSER") == "1" {
		w = axisConversionSeed(t)
	}
	if os.Getenv("CHERRIVA_STORE_POLICY_BROWSER") == "1" {
		w = storeEligibilitySeed(t)
	}
	if os.Getenv("CHERRIVA_STORES_BROWSER") == "1" {
		for i := 0; i < 45; i++ {
			w.Stores = append(w.Stores, store.Store{ID: fmt.Sprintf("store-browser-%02d", i), Name: fmt.Sprintf("فروشگاه نمونه %02d", i), Active: true, City: "تهران", Profile: &store.Profile{Province: "تهران"}})
		}
	}
	if os.Getenv("CHERRIVA_BRANDS_BROWSER") == "1" {
		w = brandManagementSeed(t)
		for i := 0; i < 45; i++ {
			w.Brands = append(w.Brands, catalog.Brand{ID: fmt.Sprintf("brand-browser-%02d", i), Name: fmt.Sprintf("برند نمونه %02d", i), EnglishName: fmt.Sprintf("Sample %02d", i), Active: i%2 == 0, Version: 1})
		}
	}
	if os.Getenv("CHERRIVA_PRODUCTS_BROWSER") == "1" {
		w = productManagementSeed(t)
		for i := 0; i < 45; i++ {
			active := i%2 == 0
			status := "draft"
			if i%3 == 0 {
				status = "published"
			}
			w.Products = append(w.Products, catalog.Product{ID: fmt.Sprintf("product-browser-%02d", i), Name: fmt.Sprintf("محصول نمونه %02d", i), EnglishName: fmt.Sprintf("Sample Product %02d", i), CategoryID: "gold", Active: &active, Status: status, Model: "shared", Version: 1})
		}
	}
	if os.Getenv("CHERRIVA_BUSINESS_BROWSER") == "1" {
		w = commerceSeed()
		w.IconMedia[0].Data = iconPNG(t, 8)
	}
	if os.Getenv("CHERRIVA_RATINGS_BROWSER") == "1" {
		w = ratingSeed(t)
		for i, rating := range []int{5, 3, 5, 5, 5, 5, 5} {
			w.ProductReviews = append(w.ProductReviews, seededReview(fmt.Sprintf("rating-%d", i), "product", "approved", rating))
		}
		w.ProductReviews = append(w.ProductReviews, seededReview("pending-rating", "product", "pending", 5), seededReview("other-rating", "other-product", "approved", 1))
		w.Products = append(w.Products, catalog.Product{ID: "empty-product", Name: "محصول بدون امتیاز آزمون", CategoryID: "regular", Status: "published", ImageID: "photo", Version: 1})
	}
	var repo backoffice.Repository = backofficememory.New(w)
	if os.Getenv("CHERRIVA_PRODUCTS_BROWSER") == "1" || os.Getenv("CHERRIVA_BUSINESS_BROWSER") == "1" || os.Getenv("CHERRIVA_RATINGS_BROWSER") == "1" {
		persistent, _ := cachePostgres(t)
		if e := persistent.Update(context.Background(), func(state *backoffice.Workspace) error { *state = w; return nil }); e != nil {
			t.Fatal(e)
		}
		repo = persistent
	}
	for _, entry := range []struct{ port, surface, app string }{{"8097", "backoffice", "backoffice-web"}, {"5199", "storefront", "storefront-web"}} {
		listener, e := net.Listen("tcp", "127.0.0.1:"+entry.port)
		if e != nil {
			t.Fatal(e)
		}
		h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test", WebSurface: entry.surface, WebRoot: filepath.Join("..", "..", "..", entry.app, "build")}, repo)
		server := &http.Server{Handler: h, ReadHeaderTimeout: 5 * time.Second}
		t.Cleanup(func() { server.Close() })
		go server.Serve(listener)
	}
	t.Log("sales/grid harness ready:8097 and5199; isolated fixture (products: PostgreSQL schema, others: memory)")
	time.Sleep(15 * time.Minute)
}
