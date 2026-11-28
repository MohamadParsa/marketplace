package shipping

import (
	"testing"
	"time"
)

func TestPreparationTimingCutoffHolidaysAndZeroDelay(t *testing.T) {
	calendar := Calendar{WorkingWeekdays: []int{0, 1, 2, 3, 6}, Holidays: []string{"2026-11-22"}}
	for _, test := range []struct {
		name, start string
		days        int
		basis       DayBasis
		want        string
	}{
		{"two working days", "2026-11-21T08:00:00+03:30", 2, WorkingDays, "2026-11-24T09:00:00+03:30"},
		{"calendar ends on holiday", "2026-11-21T08:00:00+03:30", 1, CalendarDays, "2026-11-23T09:00:00+03:30"},
		{"zero before opening", "2026-11-21T08:00:00+03:30", 0, WorkingDays, "2026-11-21T09:00:00+03:30"},
		{"zero after opening", "2026-11-21T10:00:00+03:30", 0, WorkingDays, "2026-11-21T10:00:00+03:30"},
		{"at cutoff", "2026-11-21T12:00:00+03:30", 0, WorkingDays, "2026-11-23T09:00:00+03:30"},
		{"positive after cutoff", "2026-11-21T12:00:00+03:30", 1, WorkingDays, "2026-11-24T09:00:00+03:30"},
		{"closed origin", "2026-11-20T08:00:00+03:30", 0, WorkingDays, "2026-11-21T09:00:00+03:30"},
		{"year boundary", "2026-12-30T08:00:00+03:30", 1, CalendarDays, "2027-01-02T09:00:00+03:30"},
	} {
		t.Run(test.name, func(t *testing.T) {
			start, _ := time.Parse(time.RFC3339, test.start)
			snapshot, err := ResolvePreparation(start, test.days, PreparationPolicy{DayBasis: test.basis, CalendarMode: Inherit}, calendar, DefaultTiming())
			if err != nil || snapshot.ReadyAt.Format(time.RFC3339) != test.want {
				t.Fatal(snapshot.ReadyAt, err, "want", test.want)
			}
		})
	}
}

func TestPreparationTimingValidationAndSnapshotRebase(t *testing.T) {
	for _, policy := range []TimingPolicy{{ReadyTime: "9:00", CutoffTime: "12:00"}, {ReadyTime: "13:00", CutoffTime: "12:00"}, {ReadyTime: "09:00", CutoffTime: "24:00"}, {ReadyTime: "09:00", CutoffTime: "12:00", MinSlotLeadMinutes: -1}, {ReadyTime: "09:00", CutoffTime: "12:00", MinSlotLeadMinutes: 1441}} {
		if _, err := NormalizeTiming(policy); err == nil {
			t.Fatal("accepted invalid timing", policy)
		}
	}
	start, _ := time.Parse(time.RFC3339, "2026-11-21T08:00:00+03:30")
	calendar := Calendar{Version: 4, WorkingWeekdays: []int{0, 1, 2, 3, 6}, Holidays: []string{}}
	snapshot, err := ResolvePreparation(start, 2, PreparationPolicy{Version: 2, DayBasis: WorkingDays, CalendarMode: Inherit}, calendar, DefaultTiming())
	if err != nil {
		t.Fatal(err)
	}
	snapshot.TransportBasis = CalendarDays
	calendar.WorkingWeekdays[0] = 5
	rebased, err := snapshot.Rebase(start.AddDate(0, 0, 1))
	if err != nil || rebased.ReadyAt.Format("2006-01-02") != "2026-11-24" || rebased.Calendar.Version != 4 || rebased.Calendar.WorkingWeekdays[0] != 0 || rebased.TransportBasis != CalendarDays {
		t.Fatal("snapshot/rebase", rebased, err)
	}
}
