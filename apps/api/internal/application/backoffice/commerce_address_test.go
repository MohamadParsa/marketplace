package backoffice_test

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"sync"
	"testing"
)

func recipientInput() backoffice.AddressInput {
	return backoffice.AddressInput{RecipientSelf: new(bool), Recipient: "گیرنده ساختگی", RecipientNationalID: "۰۰۱۳۵۴۷۸۹۵", Phone: "۰۹۰۰۰۰۰۰۱۱۱", Province: "تهران", City: "تهران", Street: "خیابان ساختگی", HouseNumber: "۱۲", Unit: "۴", PostalCode: "۱۲۳۴۵۶۷۸۹۰", Title: "خانه", FullAddress: "نشانی نهایی اصلاح‌شده"}
}
func TestAddressRecipientRequiredFieldsOwnershipAndConcurrency(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	valid := recipientInput()
	cases := map[string]func(*backoffice.AddressInput){"no-mode": func(i *backoffice.AddressInput) { i.RecipientSelf = nil }, "no-house": func(i *backoffice.AddressInput) { i.HouseNumber = "" }, "no-national": func(i *backoffice.AddressInput) { i.RecipientNationalID = "" }, "bad-national": func(i *backoffice.AddressInput) { i.RecipientNationalID = "1111111111" }, "no-phone": func(i *backoffice.AddressInput) { i.Phone = "" }, "bad-phone": func(i *backoffice.AddressInput) { i.Phone = "123" }, "no-postal": func(i *backoffice.AddressInput) { i.PostalCode = "" }, "no-name": func(i *backoffice.AddressInput) { i.Recipient = "" }, "no-street": func(i *backoffice.AddressInput) { i.Street = "" }, "half-coordinate": func(i *backoffice.AddressInput) { v := 35.7; i.Latitude = &v }, "bad-coordinate": func(i *backoffice.AddressInput) { v := 100.; i.Latitude = &v; i.Longitude = &v }}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			in := valid
			mutate(&in)
			if _, e := h.s.SaveAddress(ctx, h.buyer, in, h.key()); e == nil {
				t.Fatal("invalid payload accepted")
			}
		})
	}
	raw, e := h.s.SaveAddress(ctx, h.buyer, valid, "address-repeat")
	a := decodeCommerce[commerce.Address](t, raw, e)
	again, e := h.s.SaveAddress(ctx, h.buyer, valid, "address-repeat")
	if e != nil || string(again) != string(raw) {
		t.Fatal("retry", e)
	}
	if a.UserID != h.buyer.ID || a.RecipientNationalID != "0013547895" || a.Phone != "09000000111" || a.FullAddress != valid.FullAddress || a.DetailVersion != 2 {
		t.Fatal("normalization or ownership")
	}
	if rows, e := h.s.ReadAddresses(ctx, h.other); e != nil || len(rows) != 0 {
		t.Fatal("private address leaked")
	}
	in := valid
	in.ID = a.ID
	in.Version = a.Version
	if _, e := h.s.SaveAddress(ctx, h.other, in, "foreign"); e != backoffice.ErrNotFound {
		t.Fatal("foreign update", e)
	}
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, key := range []string{"concurrent-a", "concurrent-b"} {
		wg.Add(1)
		go func(k string) { defer wg.Done(); _, e := h.s.SaveAddress(ctx, h.buyer, in, k); results <- e }(key)
	}
	wg.Wait()
	close(results)
	ok, conflict := 0, 0
	for e := range results {
		if e == nil {
			ok++
		} else if e == backoffice.ErrConflict {
			conflict++
		} else {
			t.Fatal(e)
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatal(ok, conflict)
	}
}
func TestSelfRecipientFreshProfileAndHistoricalSnapshot(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	in := recipientInput()
	self := true
	in.RecipientSelf = &self
	in.Recipient = "نام جعلی درخواست"
	in.Phone = "09000000999"
	if _, e := h.s.SaveAddress(ctx, h.buyer, in, h.key()); e != backoffice.ErrProfileIncomplete {
		t.Fatal("missing national ID", e)
	}
	if e := h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Users {
			if w.Users[i].ID == h.buyer.ID {
				w.Users[i].Name = "مشتری واقعی آزمون"
				w.Users[i].NationalID = "0013547895"
			}
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	raw, e := h.s.SaveAddress(ctx, h.buyer, in, h.key())
	a := decodeCommerce[commerce.Address](t, raw, e)
	if a.Recipient != "مشتری واقعی آزمون" || a.Phone == in.Phone || a.UserID != h.buyer.ID {
		t.Fatal("stale session or spoofed self")
	}
	cart := h.add(t, h.buyer, "offer1", 1)
	qraw, e := h.s.CreateQuote(ctx, h.buyer, backoffice.QuoteInput{CartVersion: cart.Version, AddressID: a.ID, Shipping: map[string]string{"store1": "post"}, PaymentMethodID: "local"}, h.key())
	q := decodeCommerce[backoffice.QuoteView](t, qraw, e)
	// Editing a selected address invalidates an unaccepted quote.
	in.ID = a.ID
	in.Version = a.Version
	in.HouseNumber = "99"
	raw, e = h.s.SaveAddress(ctx, h.buyer, in, h.key())
	a = decodeCommerce[commerce.Address](t, raw, e)
	if _, e = h.s.AcceptQuote(ctx, h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman}, h.key()); e != backoffice.ErrCheckoutChanged {
		t.Fatal("stale consent", e)
	}
	qraw, e = h.s.CreateQuote(ctx, h.buyer, backoffice.QuoteInput{CartVersion: cart.Version, AddressID: a.ID, Shipping: map[string]string{"store1": "post"}, PaymentMethodID: "local"}, h.key())
	q = decodeCommerce[backoffice.QuoteView](t, qraw, e)
	p := h.accept(t, h.buyer, q)
	h.approve(t, h.deposit(t, h.buyer, p.ID, "local", p.Quote.TotalToman, "recipient-local-paid"))
	current, err := h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = h.s.ConfirmPurchase(ctx, h.finance, backoffice.PurchaseAdminInput{ID: p.ID, Version: current[0].Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, raw, err)
	if p.Status != "confirmed" {
		t.Fatal("fixture not confirmed")
	}
	before, _ := json.Marshal(p.Quote.Address)
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		for i := range w.Users {
			if w.Users[i].ID == h.buyer.ID {
				w.Users[i].Name = "نام تازه"
				w.Users[i].NationalID = ""
			}
		}
		return nil
	})
	rows, e := h.s.ReadPurchases(ctx, h.buyer, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	after, _ := json.Marshal(rows[0].Quote.Address)
	if string(before) != string(after) {
		t.Fatal("historical snapshot changed")
	}
	workspace, e := h.s.Read(ctx, h.admin)
	if e != nil {
		t.Fatal(e)
	}
	for _, u := range workspace.Users {
		if u.NationalID != "" {
			t.Fatal("national ID leaked to general workspace")
		}
	}
}
func TestLegacyAddressReadableButRequiresCompletionForNewCheckout(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.Addresses = append(w.Commerce.Addresses, commerce.Address{ID: "legacy", UserID: h.buyer.ID, Recipient: "گیرنده قبلی", Phone: "09000000111", Province: "تهران", City: "تهران", Street: "نشانی قدیمی", PostalCode: "1234567890", Version: 1})
		return nil
	})
	rows, e := h.s.ReadAddresses(ctx, h.buyer)
	if e != nil || len(rows) != 1 || rows[0].Recipient != "گیرنده قبلی" {
		t.Fatal("legacy read", e)
	}
	cart := h.add(t, h.buyer, "offer1", 1)
	if _, e = h.s.PreviewCheckout(ctx, h.buyer, backoffice.QuoteInput{CartVersion: cart.Version, AddressID: "legacy", Shipping: map[string]string{"store1": "post"}}); e != backoffice.ErrAddressIncomplete {
		t.Fatal("legacy bypass", e)
	}
}
