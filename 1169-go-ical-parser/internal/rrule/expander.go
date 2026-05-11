package rrule

import (
	"time"

	"github.com/example/icaltool/internal/ical"
)

type Expander struct{}

func NewExpander() *Expander {
	return &Expander{}
}

type ExpandOptions struct {
	Start time.Time
	End   time.Time
}

func (e *Expander) ExpandEvent(event *ical.Event, opts ExpandOptions) []time.Time {
	if event.RRULE == nil {
		return []time.Time{event.DTStart}
	}

	loc := event.DTStart.Location()
	var results []time.Time
	generated := 0

	current := event.DTStart

	maxIterations := 10000
	iterations := 0

	for iterations < maxIterations {
		iterations++

		if !event.RRULE.UNTIL.IsZero() {
			untilUTC := convertToSameZone(event.RRULE.UNTIL, current)
			if current.After(untilUTC) {
				break
			}
		}

		if !current.Before(opts.Start) && !current.After(opts.End) {
			if !isExcluded(current, event.EXDATEs, loc) {
				results = append(results, current)
			}
		}

		if event.RRULE.COUNT > 0 {
			generated++
			if generated >= event.RRULE.COUNT {
				break
			}
		}

		next := advanceTime(current, event.RRULE, loc)
		if next.Before(current) || next.Equal(current) {
			break
		}

		if !opts.End.IsZero() && next.After(opts.End) {
			break
		}

		current = next
	}

	results = addRDates(results, event.RDATEs, opts)

	return results
}

func (e *Expander) ExpandRRULE(rrule *ical.RRULE, dtStart time.Time, opts ExpandOptions) []time.Time {
	event := &ical.Event{
		DTStart: dtStart,
		RRULE:   rrule,
	}
	return e.ExpandEvent(event, opts)
}

func convertToSameZone(t, reference time.Time) time.Time {
	if t.Location() == reference.Location() {
		return t
	}
	if t.Location() == time.UTC {
		return t.In(reference.Location())
	}
	if reference.Location() == time.UTC {
		return t.UTC()
	}
	return t.In(reference.Location())
}

func isExcluded(t time.Time, exdates []time.Time, loc *time.Location) bool {
	for _, exd := range exdates {
		exdLocal := exd
		if exd.Location() != loc {
			exdLocal = exd.In(loc)
		}

		if t.Year() == exdLocal.Year() &&
			t.Month() == exdLocal.Month() &&
			t.Day() == exdLocal.Day() &&
			t.Hour() == exdLocal.Hour() &&
			t.Minute() == exdLocal.Minute() &&
			t.Second() == exdLocal.Second() {
			return true
		}
	}
	return false
}

func addRDates(results []time.Time, rdates []time.Time, opts ExpandOptions) []time.Time {
	for _, rdate := range rdates {
		if !rdate.Before(opts.Start) && !rdate.After(opts.End) {
			results = append(results, rdate)
		}
	}
	return results
}

func advanceTime(current time.Time, r *ical.RRULE, loc *time.Location) time.Time {
	switch r.FREQ {
	case ical.FreqDAILY:
		return advanceDaily(current, r, loc)
	case ical.FreqWEEKLY:
		return advanceWeekly(current, r, loc)
	case ical.FreqMONTHLY:
		return advanceMonthly(current, r, loc)
	case ical.FreqYEARLY:
		return advanceYearly(current, r, loc)
	default:
		return current
	}
}

func advanceDaily(current time.Time, r *ical.RRULE, loc *time.Location) time.Time {
	interval := r.INTERVAL
	if interval == 0 {
		interval = 1
	}

	next := current.AddDate(0, 0, interval)
	next = applyTimezone(next, loc)
	return next
}

func advanceWeekly(current time.Time, r *ical.RRULE, loc *time.Location) time.Time {
	interval := r.INTERVAL
	if interval == 0 {
		interval = 1
	}

	if len(r.BYDAY) > 0 {
		next := findNextWeekday(current, r.BYDAY, interval)
		next = applyTimezone(next, loc)
		return next
	}

	next := current.AddDate(0, 0, 7*interval)
	next = applyTimezone(next, loc)
	return next
}

