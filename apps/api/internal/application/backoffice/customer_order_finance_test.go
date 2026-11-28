package backoffice_test

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"testing"
)

func TestCustomerOrderActualWalletConsumptionAndOwnership(t *testing.T) {
	h, p := paidOperationsHarness(t)
	ctx := context.Background()
	rows, e := h.s.ReadPurchases(ctx, h.buyer, p.Number)
	if e != nil || len(rows) != 1 || rows[0].WalletConsumedToman != p.Quote.TotalToman {
		t.Fatal("actual debit projection", rows, e)
	}
	if _, e = h.s.ReadPurchases(ctx, h.other, p.ID); e != backoffice.ErrNotFound {
		t.Fatal("foreign order", e)
	}
	detail, e := h.s.ReadOrderWorkspace(ctx, h.ops, p.ID)
	if e != nil || detail.Purchase.WalletConsumedToman != "" {
		t.Fatal("private financial projection escaped", e)
	}
	raw, e := h.s.AdminCancelPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version, Reason: "لغو ساختگی برای تطبیق مبلغ اولیه"}, h.key())
	decodeCommerce[backoffice.PurchaseView](t, raw, e)
	rows, e = h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if e != nil || rows[0].WalletConsumedToman != p.Quote.TotalToman || rows[0].PaymentStatus != "returned-to-wallet" {
		t.Fatal("refund rewrote original debit", rows, e)
	}
	h2 := newCommerceHarness(t)
	h2.add(t, h2.buyer, "offer1", 1)
	q := h2.v2Quote(t, h2.buyer, h2.v2Input(t, h2.buyer))
	pending := h2.v2Accept(t, h2.buyer, q)
	rows, e = h2.s.ReadPurchases(ctx, h2.buyer, pending.ID)
	if e != nil || rows[0].WalletConsumedToman != "0" || rows[0].PaymentStatus != "unpaid" {
		t.Fatal("preview/hold counted as actual debit", rows, e)
	}
}
