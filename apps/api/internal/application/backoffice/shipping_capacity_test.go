package backoffice_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
	"sync"
	"testing"
	"time"
)

func quota(n int) *int { return &n }
func capacitySchedule(slotLimit int) *shipping.Schedule {
	schedule := &shipping.Schedule{Windows: []shipping.TimeWindow{{ID: "am", Start: "11:00", End: "12:00", Active: true}, {ID: "pm", Start: "13:00", End: "14:00", Active: true}}, Days: []shipping.WeeklyDay{}}
	for weekday := 0; weekday < 7; weekday++ {
		schedule.Days = append(schedule.Days, shipping.WeeklyDay{Weekday: weekday, Active: true, Slots: []shipping.WindowAssignment{{WindowID: "am", Capacity: quota(slotLimit)}, {WindowID: "pm", Capacity: quota(slotLimit)}}})
	}
	return schedule
}
func capacityHarness(t *testing.T, slots bool, daily int) *commerceHarness {
	h := newCommerceHarness(t)
	h.clock.now, _ = time.Parse(time.RFC3339, "2026-11-21T08:00:00+03:30")
	if err := h.repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		w.Commerce.Settings.ClosedWeekdays = []int{}
		w.Commerce.Settings.Holidays = []string{}
		for i := range w.Stores {
			w.Stores[i].LeadDays = 0
		}
		m := &w.Commerce.Settings.Shipping[0]
		m.DailyCapacity = quota(daily)
		m.MinDays = 1
		m.MaxDays = 3
		m.DayBasis = shipping.CalendarDays
		if slots {
			m.DeliveryMode = "slots"
			m.Schedule = capacitySchedule(1)
			m.MinDays = 0
			m.MaxDays = 0
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return h
}
func capacityQuote(t *testing.T, h *commerceHarness, actor identity.User, slot string) backoffice.QuoteView {
	t.Helper()
	cart := h.cart(t, actor)
	address := h.address(t, actor)
	in := backoffice.QuoteInput{CartVersion: cart.Version, AddressID: address.ID, PaymentMethodID: "local", Shipping: map[string]string{}}
	for _, item := range cart.Items {
		in.Shipping[item.StoreID] = "post"
		if slot != "" {
			if in.ShippingSelections == nil {
				in.ShippingSelections = map[string]shipping.DeliverySelection{}
			}
			in.ShippingSelections[item.StoreID] = shipping.DeliverySelection{Date: "2026-11-21", WindowID: slot}
		}
	}
	body, err := h.s.CreateQuote(context.Background(), actor, in, h.key())
	return decodeCommerce[backoffice.QuoteView](t, body, err)
}
func capacityRead(t *testing.T, h *commerceHarness, day string) backoffice.ShippingCapacityView {
	t.Helper()
	view, err := h.s.ReadShippingCapacity(context.Background(), h.admin, "store1", "post", day)
	if err != nil {
		t.Fatal(err)
	}
	return view
}
func TestShippingCapacityLastSlotRaceCountsOnePerShipment(t *testing.T) {
	h := capacityHarness(t, true, 2)
	h.add(t, h.buyer, "offer1", 3)
	h.add(t, h.other, "offer1", 1)
	q1 := capacityQuote(t, h, h.buyer, "am")
	q2 := capacityQuote(t, h, h.other, "am")
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i, item := range []struct {
		actor identity.User
		q     backoffice.QuoteView
	}{{h.buyer, q1}, {h.other, q2}} {
		wg.Add(1)
		go func(i int, actor identity.User, q backoffice.QuoteView) {
			defer wg.Done()
			_, err := h.s.AcceptQuote(context.Background(), actor, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman}, fmt.Sprintf("last-slot-%d", i))
			results <- err
		}(i, item.actor, item.q)
	}
	wg.Wait()
	close(results)
	success, full := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, backoffice.ErrShippingFull) {
			full++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || full != 1 {
		t.Fatal(success, full)
	}
	view := capacityRead(t, h, "2026-11-21")
	if view.Used != 1 || view.Held != 1 || view.Slots[0].Remaining != 0 || view.Slots[1].Remaining != 1 {
		t.Fatal(view)
	}
	if view.Preview == nil || len(view.Preview.Days) != 6 || view.Preview.Days[5].Date != "2026-11-26" {
		t.Fatal("admin preview differs from checkout engine", view.Preview)
	}
	w, _ := h.repo.Snapshot(context.Background())
	if len(w.Commerce.ShippingReservations) != 1 {
		t.Fatal("duplicate/partial reservation")
	}
}
func TestShippingCapacityDailyLimitAcrossDifferentSlotsAndIndependentStores(t *testing.T) {
	h := capacityHarness(t, true, 1)
	h.add(t, h.buyer, "offer1", 1)
	h.add(t, h.other, "offer1", 1)
	first := capacityQuote(t, h, h.buyer, "am")
	second := capacityQuote(t, h, h.other, "pm")
	h.accept(t, h.buyer, first)
	if _, err := h.s.AcceptQuote(context.Background(), h.other, backoffice.AcceptQuoteInput{QuoteID: second.ID, AcceptedTotalToman: second.TotalToman}, h.key()); err == nil {
		t.Fatal("daily ceiling bypassed through another slot")
	}
	if view := capacityRead(t, h, "2026-11-21"); view.Used != 1 || view.Remaining != 0 || view.Slots[1].Remaining != 0 {
		t.Fatal(view)
	}
	h.add(t, h.other, "offer2", 1)
	view, err := h.s.ReadShippingCapacity(context.Background(), h.admin, "store2", "post", "2026-11-21")
	if err != nil || view.Used != 0 || view.Remaining != 1 {
		t.Fatal("store pools merged", view, err)
	}
}
func TestShippingCapacityHoldRetryExpiryAndLateFundingNeverRenew(t *testing.T) {
	h := capacityHarness(t, true, 1)
	h.add(t, h.buyer, "offer1", 1)
	q := capacityQuote(t, h, h.buyer, "am")
	p := h.accept(t, h.buyer, q)
	w, _ := h.repo.Snapshot(context.Background())
	expiry := w.Commerce.ShippingReservations[0].ExpiresAt
	h.clock.Advance(10 * time.Minute)
	if _, err := h.s.AcceptQuote(context.Background(), h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman}, "retry-different-key"); err != nil {
		t.Fatal(err)
	}
	w, _ = h.repo.Snapshot(context.Background())
	if len(w.Commerce.ShippingReservations) != 1 || !w.Commerce.ShippingReservations[0].ExpiresAt.Equal(expiry) {
		t.Fatal("hold renewed")
	}
	d := h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "late-capacity-funding")
	h.clock.Advance(20 * time.Minute)
	if err := h.s.ReleaseExpiredCommerce(context.Background()); err != nil {
		t.Fatal(err)
	}
	h.approve(t, d)
	ps, err := h.s.ReadPurchases(context.Background(), h.buyer, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	body, err := h.s.ConfirmPurchase(context.Background(), h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: ps[0].Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, body, err)
	w, _ = h.repo.Snapshot(context.Background())
	if p.Status != "requires-reconfirmation" || w.Commerce.ShippingReservations[0].Status != "released" || len(w.Commerce.Entries) != 1 || capacityRead(t, h, "2026-11-21").Used != 0 {
		t.Fatal("late funding committed old slot", p.Status, w.Commerce.ShippingReservations)
	}
}
func TestShippingCapacityCommitReductionCancelAndAfterShippingProtection(t *testing.T) {
	for _, sent := range []bool{false, true} {
		t.Run(fmt.Sprint(sent), func(t *testing.T) {
			h := capacityHarness(t, true, 1)
			h.add(t, h.buyer, "offer1", 1)
			q := capacityQuote(t, h, h.buyer, "am")
			p := h.accept(t, h.buyer, q)
			h.repo.Update(context.Background(), func(w *backoffice.Workspace) error {
				m := &w.Commerce.Settings.Shipping[0]
				m.DailyCapacity = quota(0)
				for i := range m.Schedule.Days {
					m.Schedule.Days[i].Slots[0].Capacity = quota(0)
				}
				m.Version++
				return nil
			})
			h.approve(t, h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "commit-capacity"))
			body, err := h.s.ConfirmPurchase(context.Background(), h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
			p = decodeCommerce[backoffice.PurchaseView](t, body, err)
			if p.Status != "confirmed" {
				t.Fatal("reduction cancelled previous hold", p.Status)
			}
			view := capacityRead(t, h, "2026-11-21")
			if view.Used != 1 || view.Held != 0 || view.Committed != 1 || view.Remaining != 0 {
				t.Fatal(view)
			}
			if sent {
				h.repo.Update(context.Background(), func(w *backoffice.Workspace) error { w.Orders[0].Status = "shipped"; return nil })
			}
			body, err = h.s.AdminCancelPurchase(context.Background(), h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version, Reason: "لغو آزمون ظرفیت"}, h.key())
			if sent {
				if !errors.Is(err, backoffice.ErrConflict) || capacityRead(t, h, "2026-11-21").Used != 1 {
					t.Fatal("released after dispatch", err)
				}
				return
			}
			p = decodeCommerce[backoffice.PurchaseView](t, body, err)
			if p.Status != "cancelled" || capacityRead(t, h, "2026-11-21").Used != 0 {
				t.Fatal("not released")
			}
			body, err = h.s.AdminCancelPurchase(context.Background(), h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version, Reason: "تکرار لغو"}, h.key())
			decodeCommerce[backoffice.PurchaseView](t, body, err)
			w, _ := h.repo.Snapshot(context.Background())
			if len(w.Commerce.Entries) != 3 {
				t.Fatal("financial release repeated")
			}
		})
	}
}
func TestShippingCapacityRevisionRenameCannotResetAndWindowIsSixDates(t *testing.T) {
	h := capacityHarness(t, true, 2)
	h.add(t, h.buyer, "offer1", 1)
	q := capacityQuote(t, h, h.buyer, "am")
	h.accept(t, h.buyer, q)
	h.repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		m := &w.Commerce.Settings.Shipping[0]
		m.Version++
		m.Schedule.Windows[0].ID = "renamed-am"
		for i := range m.Schedule.Days {
			m.Schedule.Days[i].Slots[0].WindowID = "renamed-am"
		}
		return nil
	})
	view := capacityRead(t, h, "2026-11-21")
	if view.Slots[0].Remaining != 0 {
		t.Fatal("revision reset pool", view)
	}
	h.add(t, h.other, "offer1", 1)
	address := h.address(t, h.other)
	choices, err := h.s.ReadShippingChoices(context.Background(), h.other, address.ID, "")
	if err != nil || len(choices) != 1 {
		t.Fatal(choices, err)
	}
	days := choices[0].Methods[0].Days
	if len(days) != 6 || days[0].Date != "2026-11-21" || days[5].Date != "2026-11-26" {
		t.Fatal("six calendar dates", days)
	}
	for _, slot := range days[0].Slots {
		if slot.Plan.Slot.WindowID == "renamed-am" {
			t.Fatal("renamed full slot returned")
		}
	}
	// Fewer open dates must not extend the fixed window to collect six free days.
	h.repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		m := &w.Commerce.Settings.Shipping[0]
		m.Schedule.Days = []shipping.WeeklyDay{m.Schedule.Days[6], m.Schedule.Days[3]}
		return nil
	})
	choices, err = h.s.ReadShippingChoices(context.Background(), h.other, address.ID, "")
	if err != nil || len(choices[0].Methods[0].Days) != 6 || choices[0].Methods[0].Days[5].Date != "2026-11-26" || choices[0].Methods[0].Days[1].Open {
		t.Fatal("extended across closed days", choices, err)
	}
}
func TestShippingCapacityRangeFullDay500AndExplicitConfiguration(t *testing.T) {
	h := capacityHarness(t, false, 500)
	h.add(t, h.buyer, "offer1", 1)
	q := capacityQuote(t, h, h.buyer, "")
	h.repo.Update(context.Background(), func(w *backoffice.Workspace) error {
		for i := 0; i < 500; i++ {
			w.Commerce.ShippingReservations = append(w.Commerce.ShippingReservations, commerce.ShippingReservation{QuoteID: fmt.Sprintf("old-%d", i), StoreID: "store1", MethodID: "post", Status: "committed", Plan: shipping.DeliveryPlan{CapacityDay: "2026-11-21"}})
		}
		return nil
	})
	if _, err := h.s.AcceptQuote(context.Background(), h.buyer, backoffice.AcceptQuoteInput{QuoteID: q.ID, AcceptedTotalToman: q.TotalToman}, h.key()); !errors.Is(err, backoffice.ErrCheckoutChanged) {
		t.Fatal("silently moved reserved range", err)
	}
	view := capacityRead(t, h, "2026-11-21")
	if view.Used != 500 || view.Remaining != 0 {
		t.Fatal(view)
	}
	address := h.address(t, h.buyer)
	choices, err := h.s.ReadShippingChoices(context.Background(), h.buyer, address.ID, "")
	if err != nil || choices[0].Methods[0].Range.DispatchDate != "2026-11-22" || choices[0].Methods[0].Range.DeliveryFrom != "2026-11-23" {
		t.Fatal(choices, err)
	}
	h.repo.Update(context.Background(), func(w *backoffice.Workspace) error { w.Commerce.Settings.Shipping[0].DailyCapacity = nil; return nil })
	choices, err = h.s.ReadShippingChoices(context.Background(), h.buyer, address.ID, "")
	if err != nil || len(choices[0].Methods) != 0 {
		t.Fatal("nil became unlimited", choices, err)
	}
}

