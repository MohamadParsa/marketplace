package shipping

import "time"

type TimingPolicy struct {
	ReadyTime          string `json:"ready_time"`
	CutoffTime         string `json:"cutoff_time"`
	MinSlotLeadMinutes int    `json:"min_slot_lead_minutes"`
	AllowSameDaySlots  bool   `json:"allow_same_day_slots"`
}

func DefaultTiming() TimingPolicy {
	return TimingPolicy{ReadyTime: "09:00", CutoffTime: "12:00", MinSlotLeadMinutes: 120, AllowSameDaySlots: true}
}

func NormalizeTiming(policy TimingPolicy) (TimingPolicy, error) {
	for _, value := range []string{policy.ReadyTime, policy.CutoffTime} {
		clock, err := time.Parse("15:04", value)
		if err != nil || clock.Format("15:04") != value {
			return policy, ErrInvalidCalendar
		}
	}
	if policy.ReadyTime > policy.CutoffTime || policy.MinSlotLeadMinutes < 0 || policy.MinSlotLeadMinutes > 1440 {
		return policy, ErrInvalidCalendar
	}
	return policy, nil
}

type PreparationSnapshot struct {
	Policy         PreparationPolicy `json:"policy"`
	Calendar       Calendar          `json:"calendar"`
	Timing         TimingPolicy      `json:"timing"`
	LeadDays       int               `json:"lead_days"`
	SourceAt       time.Time         `json:"source_at"`
	ReadyAt        time.Time         `json:"ready_at"`
	TransportBasis DayBasis          `json:"transport_basis"`
}

func AtTime(date time.Time, clock string) time.Time {
	value, _ := time.Parse("15:04", clock)
	local := date.In(Tehran())
	return time.Date(local.Year(), local.Month(), local.Day(), value.Hour(), value.Minute(), 0, 0, Tehran())
}

func ResolvePreparation(start time.Time, leadDays int, policy PreparationPolicy, calendar Calendar, timing TimingPolicy) (PreparationSnapshot, error) {
	var out PreparationSnapshot
	policy, err := NormalizePreparation(policy)
	if err != nil {
		return out, err
	}
	timing, err = NormalizeTiming(timing)
	if err != nil {
		return out, err
	}
	calendar, err = NormalizeCalendar(calendar)
	if err != nil {
		return out, err
	}
	base := start.In(Tehran())
	if !base.Before(AtTime(base, timing.CutoffTime)) {
		base = AtTime(base.AddDate(0, 0, 1), "00:00")
	}
	base, err = AddDays(base, 0, WorkingDays, calendar)
	if err != nil {
		return out, err
	}
	ready, err := AddDays(base, leadDays, policy.DayBasis, calendar)
	if err != nil {
		return out, err
	}
	// A calendar-day delay still needs an open store to dispatch the shipment.
	ready, err = AddDays(ready, 0, WorkingDays, calendar)
	if err != nil {
		return out, err
	}
	ready = AtTime(ready, timing.ReadyTime)
	if ready.Before(start) {
		ready = start.In(Tehran())
	}
	out = PreparationSnapshot{Policy: policy, Calendar: calendar, Timing: timing, LeadDays: leadDays, SourceAt: start, ReadyAt: ready}
	return out, nil
}

// Rebase uses the captured calendar and rules, never current store settings.
func (snapshot PreparationSnapshot) Rebase(start time.Time) (PreparationSnapshot, error) {
	out, err := ResolvePreparation(start, snapshot.LeadDays, snapshot.Policy, snapshot.Calendar, snapshot.Timing)
	out.TransportBasis = snapshot.TransportBasis
	return out, err
}
