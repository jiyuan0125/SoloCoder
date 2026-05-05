package cron

import (
	"time"
)

const (
	maxSearchYears = 5
	maxIterations  = 1000000
)

func (c *CronExpr) Next(after time.Time) (time.Time, bool) {
	t := after.Add(time.Second).Truncate(time.Second)
	startYear := t.Year()
	iterations := 0

	for {
		if iterations > maxIterations {
			return time.Time{}, false
		}
		iterations++

		if t.Year()-startYear > maxSearchYears {
			return time.Time{}, false
		}

		matched, skipTo := c.tryMatch(t)
		if matched {
			return t, true
		}

		if !skipTo.IsZero() {
			t = skipTo
		} else {
			t = t.Add(time.Second)
		}
	}
}

func (c *CronExpr) tryMatch(t time.Time) (bool, time.Time) {
	second := t.Second()
	minute := t.Minute()
	hour := t.Hour()
	day := t.Day()
	month := int(t.Month())
	year := t.Year()

	if !contains(c.Second.Range, second) {
		nextSec := nextValue(c.Second.Range, second)
		if nextSec > second {
			return false, time.Date(year, time.Month(month), day, hour, minute, nextSec, 0, t.Location())
		}
		return false, time.Date(year, time.Month(month), day, hour, minute+1, c.Second.Range[0], 0, t.Location())
	}

	if !contains(c.Minute.Range, minute) {
		nextMin := nextValue(c.Minute.Range, minute)
		if nextMin > minute {
			return false, time.Date(year, time.Month(month), day, hour, nextMin, c.Second.Range[0], 0, t.Location())
		}
		return false, time.Date(year, time.Month(month), day, hour+1, c.Minute.Range[0], c.Second.Range[0], 0, t.Location())
	}

	if !contains(c.Hour.Range, hour) {
		nextHour := nextValue(c.Hour.Range, hour)
		if nextHour > hour {
			return false, time.Date(year, time.Month(month), day, nextHour, c.Minute.Range[0], c.Second.Range[0], 0, t.Location())
		}
		return false, time.Date(year, time.Month(month), day+1, c.Hour.Range[0], c.Minute.Range[0], c.Second.Range[0], 0, t.Location())
	}

	if !contains(c.Month.Range, month) {
		nextMonth := nextValue(c.Month.Range, month)
		if nextMonth > month {
			return false, time.Date(year, time.Month(nextMonth), 1, c.Hour.Range[0], c.Minute.Range[0], c.Second.Range[0], 0, t.Location())
		}
		return false, time.Date(year+1, time.Month(c.Month.Range[0]), 1, c.Hour.Range[0], c.Minute.Range[0], c.Second.Range[0], 0, t.Location())
	}

	domSpecified := len(c.DayOfMonth.Range) < 31
	dowSpecified := len(c.DayOfWeek.Range) < 8
	weekday := int(t.Weekday())

	var dayMatch bool
	if domSpecified && dowSpecified {
		domMatch := contains(c.DayOfMonth.Range, day)
		dowMatch := contains(c.DayOfWeek.Range, weekday)
		dayMatch = domMatch || dowMatch
	} else if domSpecified {
		dayMatch = contains(c.DayOfMonth.Range, day)
	} else if dowSpecified {
		dayMatch = contains(c.DayOfWeek.Range, weekday)
	} else {
		dayMatch = true
	}

	if !dayMatch {
		nextDay := t.AddDate(0, 0, 1)
		return false, time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), c.Hour.Range[0], c.Minute.Range[0], c.Second.Range[0], 0, t.Location())
	}

	return true, time.Time{}
}

func nextValue(values []int, current int) int {
	for _, v := range values {
		if v > current {
			return v
		}
	}
	return values[0]
}

func (c *CronExpr) NextN(after time.Time, n int) ([]time.Time, bool) {
	var results []time.Time
	t := after

	for i := 0; i < n; i++ {
		next, found := c.Next(t)
		if !found {
			return results, false
		}
		results = append(results, next)
		t = next
	}

	return results, true
}

func (c *CronExpr) WillNeverFire() bool {
	if contains(c.Month.Range, 2) && contains(c.DayOfMonth.Range, 29) {
		only29th := true
		for _, d := range c.DayOfMonth.Range {
			if d != 29 {
				only29th = false
				break
			}
		}

		if only29th {
			onlyFebruary := true
			for _, m := range c.Month.Range {
				if m != 2 {
					onlyFebruary = false
					break
				}
			}

			if onlyFebruary {
				return false
			}
		}
	}

	for _, m := range c.Month.Range {
		daysInCurrentMonth := getDaysInMonth(2024, m)

		for _, d := range c.DayOfMonth.Range {
			if d > daysInCurrentMonth {
				hasOtherValidDay := false
				for _, d2 := range c.DayOfMonth.Range {
					if d2 <= daysInCurrentMonth {
						hasOtherValidDay = true
						break
					}
				}

				if !hasOtherValidDay {
					hasOtherMonth := false
					for _, m2 := range c.Month.Range {
						if m2 != m {
							if d <= getDaysInMonth(2024, m2) {
								hasOtherMonth = true
								break
							}
						}
					}

					if !hasOtherMonth {
						domSpecified := len(c.DayOfMonth.Range) < 31
						dowSpecified := len(c.DayOfWeek.Range) < 8

						if domSpecified && dowSpecified {
							if len(c.DayOfWeek.Range) > 0 {
								return false
							}
						}

						return true
					}
				}
			}
		}
	}

	domSpecified := len(c.DayOfMonth.Range) < 31
	dowSpecified := len(c.DayOfWeek.Range) < 8

	if domSpecified && dowSpecified {
		if len(c.DayOfMonth.Range) == 0 || len(c.DayOfWeek.Range) == 0 {
			return true
		}
	} else if domSpecified {
		if len(c.DayOfMonth.Range) == 0 {
			return true
		}
	} else if dowSpecified {
		if len(c.DayOfWeek.Range) == 0 {
			return true
		}
	}

	return false
}

func getDaysInMonth(year int, month int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.Local).Day()
}

func contains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
