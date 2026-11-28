package backoffice_test

import (
	"context"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
)

func TestPlacedOrderPaymentDoesNotInventProcessingAndPartialDelivery(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	h.add(t, h.buyer, "offer2", 1)
	q := h.quote(t, h.buyer, "local")
	p := h.accept(t, h.buyer, q)
	if p.OrderStatus != "placed" || p.FulfillmentStatus != "awaiting-payment" {
		t.Fatalf("not placed before payment: %+v", p)
	}
	w, _ := h.repo.Snapshot(ctx)
	if len(w.Orders) != 2 || len(w.Commerce.Deposits) != 0 || len(w.Commerce.Entries) != 0 {
		t.Fatal("shipment/payment boundary", w.Orders)
	}
	for _, o := range w.Orders {
		if o.OrderStatus != "placed" || o.ShippingTerms == nil {
			t.Fatal("missing prepayment order snapshot", o)
		}
	}
	d := h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "placed-payment")
	h.approve(t, d)
	b, e := h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, b, e)
	if p.OrderStatus != "placed" || p.PaymentStatus != "paid" || p.FulfillmentStatus != "awaiting-supply" {
		t.Fatal("payment invented processing", p)
	}
	before, _ := h.repo.Snapshot(ctx)
	if e = h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.Purchases[0].OrderState = ""
		w.Orders[0].Status = "delivered"
		w.Orders[0].OrderStatus = ""
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	rows, e := h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	if rows[0].OrderStatus != "processing" || rows[0].FulfillmentStatus != "mixed" {
		t.Fatal("one delivery completed entire purchase", rows)
	}
	if e = h.repo.Update(ctx, func(w *backoffice.Workspace) error { w.Orders[1].Status = "cancelled"; return nil }); e != nil {
		t.Fatal(e)
	}
	rows, e = h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	if rows[0].OrderStatus != "completed" {
		t.Fatal("delivered plus cancelled did not complete parent", rows)
	}
	after, _ := h.repo.Snapshot(ctx)
	if len(after.Commerce.Entries) != len(before.Commerce.Entries) || len(after.Commerce.Purchases[0].Events) != len(before.Commerce.Purchases[0].Events) {
		t.Fatal("normalization fabricated finance/history")
	}
}
