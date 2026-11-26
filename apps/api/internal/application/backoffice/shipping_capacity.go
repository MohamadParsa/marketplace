package backoffice

import (
	"context"
	"errors"
	"slices"
	"sort"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
)

var ErrShippingUnavailable = errors.New("shipping_unavailable")
var ErrShippingFull = errors.New("slot_full")
var ErrShippingExpired = errors.New("slot_expired")

const shippingSearchDays = 366

type ShippingSlotAvailability struct {
	Plan      shipping.DeliveryPlan `json:"plan"`
	Remaining int                   `json:"remaining"`
}
type ShippingDayAvailability struct {
	Date           string                     `json:"date"`
	Open           bool                       `json:"open"`
	DailyRemaining int                        `json:"daily_remaining"`
	Slots          []ShippingSlotAvailability `json:"slots"`
}
type ShippingMethodAvailability struct {
	Method         ShippingMethodInput       `json:"method"`
	Range          *shipping.DeliveryPlan    `json:"range,omitempty"`
	RangeRemaining int                       `json:"range_remaining,omitempty"`
	Days           []ShippingDayAvailability `json:"days"`
}
type StoreShippingChoices struct {
	StoreID   string                       `json:"store_id"`
	StoreName string                       `json:"store_name"`
	Methods   []ShippingMethodAvailability `json:"methods"`
	Reason    string                       `json:"reason,omitempty"`
}

func activeShippingReservation(r commerce.ShippingReservation, now time.Time) bool {
	return r.Status == "committed" || r.Status == "held" && now.Before(r.ExpiresAt)
}
func shippingUsage(w Workspace, storeID, methodID, day string, slot *shipping.SlotOccurrence, now time.Time, ignoreQuote, ignorePurchase string) (daily, window int) {
	if w.Commerce == nil {
		return
	}
	for _, r := range w.Commerce.ShippingReservations {
		if !activeShippingReservation(r, now) || r.StoreID != storeID || r.MethodID != methodID || r.Plan.CapacityDay != day || ignoreQuote != "" && r.QuoteID == ignoreQuote || ignorePurchase != "" && r.PurchaseID == ignorePurchase {
			continue
		}
		daily++
		// Physical overlap, not a revision/window label, keeps renamed definitions from resetting capacity.
		if slot != nil && r.Plan.Slot != nil && r.Plan.Slot.StartAt.Before(slot.EndAt) && slot.StartAt.Before(r.Plan.Slot.EndAt) {
			window++
		}
	}
	// Older confirmed range promises with a captured preparation date still consume a unit.
	for _, q := range w.Commerce.Quotes {
		if q.ID == ignoreQuote || ignorePurchase != "" && q.PurchaseID == ignorePurchase {
			continue
		}
		p := findPurchase(w.Commerce, q.PurchaseID)
		if p == nil || p.Status != "confirmed" || p.QuoteID != q.ID {
			continue
		}
		for _, ship := range q.Shipments {
			if ship.StoreID != storeID || ship.Method.ID != methodID || ship.Delivery != nil || ship.Preparation == nil || ship.Preparation.ReadyAt.In(shipping.Tehran()).Format("2006-01-02") != day {
				continue
			}
			if slices.ContainsFunc(w.Commerce.ShippingReservations, func(r commerce.ShippingReservation) bool { return r.QuoteID == q.ID && r.StoreID == storeID }) {
				continue
			}
			daily++
		}
	}
	return
}

func beforeShipping(status string) bool {
	return slices.Contains([]string{"awaiting-supply", "supplied", "packing", "ready-to-ship", "awaiting-dispatch", "on-hold"}, status)
}

type ShippingCapacitySlotView struct {
	Window    shipping.SlotOccurrence `json:"window"`
	Used      int                     `json:"used"`
	Remaining int                     `json:"remaining"`
}
type ShippingCapacityView struct {
	Preview   *ShippingMethodAvailability `json:"preview,omitempty"`
	StoreID   string                      `json:"store_id"`
	MethodID  string                      `json:"method_id"`
	Date      string                      `json:"date"`
	Active    bool                        `json:"active"`
	Open      bool                        `json:"open"`
	Limit     *int                        `json:"limit"`
	Used      int                         `json:"used"`
	Held      int                         `json:"held"`
	Committed int                         `json:"committed"`
	Remaining int                         `json:"remaining"`
	Slots     []ShippingCapacitySlotView  `json:"slots"`
}

