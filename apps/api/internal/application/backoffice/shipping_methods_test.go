package backoffice_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
)

func methodInput(id string) backoffice.ShippingMethodInput {
	capacity := 17
	return backoffice.ShippingMethodInput{ID: id, Name: "حمل آزمون", Active: true, AllCities: true, CollectionMode: "prepaid", DeliveryMode: "range", DailyCapacity: &capacity, CategoryMode: "all", FeeToman: "0", MinDays: 1, MaxDays: 3, DayBasis: shipping.CalendarDays}
}

func TestShippingMethodCRUDVersionsRacePermissionsAndFieldErrors(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	in := methodInput("new-method")
	for _, actor := range []identity.User{h.finance, h.buyer} {
		if _, err := h.s.SaveShippingMethod(ctx, actor, in, true, "forbid"); !errors.Is(err, backoffice.ErrForbidden) {
			t.Fatal("permission", err)
		}
	}
	before, _ := h.repo.Snapshot(ctx)
	saved, err := h.s.SaveShippingMethod(ctx, h.admin, in, true, "create-method")
	if err != nil || saved.Version != 1 {
		t.Fatal(saved, err)
	}
	if replay, err := h.s.SaveShippingMethod(ctx, h.admin, in, true, "create-method"); err != nil || !reflect.DeepEqual(saved, replay) {
		t.Fatal("retry", replay, err)
	}
	changed := in
	changed.Name = "تغییر"
	if _, err := h.s.SaveShippingMethod(ctx, h.admin, changed, true, "create-method"); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatal("key reuse", err)
	}
	if _, err := h.s.SaveShippingMethod(ctx, h.admin, in, true, "duplicate"); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatal("duplicate id", err)
	}
	in = saved
	in.Name = "ویرایش آزمون"
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := h.s.SaveShippingMethod(ctx, h.admin, in, false, fmt.Sprintf("update-%d", i))
			results <- err
		}(i)
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, backoffice.ErrConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatal(success, conflict)
	}
	after, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(before.Orders, after.Orders) || !reflect.DeepEqual(before.Ledger, after.Ledger) || !reflect.DeepEqual(before.Offers, after.Offers) || !reflect.DeepEqual(before.Commerce.Purchases, after.Commerce.Purchases) {
		t.Fatal("protected data mutation")
	}
	for _, test := range []struct {
		field  string
		change func(*backoffice.ShippingMethodInput)
	}{
		{"name", func(in *backoffice.ShippingMethodInput) { in.Name = "" }},
		{"collection_mode", func(in *backoffice.ShippingMethodInput) { in.CollectionMode = "unknown" }},
		{"fee_toman", func(in *backoffice.ShippingMethodInput) { in.FeeToman = "-1" }},
		{"fee_toman", func(in *backoffice.ShippingMethodInput) { in.CollectionMode = "collect"; in.FeeToman = "100" }},
		{"min_days", func(in *backoffice.ShippingMethodInput) { in.MinDays = 9; in.MaxDays = 1 }},
		{"daily_capacity", func(in *backoffice.ShippingMethodInput) { in.DailyCapacity = nil }},
		{"daily_capacity", func(in *backoffice.ShippingMethodInput) { value := -1; in.DailyCapacity = &value }},
		{"city_ids", func(in *backoffice.ShippingMethodInput) { in.AllCities = false; in.CityIDs = []string{"unknown"} }},
		{"city_ids", func(in *backoffice.ShippingMethodInput) {
			in.AllCities = false
			in.CityIDs = []string{"tehran-city", "tehran-city"}
		}},
		{"allowed_categories", func(in *backoffice.ShippingMethodInput) { in.CategoryMode = "selected" }},
		{"allowed_categories", func(in *backoffice.ShippingMethodInput) {
			in.CategoryMode = "selected"
			in.AllowedCategories = []string{"gold"}
			in.DeniedCategories = []string{"gold"}
		}},
		{"denied_categories", func(in *backoffice.ShippingMethodInput) { in.DeniedCategories = []string{"unknown"} }},
	} {
		bad := methodInput("bad")
		test.change(&bad)
		_, err := h.s.SaveShippingMethod(ctx, h.admin, bad, true, h.key())
		var field *backoffice.ShippingFieldError
		if !errors.As(err, &field) || field.Field != test.field {
			t.Fatal(test.field, err)
		}
	}
}

