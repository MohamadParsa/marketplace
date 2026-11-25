package backoffice_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

func (h *commerceHarness) credit(t *testing.T, amount string, withdrawable bool) backoffice.AdminCreditResult {
	t.Helper()
	b, e := h.s.CreditCustomerWallet(context.Background(), h.finance, backoffice.AdminCreditInput{UserID: h.buyer.ID, AmountToman: amount, Title: "اعتبار آزمون", Description: "منبع مستقل آزمون مالی", Withdrawable: withdrawable}, h.key())
	return decodeCommerce[backoffice.AdminCreditResult](t, b, e)
}

func TestWalletCreditFourSourcesFIFOAndCancellationRestriction(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	cash1 := h.credit(t, "40000", true)
	h.clock.now = h.clock.now.Add(time.Minute)
	restricted1 := h.credit(t, "20000", false)
	h.clock.now = h.clock.now.Add(time.Minute)
	cash2 := h.credit(t, "30000", true)
	h.clock.now = h.clock.now.Add(time.Minute)
	restricted2 := h.credit(t, "30000", false)
	h.add(t, h.buyer, "offer1", 1)
	in := h.v2Input(t, h.buyer)
	in.WalletToman = "120000"
	in.PaymentMethodID = ""
	q := h.v2Quote(t, h.buyer, in)
	p := h.v2Accept(t, h.buyer, q)
	if len(p.CreditSources) != 4 {
		t.Fatal("expected four source allocations", p.CreditSources)
	}
	for i, id := range []string{restricted1.ID, restricted2.ID, cash1.ID, cash2.ID} {
		if p.CreditSources[i].LotID != id {
			t.Fatal("priority/FIFO", p.CreditSources)
		}
	}
	w, e := h.s.ReadWallet(ctx, h.buyer)
	if e != nil || w.AvailableToman != "0" || w.WithdrawableToman != "0" {
		t.Fatal(w, e)
	}
	b, e := h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, b, e)
	b, e = h.s.AdminCancelPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version, Reason: "لغو آزمون"}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, b, e)
	w, e = h.s.ReadWallet(ctx, h.buyer)
	if e != nil || w.BalanceToman != "120000" || w.WithdrawableToman != "70000" || w.PurchaseOnlyToman != "50000" {
		t.Fatal("restriction lost after cancellation", w, e)
	}
	state, _ := h.repo.Snapshot(ctx)
	returned := 0
	for _, l := range state.Commerce.CreditLots {
		if l.OriginLotID != "" {
			returned++
			if l.AvailableSince().After(l.At) {
				t.Fatal("bad source age")
			}
		}
	}
	if returned != 4 {
		t.Fatal("missing original-source returns")
	}
	if _, e = h.s.RequestRefund(ctx, h.buyer, backoffice.RefundInput{AmountToman: "70000.1", Instructions: "مقصد آزمون"}, h.key()); !errors.Is(e, backoffice.ErrWithdrawalUnavailable) {
		t.Fatal("restricted balance became cash", e)
	}
	b, e = h.s.RequestRefund(ctx, h.buyer, backoffice.RefundInput{AmountToman: "70000", Instructions: "مقصد آزمون"}, h.key())
	r := decodeCommerce[backoffice.RefundView](t, b, e)
	_, e = h.s.ResolveRefund(ctx, h.finance, backoffice.RefundResolveInput{ID: r.ID, Version: r.Version, Complete: true, Reference: "POLICY-WITHDRAW-001", Date: h.clock.Now().Format("2006-01-02"), Reason: "ثبت خروج آزمایشی"}, h.key())
	if e != nil {
		t.Fatal(e)
	}
	w, e = h.s.ReadWallet(ctx, h.buyer)
	if e != nil || w.AvailableToman != "50000" || w.WithdrawableToman != "0" {
		t.Fatal(w, e)
	}
	if _, e = h.s.RequestRefund(ctx, h.buyer, backoffice.RefundInput{AmountToman: "1", Instructions: "مقصد آزمون"}, h.key()); !errors.Is(e, backoffice.ErrWithdrawalUnavailable) {
		t.Fatal(e)
	}
}

