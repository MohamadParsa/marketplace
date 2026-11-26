package backoffice_test

import (
	"context"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
	"reflect"
	"sync"
	"testing"
	"time"
)

func preparationInput() backoffice.StorePreparationInput {
	return backoffice.StorePreparationInput{LeadDays: 2, DayBasis: shipping.CalendarDays, CalendarMode: shipping.Override, WorkingWeekdays: []int{0, 1, 2, 3, 6}}
}

func TestStorePreparationPermissionsRetryValidationAndRacingVersions(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	actor := h.buyer
	actor.Surface = identity.SurfaceSeller
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.StoreMemberships = append(w.StoreMemberships, identity.StoreMembership{UserID: actor.ID, StoreID: "store1"})
		w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "prep-test", Scope: identity.ScopeStore, Permissions: []string{backoffice.StoresPreparationRead, backoffice.StoresPreparationWrite}})
		w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: actor.ID, StoreID: "store1", RoleID: "prep-test"})
		return nil
	})
	rows, err := h.s.ReadStorePreparations(ctx, actor)
	if err != nil || len(rows) != 1 || rows[0].ID != "store1" || rows[0].Policy.DayBasis != shipping.WorkingDays || rows[0].Policy.CalendarMode != shipping.Inherit {
		t.Fatal("scoped legacy read", rows, err)
	}
	in := preparationInput()
	for _, who := range []identity.User{h.finance, h.buyer} {
		if _, err := h.s.SaveStorePreparation(ctx, who, "store1", in, "forbid"); !errors.Is(err, backoffice.ErrForbidden) {
			t.Fatal("forbidden", err)
		}
	}
	if _, err := h.s.SaveStorePreparation(ctx, actor, "store2", in, "other"); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("other store", err)
	}
	before, _ := h.repo.Snapshot(ctx)
	for _, bad := range []backoffice.StorePreparationInput{{LeadDays: -1, DayBasis: shipping.CalendarDays, CalendarMode: shipping.Inherit}, {LeadDays: 366, DayBasis: shipping.CalendarDays, CalendarMode: shipping.Inherit}, {DayBasis: "invalid", CalendarMode: shipping.Inherit}, {DayBasis: shipping.WorkingDays, CalendarMode: shipping.Override}, {DayBasis: shipping.WorkingDays, CalendarMode: shipping.Override, WorkingWeekdays: []int{6, 6}}, {DayBasis: shipping.WorkingDays, CalendarMode: shipping.Inherit, WorkingWeekdays: []int{6}}} {
		if _, err := h.s.SaveStorePreparation(ctx, actor, "store1", bad, "invalid"); !errors.Is(err, backoffice.ErrInvalid) {
			t.Fatal("invalid", bad, err)
		}
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, key := range []string{"race1", "race2"} {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			_, e := h.s.SaveStorePreparation(ctx, actor, "store1", in, key)
			results <- e
		}(key)
	}
	wg.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		if err == nil {
			successes++
		} else if errors.Is(err, backoffice.ErrConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatal("race", successes, conflicts)
	}
	current, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(before.Orders, current.Orders) || !reflect.DeepEqual(before.Commerce, current.Commerce) || !reflect.DeepEqual(before.Offers, current.Offers) {
		t.Fatal("unrelated mutation")
	}
	var savedKey string
	for _, event := range current.Audit {
		if event.Action == "stores.preparation.save" {
			savedKey = event.ID
		}
	}
	row, err := h.s.SaveStorePreparation(ctx, actor, "store1", in, savedKey)
	if err != nil || row.Version != 1 || row.Policy.Version != 1 {
		t.Fatal("retry", row, err)
	}
	in.LeadDays = 3
	if _, err = h.s.SaveStorePreparation(ctx, actor, "store1", in, savedKey); !errors.Is(err, backoffice.ErrConflict) {
		t.Fatal("key reuse", err)
	}
	h.repo.Update(ctx, func(w *backoffice.Workspace) error { w.Access.Grants = nil; return nil })
	in.LeadDays = 2
	if _, err = h.s.SaveStorePreparation(ctx, actor, "store1", in, savedKey); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal("revoked retry", err)
	}
}

