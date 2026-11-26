package backoffice_test

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"reflect"
	"testing"
)

func setProductActive(t *testing.T, h *commerceHarness, active bool) {
	t.Helper()
	w, _ := h.repo.Snapshot(context.Background())
	p := w.Products[0]
	if e := h.s.Execute(context.Background(), h.admin, backoffice.Command{Action: "catalog.product.update", Target: p.ID, Version: p.Version, IdempotencyKey: h.key(), Catalog: &backoffice.CatalogInput{Name: p.Name, EnglishName: p.EnglishName, CategoryID: p.CategoryID, BrandID: p.BrandID, Published: p.Status == "published", Active: &active}}); e != nil {
		t.Fatal(e)
	}
}
func TestProductDeactivationAfterFundingReleasesHoldWithoutSpending(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "card-transfer")
	p := h.accept(t, h.buyer, q)
	h.approve(t, h.deposit(t, h.buyer, p.ID, "card-transfer", q.TotalToman, "product-funding"))
	balance := h.wallet(t, h.buyer).BalanceToman
	setProductActive(t, h, false)
	before, _ := h.repo.Snapshot(ctx)
	b, e := h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	view := decodeCommerce[backoffice.PurchaseView](t, b, e)
	after, _ := h.repo.Snapshot(ctx)
	if view.Status != "requires-reconfirmation" || h.wallet(t, h.buyer).BalanceToman != balance || len(after.Commerce.Entries) != len(before.Commerce.Entries) || after.Offers[0].Reserved != 0 {
		t.Fatal("inactive purchase consumed funds", view.Status)
	}
	if !reflect.DeepEqual(before.Commerce.Quotes[0].Lines, after.Commerce.Quotes[0].Lines) {
		t.Fatal("financial snapshot changed")
	}
}
func TestProductDeactivationPreservesConfirmedOrdersAndRejectsStaleQuote(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "card-transfer")
	setProductActive(t, h, false)
	if _, e := h.s.AcceptQuote(ctx, h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman, TermsDigest: q.TermsDigest}, h.key()); e == nil {
		t.Fatal("stale quote accepted")
	}
	setProductActive(t, h, true)
	q = h.quote(t, h.buyer, "card-transfer")
	p := h.accept(t, h.buyer, q)
	h.approve(t, h.deposit(t, h.buyer, p.ID, "card-transfer", q.TotalToman, "confirmed-product-funding"))
	b, e := h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	decodeCommerce[backoffice.PurchaseView](t, b, e)
	before, _ := h.repo.Snapshot(ctx)
	setProductActive(t, h, false)
	after, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(before.Orders, after.Orders) || !reflect.DeepEqual(before.Commerce, after.Commerce) || !reflect.DeepEqual(before.Ledger, after.Ledger) {
		t.Fatal("confirmed history rewritten")
	}
	cart := h.cart(t, h.buyer)
	for _, i := range cart.Items {
		if i.ProductID == "phone" && (!i.Inactive || i.Available || i.UnitPriceToman != "") {
			t.Fatal("inactive cart price")
		}
	}
}
