package shipping

import (
	"testing"
	"time"
)

func calendarDate(value string) time.Time {
	date, err := time.ParseInLocation("2006-01-02", value, Tehran())
	if err != nil {
		panic(err)
	}
	return date
}

func TestCalendarSeparatesPreparationAndTransitUnits(t *testing.T) {
	calendar := Calendar{Version: 1, WorkingWeekdays: []int{0, 1, 2, 3, 6}, Holidays: []string{"2026-11-22"}}
	ready, from, to, err := DeliveryDates(calendarDate("2026-11-21"), 2, WorkingDays, 1, 3, CalendarDays, calendar)
	if err != nil || ready.Format("2006-01-02") != "2026-11-24" || from.Format("2006-01-02") != "2026-11-25" || to.Format("2006-01-02") != "2026-11-27" {
		t.Fatalf("unexpected two-stage promise: ready=%v from=%v to=%v error=%v", ready, from, to, err)
	}
}

func TestCalendarInheritanceCopiesAndDoesNotRewriteOverrides(t *testing.T) {
	site := Calendar{Version: 3, WorkingWeekdays: []int{0, 1, 2, 3, 6}, Holidays: []string{"2026-12-31"}}
	policy := &PreparationPolicy{DayBasis: WorkingDays, CalendarMode: Override, WorkingWeekdays: []int{5, 6}}
	resolved, err := EffectiveCalendar(site, policy)
	if err != nil || !resolved.IsWorking(calendarDate("2026-11-27")) {
		t.Fatal("store override was not resolved", err)
	}
	policy.WorkingWeekdays[0] = 1
	site.Holidays[0] = "2027-01-01"
	if resolved.WorkingWeekdays[0] != 5 || resolved.Holidays[0] != "2026-12-31" {
		t.Fatal("resolved snapshot aliases mutable settings")
	}
	inherited, err := EffectiveCalendar(site, nil)
	if err != nil || inherited.IsWorking(calendarDate("2027-01-01")) {
		t.Fatal("legacy inheritance did not use site holidays", err)
	}
}

func TestCalendarHandlesLocalMidnightYearBoundaryAndSparseWeek(t *testing.T) {
	all := Calendar{WorkingWeekdays: []int{0, 1, 2, 3, 4, 5, 6}}
	start := time.Date(2026, 12, 31, 21, 0, 0, 0, time.UTC) // January 1 in Tehran.
	got, err := AddDays(start, 1, CalendarDays, all)
	if err != nil || got.Format("2006-01-02") != "2027-01-02" {
		t.Fatal("Tehran local date was not used", got, err)
	}
	sparse := Calendar{WorkingWeekdays: []int{6}}
	got, err = AddDays(calendarDate("2026-11-21"), 365, WorkingDays, sparse)
	if err != nil || got.Sub(calendarDate("2026-11-21")) != 365*7*24*time.Hour {
		t.Fatal("long legitimate preparation was truncated", got, err)
	}
}

func TestCalendarRejectsInvalidSettingsAndKeepsZeroDelayExplicit(t *testing.T) {
	for _, days := range [][]int{nil, {}, {1, 1}, {-1}, {7}} {
		if _, err := NormalizeWeekdays(days); err == nil {
			t.Fatalf("accepted invalid weekdays: %v", days)
		}
	}
	for _, policy := range []PreparationPolicy{{DayBasis: WorkingDays, CalendarMode: Override}, {DayBasis: WorkingDays, CalendarMode: Inherit, WorkingWeekdays: []int{1}}, {DayBasis: "invalid", CalendarMode: Inherit}} {
		if _, err := NormalizePreparation(policy); err == nil {
			t.Fatalf("accepted invalid policy: %+v", policy)
		}
	}
	calendar := Calendar{WorkingWeekdays: []int{6}}
	got, err := AddDays(calendarDate("2026-11-20"), 0, WorkingDays, calendar)
	if err != nil || got.Format("2006-01-02") != "2026-11-21" {
		t.Fatal("zero working-day delay must resolve an open date", got, err)
	}
	got, err = AddDays(calendarDate("2026-11-20"), 0, CalendarDays, calendar)
	if err != nil || got.Format("2006-01-02") != "2026-11-20" {
		t.Fatal("calendar-day arithmetic unexpectedly moved a closed date", got, err)
	}
}