func TestShippingCoverageCanonicalCitiesDescendantsMixedShipmentAndNewCity(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	in := methodInput("canonical")
	in.AllCities = false
	in.CityIDs = []string{"tehran-city"}
	in.CategoryMode = "selected"
	in.AllowedCategories = []string{"gold"}
	in.DeniedCategories = []string{"coin"}
	if _, err := h.s.SaveShippingMethod(ctx, h.admin, in, true, h.key()); err != nil {
		t.Fatal(err)
	}
	w, _ := h.repo.Snapshot(ctx)
	method := w.Commerce.Settings.Shipping[1]
	if !backoffice.ShippingAllowsCategory(w, method, "gold") || backoffice.ShippingAllowsCategory(w, method, "coin") || backoffice.ShippingAllowsCategory(w, method, "regular") || backoffice.ShippingAllowsCategory(w, method, "missing") {
		t.Fatal("tree deny priority")
	}
	method.CategoryMode = "all"
	method.DeniedCategories = []string{"gold"}
	address := commerce.Address{Province: "تهران", City: "تهران"}
	lines := []commerce.LineSnapshot{{StoreID: "store1", ProductID: "phone"}}
	if !backoffice.ShippingMethodEligible(w, method, address, "store1", lines) {
		t.Fatal("canonical city")
	}
	lines = append(lines, commerce.LineSnapshot{StoreID: "store1", ProductID: "gold-product"})
	if backoffice.ShippingMethodEligible(w, method, address, "store1", lines) {
		t.Fatal("one forbidden item escaped")
	}
	if backoffice.ShippingMethodEligible(w, method, commerce.Address{Province: "قم", City: "قم"}, "store1", lines[:1]) {
		t.Fatal("wrong city")
	}
	w.Commerce.Settings.Geography[1].Cities = append(w.Commerce.Settings.Geography[1].Cities, commerce.City{ID: "new-city", Name: "شهر تازه", Active: true})
	method.AllCities = true
	if !backoffice.ShippingMethodEligible(w, method, commerce.Address{Province: "قم", City: "شهر تازه"}, "store1", lines[:1]) {
		t.Fatal("all cities excludes new active city")
	}
	w.Commerce.Settings.Geography[1].Cities = append(w.Commerce.Settings.Geography[1].Cities, commerce.City{ID: "other-tehran", Name: "تهران", Active: true})
	legacy := method
	legacy.AllCities = false
	legacy.CityIDs = nil
	legacy.Cities = []string{"تهران"}
	if backoffice.ShippingMethodEligible(w, legacy, address, "store1", lines[:1]) {
		t.Fatal("ambiguous legacy city escaped")
	}
}

func TestShippingMethodDirectoryFiltersAndPagination(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	for i := 0; i < 4; i++ {
		in := methodInput(fmt.Sprintf("ship-%d", i))
		in.Name = fmt.Sprintf("روش %d", i)
		if i%2 == 0 {
			in.CollectionMode = "collect"
			in.FeeToman = ""
		}
		if _, err := h.s.SaveShippingMethod(ctx, h.admin, in, true, h.key()); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := h.s.ReadShippingMethods(ctx, h.admin, backoffice.ShippingMethodFilter{CollectionMode: "collect", Size: 1, Page: 1})
	if err != nil || rows.Total != 2 || len(rows.Rows) != 1 || rows.Rows[0].CollectionMode != "collect" || rows.Rows[0].FeeToman != "" {
		t.Fatal(rows, err)
	}
	all, err := h.s.ReadShippingMethods(ctx, h.admin, backoffice.ShippingMethodFilter{Size: 2, Page: 99})
	if err != nil || all.Total != 5 || all.Page != 2 || len(all.Rows) != 1 {
		t.Fatal("pagination", all, err)
	}
	legacy, err := h.s.ReadShippingMethods(ctx, h.admin, backoffice.ShippingMethodFilter{Query: "post"})
	if err != nil || len(legacy.Rows) != 1 || legacy.Rows[0].CollectionMode != "prepaid" || legacy.Rows[0].DeliveryMode != "range" || legacy.Rows[0].DayBasis != shipping.WorkingDays || legacy.Rows[0].DailyCapacity != nil {
		t.Fatal("legacy defaults", legacy, err)
	}
}

func TestShippingMethodRetryRechecksCurrentActor(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	in := methodInput("revoked")
	if _, err := h.s.SaveShippingMethod(ctx, h.admin, in, true, "retry-access"); err != nil {
		t.Fatal(err)
	}
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		for index := range w.Users {
			if w.Users[index].ID == h.admin.ID {
				w.Users[index].Active = false
			}
		}
		return nil
	})
	if _, err := h.s.SaveShippingMethod(ctx, h.admin, in, true, "retry-access"); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("revoked retry", err)
	}
}

