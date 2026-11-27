package backoffice_test

import (
	"context"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
)

func TestPaymentAttemptsFundingReceiptRetryAndSeparateOrderDebit(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "local")
	p := h.accept(t, h.buyer, q)
	read := func() backoffice.PurchaseView {
		t.Helper()
		rows, e := h.s.ReadPurchases(ctx, h.buyer, p.ID)
		if e != nil {
			t.Fatal(e)
		}
		return rows[0]
	}
	if p.PaymentStage != "requested" || len(p.PaymentAttempts) != 1 {
		t.Fatal("missing requested payment", p)
	}
	d := h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "attempt-one")
	if read().PaymentStage != "pending" {
		t.Fatal("missing pending funding")
	}
	b, e := h.s.AttachReceipt(ctx, h.buyer, d.ID, d.Version, h.key(), sampleIcon(t))
	d = decodeCommerce[backoffice.DepositView](t, b, e)
	if read().PaymentStage != "awaiting-result" {
		t.Fatal("receipt not awaiting result")
	}
	b, e = h.s.ReviewDeposit(ctx, h.finance, backoffice.DepositReviewInput{ID: d.ID, Version: d.Version, Approve: false, Reason: "رسید ساختگی نامعتبر"}, h.key())
	decodeCommerce[backoffice.DepositView](t, b, e)
	if read().PaymentStage != "failed" || h.wallet(t, h.buyer).BalanceToman != "0" || read().OrderStatus != "placed" {
		t.Fatal("rejection paid/cancelled order")
	}
	d = h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "attempt-two")
	h.approve(t, d)
	p = read()
	if p.PaymentStage != "pending" || p.PaymentStatus != "unpaid" || p.WalletBalanceToman != q.TotalToman {
		t.Fatal("funding success incorrectly paid the order", p)
	}
	b, e = h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, b, e)
	if p.PaymentStage != "paid" || p.OrderStatus != "placed" || len(p.PaymentAttempts) != 3 {
		t.Fatal("order debit/history", p)
	}
	if p.PaymentAttempts[0].Status != "paid" || p.PaymentAttempts[1].Status != "failed" || p.PaymentAttempts[2].Purpose != "wallet-funding" || p.PaymentAttempts[2].Status != "paid" {
		t.Fatal("attempt history overwritten", p.PaymentAttempts)
	}
	admin, e := h.s.ReadCommerceAdmin(ctx, h.ops)
	if e != nil {
		t.Fatal(e)
	}
	if len(admin.Purchases[0].PaymentAttempts) != 0 {
		t.Fatal("financial attempts leaked to operations")
	}
}

func TestPaymentAttemptsLateFundingDoesNotReviveExpiredOrder(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "local")
	p := h.accept(t, h.buyer, q)
	d := h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "late-funding-attempt")
	h.clock.Advance(31 * time.Minute)
	h.approve(t, d)
	rows, e := h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	if rows[0].Status != "requires-reconfirmation" || rows[0].PaymentStatus != "unpaid" || rows[0].PaymentStage != "failed" || rows[0].WalletBalanceToman != q.TotalToman {
		t.Fatal("late funding revived purchase", rows)
	}
	w, _ := h.repo.Snapshot(ctx)
	if w.Offers[0].Reserved != 0 || w.Offers[0].Stock != 5 {
		t.Fatal("late funds consumed inventory")
	}
	if len(w.Commerce.PaymentAttempts) != 2 || w.Commerce.PaymentAttempts[0].FailureCode != "reservation_expired" || w.Commerce.PaymentAttempts[1].Status != "paid" {
		t.Fatal("expiry/funding history lost", w.Commerce.PaymentAttempts)
	}
}

func TestPaymentAuthorizationIsNotSettlement(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "local")
	p := h.accept(t, h.buyer, q)
	if err := h.repo.Update(ctx, func(w *backoffice.Workspace) error { w.Commerce.PaymentAttempts[0].Status = "authorized"; return nil }); err != nil {
		t.Fatal(err)
	}
	rows, e := h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	if rows[0].PaymentStage != "authorized" || rows[0].PaymentStatus != "unpaid" || rows[0].OrderStatus != "placed" {
		t.Fatal("authorization treated as paid", rows)
	}
	_, e = h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	if e != backoffice.ErrInsufficientWallet {
		t.Fatal("authorization bypassed actual wallet debit", e)
	}
	w, _ := h.repo.Snapshot(ctx)
	if len(w.Commerce.Entries) != 0 || w.Orders[0].Payment == "paid" {
		t.Fatal("authorization fabricated settlement")
	}
}

func TestLateFundingAfterCancellationCreditsWalletOnce(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "local")
	p := h.accept(t, h.buyer, q)
	d := h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "cancelled-late-funding")
	if _, e := h.s.CancelPurchase(ctx, h.buyer, p.ID, "انصراف پیش از پرداخت", h.key()); e != nil {
		t.Fatal(e)
	}
	h.approve(t, d)
	h.approve(t, d)
	rows, e := h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	if rows[0].OrderStatus != "cancelled" || rows[0].PaymentStatus != "unpaid" || rows[0].PaymentStage != "failed" || rows[0].WalletBalanceToman != q.TotalToman {
		t.Fatal("late funds revived cancelled order", rows)
	}
	w, _ := h.repo.Snapshot(ctx)
	if len(w.Commerce.Entries) != 1 || len(w.Commerce.PaymentAttempts) != 2 || w.Offers[0].Reserved != 0 || w.Offers[0].Stock != 5 {
		t.Fatal("late duplicate credit or reservation")
	}
}
