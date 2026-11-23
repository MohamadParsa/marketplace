package bootstrap_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/reviewpostgres"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/orders"
	"strings"
	"testing"
	"time"
)

func TestBestSellingPostgresInvalidationSettingsPrivacyRestart(t *testing.T) {
	repo, db := cachePostgres(t)
	ctx := context.Background()
	cache := backoffice.NewPublicCache(&publicMemoryCache{values: map[string][]byte{}})
	h := bootstrap.NewHTTPHandlerWithPublicCache(config.Config{Environment: "test"}, repo, cache)
	path := "/storefront/catalog?sort=best-selling"
	if r := cacheRequest(h, path, ""); r.Code != 200 {
		t.Fatal(r.Code, r.Body.String())
	}
	before, _ := repo.CatalogSalesRevision(ctx)
	beforeStamp, _ := repo.CatalogStamp(ctx, "mobile")
	now := time.Now().UTC()
	e := repo.Update(ctx, func(w *backoffice.Workspace) error {
		if w.Commerce == nil {
			w.Commerce = &commerce.State{}
		}
		w.Commerce.Purchases = append(w.Commerce.Purchases, commerce.Purchase{ID: "sales-test", ConfirmedAt: &now, Status: "confirmed"})
		w.Orders = append(w.Orders, orders.Order{ID: "sales-order", CheckoutID: "sales-test", Payment: "paid", Items: []orders.Item{{ID: "x", ProductID: "smartwatch", Quantity: 5}}})
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	after, _ := repo.CatalogSalesRevision(ctx)
	stamp, _ := repo.CatalogStamp(ctx, "mobile")
	if after == before || stamp != beforeStamp {
		t.Fatal("sales version or unrelated PDP invalidation", before, after)
	}
	r := cacheRequest(h, path, "")
	var d backoffice.ListingResult
	if json.Unmarshal(r.Body.Bytes(), &d) != nil || d.Products[0].ID != "smartwatch" {
		t.Fatal("stale best selling", r.Body.String())
	}
	for _, private := range []string{"sales-test", "sales-order", "paid_at", "quantity", "user_id", "sales_read_model"} {
		if strings.Contains(r.Body.String(), private) {
			t.Fatal("private sales input leaked", private)
		}
	}
	if e = repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Commerce.Purchases {
			if w.Commerce.Purchases[i].ID == "sales-test" {
				w.Commerce.Purchases[i].Status = "cancelled"
			}
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	r = cacheRequest(h, path, "")
	json.Unmarshal(r.Body.Bytes(), &d)
	if d.Products[0].ID == "smartwatch" {
		t.Fatal("cancel did not invalidate")
	}
	admin := login(t, h, "09000000001")
	viewer := login(t, h, "09000000005")
	w, _ := repo.Snapshot(ctx)
	cmd := backoffice.Command{Action: "settings.sales-window", Target: "display", Amount: 7, Version: w.DisplaySettings.Version, IdempotencyKey: "sales-window-seven"}
	productCall(t, h, "POST", "/backoffice/commands", viewer, cmd, 403)
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 204)
	cmd.IdempotencyKey = "sales-old-version"
	productCall(t, h, "POST", "/backoffice/commands", admin, cmd, 409)
	s, e := repo.CatalogSales(ctx)
	if e != nil || s.WindowDays != 7 {
		t.Fatal(s, e)
	}
	other, e := reviewpostgres.New(ctx, db, backoffice.Workspace{})
	if e != nil {
		t.Fatal(e)
	}
	restarted, e := other.CatalogSales(ctx)
	if e != nil || restarted.WindowDays != 7 {
		t.Fatal("sales settings restart", restarted, e)
	}
}

type salesTestClock struct{ now time.Time }

func (c *salesTestClock) Now() time.Time { return c.now }

type salesTestIDs struct{}

func (salesTestIDs) Token() (string, error) { return "unused-sales-token", nil }
func TestBestSellingCacheAgesOutWithoutAnyWrite(t *testing.T) {
	repo, _ := cachePostgres(t)
	ctx := context.Background()
	now := time.Date(2026, 11, 23, 10, 0, 0, 0, time.UTC)
	paid := now.Add(-30*24*time.Hour + 30*time.Second)
	e := repo.Update(ctx, func(w *backoffice.Workspace) error {
		if w.Commerce == nil {
			w.Commerce = &commerce.State{}
		}
		w.Commerce.Purchases = append(w.Commerce.Purchases, commerce.Purchase{ID: "clock-sale", Status: "confirmed", ConfirmedAt: &paid})
		w.Orders = append(w.Orders, orders.Order{ID: "clock-order", CheckoutID: "clock-sale", Payment: "paid", Items: []orders.Item{{ID: "x", ProductID: "smartwatch", Quantity: 7}}})
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	clock := &salesTestClock{now}
	s := backoffice.NewService(repo, clock, salesTestIDs{})
	s.UsePublicCache(backoffice.NewPublicCache(&publicMemoryCache{values: map[string][]byte{}}))
	q := backoffice.ListingQuery{Sort: "best-selling", Page: 1, PageSize: 20}
	first, e := s.PublicListing(ctx, q)
	if e != nil || first.Products[0].ID != "smartwatch" {
		t.Fatal(first, e)
	}
	clock.now = now.Add(29 * time.Second)
	hit, e := s.PublicListing(ctx, q)
	if e != nil || hit.Products[0].ID != "smartwatch" {
		t.Fatal(hit, e)
	}
	before, _ := repo.CatalogSalesRevision(ctx)
	clock.now = now.Add(60 * time.Second)
	expired, e := s.PublicListing(ctx, q)
	after, _ := repo.CatalogSalesRevision(ctx)
	if e != nil || expired.Products[0].ID == "smartwatch" || before != after {
		t.Fatal("time-only invalidation failed", e, before, after)
	}
}
