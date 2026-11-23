package backoffice

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/finance"
	"github.com/cherriva/marketplace/apps/api/internal/domain/orders"
	"testing"
	"time"
)

func TestCatalogSalesOnlyPaidNetUnitsAndRollingBoundary(t *testing.T) {
	now := time.Date(2026, 11, 23, 10, 0, 0, 0, time.UTC)
	paid := now.Add(-time.Hour)
	w := Workspace{Commerce: &commerce.State{Purchases: []commerce.Purchase{{ID: "purchase", ConfirmedAt: &paid, Status: "confirmed"}, {ID: "unpaid", Status: "awaiting-payment"}}}, Orders: []orders.Order{
		{ID: "a", CheckoutID: "purchase", Payment: "paid", Items: []orders.Item{{ID: "one", ProductID: "p", Quantity: 3}, {ID: "two", ProductID: "p", Quantity: 2}, {ID: "returned", ProductID: "p", Quantity: 2}}},
		{ID: "b", CheckoutID: "unpaid", Payment: "unpaid", Items: []orders.Item{{ID: "pending", ProductID: "p", Quantity: 99}}}},
		Transfers: []finance.Transfer{{OrderID: "a", ItemID: "returned", Kind: "refund", Status: "success"}, {OrderID: "a", ItemID: "returned", Kind: "refund", Status: "success"}}}
	if n := BuildCatalogSales(w).Counts(now)["p"]; n != 5 {
		t.Fatal("net units", n)
	}
	w.Orders[0].Items[2].Status = "refunded"
	if n := BuildCatalogSales(w).Counts(now)["p"]; n != 5 {
		t.Fatal("double refund", n)
	}
	w.Commerce.Purchases[0].Status = "cancelled"
	if n := BuildCatalogSales(w).Counts(now)["p"]; n != 0 {
		t.Fatal("cancelled purchase counted", n)
	}
	boundary := now.Add(-30 * 24 * time.Hour)
	s := CatalogSales{WindowDays: 30, Sales: []CatalogSale{{"p", 7, boundary}, {"p", 3, boundary.Add(time.Nanosecond)}, {"p", 2, now}, {"p", 50, now.Add(time.Nanosecond)}}}
	if n := s.Counts(now)["p"]; n != 5 {
		t.Fatal("rolling interval", n)
	}
	s.WindowDays = 1
	if n := s.Counts(now)["p"]; n != 2 {
		t.Fatal("window setting", n)
	}
}