func TestWalletCreditValidationPermissionAndConcurrentRetry(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	in := backoffice.AdminCreditInput{UserID: h.buyer.ID, AmountToman: "۱۲۳.۴", Title: "هدیه آزمون", Description: "دلیل اجباری آزمون"}
	for _, a := range []identity.User{h.buyer, h.other, h.ops} {
		if _, e := h.s.CreditCustomerWallet(ctx, a, in, h.key()); !errors.Is(e, backoffice.ErrForbidden) {
			t.Fatal("permission", e)
		}
	}
	for _, amount := range []string{"0", "-1", "1.01", "NaN", "900000000000000.1"} {
		bad := in
		bad.AmountToman = amount
		if _, e := h.s.CreditCustomerWallet(ctx, h.finance, bad, h.key()); !errors.Is(e, backoffice.ErrInvalid) {
			t.Fatal(amount, e)
		}
	}
	for _, empty := range []string{"title", "description"} {
		bad := in
		if empty == "title" {
			bad.Title = " "
		} else {
			bad.Description = " "
		}
		if _, e := h.s.CreditCustomerWallet(ctx, h.finance, bad, h.key()); !errors.Is(e, backoffice.ErrInvalid) {
			t.Fatal(empty, e)
		}
	}
	bad := in
	bad.UserID = "missing-user"
	if _, e := h.s.CreditCustomerWallet(ctx, h.finance, bad, h.key()); !errors.Is(e, backoffice.ErrNotFound) {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, e := h.s.CreditCustomerWallet(ctx, h.finance, in, "same-credit"); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	w, e := h.s.ReadCustomerCreditWallet(ctx, h.finance, h.buyer.ID, "same-credit")
	if e != nil || w.BalanceToman != "123.4" || len(w.Credits) != 1 || w.PreviousOperation == nil || w.Credits[0].Withdrawable {
		t.Fatal(w, e)
	}
	in.AmountToman = "124"
	if _, e = h.s.CreditCustomerWallet(ctx, h.finance, in, "same-credit"); !errors.Is(e, backoffice.ErrConflict) {
		t.Fatal("key payload mismatch", e)
	}
	state, _ := h.repo.Snapshot(ctx)
	if len(state.Commerce.Entries) != 1 || len(state.Commerce.Deposits) != 0 || len(state.Commerce.CreditLots) != 1 {
		t.Fatal("double credit or fabricated deposit")
	}
	if state.Commerce.CreditLots[0].ActorID != h.finance.ID || state.Commerce.CreditLots[0].Title != "هدیه آزمون" {
		t.Fatal("missing actor/title")
	}
	if e = h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Users {
			if w.Users[i].ID == h.finance.ID {
				w.Users[i].Active = false
			}
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	if _, e = h.s.CreditCustomerWallet(ctx, h.finance, in, h.key()); !errors.Is(e, backoffice.ErrForbidden) {
		t.Fatal("revoked actor", e)
	}
}

func TestWalletCreditConcurrentPurchaseAndWithdrawal(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.credit(t, "120000", true)
	h.add(t, h.buyer, "offer1", 1)
	in := h.v2Input(t, h.buyer)
	in.PaymentMethodID = ""
	in.WalletToman = "120000"
	q := h.v2Quote(t, h.buyer, in)
	var wins atomic.Int32
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, e := h.s.AcceptQuote(ctx, h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman, TermsDigest: q.TermsDigest}, "race-purchase")
		if e == nil {
			wins.Add(1)
		} else if !errors.Is(e, backoffice.ErrInsufficientWallet) {
			t.Error(e)
		}
	}()
	go func() {
		defer wg.Done()
		_, e := h.s.RequestRefund(ctx, h.buyer, backoffice.RefundInput{AmountToman: "120000", Instructions: "مقصد آزمون"}, "race-refund")
		if e == nil {
			wins.Add(1)
		} else if !errors.Is(e, backoffice.ErrWithdrawalUnavailable) {
			t.Error(e)
		}
	}()
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatal("one balance reserved twice", wins.Load())
	}
	w, e := h.s.ReadWallet(ctx, h.buyer)
	if e != nil || w.AvailableToman != "0" || w.HeldToman != "120000" {
		t.Fatal(w, e)
	}
}

