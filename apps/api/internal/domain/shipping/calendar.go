package shipping

import (
	"errors"
	"slices"
	"time"
	_ "time/tzdata"
)

type DayBasis string

const (
	WorkingDays  DayBasis = "working"
	CalendarDays DayBasis = "calendar"
	Inherit               = "inherit"
	Override              = "override"
)

var ErrInvalidCalendar = errors.New("invalid shipping calendar")

type PreparationPolicy struct {
	Version         int      `json:"version"`
	DayBasis        DayBasis `json:"day_basis"`
	CalendarMode    string   `json:"calendar_mode"`
	WorkingWeekdays []int    `json:"working_weekdays"`
}

type Calendar struct {
	Version         int      `json:"version"`
	WorkingWeekdays []int    `json:"working_weekdays"`
	Holidays        []string `json:"holidays"`
}

var tehranLocation = func() *time.Location {
	location, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		panic(err)
	}
	return location
}()

func Tehran() *time.Location { return tehranLocation }

func ValidBasis(basis DayBasis) bool { return basis == WorkingDays || basis == CalendarDays }

func NormalizeWeekdays(days []int) ([]int, error) {
	if len(days) == 0 || len(days) > 7 {
		return nil, ErrInvalidCalendar
	}
	normalized := slices.Clone(days)
	slices.Sort(normalized)
	for index, day := range normalized {
		if day < 0 || day > 6 || index > 0 && normalized[index-1] == day {
			return nil, ErrInvalidCalendar
		}
	}
	return normalized, nil
}

func NormalizeCalendar(calendar Calendar) (Calendar, error) {
	days, err := NormalizeWeekdays(calendar.WorkingWeekdays)
	if err != nil || len(calendar.Holidays) > 366 {
		return Calendar{}, ErrInvalidCalendar
	}
	calendar.WorkingWeekdays = days
	calendar.Holidays = append([]string{}, calendar.Holidays...)
	slices.Sort(calendar.Holidays)
	for index, holiday := range calendar.Holidays {
		date, err := time.Parse("2006-01-02", holiday)
		if err != nil || date.Format("2006-01-02") != holiday || index > 0 && calendar.Holidays[index-1] == holiday {
			return Calendar{}, ErrInvalidCalendar
		}
	}
	return calendar, nil
}

func NormalizePreparation(policy PreparationPolicy) (PreparationPolicy, error) {
	if !ValidBasis(policy.DayBasis) || policy.CalendarMode != Inherit && policy.CalendarMode != Override {
		return PreparationPolicy{}, ErrInvalidCalendar
	}
	if policy.CalendarMode == Inherit {
		if len(policy.WorkingWeekdays) != 0 {
			return PreparationPolicy{}, ErrInvalidCalendar
		}
		policy.WorkingWeekdays = []int{}
		return policy, nil
	}
	days, err := NormalizeWeekdays(policy.WorkingWeekdays)
	if err != nil {
		return PreparationPolicy{}, err
	}
	policy.WorkingWeekdays = days
	return policy, nil
}

func EffectiveCalendar(site Calendar, policy *PreparationPolicy) (Calendar, error) {
	calendar, err := NormalizeCalendar(site)
	if err != nil {
		return Calendar{}, err
	}
	if policy == nil {
		return calendar, nil
	}
	normalized, err := NormalizePreparation(*policy)
	if err != nil {
		return Calendar{}, err
	}
	if normalized.CalendarMode == Override {
		calendar.WorkingWeekdays = slices.Clone(normalized.WorkingWeekdays)
	}
	return calendar, nil
}

func (calendar Calendar) IsWorking(date time.Time) bool {
	date = date.In(Tehran())
	return slices.Contains(calendar.WorkingWeekdays, int(date.Weekday())) && !slices.Contains(calendar.Holidays, date.Format("2006-01-02"))
}

// AddDays counts local dates, excluding the starting date for a positive delay.
// It deliberately does not decide readiness hours, cutoffs or slot eligibility.
func AddDays(start time.Time, days int, basis DayBasis, calendar Calendar) (time.Time, error) {
	if !ValidBasis(basis) || days < 0 || days > 365 {
		return time.Time{}, ErrInvalidCalendar
	}
	calendar, err := NormalizeCalendar(calendar)
	if err != nil {
		return time.Time{}, err
	}
	local := start.In(Tehran())
	date := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, Tehran())
	if basis == CalendarDays {
		return date.AddDate(0, 0, days), nil
	}
	remaining := days
	limit := (days + len(calendar.Holidays) + 1) * 7
	for searched := 0; searched <= limit; searched++ {
		if searched > 0 && remaining > 0 && calendar.IsWorking(date) {
			remaining--
		}
		if remaining == 0 && calendar.IsWorking(date) {
			return date, nil
		}
		date = date.AddDate(0, 0, 1)
	}
	return time.Time{}, ErrInvalidCalendar
}

func DeliveryDates(start time.Time, leadDays int, preparationBasis DayBasis, minDays, maxDays int, deliveryBasis DayBasis, calendar Calendar) (ready, from, to time.Time, err error) {
	if minDays < 0 || maxDays < minDays || maxDays > 60 {
		err = ErrInvalidCalendar
		return
	}
	ready, err = AddDays(start, leadDays, preparationBasis, calendar)
	if err != nil {
		return
	}
	from, err = AddDays(ready, minDays, deliveryBasis, calendar)
	if err != nil {
		return
	}
	to, err = AddDays(ready, maxDays, deliveryBasis, calendar)
	return
}