func TestShippingCollectExcludedFromTotalsAndHistoricalSnapshot(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	in := methodInput("collect")
	in.CollectionMode = "collect"
	in.FeeToman = ""
	method, err := h.s.SaveShippingMethod(ctx, h.admin, in, true, h.key())
	if err != nil {
		t.Fatal(err)
	}
	h.add(t, h.buyer, "offer1", 1)
	cart := h.cart(t, h.buyer)
	address := h.address(t, h.buyer)
	b, e := h.s.CreateQuote(ctx, h.buyer, backoffice.QuoteInput{CartVersion: cart.Version, AddressID: address.ID, Shipping: map[string]string{"store1": "collect"}, PaymentMethodID: "local"}, h.key())
	q := decodeCommerce[backoffice.QuoteView](t, b, e)
	if q.ShippingToman != "0" || q.TotalToman != "100000" || q.Shipments[0].CollectionMode != "collect" || q.Shipments[0].FeeToman != "" {
		t.Fatal("collect money", q)
	}
	p := h.accept(t, h.buyer, q)
	d := h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "collect-funding")
	h.approve(t, d)
	b, e = h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	_ = decodeCommerce[backoffice.PurchaseView](t, b, e)
	before, _ := h.repo.Snapshot(ctx)
	method.Name = "نام جدید"
	method.CollectionMode = "prepaid"
	method.FeeToman = "10000"
	if _, err = h.s.SaveShippingMethod(ctx, h.admin, method, false, h.key()); err != nil {
		t.Fatal(err)
	}
	after, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(before.Orders, after.Orders) || !reflect.DeepEqual(before.Commerce.Quotes, after.Commerce.Quotes) || !reflect.DeepEqual(before.Commerce.Entries, after.Commerce.Entries) || after.Orders[0].ShippingFee != 0 {
		t.Fatal("history/ledger changed")
	}
}

func TestShippingRestrictionsRecheckedAtAcceptAndFinalPayment(t *testing.T) {
	for _, stage := range []string{"accept", "payment"} {
		t.Run(stage, func(t *testing.T) {
			h := newCommerceHarness(t)
			ctx := context.Background()
			h.add(t, h.buyer, "offer1", 1)
			q := h.quote(t, h.buyer, "local")
			var p backoffice.PurchaseView
			if stage == "payment" {
				p = h.accept(t, h.buyer, q)
				d := h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "banned-funding")
				h.approve(t, d)
			}
			settings, _ := h.s.ReadCommerceSettings(ctx, h.admin)
			method := settings.Shipping[0]
			method.DeniedCategories = []string{"regular"}
			if _, err := h.s.SaveShippingMethod(ctx, h.admin, method, false, h.key()); err != nil {
				t.Fatal(err)
			}
			if stage == "accept" {
				if _, err := h.s.AcceptQuote(ctx, h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman}, h.key()); err == nil {
					t.Fatal("accepted banned method")
				}
			} else {
				b, e := h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
				p = decodeCommerce[backoffice.PurchaseView](t, b, e)
				if p.Status != "requires-reconfirmation" {
					t.Fatal("payment bypass", p)
				}
			}
		})
	}
}
