package backoffice_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
)

func TestRefundFieldErrorsKeepReservationAndFinalReceiptPostsOnce(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.credit(t, "500000", true)
	raw, e := h.s.RequestRefund(ctx, h.buyer, backoffice.RefundInput{AmountToman: "100000", Instructions: "حساب ساختگی آزمون"}, h.key())
	r := decodeCommerce[backoffice.RefundView](t, raw, e)
	valid := backoffice.RefundResolveInput{ID: r.ID, Version: r.Version, Complete: true, Reference: "۰۰۰۱۲۳۴۵۶", Date: h.clock.Now().Format("2006-01-02"), Reason: "تطبیق رسید ساختگی آزمون"}
	before, _ := h.repo.Snapshot(ctx)
	for _, tc := range []struct {
		name   string
		change func(*backoffice.RefundResolveInput)
		want   error
	}{
		{"name instead of receipt", func(in *backoffice.RefundResolveInput) { in.Reference = "مدیر سایت" }, backoffice.ErrTransferReference},
		{"short receipt", func(in *backoffice.RefundResolveInput) { in.Reference = "12" }, backoffice.ErrTransferReference},
		{"future date", func(in *backoffice.RefundResolveInput) { in.Date = "2099-01-01" }, backoffice.ErrTransferDate},
		{"missing reason", func(in *backoffice.RefundResolveInput) { in.Reason = "" }, backoffice.ErrRefundReason},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := valid
			tc.change(&in)
			if _, e := h.s.ResolveRefund(ctx, h.finance, in, h.key()); e != tc.want {
				t.Fatal(e)
			}
			after, _ := h.repo.Snapshot(ctx)
			if !reflect.DeepEqual(before.Commerce, after.Commerce) {
				t.Fatal("invalid field altered ledger or hold")
			}
		})
	}
	raw, e = h.s.ResolveRefund(ctx, h.finance, valid, h.key())
	result := decodeCommerce[backoffice.RefundView](t, raw, e)
	if result.Status != "completed" || result.Reference != "000123456" {
		t.Fatal("reference normalization", result)
	}
	if _, e = h.s.ResolveRefund(ctx, h.finance, valid, h.key()); e != nil {
		t.Fatal("completion retry", e)
	}
	wallet := h.wallet(t, h.buyer)
	if wallet.BalanceToman != "400000" || wallet.HeldToman != "0" {
		t.Fatal(wallet)
	}
	state, _ := h.repo.Snapshot(ctx)
	debits := 0
	for _, entry := range state.Commerce.Entries {
		if entry.Kind == "manual-refund" {
			debits++
		}
	}
	if debits != 1 {
		t.Fatal("completion posted twice", debits)
	}
	if _, e = h.s.ReadWallet(ctx, h.other); e != nil {
		t.Fatal(e)
	}
	other := h.wallet(t, h.other)
	if len(other.Entries) != 0 || len(other.Refunds) != 0 {
		t.Fatal("other customer's transactions leaked")
	}
}