func TestShippingCapacityApprovedCustomScheduleIsIndependentOfPublicEdits(t *testing.T) {
	h := capacityHarness(t, false, 2)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	method := methodInput("post")
	method.Name = "پست نمونه"
	method.DeliveryMode = "slots"
	method.Schedule = capacitySchedule(1)
	method.MinDays = 0
	method.MaxDays = 0
	method.DailyCapacity = quota(2)
	in := shippingCommand("draft", 0)
	in.Method = &method
	if _, err := h.s.SaveStoreShipping(ctx, h.admin, in, "custom-slots"); err != nil {
		t.Fatal(err)
	}
	address := h.address(t, h.buyer)
	choices, err := h.s.ReadShippingChoices(ctx, h.buyer, address.ID, "")
	if err != nil || len(choices[0].Methods) != 0 {
		t.Fatal("draft slots leaked", choices, err)
	}
	if _, err = h.s.SaveStoreShipping(ctx, h.admin, shippingCommand("submit", 1), "submit-slots"); err != nil {
		t.Fatal(err)
	}
	approve := shippingCommand("approve", 2)
	approve.Reason = "برنامهٔ اختصاصی بررسی شد"
	if _, err = h.s.SaveStoreShipping(ctx, h.admin, approve, "approve-slots"); err != nil {
		t.Fatal(err)
	}
	q := capacityQuote(t, h, h.buyer, "am")
	p := h.accept(t, h.buyer, q)
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		base := &w.Commerce.Settings.Shipping[0]
		base.DeliveryMode = "slots"
		base.Schedule = capacitySchedule(1)
		base.Schedule.Windows[0].Start = "15:00"
		base.Schedule.Windows[0].End = "16:00"
		base.Version++
		return nil
	})
	choices, err = h.s.ReadShippingChoices(ctx, h.buyer, address.ID, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, slot := range choices[0].Methods[0].Days[0].Slots {
		if slot.Plan.Slot.WindowID == "am" {
			found = true
			if slot.Plan.Slot.StartAt.In(shipping.Tehran()).Format("15:04") != "11:00" {
				t.Fatal("public schedule overwrote custom live")
			}
		}
	}
	if !found {
		t.Fatal("custom missing", choices)
	}
	w, _ := h.repo.Snapshot(ctx)
	if w.Commerce.Quotes[0].Shipments[0].Delivery.Slot.StartAt.In(shipping.Tehran()).Format("15:04") != "11:00" {
		t.Fatal("historical plan changed")
	}
}

