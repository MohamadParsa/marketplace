package backoffice_test

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"testing"
)

func TestWalletRequestCapsToNetOrderBeforeBalanceValidation(t *testing.T) {
	for _, mode := range []string{"plain", "discount", "zero"} {
		t.Run(mode, func(t *testing.T) {
			h := newCommerceHarness(t)
			ctx := context.Background()
			h.credit(t, "120000", false)
			h.add(t, h.buyer, "offer1", 1)
			in := h.v2Input(t, h.buyer)
			in.WalletToman = "۹۹۹۹۹۹۹"
			in.PaymentMethodID = ""
			expected := "120000"
			if mode != "plain" {
				rule := couponDraft()
				if mode == "zero" {
					rule.Target = "order"
					rule.Kind = "fixed"
					rule.Value = "150000"
					expected = "0"
				} else {
					expected = "110000"
				}
				h.coupon(t, rule)
				in.CouponCode = rule.Code
			}
			q := h.v2Quote(t, h.buyer, in)
			if q.TotalToman != expected || q.WalletSelectedToman != expected || q.FundingToman != "0" {
				t.Fatal(q)
			}
			p := h.v2Accept(t, h.buyer, q)
			b, e := h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
			p = decodeCommerce[backoffice.PurchaseView](t, b, e)
			if p.PaymentStatus != "paid" {
				t.Fatal(p)
			}
			state, _ := h.repo.Snapshot(ctx)
			debits := 0
			for _, entry := range state.Commerce.Entries {
				if entry.Kind == "order-payment" {
					debits++
					if entry.Amount != -state.Commerce.Quotes[len(state.Commerce.Quotes)-1].Total {
						t.Fatal("wrong debit", entry)
					}
				}
			}
			if mode == "zero" && debits != 0 || mode != "zero" && debits != 1 {
				t.Fatal("unexpected debit count", debits)
			}
		})
	}
}

func TestWalletCapPreservesInsufficientAndInvalidAmountErrors(t *testing.T) {
	h := newCommerceHarness(t)
	h.credit(t, "100", false)
	h.add(t, h.buyer, "offer1", 1)
	in := h.v2Input(t, h.buyer)
	in.WalletToman = "9999999"
	if _, e := h.s.CreateQuote(context.Background(), h.buyer, in, h.key()); e != backoffice.ErrInsufficientWallet {
		t.Fatal(e)
	}
	for _, amount := range []string{"-1", "broken"} {
		in.WalletToman = amount
		if _, e := h.s.CreateQuote(context.Background(), h.buyer, in, h.key()); e != backoffice.ErrInvalid {
			t.Fatal(amount, e)
		}
	}
}