func (s *Service) ReadShippingCapacity(ctx context.Context, a identity.User, storeID, methodID, day string) (ShippingCapacityView, error) {
	out := ShippingCapacityView{StoreID: storeID, MethodID: methodID, Date: day, Slots: []ShippingCapacitySlotView{}}
	if a.Surface != identity.SurfaceBackoffice && a.Surface != identity.SurfaceSeller {
		return out, ErrForbidden
	}
	if day == "" {
		day = s.clock.Now().In(shipping.Tehran()).Format("2006-01-02")
		out.Date = day
	}
	date, err := time.ParseInLocation("2006-01-02", day, shipping.Tehran())
	if err != nil || date.Format("2006-01-02") != day {
		return out, ErrInvalid
	}
	w, err := s.repository.Snapshot(ctx)
	if err != nil {
		return out, err
	}
	actor, err := catalogActor(w, a)
	if err != nil || !storeShippingAllowed(w, actor, storeID, StoresShippingRead) {
		return out, ErrForbidden
	}
	st, err := storeByID(&w, storeID)
	if err != nil {
		return out, err
	}
	for _, base := range preparationSettings(w).Shipping {
		if base.ID != methodID {
			continue
		}
		method, active := EffectiveStoreShipping(w, storeID, base)
		if !active {
			method = base
			b := storeShippingBinding(storeShippingPolicy(w, storeID), methodID)
			if b.Mode == "custom" && b.Live != nil {
				method = b.Live.Method
			}
		}
		out.Active = active
		out.Limit = method.DailyCapacity
		calendar, err := siteShippingCalendar(preparationSettings(w))
		if err != nil {
			return out, err
		}
		calendar, err = shipping.EffectiveCalendar(calendar, st.Preparation)
		if err != nil {
			return out, ErrInvalid
		}
		out.Open = active && calendar.IsWorking(date)
		out.Used, _ = shippingUsage(w, storeID, methodID, day, nil, s.clock.Now(), "", "")
		if w.Commerce != nil {
			for _, r := range w.Commerce.ShippingReservations {
				if r.StoreID == storeID && r.MethodID == methodID && r.Plan.CapacityDay == day && r.Status == "held" && s.clock.Now().Before(r.ExpiresAt) {
					out.Held++
				}
			}
		}
		out.Committed = out.Used - out.Held
		if out.Limit != nil {
			out.Remaining = max(0, *out.Limit-out.Used)
		}
		out.Open = out.Open && out.Remaining > 0 && day >= s.clock.Now().In(shipping.Tehran()).Format("2006-01-02")
		futureSlot := false
		if method.Schedule != nil {
			for _, slot := range method.Schedule.Occurrences(date) {
				_, used := shippingUsage(w, storeID, methodID, day, &slot, s.clock.Now(), "", "")
				out.Slots = append(out.Slots, ShippingCapacitySlotView{Window: slot, Used: used, Remaining: min(out.Remaining, max(0, slot.Capacity-used))})
				if s.clock.Now().Before(slot.StartAt) && slot.Capacity > used {
					futureSlot = true
				}
			}
		}
		if method.DeliveryMode == "slots" {
			out.Open = out.Open && futureSlot
		}
		if active {
			prep, _, _, err := shipmentPreparation(preparationSettings(w), *st, method, s.clock.Now())
			if err == nil {
				preview, e := shippingAvailability(w, storeID, method, prep, s.clock.Now(), "", "")
				if e == nil {
					out.Preview = &preview
				}
			}
		}
		return out, nil
	}
	return out, ErrNotFound
}
func planForDay(method commerce.ShippingMethod, prep shipping.PreparationSnapshot, date time.Time, slot *shipping.SlotOccurrence) (shipping.DeliveryPlan, error) {
	day := date.In(shipping.Tehran()).Format("2006-01-02")
	plan := shipping.DeliveryPlan{CapacityDay: day, DispatchDate: day, MethodVersion: method.Version, StoreRevision: method.StoreRevision, Slot: slot}
	if slot != nil {
		plan.DispatchDate = prep.ReadyAt.In(shipping.Tehran()).Format("2006-01-02")
		plan.DeliveryFrom = day
		plan.DeliveryTo = day
		return plan, nil
	}
	_, from, to, err := shipping.DeliveryDates(date, 0, shipping.CalendarDays, method.MinDays, method.MaxDays, prep.TransportBasis, prep.Calendar)
	if err != nil {
		return plan, ErrInvalid
	}
	plan.DeliveryFrom = from.Format("2006-01-02")
	plan.DeliveryTo = to.Format("2006-01-02")
	return plan, nil
}

