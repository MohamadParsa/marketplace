package backoffice_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
)

func TestShippingTermsRequireNewConsentBeforeAcceptAndPayment(t *testing.T) {
	changes := map[string]func(*backoffice.Workspace){
		"fee":  func(w *backoffice.Workspace) { w.Commerce.Settings.Shipping[0].Fee++ },
		"name": func(w *backoffice.Workspace) { w.Commerce.Settings.Shipping[0].Name = "نام تازه" },
		"collect": func(w *backoffice.Workspace) {
			w.Commerce.Settings.Shipping[0].CollectionMode = "collect"
			w.Commerce.Settings.Shipping[0].Fee = 0
		},
		"minimum":         func(w *backoffice.Workspace) { w.Commerce.Settings.Shipping[0].MinDays = 2 },
		"maximum":         func(w *backoffice.Workspace) { w.Commerce.Settings.Shipping[0].MaxDays = 4 },
		"transport-basis": func(w *backoffice.Workspace) { w.Commerce.Settings.Shipping[0].DayBasis = shipping.WorkingDays },
		"lead":            func(w *backoffice.Workspace) { w.Stores[0].LeadDays = 2 },
		"calendar":        func(w *backoffice.Workspace) { w.Commerce.Settings.Holidays = []string{"2026-11-23"} },
		"timing": func(w *backoffice.Workspace) {
			v := shipping.DefaultTiming()
			v.CutoffTime = "13:00"
			w.Commerce.Settings.ShippingTiming = &v
		},
	}
	for name, change := range changes {
		for _, stage := range []string{"accept", "payment"} {
			t.Run(name+"/"+stage, func(t *testing.T) {
				h := capacityHarness(t, false, 10)
				ctx := context.Background()
				h.add(t, h.buyer, "offer1", 1)
				q := capacityQuote(t, h, h.buyer, "")
				var p backoffice.PurchaseView
				if stage == "payment" {
					p = h.accept(t, h.buyer, q)
					h.approve(t, h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "funding"))
				}
				before, _ := h.repo.Snapshot(ctx)
				if err := h.repo.Update(ctx, func(w *backoffice.Workspace) error { change(w); return nil }); err != nil {
					t.Fatal(err)
				}
				if stage == "accept" {
					_, err := h.s.AcceptQuote(ctx, h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman}, h.key())
					if !errors.Is(err, backoffice.ErrCheckoutChanged) {
						t.Fatal("stale terms accepted", err)
					}
				} else {
					body, err := h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
					updated := decodeCommerce[backoffice.PurchaseView](t, body, err)
					if updated.Status != "requires-reconfirmation" {
						t.Fatal(updated)
					}
					after, _ := h.repo.Snapshot(ctx)
					if !reflect.DeepEqual(before.Commerce.Entries, after.Commerce.Entries) || after.Commerce.ShippingReservations[0].Status != "released" || after.Orders[0].Payment == "paid" {
						t.Fatal("financial side effect")
					}
				}
			})
		}
	}
}

func TestShippingCollectMixedShipmentsCouponWalletCancelAndHistoricalTerms(t *testing.T) {
	for _, fee := range []commerce.Amount{0, 200000} {
		t.Run(commerce.Toman(fee), func(t *testing.T) {
			h := capacityHarness(t, true, 10)
			ctx := context.Background()
			if err := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
				first := &w.Commerce.Settings.Shipping[0]
				first.CollectionMode = "collect"
				first.Fee = 0
				second := *first
				second.ID = "range"
				second.Name = "ارسال پیش‌کرایه"
				second.DeliveryMode = "range"
				second.Schedule = nil
				second.CollectionMode = "prepaid"
				second.Fee = fee
				second.MinDays = 1
				second.MaxDays = 3
				w.Commerce.Settings.Shipping = append(w.Commerce.Settings.Shipping, second)
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			h.approve(t, h.deposit(t, h.buyer, "", "card-transfer", "200000", "old-wallet"))
			h.add(t, h.buyer, "offer1", 1)
			h.add(t, h.buyer, "gold-offer", 1)
			h.add(t, h.buyer, "offer2", 1)
			in := h.v2Input(t, h.buyer)
			in.Shipping = map[string]string{"store1": "post", "store2": "range"}
			in.ShippingSelections = map[string]shipping.DeliverySelection{"store1": {Date: "2026-11-21", WindowID: "am"}}
			coupon := couponDraft()
			coupon.Target = "shipping"
			h.coupon(t, coupon)
			in.CouponCode = coupon.Code
			if fee == 0 {
				if _, err := h.s.CreateQuote(ctx, h.buyer, in, h.key()); !errors.Is(err, backoffice.ErrCoupon) {
					t.Fatal("collect cannot create shipping discount", err)
				}
			} else {
				shippingDiscount := h.v2Quote(t, h.buyer, in)
				if shippingDiscount.DiscountToman != commerce.Toman(fee/10) {
					t.Fatal("shipping coupon included collect", shippingDiscount)
				}
			}
			coupon.Code = "ORDER10"
			coupon.Target = "order"
			h.coupon(t, coupon)
			in.CouponCode = coupon.Code
			in.WalletToman = "20000"
			q := h.v2Quote(t, h.buyer, in)
			if len(q.Shipments) != 2 || q.ShippingToman != commerce.Toman(fee) || q.DiscountToman != commerce.Toman((3200000+fee)/10) {
				t.Fatal(q)
			}
			p := h.v2Accept(t, h.buyer, q)
			h.approve(t, h.deposit(t, h.buyer, p.ID, "card-transfer", q.FundingToman, "new-funding"))
			body, err := h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
			p = decodeCommerce[backoffice.PurchaseView](t, body, err)
			if p.Status != "confirmed" {
				t.Fatal(p)
			}
			captured, _ := h.repo.Snapshot(ctx)
			for _, order := range captured.Orders {
				if order.ShippingTerms == nil || order.Delivery == nil {
					t.Fatal("missing order terms")
				}
				if order.StoreID == "store1" && (order.ShippingFee != 0 || order.ShippingTerms.CollectionMode != "collect" || order.Delivery.Slot == nil) {
					t.Fatal(order)
				}
			}
			h.repo.Update(ctx, func(w *backoffice.Workspace) error {
				w.Commerce.Settings.Shipping[0].Name = "تغییر بعد از قطعیت"
				w.Commerce.Settings.Shipping[0].Fee = 999999
				return nil
			})
			after, _ := h.repo.Snapshot(ctx)
			if !reflect.DeepEqual(captured.Orders, after.Orders) || !reflect.DeepEqual(captured.Commerce.Quotes, after.Commerce.Quotes) {
				t.Fatal("history changed")
			}
			beforeBalance := h.wallet(t, h.buyer).BalanceToman
			body, err = h.s.AdminCancelPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version, Reason: "لغو آزمون"}, h.key())
			p = decodeCommerce[backoffice.PurchaseView](t, body, err)
			state, _ := h.repo.Snapshot(ctx)
			var refunded commerce.Amount
			for _, entry := range state.Commerce.Entries {
				if entry.Kind == "order-cancel-credit" {
					refunded += entry.Amount
				}
			}
			if commerce.Toman(refunded) != q.TotalToman {
				t.Fatal("unexpected shipping refund", refunded)
			}
			expected, _ := commerce.ParseToman(beforeBalance)
			total, _ := commerce.ParseToman(q.TotalToman)
			if h.wallet(t, h.buyer).BalanceToman != commerce.Toman(expected+total) {
				t.Fatal("refund must equal captured platform total")
			}
		})
	}
}
