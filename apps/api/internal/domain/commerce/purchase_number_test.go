package commerce_test

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"testing"
	"time"
)

func TestPurchaseNumbersBackfillStableAndContinue(t *testing.T) {
	c := commerce.State{Purchases: []commerce.Purchase{{ID: "newer", CreatedAt: time.Unix(2, 0)}, {ID: "older", CreatedAt: time.Unix(1, 0)}}}
	c.EnsurePurchaseNumbers()
	if c.Purchases[1].Number != 1000 || c.Purchases[0].Number != 1001 || c.NextPurchaseNumber != 1002 {
		t.Fatal(c)
	}
	c.Purchases[1].Status = "cancelled"
	c.EnsurePurchaseNumbers()
	if c.Purchases[1].Number != 1000 || c.NextPurchaseNumber != 1002 {
		t.Fatal("cancelled/repeated normalization changed numbers")
	}
	c.Purchases = append(c.Purchases, commerce.Purchase{ID: "latest", CreatedAt: time.Unix(3, 0)})
	c.EnsurePurchaseNumbers()
	if c.Purchases[2].Number != 1002 {
		t.Fatal("counter reused")
	}
}