func TestShippingCapacityFailedPaymentRollsBackCommitForAllShipments(t *testing.T) {
	h := capacityHarness(t, false, 2)
	h.add(t, h.buyer, "offer1", 1)
	h.add(t, h.buyer, "offer2", 1)
	q := capacityQuote(t, h, h.buyer, "")
	p := h.accept(t, h.buyer, q)
	if _, err := h.s.ConfirmPurchase(context.Background(), h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key()); !errors.Is(err, backoffice.ErrInsufficientWallet) {
		t.Fatal(err)
	}
	w, _ := h.repo.Snapshot(context.Background())
	if len(w.Commerce.ShippingReservations) != 2 || len(w.Commerce.Entries) != 0 {
		t.Fatal("partial payment state")
	}
	for _, r := range w.Commerce.ShippingReservations {
		if r.Status != "held" || r.CommittedAt != nil {
			t.Fatal("commit escaped financial rollback", r)
		}
	}
}

func TestShippingCapacityProgramChangeRequiresFreshChoiceBeforePayment(t *testing.T) {
	h := capacityHarness(t, true, 1)
	ctx := context.Background()
	h.add(t, h.buyer, "offer1", 1)
	q := capacityQuote(t, h, h.buyer, "am")
	p := h.accept(t, h.buyer, q)
	h.approve(t, h.deposit(t, h.buyer, p.ID, "local", q.TotalToman, "changed-program"))
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		m := &w.Commerce.Settings.Shipping[0]
		m.Schedule.Windows[0].Start = "11:30"
		m.Schedule.Windows[0].End = "12:30"
		m.Version++
		return nil
	})
	body, err := h.s.ConfirmPurchase(ctx, h.admin, backoffice.PurchaseAdminInput{ID: p.ID, Version: p.Version}, h.key())
	p = decodeCommerce[backoffice.PurchaseView](t, body, err)
	w, _ := h.repo.Snapshot(ctx)
	if p.Status != "requires-reconfirmation" || w.Commerce.ShippingReservations[0].Status != "released" || len(w.Commerce.Entries) != 1 || w.Commerce.Quotes[0].Shipments[0].Delivery.Slot.StartAt.In(shipping.Tehran()).Format("15:04") != "11:00" {
		t.Fatal("silent slot replacement or debit", p.Status, w.Commerce.ShippingReservations)
	}
}

