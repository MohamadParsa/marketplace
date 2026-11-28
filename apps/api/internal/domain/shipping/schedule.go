package shipping

import (
	"errors"
	"slices"
	"sort"
	"strings"
	"time"
)

var ErrInvalidSchedule = errors.New("invalid shipping schedule")

type TimeWindow struct {
	ID     string `json:"id"`
	Start  string `json:"start"`
	End    string `json:"end"`
	Active bool   `json:"active"`
}
type WindowAssignment struct {
	WindowID string `json:"window_id"`
	Capacity *int   `json:"capacity"`
}
type WeeklyDay struct {
	Weekday int                `json:"weekday"`
	Active  bool               `json:"active"`
	Slots   []WindowAssignment `json:"slots"`
}
type Schedule struct {
	Windows []TimeWindow `json:"windows"`
	Days    []WeeklyDay  `json:"days"`
}
type DeliverySelection struct {
	Date     string `json:"date"`
	WindowID string `json:"window_id"`
}
type SlotOccurrence struct {
	WindowID string    `json:"window_id"`
	StartAt  time.Time `json:"start_at"`
	EndAt    time.Time `json:"end_at"`
	Capacity int       `json:"capacity"`
}
type DeliveryPlan struct {
	CapacityDay   string          `json:"capacity_day"`
	DispatchDate  string          `json:"dispatch_date"`
	DeliveryFrom  string          `json:"delivery_from"`
	DeliveryTo    string          `json:"delivery_to"`
	MethodVersion int             `json:"method_version"`
	StoreRevision int             `json:"store_revision"`
	Slot          *SlotOccurrence `json:"slot,omitempty"`
}

func NormalizeSchedule(in *Schedule) (*Schedule, error) {
	if in == nil {
		return nil, nil
	}
	if len(in.Windows) > 200 || len(in.Days) > 7 {
		return nil, ErrInvalidSchedule
	}
	out := &Schedule{Windows: append([]TimeWindow{}, in.Windows...), Days: []WeeklyDay{}}
	seen := map[string]bool{}
	for _, window := range out.Windows {
		if strings.TrimSpace(window.ID) == "" || len(window.ID) > 80 || strings.ContainsRune(window.ID, '\x00') || seen[window.ID] {
			return nil, ErrInvalidSchedule
		}
		seen[window.ID] = true
		for _, clock := range []string{window.Start, window.End} {
			parsed, err := time.Parse("15:04", clock)
			if err != nil || parsed.Format("15:04") != clock {
				return nil, ErrInvalidSchedule
			}
		}
		if window.Start >= window.End {
			return nil, ErrInvalidSchedule
		}
	}
	sort.Slice(out.Windows, func(i, j int) bool {
		if out.Windows[i].Start == out.Windows[j].Start {
			return out.Windows[i].ID < out.Windows[j].ID
		}
		return out.Windows[i].Start < out.Windows[j].Start
	})
	weekdays := map[int]bool{}
	for _, day := range in.Days {
		if day.Weekday < 0 || day.Weekday > 6 || weekdays[day.Weekday] || len(day.Slots) > 200 {
			return nil, ErrInvalidSchedule
		}
		weekdays[day.Weekday] = true
		next := WeeklyDay{Weekday: day.Weekday, Active: day.Active, Slots: []WindowAssignment{}}
		assigned := map[string]bool{}
		activeWindows := []TimeWindow{}
		for _, slot := range day.Slots {
			index := slices.IndexFunc(out.Windows, func(w TimeWindow) bool { return w.ID == slot.WindowID })
			if index < 0 || assigned[slot.WindowID] || slot.Capacity == nil || *slot.Capacity < 0 || *slot.Capacity > 1000000 {
				return nil, ErrInvalidSchedule
			}
			assigned[slot.WindowID] = true
			capacity := *slot.Capacity
			next.Slots = append(next.Slots, WindowAssignment{WindowID: slot.WindowID, Capacity: &capacity})
			if day.Active && out.Windows[index].Active {
				activeWindows = append(activeWindows, out.Windows[index])
			}
		}
		sort.Slice(activeWindows, func(i, j int) bool { return activeWindows[i].Start < activeWindows[j].Start })
		for i := 1; i < len(activeWindows); i++ {
			if activeWindows[i].Start < activeWindows[i-1].End {
				return nil, ErrInvalidSchedule
			}
		}
		sort.Slice(next.Slots, func(i, j int) bool { return next.Slots[i].WindowID < next.Slots[j].WindowID })
		out.Days = append(out.Days, next)
	}
	sort.Slice(out.Days, func(i, j int) bool { return out.Days[i].Weekday < out.Days[j].Weekday })
	return out, nil
}

func (s Schedule) Occurrences(date time.Time) []SlotOccurrence {
	out := []SlotOccurrence{}
	windows := map[string]TimeWindow{}
	for _, window := range s.Windows {
		windows[window.ID] = window
	}
	for _, day := range s.Days {
		if !day.Active || day.Weekday != int(date.In(Tehran()).Weekday()) {
			continue
		}
		for _, slot := range day.Slots {
			window, exists := windows[slot.WindowID]
			if exists && window.Active && slot.Capacity != nil && *slot.Capacity > 0 {
				out = append(out, SlotOccurrence{WindowID: window.ID, StartAt: AtTime(date, window.Start), EndAt: AtTime(date, window.End), Capacity: *slot.Capacity})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartAt.Before(out[j].StartAt) })
	return out
}

// Transport starts after preparation; adding calendar dates must retain its clock.
func EarliestSlot(snapshot PreparationSnapshot, minDays int) (time.Time, error) {
	ready := snapshot.ReadyAt
	if minDays > 0 {
		date, err := AddDays(ready, minDays, snapshot.TransportBasis, snapshot.Calendar)
		if err != nil {
			return time.Time{}, err
		}
		clock := ready.In(Tehran())
		ready = time.Date(date.Year(), date.Month(), date.Day(), clock.Hour(), clock.Minute(), clock.Second(), clock.Nanosecond(), Tehran())
	}
	return ready.Add(time.Duration(snapshot.Timing.MinSlotLeadMinutes) * time.Minute), nil
}