func advanceMonthly(current time.Time, r *ical.RRULE, loc *time.Location) time.Time {
	interval := r.INTERVAL
	if interval == 0 {
		interval = 1
	}

	if len(r.BYDAY) > 0 {
		next := findNextMonthlyByDay(current, r.BYDAY, r.BYSETPOS, interval)
		next = applyTimezone(next, loc)
		return next
	}

	if len(r.BYMONTHDAY) > 0 {
		next := findNextMonthlyByMonthday(current, r.BYMONTHDAY, interval)
		next = applyTimezone(next, loc)
		return next
	}

	next := current.AddDate(0, interval, 0)
	next = applyTimezone(next, loc)
	return next
}

func advanceYearly(current time.Time, r *ical.RRULE, loc *time.Location) time.Time {
	interval := r.INTERVAL
	if interval == 0 {
		interval = 1
	}

	if len(r.BYMONTH) > 0 && len(r.BYDAY) > 0 {
		next := findNextYearlyByMonthAndDay(current, r.BYMONTH, r.BYDAY, r.BYSETPOS, interval)
		next = applyTimezone(next, loc)
		return next
	}

	if len(r.BYMONTH) > 0 {
		next := findNextYearlyByMonth(current, r.BYMONTH, r.BYMONTHDAY, interval)
		next = applyTimezone(next, loc)
		return next
	}

	next := current.AddDate(interval, 0, 0)
	next = applyTimezone(next, loc)
	return next
}

func findNextWeekday(current time.Time, days []ical.WeekdayPos, interval int) time.Time {
	today := current.Weekday()
	todayNum := int(today)
	baseDate := current

	minOffset := 7 * interval
	found := false

	for _, d := range days {
		if d.Pos != 0 {
			continue
		}

		dayNum := int(d.Day)
		diff := dayNum - todayNum

		var offset int
		if diff > 0 {
			offset = diff + 7*(interval-1)
		} else {
			offset = diff + 7*interval
		}

		if offset > 0 && offset < minOffset {
			minOffset = offset
			found = true
		}
	}

	if found {
		return baseDate.AddDate(0, 0, minOffset)
	}

	return baseDate.AddDate(0, 0, 7*interval)
}

func findNextMonthlyByDay(current time.Time, days []ical.WeekdayPos, setpos []int, interval int) time.Time {
	year, month, _ := current.Date()

	for i := 0; i < 12; i++ {
		nextMonth := time.Date(year, month+time.Month(i*interval+1), 1, 0, 0, 0, 0, current.Location())

		candidates := findWeekdaysInMonth(nextMonth, days, setpos)

		for _, cand := range candidates {
			candTime := time.Date(cand.Year(), cand.Month(), cand.Day(),
				current.Hour(), current.Minute(), current.Second(),
				current.Nanosecond(), current.Location())
			if candTime.After(current) {
				return candTime
			}
		}
	}

	return current.AddDate(0, interval, 0)
}

func findWeekdaysInMonth(startOfMonth time.Time, days []ical.WeekdayPos, setpos []int) []time.Time {
	var result []time.Time
	firstDay := startOfMonth
	lastDay := firstDay.AddDate(0, 1, -1)

	for _, dayPos := range days {
		if dayPos.Pos != 0 {
			found := findNthWeekdayInMonth(startOfMonth, dayPos.Day, dayPos.Pos)
			if !found.IsZero() {
				result = append(result, found)
			}
		} else {
			for d := firstDay; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
				if d.Weekday() == dayPos.Day {
					result = append(result, d)
				}
			}
		}
	}

	if len(setpos) > 0 {
		filtered := make([]time.Time, 0)
		for _, pos := range setpos {
			idx := pos
			if pos < 0 {
				idx = len(result) + pos
			} else {
				idx = pos - 1
			}
			if idx >= 0 && idx < len(result) {
				filtered = append(filtered, result[idx])
			}
		}
		return filtered
	}

	return result
}

