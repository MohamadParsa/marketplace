package shipping_test

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
	"testing"
	"time"
)

func slotCapacity(value int) *int { return &value }
func validSchedule() *shipping.Schedule {
	return &shipping.Schedule{Windows: []shipping.TimeWindow{{ID: "am", Start: "10:30", End: "12:15", Active: true}, {ID: "pm", Start: "12:15", End: "16:40", Active: true}}, Days: []shipping.WeeklyDay{{Weekday: 6, Active: true, Slots: []shipping.WindowAssignment{{WindowID: "am", Capacity: slotCapacity(3)}, {WindowID: "pm", Capacity: slotCapacity(7)}}}, {Weekday: 0, Active: true, Slots: []shipping.WindowAssignment{{WindowID: "am", Capacity: slotCapacity(1)}}}}}
}
func TestScheduleRejectsInvalidDefinitionsAndAssignments(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*shipping.Schedule)
	}{
		{"reverse", func(s *shipping.Schedule) { s.Windows[0].Start = "13:00" }},
		{"format", func(s *shipping.Schedule) { s.Windows[0].Start = "9:00" }},
		{"duplicate-id", func(s *shipping.Schedule) { s.Windows[1].ID = "am" }},
		{"overlap", func(s *shipping.Schedule) { s.Windows[1].Start = "12:14" }},
		{"weekday", func(s *shipping.Schedule) { s.Days[0].Weekday = 7 }},
		{"duplicate-day", func(s *shipping.Schedule) { s.Days[1].Weekday = 6 }},
		{"duplicate-assignment", func(s *shipping.Schedule) { s.Days[0].Slots = append(s.Days[0].Slots, s.Days[0].Slots[0]) }},
		{"unknown-window", func(s *shipping.Schedule) { s.Days[0].Slots[0].WindowID = "missing" }},
		{"missing-capacity", func(s *shipping.Schedule) { s.Days[0].Slots[0].Capacity = nil }},
		{"negative-capacity", func(s *shipping.Schedule) { s.Days[0].Slots[0].Capacity = slotCapacity(-1) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			schedule := validSchedule()
			test.change(schedule)
			if _, err := shipping.NormalizeSchedule(schedule); err == nil {
				t.Fatal("accepted invalid schedule")
			}
		})
	}
	schedule, err := shipping.NormalizeSchedule(validSchedule())
	if err != nil {
		t.Fatal(err)
	}
	sat, _ := time.ParseInLocation("2006-01-02", "2026-11-21", shipping.Tehran())
	if len(schedule.Occurrences(sat)) != 2 || len(schedule.Occurrences(sat.AddDate(0, 0, 1))) != 1 || len(schedule.Occurrences(sat.AddDate(0, 0, 2))) != 0 {
		t.Fatal("weekly reuse")
	}
	schedule.Windows[0].Active = false
	if len(schedule.Occurrences(sat)) != 1 {
		t.Fatal("inactive definition")
	}
}
func TestEarliestSlotRetainsClockAndMinimumTransport(t *testing.T) {
	ready, _ := time.Parse(time.RFC3339Nano, "2026-11-21T11:45:30.123456+03:30")
	snapshot := shipping.PreparationSnapshot{ReadyAt: ready, TransportBasis: shipping.WorkingDays, Calendar: shipping.Calendar{WorkingWeekdays: []int{0, 1, 2, 3, 6}, Holidays: []string{"2026-11-22"}}, Timing: shipping.TimingPolicy{MinSlotLeadMinutes: 120}}
	earliest, err := shipping.EarliestSlot(snapshot, 1)
	if err != nil || earliest.Format(time.RFC3339Nano) != "2026-11-23T13:45:30.123456+03:30" {
		t.Fatal(earliest, err)
	}
}
