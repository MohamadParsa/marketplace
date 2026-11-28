package backoffice_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/orders"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
)

func TestSupplyOverdueUsesPaidSnapshotAtBoundaryWithoutSideEffects(t *testing.T) {
	h := capacityHarness(t, false, 10)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	h.add(t, h.buyer, "offer2", 1)
	p := h.accept(t, h.buyer, capacityQuote(t, h, h.buyer, ""))
	h.approve(t, h.deposit(t, h.buyer, p.ID, "local", p.Quote.TotalToman, "overdue-fund"))
	body, err := h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, body, err)
	if p.Status != "confirmed" {
		t.Fatal(p)
	}
	state, _ := h.repo.Snapshot(ctx)
	deadline := state.Orders[0].Preparation.ReadyAt
	h.clock.now = deadline.Add(-time.Nanosecond)
	view, err := h.s.ReadSupplyOverdue(ctx, h.admin, "", "")
	if err != nil || len(view.Rows) != 0 {
		t.Fatal(view, err)
	}
	h.clock.now = deadline
	view, err = h.s.ReadSupplyOverdue(ctx, h.admin, "", "")
	if err != nil || len(view.Rows) != 2 || view.Rows[0].OverdueMinutes != 0 {
		t.Fatal(view, err)
	}
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Commerce.Settings.ClosedWeekdays = []int{0, 1, 2, 3, 4, 5}
		w.Commerce.Settings.Holidays = []string{"2026-11-21"}
		w.Stores[0].LeadDays = 365
		w.Commerce.Settings.Version++
		return nil
	})
	before, _ := h.repo.Snapshot(ctx)
	for i := 0; i < 4; i++ {
		view, err = h.s.ReadSupplyOverdue(ctx, h.admin, "store1", "")
		if err != nil || len(view.Rows) != 1 || !view.Rows[0].Deadline.Equal(deadline) || !view.Rows[0].PaidAt.Equal(state.Orders[0].Preparation.SourceAt) {
			t.Fatal(view, err)
		}
	}
	after, _ := h.repo.Snapshot(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("read created duplicate alerts or changed order/financial data")
	}
	h.repo.Update(ctx, func(w *backoffice.Workspace) error { w.Orders[0].Status = "supplied"; return nil })
	view, err = h.s.ReadSupplyOverdue(ctx, h.admin, "store1", "")
	if err != nil || len(view.Rows) != 0 {
		t.Fatal("supplied order stayed in queue", view, err)
	}
}

func TestSupplyOverdueScopesMissingDeadlineAndHistoricalHoliday(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	source := time.Date(2026, 11, 25, 11, 0, 0, 0, shipping.Tehran()) // Wednesday; Thu/Fri and holiday Saturday excluded.
	snapshot, err := shipping.ResolvePreparation(source, 1, shipping.PreparationPolicy{DayBasis: shipping.WorkingDays, CalendarMode: shipping.Inherit}, shipping.Calendar{Version: 7, WorkingWeekdays: []int{0, 1, 2, 3, 6}, Holidays: []string{"2026-11-28"}}, shipping.DefaultTiming())
	if err != nil || snapshot.ReadyAt.Format("2006-01-02") != "2026-11-29" {
		t.Fatal(snapshot, err)
	}
	var seller identity.User
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		w.Orders = []orders.Order{{ID: "late-one", StoreID: "store1", PurchaseNumber: "1200", Status: "awaiting-supply", Payment: "paid", Preparation: &snapshot}, {ID: "late-two", StoreID: "store2", PurchaseNumber: "1201", Status: "awaiting-supply", Payment: "paid", Preparation: &snapshot}, {ID: "unpaid", StoreID: "store1", Status: "awaiting-supply", Payment: "unpaid", Preparation: &snapshot}, {ID: "cancelled", StoreID: "store1", Status: "cancelled", Payment: "paid", Preparation: &snapshot}, {ID: "legacy-one", StoreID: "store1", Status: "awaiting-supply", Payment: "paid"}, {ID: "legacy-two", StoreID: "store2", Status: "awaiting-supply", Payment: "paid"}}
		w.Access.Roles = append(w.Access.Roles, identity.RoleDefinition{ID: "deadline-reader", Scope: identity.ScopeStore, Permissions: []string{"orders.read"}, Version: 1})
		w.Access.Grants = append(w.Access.Grants, identity.RoleGrant{UserID: "seller-test", RoleID: "deadline-reader", StoreID: "store1"})
		w.StoreMemberships = append(w.StoreMemberships, identity.StoreMembership{UserID: "seller-test", StoreID: "store1"})
		for _, u := range w.Users {
			if u.ID == "seller-test" {
				seller = identity.User{ID: u.ID, Active: u.Active, AccessVersion: u.AccessVersion, Surface: identity.SurfaceSeller}
			}
		}
		return nil
	})
	h.clock.now = snapshot.ReadyAt.Add(61 * time.Minute)
	view, err := h.s.ReadSupplyOverdue(ctx, seller, "", "")
	if err != nil || len(view.Rows) != 1 || view.Rows[0].OrderID != "late-one" || view.Rows[0].OverdueMinutes != 61 || view.MissingDeadlineCount != 1 {
		t.Fatal(view, err)
	}
	if _, err = h.s.ReadSupplyOverdue(ctx, seller, "store2", ""); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal(err)
	}
	if _, err = h.s.ReadSupplyOverdue(ctx, h.buyer, "", ""); !errors.Is(err, backoffice.ErrForbidden) {
		t.Fatal(err)
	}
	view, err = h.s.ReadSupplyOverdue(ctx, h.admin, "", "۱۲۰۰")
	if err != nil || len(view.Rows) != 1 || view.Rows[0].OrderID != "late-one" {
		t.Fatal(view, err)
	}
	h.repo.Update(ctx, func(w *backoffice.Workspace) error { w.Access.Grants = nil; return nil })
	view, err = h.s.ReadSupplyOverdue(ctx, seller, "", "")
	if err != nil || len(view.Rows) != 0 || view.MissingDeadlineCount != 0 {
		t.Fatal("stale scope leaked", view, err)
	}
}