func findNthWeekdayInMonth(startOfMonth time.Time, weekday time.Weekday, n int) time.Time {
	lastDay := startOfMonth.AddDate(0, 1, -1)

	if n > 0 {
		count := 0
		for d := startOfMonth; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
			if d.Weekday() == weekday {
				count++
				if count == n {
					return d
				}
			}
		}
	} else if n < 0 {
		count := 0
		for d := lastDay; !d.Before(startOfMonth); d = d.AddDate(0, 0, -1) {
			if d.Weekday() == weekday {
				count++
				if count == -n {
					return d
				}
			}
		}
	}

	return time.Time{}
}

func findNextMonthlyByMonthday(current time.Time, monthdays []int, interval int) time.Time {
	year, month, day := current.Date()

	for i := 0; i < 12; i++ {
		baseMonth := month
		if i > 0 {
			baseMonth += time.Month(interval)
		}

		firstDayOfMonth := time.Date(year, baseMonth, 1, 0, 0, 0, 0, current.Location())
		lastDay := firstDayOfMonth.AddDate(0, 1, -1)

		for _, md := range monthdays {
			var dayNum int
			if md > 0 {
				dayNum = md
			} else if md < 0 {
				dayNum = lastDay.Day() + md + 1
			}

			if dayNum > 0 && dayNum <= lastDay.Day() {
				cand := time.Date(year, baseMonth, dayNum,
					current.Hour(), current.Minute(), current.Second(),
					current.Nanosecond(), current.Location())

				if i == 0 {
					if dayNum > day && cand.After(current) {
						return cand
					}
				} else {
					return cand
				}
			}
		}
	}

	return current.AddDate(0, interval, 0)
}

func findNextYearlyByMonthAndDay(current time.Time, months []int, days []ical.WeekdayPos, setpos []int, interval int) time.Time {
	year := current.Year()

	for i := 0; i < interval+1; i++ {
		checkYear := year + i

		for _, m := range months {
			if m < 1 || m > 12 {
				continue
			}

			if i == 0 && time.Month(m) < current.Month() {
				continue
			}

			startOfMonth := time.Date(checkYear, time.Month(m), 1, 0, 0, 0, 0, current.Location())

			candidates := findWeekdaysInMonth(startOfMonth, days, setpos)

			for _, cand := range candidates {
				candTime := time.Date(cand.Year(), cand.Month(), cand.Day(),
					current.Hour(), current.Minute(), current.Second(),
					current.Nanosecond(), current.Location())
				if candTime.After(current) {
					return candTime
				}
			}
		}
	}

	return current.AddDate(interval, 0, 0)
}

func findNextYearlyByMonth(current time.Time, months []int, monthdays []int, interval int) time.Time {
	year := current.Year()

	for i := 0; i < interval+1; i++ {
		checkYear := year + i

		for _, m := range months {
			if m < 1 || m > 12 {
				continue
			}

			if i == 0 && time.Month(m) < current.Month() {
				continue
			}

			startOfMonth := time.Date(checkYear, time.Month(m), 1, 0, 0, 0, 0, current.Location())
			lastDay := startOfMonth.AddDate(0, 1, -1)

			for _, md := range monthdays {
				var dayNum int
				if md > 0 {
					dayNum = md
				} else if md < 0 {
					dayNum = lastDay.Day() + md + 1
				}

				if dayNum > 0 && dayNum <= lastDay.Day() {
					cand := time.Date(checkYear, time.Month(m), dayNum,
						current.Hour(), current.Minute(), current.Second(),
						current.Nanosecond(), current.Location())
					if cand.After(current) {
						return cand
					}
				}
			}
		}
	}

	return current.AddDate(interval, 0, 0)
}

func applyTimezone(t time.Time, loc *time.Location) time.Time {
	if loc == nil {
		return t
	}
	return t.In(loc)
}