func shippingAvailability(w Workspace, storeID string, method commerce.ShippingMethod, prep shipping.PreparationSnapshot, now time.Time, ignoreQuote, ignorePurchase string) (ShippingMethodAvailability, error) {
	out := ShippingMethodAvailability{Method: shippingMethodInput(method), Days: []ShippingDayAvailability{}}
	out.Method.StoreIDs = []string{storeID}
	if !method.Active || method.DailyCapacity == nil || *method.DailyCapacity <= 0 {
		return out, ErrShippingUnavailable
	}
	start := shipping.AtTime(prep.ReadyAt, "00:00")
	if method.DeliveryMode != "slots" {
		for offset := 0; offset < shippingSearchDays; offset++ {
			date := start.AddDate(0, 0, offset)
			if !prep.Calendar.IsWorking(date) {
				continue
			}
			used, _ := shippingUsage(w, storeID, method.ID, date.Format("2006-01-02"), nil, now, ignoreQuote, ignorePurchase)
			if used >= *method.DailyCapacity {
				continue
			}
			plan, err := planForDay(method, prep, date, nil)
			if err != nil {
				return out, err
			}
			out.Range = &plan
			out.RangeRemaining = *method.DailyCapacity - used
			return out, nil
		}
		return out, ErrShippingUnavailable
	}
	schedule, err := shipping.NormalizeSchedule(method.Schedule)
	if err != nil || schedule == nil {
		return out, ErrShippingUnavailable
	}
	possible := false
	for weekday := 0; weekday < 7; weekday++ {
		if !slices.Contains(prep.Calendar.WorkingWeekdays, weekday) {
			continue
		}
		date := start
		for int(date.Weekday()) != weekday {
			date = date.AddDate(0, 0, 1)
		}
		if len(schedule.Occurrences(date)) > 0 {
			possible = true
			break
		}
	}
	if !possible {
		return out, ErrShippingUnavailable
	}
	earliest, err := shipping.EarliestSlot(prep, method.MinDays)
	if err != nil {
		return out, ErrInvalid
	}
	if earliest.After(start) {
		start = shipping.AtTime(earliest, "00:00")
	}
	dayOptions := func(date time.Time) ShippingDayAvailability {
		day := ShippingDayAvailability{Date: date.Format("2006-01-02"), Slots: []ShippingSlotAvailability{}}
		used, _ := shippingUsage(w, storeID, method.ID, day.Date, nil, now, ignoreQuote, ignorePurchase)
		day.DailyRemaining = max(0, *method.DailyCapacity-used)
		if !prep.Calendar.IsWorking(date) || !prep.Timing.AllowSameDaySlots && day.Date == now.In(shipping.Tehran()).Format("2006-01-02") || day.DailyRemaining == 0 {
			return day
		}
		for _, occurrence := range schedule.Occurrences(date) {
			if occurrence.StartAt.Before(earliest) || !now.Before(occurrence.StartAt) {
				continue
			}
			_, used := shippingUsage(w, storeID, method.ID, day.Date, &occurrence, now, ignoreQuote, ignorePurchase)
			remaining := min(day.DailyRemaining, occurrence.Capacity-used)
			if remaining <= 0 {
				continue
			}
			copy := occurrence
			plan, e := planForDay(method, prep, date, &copy)
			if e != nil {
				continue
			}
			day.Slots = append(day.Slots, ShippingSlotAvailability{Plan: plan, Remaining: remaining})
		}
		day.Open = len(day.Slots) > 0
		return day
	}
	for offset := 0; offset < shippingSearchDays; offset++ {
		date := start.AddDate(0, 0, offset)
		first := dayOptions(date)
		if !first.Open {
			continue
		}
		out.Days = append(out.Days, first)
		for next := 1; next <= 5; next++ {
			out.Days = append(out.Days, dayOptions(date.AddDate(0, 0, next)))
		}
		return out, nil
	}
	return out, ErrShippingUnavailable
}

