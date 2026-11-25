package bootstrap_test

import (
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
)

// Explicit opt-in, memory-only harness. It never reads a database connection.
func TestAdjustmentsBrowserHarness(t *testing.T) {
	if os.Getenv("CHERRIVA_ADJUSTMENTS_BROWSER") != "1" {
		t.Skip("manual isolated browser harness")
	}
	w := commerceSeed()
	w.IconMedia[0].Data = iconPNG(t, 12)
	w.Attributes = []catalog.AttributeDefinition{
		{ID: "color", Code: "color", Name: "رنگ", Kind: "enum", Options: []string{"مشکی", "سفید"}, Version: 1},
		{ID: "warranty", Code: "warranty", Name: "گارانتی", Kind: "enum", Options: []string{"شرکت الف ۶ماه", "شرکت ب ۶ماه", "گارانتی یک ساله"}, Version: 1},
	}
	w.Categories[0].Bindings = []catalog.AttributeBinding{{AttributeID: "color", Axis: true}, {AttributeID: "warranty", Axis: true}}
	w.Variants[0].Values = []catalog.AttributeValue{{AttributeID: "color", Value: "مشکی"}, {AttributeID: "warranty", Value: "شرکت الف ۶ماه"}}
	w.Variants = append(w.Variants, catalog.Variant{ID: "variant-b", ProductID: "product", Version: 1, Values: []catalog.AttributeValue{{AttributeID: "color", Value: "مشکی"}, {AttributeID: "warranty", Value: "شرکت ب ۶ماه"}}}, catalog.Variant{ID: "variant-white", ProductID: "product", Version: 1, Values: []catalog.AttributeValue{{AttributeID: "color", Value: "سفید"}, {AttributeID: "warranty", Value: "شرکت الف ۶ماه"}}})
	w.Offers = append(w.Offers, catalog.Offer{ID: "offer-b", ProductID: "product", VariantID: "variant-b", StoreID: "s1", Price: 900000, Stock: 10, Version: 1}, catalog.Offer{ID: "offer-white", ProductID: "product", VariantID: "variant-white", StoreID: "s1", Price: 1100000, Stock: 10, Version: 1})
	if os.Getenv("CHERRIVA_WALLET_POLICY_BROWSER") == "1" {
		w.Offers[0].RegularPrice = 1200000
		w.Offers = append(w.Offers, catalog.Offer{ID: "offer-second-seller", ProductID: "product", VariantID: "variant", StoreID: "s2", Price: 1000000, RegularPrice: 1300000, Stock: 10, Version: 1})
	}
	repo := backofficememory.New(w)
	for _, surface := range []struct{ port, name, root string }{{"8096", "backoffice", "backoffice-web"}, {"5196", "storefront", "storefront-web"}} {
		listener, e := net.Listen("tcp", "127.0.0.1:"+surface.port)
		if e != nil {
			t.Fatal(e)
		}
		h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test", WebSurface: surface.name, WebRoot: filepath.Join("..", "..", "..", surface.root, "build")}, repo)
		server := &http.Server{Handler: h, ReadHeaderTimeout: 5 * time.Second}
		t.Cleanup(func() { server.Close() })
		go server.Serve(listener)
	}
	t.Log("isolated browser harness ready: 8096 backoffice, 5196 storefront; memory only")
	time.Sleep(15 * time.Minute)
}