func TestStorePreparationTwoStageCheckoutAndPaidSnapshot(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	start, _ := time.Parse(time.RFC3339, "2026-11-21T08:00:00+03:30")
	h.clock.now = start
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.Settings.ClosedWeekdays = []int{4, 5}
		w.Commerce.Settings.Holidays = []string{"2026-11-22"}
		w.Commerce.Settings.Shipping[0].DayBasis = shipping.CalendarDays
		return nil
	})
	in := preparationInput()
	in.DayBasis = shipping.WorkingDays
	in.CalendarMode = shipping.Inherit
	in.WorkingWeekdays = nil
	row, err := h.s.SaveStorePreparation(ctx, h.admin, "store1", in, h.key())
	if err != nil {
		t.Fatal(err)
	}
	in.Version = row.Version
	preview, err := h.s.PreviewStorePreparation(ctx, h.admin, "store1", in)
	if err != nil || preview.ReadyAt.Format("2006-01-02") != "2026-11-24" {
		t.Fatal("shared preview", preview, err)
	}
	h.clock.now = start.Add(3*time.Hour + 50*time.Minute)
	h.add(t, h.buyer, "offer1", 1)
	q := h.quote(t, h.buyer, "local")
	if q.Shipments[0].DeliveryFrom != "2026-11-25" || q.Shipments[0].DeliveryTo != "2026-11-27" || q.Shipments[0].LeadDayBasis != shipping.WorkingDays || q.Shipments[0].TransportDayBasis != shipping.CalendarDays {
		t.Fatal("two stage", q.Shipments)
	}
	p := h.accept(t, h.buyer, q)
	d := h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "prep-funding")
	h.approve(t, d)
	h.clock.Advance(10 * time.Minute) // Exactly the cutoff: payment, not the deposit, starts preparation.
	b, e := h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, b, e)
	paid, _ := h.repo.Snapshot(ctx)
	snap := paid.Orders[0].Preparation
	if snap == nil || !snap.SourceAt.Equal(h.clock.Now()) || snap.ReadyAt.Format("2006-01-02") != "2026-11-25" || paid.Orders[0].DeliveryFrom != "2026-11-26" {
		t.Fatal("paid origin", snap, paid.Orders[0])
	}
	in.DayBasis = shipping.CalendarDays
	in.LeadDays = 9
	if _, err = h.s.SaveStorePreparation(ctx, h.admin, "store1", in, h.key()); err != nil {
		t.Fatal(err)
	}
	after, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(paid.Orders, after.Orders) || !reflect.DeepEqual(paid.Commerce.Quotes, after.Commerce.Quotes) {
		t.Fatal("history rewritten")
	}
}

func TestStorePreparationSiteChangesOnlyReplaceInheritedWeekAndValidateSettings(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	in := preparationInput()
	if _, err := h.s.SaveStorePreparation(ctx, h.admin, "store2", in, h.key()); err != nil {
		t.Fatal(err)
	}
	settings, err := h.s.ReadCommerceSettings(ctx, h.admin)
	if err != nil {
		t.Fatal(err)
	}
	timing := shipping.TimingPolicy{ReadyTime: "08:30", CutoffTime: "13:00", MinSlotLeadMinutes: 90, AllowSameDaySlots: false}
	input := backoffice.CommerceSettingsInput{Version: settings.Version, Shipping: &backoffice.ShippingSettingsInput{Methods: settings.Shipping, ClosedWeekdays: []int{0, 1, 2, 3, 4, 5}, Holidays: []string{"2026-12-06"}, Timing: &timing}}
	b, e := h.s.SaveCommerceSettings(ctx, h.admin, input, h.key())
	_ = decodeCommerce[backoffice.CommerceSettingsView](t, b, e)
	rows, err := h.s.ReadStorePreparations(ctx, h.admin)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.ID == "store1" && !reflect.DeepEqual(row.EffectiveCalendar.WorkingWeekdays, []int{6}) {
			t.Fatal("inherit", row)
		}
		if row.ID == "store2" && !reflect.DeepEqual(row.EffectiveCalendar.WorkingWeekdays, []int{0, 1, 2, 3, 6}) {
			t.Fatal("override", row)
		}
		if row.Timing != timing || !reflect.DeepEqual(row.EffectiveCalendar.Holidays, []string{"2026-12-06"}) {
			t.Fatal("central timing/holidays", row)
		}
	}
	current, _ := h.repo.Snapshot(ctx)
	input.Version = current.Commerce.Settings.Version
	for _, holidays := range [][]string{{"2026-12-06", "2026-12-06"}, {"2026-02-30"}} {
		input.Shipping.Holidays = holidays
		if _, err := h.s.SaveCommerceSettings(ctx, h.admin, input, h.key()); !errors.Is(err, backoffice.ErrInvalid) {
			t.Fatal("holidays", err)
		}
	}
	input.Shipping.Holidays = nil
	input.Shipping.ClosedWeekdays = []int{0, 1, 2, 3, 4, 5, 6}
	if _, err := h.s.SaveCommerceSettings(ctx, h.admin, input, h.key()); !errors.Is(err, backoffice.ErrInvalid) {
		t.Fatal("no working day", err)
	}
	after, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(current.Commerce.Settings, after.Commerce.Settings) {
		t.Fatal("invalid settings persisted")
	}
}