func TestShippingCapacityThreeHundredCeilingAndReadiness(t *testing.T) {
	h := capacityHarness(t, true, 300)
	ctx := context.Background()
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		m := &w.Commerce.Settings.Shipping[0]
		m.Schedule.Windows = append(m.Schedule.Windows, shipping.TimeWindow{ID: "evening", Start: "17:00", End: "18:00", Active: true})
		for i := range m.Schedule.Days {
			m.Schedule.Days[i].Slots = []shipping.WindowAssignment{{WindowID: "am", Capacity: quota(100)}, {WindowID: "pm", Capacity: quota(100)}, {WindowID: "evening", Capacity: quota(100)}}
		}
		date, _ := time.ParseInLocation("2006-01-02", "2026-11-21", shipping.Tehran())
		for _, slot := range m.Schedule.Occurrences(date) {
			for i := 0; i < 100; i++ {
				copy := slot
				w.Commerce.ShippingReservations = append(w.Commerce.ShippingReservations, commerce.ShippingReservation{QuoteID: fmt.Sprintf("capacity-%s-%d", slot.WindowID, i), StoreID: "store1", MethodID: "post", Status: "committed", Plan: shipping.DeliveryPlan{CapacityDay: "2026-11-21", Slot: &copy}})
			}
		}
		return nil
	})
	view := capacityRead(t, h, "2026-11-21")
	if view.Used != 300 || view.Remaining != 0 {
		t.Fatal(view)
	}
	for _, slot := range view.Slots {
		if slot.Used != 100 || slot.Remaining != 0 {
			t.Fatal(slot)
		}
	}
	h.add(t, h.buyer, "offer1", 1)
	address := h.address(t, h.buyer)
	choices, err := h.s.ReadShippingChoices(ctx, h.buyer, address.ID, "")
	if err != nil || choices[0].Methods[0].Days[0].Date != "2026-11-22" {
		t.Fatal("301st offered on full day", choices, err)
	}
	h = capacityHarness(t, true, 2)
	h.add(t, h.buyer, "offer1", 1)
	address = h.address(t, h.buyer)
	h.repo.Update(ctx, func(w *backoffice.Workspace) error {
		policy := shipping.DefaultTiming()
		policy.ReadyTime = "10:01"
		w.Commerce.Settings.ShippingTiming = &policy
		return nil
	})
	choices, err = h.s.ReadShippingChoices(ctx, h.buyer, address.ID, "")
	if err != nil || len(choices[0].Methods[0].Days[0].Slots) != 1 || choices[0].Methods[0].Days[0].Slots[0].Plan.Slot.WindowID != "pm" {
		t.Fatal("insufficient preparation opportunity", choices, err)
	}
}
