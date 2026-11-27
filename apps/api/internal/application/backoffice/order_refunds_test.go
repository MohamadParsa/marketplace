package backoffice_test

import (
	"bytes"
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"sync"
	"testing"
)

func caseRefundInput(t *testing.T, h *commerceHarness, p backoffice.PurchaseView, r backoffice.OrderRequestView, method string) backoffice.OrderRefundInput {
	t.Helper()
	d, e := h.s.ReadOrderRequests(context.Background(), h.finance, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	for _, entry := range d.Requests {
		if entry.ID == r.ID {
			return backoffice.OrderRefundInput{ID: r.ID, Version: entry.Version, Method: method, AcceptedAmountToman: entry.RefundableToman, Instructions: "مقصد بانکی ساختگی تأییدشده", Reason: "تأیید مالی ساختگی"}
		}
	}
	t.Fatal("case missing")
	return backoffice.OrderRefundInput{}
}
func TestOrderRefundPartialDiscountSourcesBankHoldAndShippingOnce(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	rule := couponDraft()
	h.coupon(t, rule)
	restricted := h.credit(t, "150000", false)
	cash := h.credit(t, "140000", true)
	h.add(t, h.buyer, "offer1", 3)
	in := h.v2Input(t, h.buyer)
	in.CouponCode = rule.Code
	in.WalletToman = "290000"
	in.PaymentMethodID = ""
	q := h.v2Quote(t, h.buyer, in)
	if q.TotalToman != "290000" {
		t.Fatal(q)
	}
	p := h.v2Accept(t, h.buyer, q)
	b, e := h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, b, e)
	for index, method := range []string{"wallet", "bank", "wallet"} {
		r := createCase(t, h, p, "cancel", 0, 1)
		r = reviewCase(t, h, r, "approve", nil)
		input := caseRefundInput(t, h, p, r, method)
		expected := "90000"
		if index == 2 {
			expected = "110000"
		}
		if input.AcceptedAmountToman != expected {
			t.Fatal("net unit / shipping", input)
		}
		if _, e := h.s.RefundOrderRequest(ctx, h.ops, input, h.key()); e != backoffice.ErrForbidden {
			t.Fatal("ops money", e)
		}
		key := h.key()
		b, e = h.s.RefundOrderRequest(ctx, h.finance, input, key)
		f := decodeCommerce[backoffice.OrderRefundSummary](t, b, e)
		again, e := h.s.RefundOrderRequest(ctx, h.finance, input, key)
		if e != nil || !bytes.Equal(b, again) {
			t.Fatal("refund replay", e)
		}
		if _, e = h.s.RefundOrderRequest(ctx, h.finance, input, h.key()); e != backoffice.ErrConflict {
			t.Fatal("double credit", e)
		}
		if index == 1 {
			wallet, e := h.s.ReadWallet(ctx, h.buyer)
			if e != nil || wallet.HeldToman != "90000" || wallet.AvailableToman != "90000" {
				t.Fatal("unheld bank credit", wallet, e)
			}
			if f.Withdrawal == nil {
				t.Fatal("missing bank link")
			}
			resolve := backoffice.RefundResolveInput{ID: f.Withdrawal.ID, Version: f.Withdrawal.Version, Complete: true, Reference: "BANK-138-ONE", Date: h.clock.Now().Format("2006-01-02"), Reason: "ثبت انتقال صرفاً ساختگی", VerifiedAmountToman: "90000"}
			wrong := resolve
			wrong.VerifiedAmountToman = "1"
			if _, e = h.s.ResolveRefund(ctx, h.finance, wrong, h.key()); e != backoffice.ErrInvalid {
				t.Fatal("amount proof", e)
			}
			b, e = h.s.ResolveRefund(ctx, h.finance, resolve, "bank-complete")
			decodeCommerce[backoffice.RefundView](t, b, e)
			b, e = h.s.ResolveRefund(ctx, h.finance, resolve, "bank-complete")
			decodeCommerce[backoffice.RefundView](t, b, e)
		}
	}
	state, _ := h.repo.Snapshot(ctx)
	var total, shipping commerce.Amount
	for _, f := range state.Commerce.OrderRefunds {
		total += f.Amount
		shipping += f.Shipping
		if f.SettlementStatus != "review-required" {
			t.Fatal("missing settlement review")
		}
	}
	if commerce.Toman(total) != q.TotalToman || commerce.Toman(shipping) != "20000" || len(state.Commerce.OrderRefunds) != 3 || state.Offers[0].Stock != 5 {
		t.Fatal("wrong totals", total, shipping)
	}
	wallet, e := h.s.ReadWallet(ctx, h.buyer)
	if e != nil || wallet.BalanceToman != "200000" || wallet.PurchaseOnlyToman != "150000" || wallet.WithdrawableToman != "50000" {
		t.Fatal("restriction conversion", wallet, e)
	}
	originals := map[string]commerce.CreditLot{}
	for _, l := range state.Commerce.CreditLots {
		originals[l.ID] = l
	}
	for _, l := range state.Commerce.CreditLots {
		if l.OriginLotID == "" {
			continue
		}
		source := originals[l.OriginLotID]
		if l.CanWithdraw() != source.CanWithdraw() || !l.AvailableSince().Equal(source.AvailableSince()) {
			t.Fatal("FIFO/source changed")
		}
	}
	if originals[restricted.ID].Spent != originals[restricted.ID].Amount || originals[cash.ID].Spent != originals[cash.ID].Amount {
		t.Fatal("original rewritten")
	}
	if counts := backoffice.BuildCatalogSales(state).Counts(h.clock.Now()); counts["phone"] != 0 {
		t.Fatal("sales not adjusted", counts)
	}
	coupons, e := h.s.ReadCoupons(ctx, h.admin)
	if e != nil || coupons.Items[0].Consumed != 1 {
		t.Fatal("coupon restored", coupons, e)
	}
}