func selectedDelivery(availability ShippingMethodAvailability, selection shipping.DeliverySelection, now time.Time) (shipping.DeliveryPlan, error) {
	if availability.Range != nil {
		if selection.Date != "" || selection.WindowID != "" {
			return shipping.DeliveryPlan{}, ErrInvalid
		}
		return *availability.Range, nil
	}
	date, err := time.ParseInLocation("2006-01-02", selection.Date, shipping.Tehran())
	if err != nil || date.Format("2006-01-02") != selection.Date || selection.WindowID == "" {
		return shipping.DeliveryPlan{}, ErrInvalid
	}
	if selection.Date < now.In(shipping.Tehran()).Format("2006-01-02") {
		return shipping.DeliveryPlan{}, ErrShippingExpired
	}
	if availability.Method.Schedule == nil || !slices.ContainsFunc(availability.Method.Schedule.Windows, func(w shipping.TimeWindow) bool { return w.ID == selection.WindowID }) {
		return shipping.DeliveryPlan{}, ErrInvalid
	}
	for _, day := range availability.Days {
		if day.Date == selection.Date {
			for _, slot := range day.Slots {
				if slot.Plan.Slot != nil && slot.Plan.Slot.WindowID == selection.WindowID {
					return slot.Plan, nil
				}
			}
			return shipping.DeliveryPlan{}, ErrShippingFull
		}
	}
	if len(availability.Days) > 0 && selection.Date < availability.Days[0].Date {
		return shipping.DeliveryPlan{}, ErrShippingFull
	}
	return shipping.DeliveryPlan{}, ErrInvalid
}

func (s *Service) ReadShippingChoices(ctx context.Context, a identity.User, addressID, purchaseID string) ([]StoreShippingChoices, error) {
	out := []StoreShippingChoices{}
	if a.Surface != identity.SurfaceCustomer {
		return out, ErrForbidden
	}
	if err := s.ReleaseExpiredCommerce(ctx); err != nil {
		return out, err
	}
	w, err := s.repository.Snapshot(ctx)
	if err != nil {
		return out, err
	}
	if err = customerActor(w, a); err != nil {
		return out, err
	}
	c := commerceState(&w)
	var address commerce.Address
	found := false
	for _, v := range c.Addresses {
		if v.ID == addressID && v.UserID == a.ID {
			address = v
			found = true
		}
	}
	if !found {
		return out, ErrNotFound
	}
	if err = validateCheckoutAddress(w, a, address); err != nil {
		return out, err
	}
	if _, _, ok := canonicalDestination(c.Settings.Geography, address.Province, address.City); !ok {
		return out, ErrAddressUnavailable
	}
	lines := []commerce.LineSnapshot{}
	if purchaseID != "" {
		p := findPurchase(c, purchaseID)
		if p == nil || p.UserID != a.ID {
			return out, ErrNotFound
		}
		if p.Status == "confirmed" || p.Status == "cancelled" {
			return out, ErrConflict
		}
		q := findQuote(c, p.QuoteID)
		if q == nil {
			return out, ErrConflict
		}
		lines = append(lines, q.Lines...)
	} else {
		index := cartIndex(c, CartOwner{Actor: a})
		if index < 0 {
			return out, ErrNotFound
		}
		for _, line := range c.Carts[index].Lines {
			item, e := cartItem(w, line)
			if e != nil || !item.Available {
				return out, ErrStock
			}
			lines = append(lines, commerce.LineSnapshot{ProductID: item.ProductID, StoreID: item.StoreID, OfferID: item.OfferID, Quantity: item.Quantity})
		}
	}
	if len(lines) == 0 {
		return out, ErrInvalid
	}
	for _, st := range w.Stores {
		if !slices.ContainsFunc(lines, func(l commerce.LineSnapshot) bool { return l.StoreID == st.ID }) {
			continue
		}
		row := StoreShippingChoices{StoreID: st.ID, StoreName: st.Name, Methods: []ShippingMethodAvailability{}}
		if st.Active {
			for _, base := range c.Settings.Shipping {
				method, allowed := EffectiveStoreShipping(w, st.ID, base)
				if !allowed || !storeShippingEligible(w, base, method, address, st.ID, lines) {
					continue
				}
				prep, _, _, e := shipmentPreparation(c.Settings, st, method, s.clock.Now())
				if e != nil {
					continue
				}
				available, e := shippingAvailability(w, st.ID, method, prep, s.clock.Now(), "", purchaseID)
				if e == nil {
					row.Methods = append(row.Methods, available)
				}
			}
		}
		if len(row.Methods) == 0 {
			row.Reason = "برای اقلام و مقصد این مرسوله، روش مجاز دارای ظرفیت و برنامهٔ تحویل در دسترس نیست؛ آدرس را تغییر دهید یا با پشتیبانی پیگیری کنید."
		}
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StoreID < out[j].StoreID })
	return out, nil
}