func TestWalletCreditLegacyBalanceHoldsAndExpiry(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	legacy := commerce.WalletEntry{ID: "legacy-entry", UserID: h.buyer.ID, Source: "legacy-cash", Kind: "deposit-credit", Amount: 2000000, At: h.clock.Now().Add(-24 * time.Hour)}
	if e := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.Entries = []commerce.WalletEntry{legacy}
		w.Commerce.Refunds = []commerce.RefundRequest{{ID: "legacy-refund", UserID: h.buyer.ID, Amount: 300000, Status: "pending", Version: 1, CreatedAt: h.clock.Now()}}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	h.credit(t, "50000", false)
	w, e := h.s.ReadWallet(ctx, h.buyer)
	if e != nil || w.BalanceToman != "250000" || w.AvailableToman != "220000" || w.WithdrawableToman != "170000" {
		t.Fatal("legacy policy", w, e)
	}
	state, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(state.Commerce.Entries[0], legacy) || len(state.Commerce.Refunds[0].Allocations) != 1 {
		t.Fatal("ledger changed or legacy hold lost")
	}
	h.add(t, h.buyer, "offer1", 1)
	in := h.v2Input(t, h.buyer)
	in.PaymentMethodID = ""
	in.WalletToman = "120000"
	q := h.v2Quote(t, h.buyer, in)
	h.v2Accept(t, h.buyer, q)
	h.clock.now = h.clock.now.Add(31 * time.Minute)
	if e = h.s.ReleaseExpiredCommerce(ctx); e != nil {
		t.Fatal(e)
	}
	w, e = h.s.ReadWallet(ctx, h.buyer)
	if e != nil || w.AvailableToman != "220000" || w.WithdrawableToman != "170000" || w.PurchaseOnlyToman != "50000" {
		t.Fatal("expiry changed restrictions", w, e)
	}
	_, e = h.s.ResolveRefund(ctx, h.finance, backoffice.RefundResolveInput{ID: "legacy-refund", Version: 1, Complete: false, Reason: "رد آزمون"}, h.key())
	if e != nil {
		t.Fatal(e)
	}
	w, e = h.s.ReadWallet(ctx, h.buyer)
	if e != nil || w.WithdrawableToman != "200000" {
		t.Fatal(w, e)
	}
}

func TestWalletCreditLegacyConfirmedOrderCancellationKeepsCash(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.approve(t, h.deposit(t, h.buyer, "", "card-transfer", "200000", "LEGACY-CANCEL-CASH"))
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "card-transfer")
	p := h.accept(t, h.buyer, q)
	b, e := h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, b, e)
	// Isolated memory fixture represents an already-paid pre-policy order.
	if e = h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.CreditAccounts = nil
		w.Commerce.CreditLots = nil
		for i := range w.Commerce.Quotes {
			if w.Commerce.Quotes[i].ID == q.ID {
				w.Commerce.Quotes[i].Credit = nil
			}
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	before, _ := h.repo.Snapshot(ctx)
	ledger, _ := json.Marshal(before.Commerce.Entries)
	_, e = h.s.AdminCancelPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version, Reason: "لغو سفارش قدیمی آزمون"}, h.key())
	if e != nil {
		t.Fatal(e)
	}
	w, e := h.s.ReadWallet(ctx, h.buyer)
	if e != nil || w.BalanceToman != "200000" || w.WithdrawableToman != "200000" {
		t.Fatal(w, e)
	}
	after, _ := h.repo.Snapshot(ctx)
	kept, _ := json.Marshal(after.Commerce.Entries[:len(before.Commerce.Entries)])
	if string(ledger) != string(kept) || len(after.Commerce.Entries) != len(before.Commerce.Entries)+1 {
		t.Fatal("legacy ledger was rewritten")
	}
}

func TestOrderSavingsSeparateFromCouponAndImmutableAfterConfirmation(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	if e := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Offers {
			switch w.Offers[i].ID {
			case "offer1":
				w.Offers[i].RegularPrice = 1200000
			case "gold-offer":
				w.Offers[i].RegularPrice = 1500000
			}
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	h.credit(t, "1000000", false)
	h.coupon(t, couponDraft())
	h.add(t, h.buyer, "offer1", 2)
	h.add(t, h.buyer, "gold-offer", 2)
	in := h.v2Input(t, h.buyer)
	in.CouponCode = "TEST10"
	in.WalletToman = "380000"
	in.PaymentMethodID = ""
	q := h.v2Quote(t, h.buyer, in)
	if q.DirectDiscountToman != "140000" || q.DiscountToman != "40000" || q.TotalToman != "380000" {
		t.Fatal(q)
	}
	p := h.v2Accept(t, h.buyer, q)
	b, e := h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, b, e)
	before, _ := json.Marshal(p.Quote)
	if e = h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Offers {
			w.Offers[i].Price = 500000
			w.Offers[i].RegularPrice = 700000
			w.Offers[i].Version++
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	orders, e := h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	after, _ := json.Marshal(orders[0].Quote)
	if string(before) != string(after) {
		t.Fatal("historical savings changed")
	}
	for _, reference := range []int64{0, 900000, 1000000} {
		t.Run(fmt.Sprint(reference), func(t *testing.T) {
			fresh := newCommerceHarness(t)
			_ = fresh.repo.Update(ctx, func(w *backoffice.Workspace) error { w.Offers[0].RegularPrice = reference; return nil })
			fresh.add(t, fresh.buyer, "offer1", 1)
			view := fresh.v2Quote(t, fresh.buyer, fresh.v2Input(t, fresh.buyer))
			if view.DirectDiscountToman != "" {
				t.Fatal("invented savings", view.DirectDiscountToman)
			}
		})
	}
}