func TestOrderRefundBankRejectStaysWalletAndConcurrentRefundOnce(t *testing.T) {
	h, p := paidOperationsHarness(t)
	ctx := context.Background()
	r := reviewCase(t, h, createCase(t, h, p, "cancel", 0, 1), "approve", nil)
	in := caseRefundInput(t, h, p, r, "bank")
	var wg sync.WaitGroup
	wg.Add(2)
	results := make(chan error, 2)
	for _, key := range []string{"finance-a", "finance-b"} {
		go func(k string) { defer wg.Done(); _, e := h.s.RefundOrderRequest(ctx, h.finance, in, k); results <- e }(key)
	}
	wg.Wait()
	close(results)
	success := 0
	for e := range results {
		if e == nil {
			success++
		} else if e != backoffice.ErrConflict {
			t.Fatal(e)
		}
	}
	if success != 1 {
		t.Fatal(success)
	}
	state, _ := h.repo.Snapshot(ctx)
	if len(state.Commerce.OrderRefunds) != 1 || len(state.Commerce.Refunds) != 1 {
		t.Fatal("double bank")
	}
	withdraw := state.Commerce.Refunds[0]
	b, e := h.s.ResolveRefund(ctx, h.finance, backoffice.RefundResolveInput{ID: withdraw.ID, Version: withdraw.Version, Reason: "انتقال انجام نشد؛ وجه در کیف پول باقی بماند"}, h.key())
	decodeCommerce[backoffice.RefundView](t, b, e)
	wallet, e := h.s.ReadWallet(ctx, h.buyer)
	if e != nil || wallet.HeldToman != "0" || wallet.AvailableToman != in.AcceptedAmountToman {
		t.Fatal(wallet, e)
	}
	if _, e = h.s.RefundOrderRequest(ctx, h.finance, in, h.key()); e != backoffice.ErrConflict {
		t.Fatal("bank rejection allowed second order credit", e)
	}
}

func TestOrderRefundRestrictedBankAndUnpaidBlocked(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.credit(t, "120000", false)
	h.add(t, h.buyer, "offer1", 1)
	in := h.v2Input(t, h.buyer)
	in.WalletToman = "120000"
	in.PaymentMethodID = ""
	p := h.v2Accept(t, h.buyer, h.v2Quote(t, h.buyer, in))
	if _, e := h.s.RefundOrderRequest(ctx, h.finance, backoffice.OrderRefundInput{ID: "missing", Method: "wallet"}, h.key()); e != backoffice.ErrNotFound {
		t.Fatal(e)
	}
	d, _ := h.s.ReadOrderRequests(ctx, h.buyer, p.ID)
	line := d.Items[0]
	if _, e := h.s.CreateOrderRequest(ctx, h.buyer, backoffice.OrderRequestInput{PurchaseID: p.ID, Version: p.Version, Kind: "cancel", Reason: "ساختگی", Items: []backoffice.RequestedItemInput{{OrderID: line.OrderID, ItemID: line.ItemID, Quantity: 1}}}, h.key()); e != backoffice.ErrConflict {
		t.Fatal("unpaid case", e)
	}
	b, e := h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, b, e)
	r := reviewCase(t, h, createCase(t, h, p, "cancel", 0, 1), "approve", nil)
	input := caseRefundInput(t, h, p, r, "bank")
	if _, e = h.s.RefundOrderRequest(ctx, h.finance, input, h.key()); e != backoffice.ErrWithdrawalUnavailable {
		t.Fatal("restricted bank", e)
	}
	input.Method = "wallet"
	b, e = h.s.RefundOrderRequest(ctx, h.finance, input, h.key())
	decodeCommerce[backoffice.OrderRefundSummary](t, b, e)
}

func TestReturnRefundAfterDeliveryExcludesShippingAndDoesNotRestock(t *testing.T) {
	h, p := paidOperationsHarness(t)
	ctx := context.Background()
	id := p.OrderIDs[0]
	for _, action := range []string{"start", "supply", "pack", "ship", "deliver"} {
		operationStep(t, h, h.admin, id, action)
	}
	r := reviewCase(t, h, createCase(t, h, p, "return", 0, 1), "approve", nil)
	input := caseRefundInput(t, h, p, r, "wallet")
	if input.AcceptedAmountToman != "100000" {
		t.Fatal("shipping refunded after dispatch", input)
	}
	b, e := h.s.RefundOrderRequest(ctx, h.finance, input, h.key())
	f := decodeCommerce[backoffice.OrderRefundSummary](t, b, e)
	if f.ShippingToman != "0" {
		t.Fatal(f)
	}
	state, _ := h.repo.Snapshot(ctx)
	if state.Offers[0].Stock != 4 || state.Commerce.ShippingReservations[0].Status != "committed" {
		t.Fatal("financial changed stock/capacity")
	}
	if counts := backoffice.BuildCatalogSales(state).Counts(h.clock.Now()); counts["phone"] != 1 {
		t.Fatal("net units", counts)
	}
	v, e := h.s.ReadOrderOperations(ctx, h.admin, id, "")
	if e != nil || v.OrderStatus != "completed" {
		t.Fatal("delivery history", v, e)
	}
}
