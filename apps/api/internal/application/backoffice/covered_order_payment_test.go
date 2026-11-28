package backoffice_test

import (
	"context"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"sync"
	"testing"
)

func TestCoveredPlacementConcurrentKeysConsumeCreditAndStockOnce(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.credit(t, "200000", false)
	h.add(t, h.buyer, "offer1", 1)
	in := h.v2Input(t, h.buyer)
	in.WalletToman = "999999"
	in.PaymentMethodID = ""
	q := h.v2Quote(t, h.buyer, in)
	accept := backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman, TermsDigest: q.TermsDigest}
	forged := accept
	forged.AcceptedTotalToman = "0"
	if _, e := h.s.AcceptQuote(ctx, h.buyer, forged, "forged-zero"); e != backoffice.ErrCheckoutChanged {
		t.Fatal(e)
	}
	if _, e := h.s.AcceptQuote(ctx, h.other, accept, "foreign"); e != backoffice.ErrNotFound {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	results := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, e := h.s.AcceptQuote(ctx, h.buyer, accept, fmt.Sprintf("covered-%d", i))
			results <- e
		}(i)
	}
	wg.Wait()
	close(results)
	for e := range results {
		if e != nil {
			t.Fatal(e)
		}
	}
	w, _ := h.repo.Snapshot(ctx)
	debits := 0
	for _, e := range w.Commerce.Entries {
		if e.Kind == "order-payment" {
			debits++
		}
	}
	if len(w.Commerce.Purchases) != 1 || len(w.Orders) != 1 || debits != 1 || w.Offers[0].Stock != 4 || w.Offers[0].Reserved != 0 {
		t.Fatal("duplicate effect")
	}
	if w.Orders[0].OrderStatus != "placed" || w.Orders[0].Payment != "paid" {
		t.Fatal("automatically started processing")
	}
	wallet := h.wallet(t, h.buyer)
	if wallet.BalanceToman != "80000" || wallet.HeldToman != "0" || wallet.PurchaseOnlyToman != "80000" {
		t.Fatal("source restriction or debit changed", wallet)
	}
}