func shippingProgramValid(w Workspace, ship commerce.ShipmentSnapshot) error {
	if ship.Delivery == nil {
		return ErrShippingUnavailable
	}
	current := commerce.ShippingMethod{}
	allowed := false
	for _, base := range preparationSettings(w).Shipping {
		if base.ID == ship.Method.ID {
			current, allowed = EffectiveStoreShipping(w, ship.StoreID, base)
			break
		}
	}
	if !allowed || shippingMethodInput(current).DeliveryMode != shippingMethodInput(ship.Method).DeliveryMode {
		return ErrShippingUnavailable
	}
	if ship.Delivery.Slot != nil {
		if current.Schedule == nil {
			return ErrShippingUnavailable
		}
		date, err := time.ParseInLocation("2006-01-02", ship.Delivery.CapacityDay, shipping.Tehran())
		if err != nil {
			return ErrInvalid
		}
		matches := false
		for _, day := range current.Schedule.Days {
			if day.Active && day.Weekday == int(date.Weekday()) {
				for _, assignment := range day.Slots {
					for _, window := range current.Schedule.Windows {
						if window.ID == assignment.WindowID && window.Active && shipping.AtTime(date, window.Start).Equal(ship.Delivery.Slot.StartAt) && shipping.AtTime(date, window.End).Equal(ship.Delivery.Slot.EndAt) {
							matches = true
						}
					}
				}
			}
		}
		if !matches {
			return ErrShippingUnavailable
		}
	}
	return nil
}

func holdShipping(w *Workspace, q *commerce.Quote, purchaseID string, now time.Time) error {
	c := commerceState(w)
	for _, ship := range q.Shipments {
		if err := shippingProgramValid(*w, ship); err != nil {
			return err
		}
		for _, r := range c.ShippingReservations {
			if r.QuoteID == q.ID && r.StoreID == ship.StoreID {
				return ErrConflict
			}
		}
		used, slotUsed := shippingUsage(*w, ship.StoreID, ship.Method.ID, ship.Delivery.CapacityDay, ship.Delivery.Slot, now, "", "")
		if ship.Method.DailyCapacity == nil || used >= *ship.Method.DailyCapacity || ship.Delivery.Slot != nil && slotUsed >= ship.Delivery.Slot.Capacity {
			return ErrShippingFull
		}
		c.ShippingReservations = append(c.ShippingReservations, commerce.ShippingReservation{QuoteID: q.ID, PurchaseID: purchaseID, StoreID: ship.StoreID, MethodID: ship.Method.ID, Plan: *ship.Delivery, Status: "held", CreatedAt: now, ExpiresAt: now.Add(commerce.ReservationMinutes * time.Minute)})
	}
	return nil
}
func releaseShipping(w *Workspace, quoteID string, committed bool, now time.Time) {
	if w.Commerce == nil {
		return
	}
	for i := range w.Commerce.ShippingReservations {
		r := &w.Commerce.ShippingReservations[i]
		if r.QuoteID == quoteID && (r.Status == "held" || committed && r.Status == "committed") {
			r.Status = "released"
			r.ReleasedAt = &now
		}
	}
}
func commitShipping(w *Workspace, q commerce.Quote, now time.Time) error {
	if w.Commerce == nil {
		return ErrShippingExpired
	}
	indices := []int{}
	for _, ship := range q.Shipments {
		if err := shippingProgramValid(*w, ship); err != nil {
			return err
		}
		index := slices.IndexFunc(w.Commerce.ShippingReservations, func(r commerce.ShippingReservation) bool {
			return r.QuoteID == q.ID && r.StoreID == ship.StoreID && r.MethodID == ship.Method.ID
		})
		if index < 0 || w.Commerce.ShippingReservations[index].Status != "held" || !now.Before(w.Commerce.ShippingReservations[index].ExpiresAt) {
			return ErrShippingExpired
		}
		indices = append(indices, index)
	}
	for _, index := range indices {
		w.Commerce.ShippingReservations[index].Status = "committed"
		w.Commerce.ShippingReservations[index].CommittedAt = &now
	}
	return nil
}
